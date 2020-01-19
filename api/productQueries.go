package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var productQueryFields = graphql.Fields{
	"products": &graphql.Field{
		Type:        graphql.NewList(ProductType),
		Description: "Get list of products",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			pathMap := map[string]string{
				"path": "products",
			}
			cachepathBytes, err := json.Marshal(pathMap)
			if err != nil {
				return nil, err
			}
			cachepath := string(cachepathBytes)
			if !isDebug() {
				cachedres, err := redisClient.Get(cachepath).Result()
				if err != nil {
					if err != redis.Nil {
						return nil, err
					}
				} else {
					var products []map[string]interface{}
					json.UnmarshalFromString(cachedres, &products)
					return products, nil
				}
			}
			cursor, err := productCollection.Find(ctxMongo, bson.M{}, nil)
			if err != nil {
				return nil, err
			}
			defer cursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			var products = []map[string]interface{}{}
			for cursor.Next(ctxMongo) {
				productDataPrimitive := &bson.D{}
				err = cursor.Decode(productDataPrimitive)
				if err != nil {
					return nil, err
				}
				productData, err := processProductFromDB(productDataPrimitive.Map())
				if err != nil {
					return nil, err
				}
				products = append(products, productData)
			}
			productsResBytes, err := json.Marshal(products)
			if err != nil {
				return nil, err
			}
			err = redisClient.Set(cachepath, string(productsResBytes), cacheTime).Err()
			if err != nil {
				return nil, err
			}
			return products, nil
		},
	},
	"project": &graphql.Field{
		Type:        ProjectType,
		Description: "Get a Project",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			productIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			productID, err := primitive.ObjectIDFromHex(productIDString)
			if err != nil {
				return nil, err
			}
			productData, err := getProduct(productID, !isDebug())
			if err != nil {
				return nil, err
			}
			return productData, nil
		},
	},
}
