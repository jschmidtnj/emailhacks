package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CouponType coupon object for discounts
var CouponType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Product",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"secret": &graphql.Field{
			Type: graphql.String,
		},
		"amount": &graphql.Field{
			Type: graphql.Int,
		},
		"percent": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

func getCoupon(couponID primitive.ObjectID, useCache bool) (map[string]interface{}, error) {
	pathMap := map[string]string{
		"path": "coupon",
		"id":   couponID.Hex(),
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	var couponData map[string]interface{}
	cachepath := string(cachepathBytes)
	if useCache {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			json.UnmarshalFromString(cachedres, &couponData)
			return couponData, nil
		}
	}
	couponDataCursor, err := couponCollection.Find(ctxMongo, bson.M{
		"_id": couponID,
	})
	defer couponDataCursor.Close(ctxMongo)
	if err != nil {
		return nil, err
	}
	var foundCoupon = false
	for couponDataCursor.Next(ctxMongo) {
		foundCoupon = true
		couponPrimitive := &bson.D{}
		err = couponDataCursor.Decode(couponPrimitive)
		if err != nil {
			return nil, err
		}
		couponData = couponPrimitive.Map()
		couponData["id"] = couponData["_id"]
		delete(couponData, "_id")
		break
	}
	if !foundCoupon {
		return nil, errors.New("product not found with given id")
	}
	couponResBytes, err := json.Marshal(couponData)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(couponResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return couponData, nil
}
