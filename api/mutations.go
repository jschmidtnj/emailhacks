package main

import (
	"errors"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	if accessObj["type"] == nil {
		return errors.New("no type field given")
	}
	accessType, ok := accessObj["type"].(string)
	if !ok {
		return errors.New("cannot cast type to string")
	}
	if !findInArray(accessType, validAccessTypes) {
		return errors.New("invalid access type given")
	}
	return nil
}

func changeUserFormAccess(formID primitive.ObjectID, access []map[string]interface{}) error {
	for _, accessUser := range access {
		if err := checkAccessObj(accessUser); err != nil {
			return err
		}
	}
	formIDString := formID.Hex()
	for _, accessUser := range access {
		formUpdateData := bson.M{}
		if accessUser["type"] != noAccessLevel {
			formUpdateData["$addToSet"] = bson.M{
				"access": bson.M{
					"id":   accessUser["id"],
					"type": accessUser["type"],
				},
			}
		} else {
			formUpdateData["$pull"] = bson.M{
				"access.id": accessUser["id"],
			}
		}
		_, err := formCollection.UpdateOne(ctxMongo, bson.M{
			"_id": formID,
		}, formUpdateData)
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
				"forms": bson.M{
					"id":   formIDString,
					"type": accessUser["type"],
				},
			}
		} else {
			accountUpdateData["$pull"] = bson.M{
				"forms.id": formIDString,
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
			"addForm":                formMutationFields["addForm"],
			"updateForm":             formMutationFields["updateForm"],
			"updateFormOrganization": formMutationFields["updateFormOrganization"],
			"deleteForm":             formMutationFields["deleteForm"],
			"deleteUser":             userMutationFields["deleteUser"],
			"deleteAccount":          userMutationFields["deleteAccount"],
			"addProject":             projectMutationFields["addProject"],
			"updateProject":          projectMutationFields["updateProject"],
			"deleteProject":          projectMutationFields["deleteProject"],
			"addBlog":                blogMutationFields["addBlog"],
			"updateBlog":             blogMutationFields["updateBlog"],
			"deleteBlog":             blogMutationFields["deleteBlog"],
		},
	})
}
