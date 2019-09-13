package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func rootQuery() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"hello": &graphql.Field{
				Type:        graphql.String,
				Description: "Say Hi",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "Hello World!", nil
				},
			},
			"account": &graphql.Field{
				Type:        AccountType,
				Description: "Get your account info",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					accountdata, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
					if err != nil {
						return nil, err
					}
					if accountdata["email"] == nil {
						return nil, errors.New("email not found in token")
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
						userData["date"] = objectidtimestamp(id).Format(dateFormat)
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
						userData["date"] = objectidtimestamp(id).Format(dateFormat)
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
		},
	})
}
