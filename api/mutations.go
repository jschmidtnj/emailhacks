package main

import (
	"errors"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getFormattedAccessGQLData(itemData map[string]interface{}, changedAccess []map[string]interface{}, userIDString string) ([]map[string]interface{}, []string, []string, error) {
	if itemData["access"] == nil || len(userIDString) == 0 {
		return []map[string]interface{}{}, []string{}, []string{}, nil
	}
	userData := itemData["access"].(map[string]bson.M)[userIDString]
	tags := []string{}
	categories := []string{}
	if userData != nil {
		// you're not an admin
		var err error
		tags, err = interfaceListToStringList(userData["tags"].(bson.A))
		if err != nil {
			return nil, nil, nil, err
		}
		categories, err = interfaceListToStringList(userData["categories"].(bson.A))
		if err != nil {
			return nil, nil, nil, err
		}
	}
	newAccessMap := itemData["access"].(map[string]bson.M)
	if changedAccess != nil {
		for _, accessUser := range changedAccess {
			currentUserID := accessUser["id"].(string)
			if newAccessMap[currentUserID] != nil {
				if accessUser["type"].(string) == noAccessLevel {
					delete(newAccessMap, currentUserID)
				}
			} else {
				newAccessMap[currentUserID] = bson.M{
					"type": accessUser["type"].(string),
				}
			}
			delete(newAccessMap[currentUserID], "categories")
			delete(newAccessMap[currentUserID], "tags")
		}
	}
	newAccess := make([]map[string]interface{}, len(newAccessMap))
	var i = 0
	for id, accessElem := range newAccessMap {
		newAccess[i] = bson.M{
			"id":   id,
			"type": accessElem["type"],
		}
		i++
	}
	return newAccess, tags, categories, nil
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
			if currentUserIDString == userIDString {
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

func changeUserProjectAccess(projectID primitive.ObjectID, access []map[string]interface{}) error {
	for _, accessUser := range access {
		if err := checkAccessObj(accessUser); err != nil {
			return err
		}
	}
	for _, accessUser := range access {
		projectUpdateData := bson.M{}
		if accessUser["type"] != noAccessLevel {
			projectUpdateData["$addToSet"] = bson.M{
				"access": bson.M{
					"id":   accessUser["id"],
					"type": accessUser["type"],
				},
			}
		} else {
			projectUpdateData["$pull"] = bson.M{
				"access.id": accessUser["id"],
			}
		}
		_, err := projectCollection.UpdateOne(ctxMongo, bson.M{
			"_id": projectID,
		}, projectUpdateData)
		if err != nil {
			return err
		}
	}
	return nil
}

func rootMutation() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addProduct":     productMutationFields["addProduct"],
			"updateProduct":  productMutationFields["updateProduct"],
			"deleteProduct":  productMutationFields["deleteProduct"],
			"addCoupon":      couponMutationFields["addCoupon"],
			"deleteCoupon":   couponMutationFields["deleteCoupon"],
			"addResponse":    responseMutationFields["addResponse"],
			"updateResponse": responseMutationFields["updateResponse"],
			"deleteResponse": responseMutationFields["deleteResponse"],
			"addForm":        formMutationFields["addForm"],
			"updateForm":     formMutationFields["updateForm"],
			"updateFormPart": formMutationFields["updateFormPart"],
			"deleteForm":     formMutationFields["deleteForm"],
			"deleteUser":     userMutationFields["deleteUser"],
			"deleteAccount":  userMutationFields["deleteAccount"],
			"addProject":     projectMutationFields["addProject"],
			"updateProject":  projectMutationFields["updateProject"],
			"deleteProject":  projectMutationFields["deleteProject"],
			"addBlog":        blogMutationFields["addBlog"],
			"updateBlog":     blogMutationFields["updateBlog"],
			"deleteBlog":     blogMutationFields["deleteBlog"],
		},
	})
}
