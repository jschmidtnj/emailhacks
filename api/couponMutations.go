package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/stripe/stripe-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var couponMutationFields = graphql.Fields{
	"addCoupon": &graphql.Field{
		Type:        CouponType,
		Description: "Create a Coupon",
		Args: graphql.FieldConfigArgument{
			"secret": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"amount": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"percent": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["secret"] == nil {
				return nil, errors.New("secret was not provided")
			}
			secret, ok := params.Args["secret"].(string)
			if !ok {
				return nil, errors.New("problem casting secret to string")
			}
			if params.Args["amount"] == nil {
				return nil, errors.New("amount was not provided")
			}
			amount, ok := params.Args["amount"].(int)
			if !ok {
				return nil, errors.New("problem casting amount to int")
			}
			if amount <= 0 {
				return nil, errors.New("invalid max forms given")
			}
			if params.Args["percent"] == nil {
				return nil, errors.New("percent was not specified")
			}
			percent, ok := params.Args["percent"].(bool)
			if !ok {
				return nil, errors.New("problem casting percent to bool")
			}
			couponData := bson.M{
				"secret":  secret,
				"amount":  amount,
				"percent": percent,
			}
			couponCreateRes, err := couponCollection.InsertOne(ctxMongo, couponData)
			if err != nil {
				return nil, err
			}
			couponID := couponCreateRes.InsertedID.(primitive.ObjectID)
			couponIDString := couponID.Hex()
			couponParams := &stripe.CouponParams{
				ID:       &secret,
				Duration: &defaultCouponDuration,
				Params: stripe.Params{
					Metadata: map[string]string{
						"id": couponIDString,
					},
				},
			}
			if percent {
				couponParams.PercentOff = stripe.Float64(float64(amount))
			} else {
				couponParams.AmountOff = stripe.Int64(int64(amount))
			}
			_, err = stripeClient.Coupons.New(couponParams)
			if err != nil {
				return nil, err
			}
			couponData["id"] = couponIDString
			return couponData, nil
		},
	},
	"deleteCoupon": &graphql.Field{
		Type:        CouponType,
		Description: "Delete a Coupon",
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
				return nil, errors.New("coupon id not provided")
			}
			couponIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast coupon id to string")
			}
			couponID, err := primitive.ObjectIDFromHex(couponIDString)
			if err != nil {
				return nil, err
			}
			couponData, err := getCoupon(couponID, false)
			if err != nil {
				return nil, err
			}
			if _, err := stripeClient.Coupons.Del(couponData["secret"].(string), nil); err != nil {
				return nil, err
			}
			_, err = couponCollection.DeleteOne(ctxMongo, bson.M{
				"_id": couponID,
			})
			if err != nil {
				return nil, err
			}
			return couponData, nil
		},
	},
}
