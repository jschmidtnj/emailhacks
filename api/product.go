package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/stripe/stripe-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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

func processProductFromDB(productData bson.M) (bson.M, error) {
	productData["id"] = productData["_id"]
	delete(productData, "_id")
	plansArray, ok := productData["plans"].(primitive.A)
	if !ok {
		return nil, errors.New("cannot cast plans to array")
	}
	for i, plan := range plansArray {
		primitivePlan, ok := plan.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast plan to primitive D")
		}
		plansArray[i] = primitivePlan.Map()
	}
	productData["plans"] = plansArray
	return productData, nil
}

func getProductFromUserData(userData map[string]interface{}) (map[string]interface{}, error) {
	var useDefaultPlan = false
	if userData["plan"] != nil {
		plan, ok := userData["plan"].(string)
		useDefaultPlan = !ok || len(plan) == 0
	}
	var productData map[string]interface{}
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

func getProduct(productID primitive.ObjectID, useCache bool) (map[string]interface{}, error) {
	pathMap := map[string]string{
		"path": "product",
		"id":   productID.Hex(),
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	var productData map[string]interface{}
	cachepath := string(cachepathBytes)
	if useCache {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			json.UnmarshalFromString(cachedres, &productData)
			return productData, nil
		}
	}
	var productDataCursor *mongo.Cursor
	if productID == primitive.NilObjectID {
		productDataCursor, err = productCollection.Find(ctxMongo, bson.M{
			"name": defaultPlanName,
		})
	} else {
		productDataCursor, err = productCollection.Find(ctxMongo, bson.M{
			"_id": productID,
		})
	}
	defer productDataCursor.Close(ctxMongo)
	if err != nil {
		return nil, err
	}
	var foundProduct = false
	for productDataCursor.Next(ctxMongo) {
		foundProduct = true
		productPrimitive := &bson.D{}
		err = productDataCursor.Decode(productPrimitive)
		if err != nil {
			return nil, err
		}
		productData, err = processProductFromDB(productPrimitive.Map())
		if err != nil {
			return nil, err
		}
		break
	}
	if !foundProduct {
		return nil, errors.New("product not found with given id")
	}
	productResBytes, err := json.Marshal(productData)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(productResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return productData, nil
}

// user purchase a product
func purchase(userID primitive.ObjectID, productID primitive.ObjectID, couponIDString string, couponAmount int, interval string, cardToken string, formatDate bool) (map[string]interface{}, error) {
	productData, err := getProduct(productID, !isDebug())
	if err != nil {
		return nil, err
	}
	productIDString := productID.Hex()
	plans, ok := productData["plans"].(bson.A)
	if !ok {
		return nil, errors.New("cannot cast plans to array")
	}
	var foundPlan = false
	var planIDString string
	var amount int
	for _, plan := range plans {
		planObj := plan.(bson.M)
		planInterval := planObj["interval"].(string)
		if planInterval == interval {
			foundPlan = true
			if interval == singlePurchase {
				planIDString = planObj["stripeid"].(string)
			} else {
				amount = planObj["amount"].(int)
			}
			break
		}
	}
	if !foundPlan {
		return nil, errors.New("could not find plan")
	}
	userData, err := getAccount(userID, formatDate, true)
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
		_, err := stripeClient.Charges.New(&stripe.ChargeParams{
			Customer: &userStripeID,
			Currency: &defaultCurrency,
			Amount:   stripe.Int64(int64(amount - couponAmount)),
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
