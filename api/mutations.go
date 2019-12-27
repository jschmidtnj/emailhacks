package main

import (
	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func addUserFormAccess(userID primitive.ObjectID, accessType string, formIDString string) error {
	_, err := userCollection.UpdateOne(ctxMongo, bson.M{
		"_id": userID,
	}, bson.M{
		"$addToSet": bson.M{
			"forms": bson.M{
				"id":   formIDString,
				"type": accessType,
			},
		},
	})
	if err != nil {
		return err
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
			"updateFormAccess":       formMutationFields["updateFormAccess"],
			"deleteForm":             formMutationFields["deleteForm"],
			"deleteUser":             userMutationFields["deleteUser"],
			"deleteAccount":          userMutationFields["deleteAccount"],
			"addBlog":                blogMutationFields["addBlog"],
			"updateBlog":             blogMutationFields["updateBlog"],
			"deleteBlog":             blogMutationFields["deleteBlog"],
		},
	})
}
