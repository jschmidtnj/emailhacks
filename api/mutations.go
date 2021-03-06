package main

import (
	"errors"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getFormattedAccessGQLData(currentAccess interface{}, changedAccess []map[string]interface{}, userIDString string) ([]*Access, []string, []string, string, error) {
	if currentAccess == nil || len(userIDString) == 0 {
		return []*Access{}, []string{}, []string{}, "", nil
	}
	currentAccessMap, ok := currentAccess.(map[string]bson.M)
	if !ok {
		return nil, nil, nil, "", errors.New("cannot convert current access to map")
	}
	userData := currentAccessMap[userIDString]
	tags := []string{}
	categories := []string{}
	var accessType string
	if userData != nil {
		// you're not an admin
		var err error
		tags, err = interfaceListToStringList(userData["tags"].(bson.A))
		if err != nil {
			return nil, nil, nil, "", err
		}
		categories, err = interfaceListToStringList(userData["categories"].(bson.A))
		if err != nil {
			return nil, nil, nil, "", err
		}
		var ok bool
		accessType, ok = userData["type"].(string)
		if !ok {
			return nil, nil, nil, "", err
		}
	}
	newAccessMap := make(map[string]*Access, len(currentAccessMap))
	for currentUserID, accessUser := range currentAccessMap {
		newAccessMap[currentUserID] = &Access{
			Type: accessUser["type"].(string),
		}
	}
	if changedAccess != nil {
		for _, accessUser := range changedAccess {
			currentUserID := accessUser["id"].(string)
			if newAccessMap[currentUserID] != nil {
				if accessUser["type"].(string) == noAccessLevel {
					delete(newAccessMap, currentUserID)
				}
			} else {
				newAccessMap[currentUserID] = &Access{
					Type: accessUser["type"].(string),
				}
			}
		}
	}
	newAccess := make([]*Access, len(newAccessMap))
	var i = 0
	for id, accessElem := range newAccessMap {
		newAccess[i] = &Access{
			ID:   id,
			Type: accessElem.Type,
		}
		i++
	}
	return newAccess, tags, categories, accessType, nil
}

func checkAccessObj(accessObj map[string]interface{}) error {
	if accessObj["id"] == nil {
		return errors.New("no id field given")
	}
	userIDString, ok := accessObj["id"].(string)
	if !ok {
		return errors.New("cannot cast user id to string")
	}
	_, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return err
	}
	if accessObj["type"] != nil {
		accessType, ok := accessObj["type"].(string)
		if !ok {
			return errors.New("cannot cast type to string")
		}
		if !findInArray(accessType, validAccessTypes) {
			return errors.New("invalid access type given")
		}
	}
	if accessObj["categories"] != nil {
		categoriesInterface, ok := accessObj["categories"].([]interface{})
		if !ok {
			return errors.New("problem casting categories to interface array")
		}
		_, err = interfaceListToStringList(categoriesInterface)
		if err != nil {
			return err
		}
	}
	if accessObj["tags"] != nil {
		tagsInterface, ok := accessObj["tags"].([]interface{})
		if !ok {
			return errors.New("problem casting categories to interface array")
		}
		_, err = interfaceListToStringList(tagsInterface)
		if err != nil {
			return err
		}
	}
	return nil
}

func changeUserAccessData(itemID primitive.ObjectID, itemType string, userIDString string, categories []string, tags []string, access []map[string]interface{}) (primitive.M, error) {
	for _, accessUser := range access {
		if err := checkAccessObj(accessUser); err != nil {
			return nil, err
		}
	}
	itemUpdateData := bson.M{
		"$set": bson.M{
			"access": bson.M{},
		},
		"$setOnInsert": bson.M{
			"access": bson.M{},
		},
		"$unset": bson.M{
			"access": bson.M{},
		},
		"$upsert": true,
	}
	for _, accessUser := range access {
		currentUserIDString := accessUser["id"].(string)
		if accessUser["type"] != nil || currentUserIDString == userIDString {
			if accessUser["type"] != nil && accessUser["type"].(string) != noAccessLevel {
				itemUpdateData["$set"].(bson.M)["access"].(bson.M)[currentUserIDString] = bson.M{
					"type": accessUser["type"].(string),
				}
				if currentUserIDString != userIDString {
					itemUpdateData["$setOnInsert"].(bson.M)["access"].(bson.M)[currentUserIDString] = bson.M{
						"categories": bson.A{},
						"tags":       bson.A{},
					}
				}
			} else if currentUserIDString != userIDString {
				itemUpdateData["$unset"].(bson.M)["access"] = bson.M{
					currentUserIDString: 1,
				}
			}
			if currentUserIDString == userIDString && categories != nil && tags != nil {
				itemUpdateData["$set"].(bson.M)["access"].(bson.M)[currentUserIDString] = bson.M{
					"categories": categories,
					"tags":       tags,
				}
			}
		}
	}
	if categories != nil {
		itemUpdateData["$set"].(bson.M)["access"].(bson.M)[userIDString].(bson.M)["categories"] = categories
	}
	if tags != nil {
		itemUpdateData["$set"].(bson.M)["access"].(bson.M)[userIDString].(bson.M)["tags"] = tags
	}
	return itemUpdateData, nil
}

func rootMutation() *graphql.Object {
	fields := graphql.Fields{}
	for key := range productMutationFields {
		fields[key] = productMutationFields[key]
	}
	for key := range couponMutationFields {
		fields[key] = couponMutationFields[key]
	}
	for key := range responseMutationFields {
		fields[key] = responseMutationFields[key]
	}
	for key := range formMutationFields {
		fields[key] = formMutationFields[key]
	}
	for key := range userMutationFields {
		fields[key] = userMutationFields[key]
	}
	for key := range projectMutationFields {
		fields[key] = projectMutationFields[key]
	}
	for key := range blogMutationFields {
		fields[key] = blogMutationFields[key]
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: fields,
	})
}
