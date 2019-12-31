package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func deleteAccount(idstring string, formatDate bool) (interface{}, error) {
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
		userPrimitive := &bson.D{}
		err = cursor.Decode(userPrimitive)
		if err != nil {
			return nil, err
		}
		userData = userPrimitive.Map()
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
		return nil, errors.New("user not found with given id")
	}
	_, err = userCollection.DeleteOne(ctxMongo, bson.M{
		"_id": id,
	})
	if err != nil {
		return nil, err
	}
	return userData, nil
}

var userMutationFields = graphql.Fields{
	"deleteUser": &graphql.Field{
		Type:        AccountType,
		Description: "Delete a User",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			_, err := validateAdmin(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("user id not provided")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			return deleteAccount(idstring, formatDate)
		},
	},
	"deleteAccount": &graphql.Field{
		Type:        AccountType,
		Description: "Delete a User",
		Args: graphql.FieldConfigArgument{
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			idstring, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			return deleteAccount(idstring, formatDate)
		},
	},
}
