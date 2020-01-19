package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var couponQueryFields = graphql.Fields{
	"coupons": &graphql.Field{
		Type:        graphql.NewList(CouponType),
		Description: "Get list of coupons",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
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
	"coupon": &graphql.Field{
		Type:        CouponType,
		Description: "Get a Coupon",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			couponIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			couponID, err := primitive.ObjectIDFromHex(couponIDString)
			if err != nil {
				return nil, err
			}
			couponData, err := getCoupon(couponID, !isDebug())
			if err != nil {
				return nil, err
			}
			return couponData, nil
		},
	},
	"checkCoupon": &graphql.Field{
		Type:        CouponType,
		Description: "Get a Coupon",
		Args: graphql.FieldConfigArgument{
			"secret": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := getTokenData(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["secret"] == nil {
				return nil, errors.New("no secret argument found")
			}
			secret, ok := params.Args["secret"].(string)
			if !ok {
				return nil, errors.New("cannot cast secret to string")
			}
			return checkCoupon(secret)
		},
	},
}

func checkCoupon(secret string) (map[string]interface{}, error) {
	pathMap := map[string]string{
		"path":   "coupon",
		"secret": secret,
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	var couponData map[string]interface{}
	cachepath := string(cachepathBytes)
	if !isDebug() {
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
		"secret": secret,
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
