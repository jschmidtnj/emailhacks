package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userQueryFields = graphql.Fields{
	"account": &graphql.Field{
		Type:        AccountType,
		Description: "Get your account info",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accountdata, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if accountdata["email"] == nil {
				return nil, errors.New("email not found in token")
			}
			var account Account
			accountData := map[string]interface{}{}
			err = userCollection.FindOne(ctxMongo, bson.M{
				"email": accountdata["email"].(string),
			}).Decode(&accountData)
			if err != nil {
				return nil, err
			}
			accountID := accountData["_id"].(primitive.ObjectID)
			delete(accountData, "_id")
			if err = mapstructure.Decode(accountData, &account); err != nil {
				return nil, err
			}
			account.Created = objectidTimestamp(accountID).Unix()
			account.ID = accountID.Hex()
			return &account, nil
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
			account, err := getAccount(id, false)
			if err != nil {
				return nil, err
			}
			return account, nil
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

func getAccounts() (*[]Account, error) {
	cursor, err := userCollection.Find(ctxMongo, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxMongo)
	accounts := []Account{}
	for cursor.Next(ctxMongo) {
		accountData := bson.M{}
		err = cursor.Decode(&accountData)
		if err != nil {
			return nil, err
		}
		accountID := accountData["_id"].(primitive.ObjectID)
		var currentAccount Account
		if err = mapstructure.Decode(accountData, &currentAccount); err != nil {
			return nil, err
		}
		currentAccount.Created = objectidTimestamp(accountID).Unix()
		currentAccount.ID = accountID.Hex()
		accounts = append(accounts, currentAccount)
	}
	return &accounts, nil
}
