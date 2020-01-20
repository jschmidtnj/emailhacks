package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/stripe/stripe-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Product type
type Product struct {
	ID          string  `json:"id"`
	StripeID    string  `json:"stripeid"`
	Name        string  `json:"name"`
	Plans       []*Plan `json:"plans"`
	MaxProjects int64   `json:"maxprojects"`
	MaxForms    int64   `json:"maxforms"`
	MaxStorage  int64   `json:"maxstorage"`
}

// ProductType product object for purchasing
var ProductType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Product",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"stripeid": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"plans": &graphql.Field{
			Type: graphql.NewList(PlanType),
		},
		"maxprojects": &graphql.Field{
			Type:        graphql.Int,
			Description: "maximum number of projects a user can make",
		},
		"maxforms": &graphql.Field{
			Type:        graphql.Int,
			Description: "maximum number of forms a user can make",
		},
		"maxstorage": &graphql.Field{
			Type:        graphql.Int,
			Description: "maximum amount of file storage in Gb",
		},
	},
})

func getProductFromUserData(userData map[string]interface{}) (*Product, error) {
	var useDefaultPlan = false
	if userData["plan"] != nil {
		plan, ok := userData["plan"].(string)
		useDefaultPlan = !ok || len(plan) == 0
	}
	var productData *Product
	var err error
	if useDefaultPlan {
		productData, err = getProduct(primitive.NilObjectID, !isDebug())
	} else {
		productID, err := primitive.ObjectIDFromHex(userData["plan"].(string))
		if err != nil {
			return nil, err
		}
		productData, err = getProduct(productID, !isDebug())
	}
	if err != nil {
		return nil, err
	}
	return productData, nil
}

func getProduct(productID primitive.ObjectID, useCache bool) (*Product, error) {
	pathMap := map[string]string{
		"path": "product",
		"id":   productID.Hex(),
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	var product Product
	cachepath := string(cachepathBytes)
	if useCache {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			json.UnmarshalFromString(cachedres, &product)
			return &product, nil
		}
	}
	var mongoRes *mongo.SingleResult
	if productID == primitive.NilObjectID {
		mongoRes = productCollection.FindOne(ctxMongo, bson.M{
			"name": defaultPlanName,
		})
	} else {
		mongoRes = productCollection.FindOne(ctxMongo, bson.M{
			"_id": productID,
		})
	}
	var productData map[string]interface{}
	if err = mongoRes.Decode(&productData); err != nil {
		return nil, err
	}
	productID = productData["_id"].(primitive.ObjectID)
	if err = mapstructure.Decode(productData, &product); err != nil {
		return nil, err
	}
	product.ID = productID.Hex()
	productResBytes, err := json.Marshal(product)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(productResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// user purchase a product
func purchase(userID primitive.ObjectID, productID primitive.ObjectID, couponIDString string, couponAmount int64, couponPercent bool, interval string, cardToken string) (map[string]interface{}, error) {
	productData, err := getProduct(productID, !isDebug())
	if err != nil {
		return nil, err
	}
	productIDString := productID.Hex()
	var foundPlan = false
	var planIDString string
	var amount int64
	for _, plan := range productData.Plans {
		planInterval := plan.Interval
		if planInterval == interval {
			foundPlan = true
			if interval == singlePurchase {
				planIDString = plan.StripeID
			} else {
				amount = plan.Amount
			}
			break
		}
	}
	if !foundPlan {
		return nil, errors.New("could not find plan")
	}
	userData, err := getAccount(userID, true)
	if err != nil {
		return nil, err
	}
	userIDString := userID.Hex()
	var userStripeID string
	var newCustomer = true
	if userData["stripeid"] != nil {
		var ok bool
		userStripeID, ok = userData["stripeid"].(string)
		newCustomer = !ok || len(userStripeID) == 0
	}
	userEmail := userData["email"].(string)
	if newCustomer {
		newCustomer, err := stripeClient.Customers.New(&stripe.CustomerParams{
			Email: &userEmail,
			Source: &stripe.SourceParams{
				Token: &cardToken,
			},
			Params: stripe.Params{
				Metadata: map[string]string{
					"id": userIDString,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		userStripeID = newCustomer.ID
	} else if userData["subscriptionid"] != nil {
		stripeSubscriptionID, ok := userData["subscriptionid"].(string)
		if ok && len(stripeSubscriptionID) > 0 {
			if _, err = stripeClient.Subscriptions.Cancel(stripeSubscriptionID, nil); err != nil {
				return nil, err
			}
		}
	}
	var newPlan string
	userUpdateData := bson.M{
		"$set":      bson.M{},
		"$addToSet": bson.M{},
	}
	if interval != singlePurchase {
		subscriptionParams := &stripe.SubscriptionParams{
			Customer:              &userStripeID,
			BillingCycleAnchorNow: stripe.Bool(true),
			Items: []*stripe.SubscriptionItemsParams{&stripe.SubscriptionItemsParams{
				Plan: &planIDString,
			},
			},
		}
		if len(couponIDString) > 0 {
			subscriptionParams.Coupon = &couponIDString
		}
		stripeSubscription, err := stripeClient.Subscriptions.New(subscriptionParams)
		if err != nil {
			return nil, err
		}
		newPlan = productID.Hex()
		userUpdateData["$set"].(bson.M)["plan"] = newPlan
		userUpdateData["$set"].(bson.M)["subscriptionid"] = stripeSubscription.ID
	} else {
		var newPrice int64
		if couponPercent {
			if couponAmount >= 100 {
				newPrice = 0
			} else if couponAmount <= 0 {
				newPrice = amount
			} else {
				newPrice = int64(float64(couponAmount) / 100 * float64(amount))
			}
		} else {
			if couponAmount >= 100 {
				newPrice = 0
			} else if couponAmount <= 0 {
				newPrice = amount
			} else {
				newPrice = amount - couponAmount
			}
		}
		_, err := stripeClient.Charges.New(&stripe.ChargeParams{
			Customer: &userStripeID,
			Currency: &defaultCurrency,
			Amount:   stripe.Int64(newPrice),
			Source: &stripe.SourceParams{
				Token: &cardToken,
			},
			Params: stripe.Params{
				Metadata: map[string]string{
					"id": productIDString,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		userUpdateData["$addToSet"].(bson.M)["purchases"] = productIDString
	}
	if newCustomer {
		userUpdateData["$set"].(bson.M)["stripeid"] = userStripeID
	}
	// update user
	_, err = userCollection.UpdateOne(ctxMongo, bson.M{
		"_id": userID,
	}, userUpdateData)
	if err != nil {
		return nil, err
	}
	if interval != singlePurchase {
		userData["plan"] = newPlan
	}
	return userData, nil
}
