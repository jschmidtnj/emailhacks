package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func deleteAccount(idstring string) (interface{}, error) {
	id, err := primitive.ObjectIDFromHex(idstring)
	if err != nil {
		return nil, err
	}
	// delete all projects that this user is admin of
	sourceContext := elastic.NewFetchSourceContext(true).Include("id")
	mustQueries := []elastic.Query{
		elastic.NewTermsQuery("owner", idstring),
	}
	query := elastic.NewBoolQuery().Must(mustQueries...)
	searchResult, err := elasticClient.Search().
		Index(projectElasticIndex).
		Query(query).
		Pretty(isDebug()).
		FetchSourceContext(sourceContext).
		Do(ctxElastic)
	if err != nil {
		return nil, err
	}
	for _, hit := range searchResult.Hits.Hits {
		projectID, err := primitive.ObjectIDFromHex(hit.Id)
		if err != nil {
			return nil, err
		}
		if err = deleteProject(projectID); err != nil {
			return nil, err
		}
	}
	searchResult, err = elasticClient.Search().
		Index(formElasticIndex).
		Query(query).
		Pretty(isDebug()).
		FetchSourceContext(sourceContext).
		Do(ctxElastic)
	if err != nil {
		return nil, err
	}
	for _, hit := range searchResult.Hits.Hits {
		formID, err := primitive.ObjectIDFromHex(hit.Id)
		if err != nil {
			return nil, err
		}
		if _, err = deleteForm(formID, nil, ""); err != nil {
			return nil, err
		}
	}
	account, err := getAccount(id, true)
	if err != nil {
		return nil, err
	}
	if len(account.SubscriptionID) > 0 {
		if _, err = stripeClient.Subscriptions.Cancel(account.SubscriptionID, nil); err != nil {
			return nil, err
		}
	}
	_, err = userCollection.DeleteOne(ctxMongo, bson.M{
		"_id": id,
	})
	if err != nil {
		return nil, err
	}
	account.SubscriptionID = ""
	account.StripeID = ""
	return account, nil
}

var userMutationFields = graphql.Fields{
	"cancelSubscription": &graphql.Field{
		Type:        AccountType,
		Description: "Purchase a Product",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idString)
			if err != nil {
				return nil, err
			}
			account, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			if len(account.SubscriptionID) > 0 {
				if _, err = stripeClient.Subscriptions.Cancel(account.SubscriptionID, nil); err != nil {
					return nil, err
				}
			}
			userCollection.UpdateOne(ctxMongo, bson.M{
				"_id": id,
			}, bson.M{
				"$set": bson.M{
					"plan":           "",
					"subscriptionid": "",
				},
			})
			account.SubscriptionID = ""
			account.StripeID = ""
			return account, nil
		},
	},
	"purchase": &graphql.Field{
		Type:        AccountType,
		Description: "Purchase a Product",
		Args: graphql.FieldConfigArgument{
			"product": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"interval": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"cardToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"coupon": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idString)
			if err != nil {
				return nil, err
			}
			if params.Args["product"] == nil {
				return nil, errors.New("must specify a product id to purchase")
			}
			productIDString, ok := params.Args["product"].(string)
			if !ok {
				return nil, errors.New("cannot cast product id to string")
			}
			productID, err := primitive.ObjectIDFromHex(productIDString)
			if err != nil {
				return nil, err
			}
			if params.Args["cardToken"] == nil {
				return nil, errors.New("must provide a card token to purchase")
			}
			cardToken, ok := params.Args["cardToken"].(string)
			if !ok {
				return nil, errors.New("cannot cast card token to string")
			}
			if params.Args["interval"] == nil {
				return nil, errors.New("must specify a product interval to purchase")
			}
			interval, ok := params.Args["interval"].(string)
			if !ok {
				return nil, errors.New("cannot cast interval to string")
			}
			if !findInArray(interval, validIntervals) {
				return nil, errors.New("invalid product interval provided")
			}
			var couponIDString = ""
			var couponAmount int64 = 0
			var couponPercent = false
			if params.Args["coupon"] != nil {
				secret, ok := params.Args["coupon"].(string)
				if !ok {
					return nil, errors.New("cannot cast coupon to string")
				}
				couponData, err := checkCoupon(secret)
				if err != nil {
					return nil, err
				}
				couponIDString = couponData.ID
				couponAmount = couponData.Amount
				couponPercent = couponData.Percent
			}
			account, err := purchase(id, productID, couponIDString, couponAmount, couponPercent, interval, cardToken)
			if err != nil {
				return nil, err
			}
			account.SubscriptionID = ""
			account.StripeID = ""
			return account, nil
		},
	},
	"changeBilling": &graphql.Field{
		Type:        AccountType,
		Description: "Change Billing",
		Args: graphql.FieldConfigArgument{
			"firstname": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"lastname": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"company": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"address1": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"address2": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"city": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"state": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"zip": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"country": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"phone": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"email": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idString)
			if err != nil {
				return nil, err
			}
			updatedBillingData := bson.M{}
			if params.Args["firstname"] != nil {
				firstname, ok := params.Args["firstname"].(string)
				if !ok {
					return nil, errors.New("cannot cast first name to string")
				}
				updatedBillingData["firstname"] = firstname
			}
			if params.Args["lastname"] != nil {
				lastname, ok := params.Args["lastname"].(string)
				if !ok {
					return nil, errors.New("cannot cast last name to string")
				}
				updatedBillingData["lastname"] = lastname
			}
			if params.Args["company"] != nil {
				company, ok := params.Args["company"].(string)
				if !ok {
					return nil, errors.New("cannot cast company to string")
				}
				updatedBillingData["company"] = company
			}
			if params.Args["address1"] != nil {
				address1, ok := params.Args["address1"].(string)
				if !ok {
					return nil, errors.New("cannot cast address 1 to string")
				}
				updatedBillingData["address1"] = address1
			}
			if params.Args["address2"] != nil {
				address2, ok := params.Args["address2"].(string)
				if !ok {
					return nil, errors.New("cannot cast address 2 to string")
				}
				updatedBillingData["address2"] = address2
			}
			if params.Args["city"] != nil {
				city, ok := params.Args["city"].(string)
				if !ok {
					return nil, errors.New("cannot cast city to string")
				}
				updatedBillingData["city"] = city
			}
			if params.Args["state"] != nil {
				state, ok := params.Args["state"].(string)
				if !ok {
					return nil, errors.New("cannot cast state to string")
				}
				updatedBillingData["state"] = state
			}
			if params.Args["zip"] != nil {
				zip, ok := params.Args["zip"].(string)
				if !ok {
					return nil, errors.New("cannot cast zip to string")
				}
				updatedBillingData["zip"] = zip
			}
			if params.Args["country"] != nil {
				country, ok := params.Args["country"].(string)
				if !ok {
					return nil, errors.New("cannot cast country to string")
				}
				updatedBillingData["country"] = country
			}
			if params.Args["phone"] != nil {
				phone, ok := params.Args["phone"].(string)
				if !ok {
					return nil, errors.New("cannot cast phone to string")
				}
				updatedBillingData["phone"] = phone
			}
			if params.Args["email"] != nil {
				email, ok := params.Args["email"].(string)
				if !ok {
					return nil, errors.New("cannot cast email to string")
				}
				updatedBillingData["email"] = email
			}
			_, err = userCollection.UpdateOne(ctxMongo, bson.M{
				"_id": id,
			}, bson.M{
				"$set": bson.M{
					"billing": updatedBillingData,
				},
			})
			account, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			account.SubscriptionID = ""
			account.StripeID = ""
			return account, nil
		},
	},
	"addOrganization": &graphql.Field{
		Type:        AccountType,
		Description: "Add Organization",
		Args: graphql.FieldConfigArgument{
			"type": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"color": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idString)
			if err != nil {
				return nil, err
			}
			if params.Args["type"] == nil {
				return nil, errors.New("must specify organization type")
			}
			organizationType, ok := params.Args["type"].(string)
			if !ok {
				return nil, errors.New("cannot cast organization type to string")
			}
			if !findInArray(organizationType, validOrganization) {
				return nil, errors.New("invalid organization type given")
			}
			if params.Args["name"] == nil {
				return nil, errors.New("cannot find organization name")
			}
			name, ok := params.Args["name"].(string)
			if !ok {
				return nil, errors.New("cannot cast organization name to string")
			}
			if params.Args["color"] == nil {
				return nil, errors.New("cannot find organization color")
			}
			color, ok := params.Args["color"].(string)
			if !ok {
				return nil, errors.New("cannot cast organization color to string")
			}
			if !validHexcode.MatchString(color) {
				return nil, errors.New("invalid hex code for color")
			}
			_, err = userCollection.UpdateOne(ctxMongo, bson.M{
				"_id": id,
			}, bson.M{
				"$push": bson.M{
					organizationType: bson.M{
						"name":  name,
						"color": color,
					},
				},
			})
			account, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			account.SubscriptionID = ""
			account.StripeID = ""
			return account, nil
		},
	},
	"removeOrganization": &graphql.Field{
		Type:        AccountType,
		Description: "Remove Organization",
		Args: graphql.FieldConfigArgument{
			"type": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idString)
			if err != nil {
				return nil, err
			}
			if params.Args["type"] == nil {
				return nil, errors.New("must specify organization type")
			}
			organizationType, ok := params.Args["type"].(string)
			if !ok {
				return nil, errors.New("cannot cast organization type to string")
			}
			if !findInArray(organizationType, validOrganization) {
				return nil, errors.New("invalid organization type given")
			}
			if params.Args["name"] == nil {
				return nil, errors.New("cannot find organization name")
			}
			name, ok := params.Args["name"].(string)
			if !ok {
				return nil, errors.New("cannot cast organization name to string")
			}
			_, err = userCollection.UpdateOne(ctxMongo, bson.M{
				"_id": id,
			}, bson.M{
				"$pull": bson.M{
					organizationType: bson.M{
						"name": name,
					},
				},
			})
			account, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			account.SubscriptionID = ""
			account.StripeID = ""
			return account, nil
		},
	},
	"deleteUser": &graphql.Field{
		Type:        AccountType,
		Description: "Delete a User",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			_, err := validateAdmin(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("user id not provided")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			return deleteAccount(idstring)
		},
	},
	"deleteAccount": &graphql.Field{
		Type:        AccountType,
		Description: "Delete a User",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idstring, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			return deleteAccount(idstring)
		},
	},
}
