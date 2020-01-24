package main

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrganizeType type for a name and count of that object
var OrganizeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Organize",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"color": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// AccountType account type object for user accounts graphql
var AccountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Account",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.Int,
		},
		"updated": &graphql.Field{
			Type: graphql.Int,
		},
		"emailverified": &graphql.Field{
			Type: graphql.Boolean,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"categories": &graphql.Field{
			Type: graphql.NewList(OrganizeType),
		},
		"tags": &graphql.Field{
			Type: graphql.NewList(OrganizeType),
		},
		"plan": &graphql.Field{
			Type:        graphql.String,
			Description: "current plan",
		},
		"purchases": &graphql.Field{
			Type:        graphql.NewList(graphql.String),
			Description: "one-time purchases",
		},
	},
})

// PublicAccountType data publically available
var PublicAccountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PublicAccount",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func processUserFromDB(userData bson.M, updated bool) (bson.M, error) {
	id := userData["_id"].(primitive.ObjectID)
	userData["created"] = objectidTimestamp(id).Unix()
	var updatedTimestamp time.Time
	if updated {
		updatedTimestamp = time.Now()
	} else {
		updatedInt, ok := userData["updated"].(int64)
		if !ok {
			return nil, errors.New("cannot cast updated time to int")
		}
		updatedTimestamp = intTimestamp(updatedInt)
	}
	userData["updated"] = updatedTimestamp.Unix()
	userData["id"] = id.Hex()
	delete(userData, "_id")
	categoriesArray, ok := userData["categories"].(primitive.A)
	if !ok {
		return nil, errors.New("cannot cast categories to array")
	}
	for i, category := range categoriesArray {
		primitiveCategory, ok := category.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast category to primitive D")
		}
		categoriesArray[i] = primitiveCategory.Map()
	}
	userData["categories"] = categoriesArray
	tagsArray, ok := userData["tags"].(primitive.A)
	if !ok {
		return nil, errors.New("cannot cast tags to array")
	}
	for i, tag := range tagsArray {
		primitiveTag, ok := tag.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast tag to primitive D")
		}
		tagsArray[i] = primitiveTag.Map()
	}
	userData["tags"] = tagsArray
	return userData, nil
}

func getAccount(accountID primitive.ObjectID, updated bool) (map[string]interface{}, error) {
	userDataCursor, err := userCollection.Find(ctxMongo, bson.M{
		"_id": accountID,
	})
	defer userDataCursor.Close(ctxMongo)
	if err != nil {
		return nil, err
	}
	var userData map[string]interface{}
	var foundUser = false
	for userDataCursor.Next(ctxMongo) {
		foundUser = true
		userPrimitive := &bson.D{}
		err = userDataCursor.Decode(userPrimitive)
		if err != nil {
			return nil, err
		}
		userData, err = processUserFromDB(userPrimitive.Map(), updated)
		if err != nil {
			return nil, err
		}
		break
	}
	if !foundUser {
		return nil, errors.New("user not found with given id")
	}
	return userData, nil
}
