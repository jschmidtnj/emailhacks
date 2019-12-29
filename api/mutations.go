package main

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getFormattedGQLData(itemData map[string]interface{}, changedAccess []map[string]interface{}, userIDString string) ([]map[string]interface{}, []string, []string) {
	userData := itemData["access"].(map[string]interface{})[userIDString].(map[string]interface{})
	tags, _ := interfaceListToStringList(userData["tags"].([]interface{}))
	categories, _ := interfaceListToStringList(userData["categories"].([]interface{}))
	newAccessMap := itemData["access"].(map[string]interface{})
	if changedAccess != nil {
		for _, accessUser := range changedAccess {
			currentUserId := accessUser["id"].(string)
			if newAccessMap[currentUserId] != nil {
				if accessUser["type"].(string) == noAccessLevel {
					delete(newAccessMap, currentUserId)
				}
			} else {
				newAccessMap[currentUserId] = bson.M{
					"type": accessUser["type"].(string),
				}
			}
			delete(newAccessMap[currentUserId].(map[string]interface{}), "categories")
			delete(newAccessMap[currentUserId].(map[string]interface{}), "tags")
		}
	}
	newAccess := make([]map[string]interface{}, len(newAccessMap))
	var i = 0
	for id, accessElem := range newAccessMap {
		accessElemMap := accessElem.(map[string]interface{})
		newAccess[i] = bson.M{
			"id":   id,
			"type": accessElemMap["type"],
		}
		i++
	}
	return newAccess, tags, categories
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

func deleteForm(formID primitive.ObjectID, formData bson.M, formatDate bool) error {
	if formData == nil {
		var err error
		formData, err = getForm(formID, formatDate, false)
		if err != nil {
			return err
		}
	}
	_, err := formCollection.DeleteOne(ctxMongo, bson.M{
		"_id": formID,
	})
	if err != nil {
		return err
	}
	primativefiles, ok := formData["files"].(primitive.A)
	if !ok {
		return errors.New("cannot convert files to primitive")
	}
	formIDString := formID.Hex()
	for _, primativefile := range primativefiles {
		filedatadoc, ok := primativefile.(primitive.D)
		if !ok {
			return errors.New("cannot convert file to primitive doc")
		}
		filedata := filedatadoc.Map()
		fileid, ok := filedata["id"].(string)
		if !ok {
			return errors.New("cannot convert file id to string")
		}
		filetype, ok := filedata["type"].(string)
		if !ok {
			return errors.New("cannot convert file type to string")
		}
		fileobj := storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + originalPath)
		if err := fileobj.Delete(ctxStorage); err != nil {
			return err
		}
		if filetype == "image/gif" {
			fileobj = storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + placeholderPath + originalPath)
			blurobj := storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + placeholderPath + blurPath)
			if err := fileobj.Delete(ctxStorage); err != nil {
				return err
			}
			if err := blurobj.Delete(ctxStorage); err != nil {
				return err
			}
		} else {
			var hasblur = false
			for _, blurtype := range haveblur {
				if blurtype == filetype {
					hasblur = true
					break
				}
			}
			if hasblur {
				fileobj = storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + blurPath)
				if err := fileobj.Delete(ctxStorage); err != nil {
					return err
				}
			}
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
	itemIDString := itemID.Hex()
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
	var userItemsIndex string
	if itemType == formType {
		userItemsIndex = "forms"
	} else {
		userItemsIndex = "projects"
	}
	for _, accessUser := range access {
		var changedCurrentUser = false
		currentUserIDString := accessUser["id"].(string)
		accountUpdateData := bson.M{}
		if accessUser["type"] != nil || currentUserIDString == userIDString {
			changedCurrentUser = true
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
				accountUpdateData["$addToSet"] = bson.M{
					userItemsIndex: bson.M{
						"id":   itemIDString,
						"type": accessUser["type"],
					},
				}
			} else if currentUserIDString != userIDString {
				itemUpdateData["$unset"].(bson.M)["access"] = bson.M{
					currentUserIDString: 1,
				}
				accountUpdateData["$pull"] = bson.M{
					fmt.Sprintf("%s.id", userItemsIndex): itemIDString,
				}
			}
			if currentUserIDString == userIDString {
				itemUpdateData["$set"].(bson.M)["access"].(bson.M)[currentUserIDString] = bson.M{
					"categories": categories,
					"tags":       tags,
				}
			}
		}
		if changedCurrentUser {
			currentUserID, err := primitive.ObjectIDFromHex(currentUserIDString)
			if err != nil {
				return nil, err
			}
			_, err = userCollection.UpdateOne(ctxMongo, bson.M{
				"_id": currentUserID,
			}, accountUpdateData)
			if err != nil {
				return nil, err
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
	projectIDString := projectID.Hex()
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
		_, err := formCollection.UpdateOne(ctxMongo, bson.M{
			"_id": projectID,
		}, projectUpdateData)
		if err != nil {
			return err
		}
		userIDString, ok := accessUser["id"].(string)
		if !ok {
			return errors.New("cannot cast user id to string")
		}
		userID, err := primitive.ObjectIDFromHex(userIDString)
		if err != nil {
			return err
		}
		accountUpdateData := bson.M{}
		if accessUser["type"] != noAccessLevel {
			accountUpdateData["$addToSet"] = bson.M{
				"projects": bson.M{
					"id":   projectIDString,
					"type": accessUser["type"],
				},
			}
		} else {
			accountUpdateData["$pull"] = bson.M{
				"forms.id": projectIDString,
			}
		}
		_, err = userCollection.UpdateOne(ctxMongo, bson.M{
			"_id": userID,
		}, accountUpdateData)
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
			"addForm":       formMutationFields["addForm"],
			"updateForm":    formMutationFields["updateForm"],
			"deleteForm":    formMutationFields["deleteForm"],
			"deleteUser":    userMutationFields["deleteUser"],
			"deleteAccount": userMutationFields["deleteAccount"],
			"addProject":    projectMutationFields["addProject"],
			"updateProject": projectMutationFields["updateProject"],
			"deleteProject": projectMutationFields["deleteProject"],
			"addBlog":       blogMutationFields["addBlog"],
			"updateBlog":    blogMutationFields["updateBlog"],
			"deleteBlog":    blogMutationFields["deleteBlog"],
		},
	})
}
