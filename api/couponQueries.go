package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
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
					var coupons []Coupon
					json.UnmarshalFromString(cachedres, &coupons)
					return coupons, nil
				}
			}
			cursor, err := couponCollection.Find(ctxMongo, bson.M{}, nil)
			if err != nil {
				return nil, err
			}
			defer cursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			var coupons = []Coupon{}
			for cursor.Next(ctxMongo) {
				var couponData map[string]interface{}
				if err = cursor.Decode(couponData); err != nil {
					return nil, err
				}
				couponID := couponData["_id"].(primitive.ObjectID)
				var currentCoupon Coupon
				if err = mapstructure.Decode(couponData, &currentCoupon); err != nil {
					return nil, err
				}
				currentCoupon.ID = couponID.Hex()
				coupons = append(coupons, currentCoupon)
			}
			productsResBytes, err := json.Marshal(coupons)
			if err != nil {
				return nil, err
			}
			err = redisClient.Set(cachepath, string(productsResBytes), cacheTime).Err()
			if err != nil {
				return nil, err
			}
			return coupons, nil
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

func checkCoupon(secret string) (*Coupon, error) {
	pathMap := map[string]string{
		"path":   "coupon",
		"secret": secret,
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	var coupon Coupon
	cachepath := string(cachepathBytes)
	if !isDebug() {
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
	var couponData map[string]interface{}
	err = couponCollection.FindOne(ctxMongo, bson.M{
		"secret": secret,
	}).Decode(&couponData)
	if err != nil {
		return nil, err
	}
	couponID := couponData["_id"].(primitive.ObjectID)
	if err = mapstructure.Decode(couponData, &coupon); err != nil {
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
