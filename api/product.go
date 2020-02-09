package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Currency object
type Currency struct {
	Name         string  `json:"name"`
	ExchangeRate float64 `json:"exchangerate"`
}

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

func getProductFromUserData(account *Account) (*Product, error) {
	var useDefaultPlan = false
	useDefaultPlan = len(account.Plan) == 0
	var productData *Product
	var err error
	if useDefaultPlan {
		productData, err = getProduct(primitive.NilObjectID, !isDebug())
	} else {
		productID, err := primitive.ObjectIDFromHex(account.Plan)
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
	if useCache && !isDebug() {
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
