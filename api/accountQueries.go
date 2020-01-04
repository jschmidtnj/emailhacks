package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userQueryFields = graphql.Fields{
	"account": &graphql.Field{
		Type:        AccountType,
		Description: "Get your account info",
		Args: graphql.FieldConfigArgument{
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accountdata, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if accountdata["email"] == nil {
				return nil, errors.New("email not found in token")
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			cursor, err := userCollection.Find(ctxMongo, bson.M{"email": accountdata["email"].(string)})
			defer cursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			var userData map[string]interface{}
			var foundstuff = false
			for cursor.Next(ctxMongo) {
				userDataPrimitive := &bson.D{}
				err = cursor.Decode(userDataPrimitive)
				if err != nil {
					return nil, err
				}
				userData = userDataPrimitive.Map()
				id := userData["_id"].(primitive.ObjectID)
				if formatDate {
					userData["created"] = objectidTimestamp(id).Format(dateFormat)
				} else {
					userData["created"] = objectidTimestamp(id).Unix()
				}
				updatedInt, ok := userData["updated"].(int64)
				if !ok {
					return nil, errors.New("cannot cast updated time to int")
				}
				if formatDate {
					userData["updated"] = intTimestamp(updatedInt).Format(dateFormat)
				} else {
					userData["updated"] = intTimestamp(updatedInt).Unix()
				}
				userData["id"] = id.Hex()
				delete(userData, "_id")
				foundstuff = true
				break
			}
			if !foundstuff {
				return nil, errors.New("account data not found")
			}
			return userData, nil
		},
	},
	"user": &graphql.Field{
		Type:        AccountType,
		Description: "Get a user by id as admin",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accountdata, err := validateAdmin(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if accountdata["id"] == nil {
				return nil, errors.New("email not found in token")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idstring)
			if err != nil {
				return nil, err
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			cursor, err := userCollection.Find(ctxMongo, bson.M{
				"_id": id,
			})
			defer cursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			var userData map[string]interface{}
			var foundstuff = false
			for cursor.Next(ctxMongo) {
				userDataPrimitive := &bson.D{}
				err = cursor.Decode(userDataPrimitive)
				if err != nil {
					return nil, err
				}
				userData = userDataPrimitive.Map()
				if formatDate {
					userData["created"] = objectidTimestamp(id).Format(dateFormat)
				} else {
					userData["created"] = objectidTimestamp(id).Unix()
				}
				updatedInt, ok := userData["updated"].(int64)
				if !ok {
					return nil, errors.New("cannot cast updated time to int")
				}
				if formatDate {
					userData["updated"] = intTimestamp(updatedInt).Format(dateFormat)
				} else {
					userData["updated"] = intTimestamp(updatedInt).Unix()
				}
				userData["id"] = idstring
				delete(userData, "_id")
				foundstuff = true
				break
			}
			if !foundstuff {
				return nil, errors.New("account data not found")
			}
			return userData, nil
		},
	},
	"userPublic": &graphql.Field{
		Type:        PublicAccountType,
		Description: "Get public user data by email",
		Args: graphql.FieldConfigArgument{
			"email": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			_, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if params.Args["email"] == nil {
				return nil, errors.New("email not provided")
			}
			emailString, ok := params.Args["email"].(string)
			if !ok {
				return nil, errors.New("cannot cast email to string")
			}
			cursor, err := userCollection.Find(ctxMongo, bson.M{
				"email": emailString,
			})
			defer cursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			publicUserData := map[string]interface{}{}
			var foundstuff = false
			for cursor.Next(ctxMongo) {
				userDataPrimitive := &bson.D{}
				err = cursor.Decode(userDataPrimitive)
				if err != nil {
					return nil, err
				}
				userData := userDataPrimitive.Map()
				userID, ok := userData["_id"].(string)
				if !ok {
					return nil, errors.New("cannot cast id to string")
				}
				publicUserData["id"] = userID
				publicUserData["email"] = emailString
				foundstuff = true
				break
			}
			if !foundstuff {
				return nil, errors.New("account data not found")
			}
			return publicUserData, nil
		},
	},
}
