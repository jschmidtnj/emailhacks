package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Coupon used for discounts on products
type Coupon struct {
	ID      string `json:"id"`
	Secret  string `json:"secret"`
	Amount  int64  `json:"amount"` // in USD - global exchange rate to convert
	Percent bool   `json:"percent"`
}

// CouponType coupon object for discounts
var CouponType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Coupon",
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

func getCoupon(couponID primitive.ObjectID, useCache bool) (*Coupon, error) {
	pathMap := map[string]string{
		"path": "coupon",
		"id":   couponID.Hex(),
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	var coupon Coupon
	cachepath := string(cachepathBytes)
	if useCache {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			json.UnmarshalFromString(cachedres, &coupon)
			return &coupon, nil
		}
	}
	err = couponCollection.FindOne(ctxMongo, bson.M{
		"_id": couponID,
	}).Decode(&coupon)
	if err != nil {
		return nil, err
	}
	coupon.ID = couponID.Hex()
	couponResBytes, err := json.Marshal(coupon)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(couponResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}
