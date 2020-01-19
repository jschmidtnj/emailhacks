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
	userData, err := getAccount(id, true)
	if err != nil {
		return nil, err
	}
	stripeSubscriptionID, ok := userData["subscriptionid"].(string)
	if ok && len(stripeSubscriptionID) > 0 {
		if _, err = stripeClient.Subscriptions.Cancel(stripeSubscriptionID, nil); err != nil {
			return nil, err
		}
	}
	_, err = userCollection.DeleteOne(ctxMongo, bson.M{
		"_id": id,
	})
	if err != nil {
		return nil, err
	}
	delete(userData, "subscriptionid")
	delete(userData, "stripeid")
	return userData, nil
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
			userData, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			stripeSubscriptionID, ok := userData["subscriptionid"].(string)
			if ok && len(stripeSubscriptionID) > 0 {
				if _, err = stripeClient.Subscriptions.Cancel(stripeSubscriptionID, nil); err != nil {
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
			delete(userData, "subscriptionid")
			delete(userData, "stripeid")
			return userData, nil
		},
	},
	"purchaseProduct": &graphql.Field{
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
			var couponAmount = 0
			if params.Args["coupon"] != nil {
				secret, ok := params.Args["coupon"].(string)
				if !ok {
					return nil, errors.New("cannot cast coupon to string")
				}
				couponData, err := checkCoupon(secret)
				if err != nil {
					return nil, err
				}
				couponIDString = couponData["id"].(string)
				couponAmount = couponData["amount"].(int)
			}
			userData, err := purchase(id, productID, couponIDString, couponAmount, interval, cardToken)
			if err != nil {
				return nil, err
			}
			delete(userData, "subscriptionid")
			delete(userData, "stripeid")
			return userData, nil
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
			userData, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			delete(userData, "subscriptionid")
			delete(userData, "stripeid")
			return userData, nil
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
			userData, err := getAccount(id, true)
			if err != nil {
				return nil, err
			}
			delete(userData, "subscriptionid")
			delete(userData, "stripeid")
			return userData, nil
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
