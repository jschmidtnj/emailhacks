package main

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/olivere/elastic/v7"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var projectMutationFields = graphql.Fields{
	"addProject": &graphql.Field{
		Type:        ProjectType,
		Description: "Create a Project",
		Args: graphql.FieldConfigArgument{
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"tags": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			_, err = primitive.ObjectIDFromHex(userIDString)
			if err != nil {
				return nil, err
			}
			if params.Args["name"] == nil {
				return nil, errors.New("name not provided")
			}
			if params.Args["tags"] == nil {
				return nil, errors.New("tags not provided")
			}
			if params.Args["categories"] == nil {
				return nil, errors.New("categories not provided")
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			name, ok := params.Args["name"].(string)
			if !ok {
				return nil, errors.New("problem casting name to string")
			}
			categoriesInterface, ok := params.Args["categories"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting categories to interface array")
			}
			categories, err := interfaceListToStringList(categoriesInterface)
			if err != nil {
				return nil, err
			}
			tagsInterface, ok := params.Args["tags"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting tags to interface array")
			}
			tags, err := interfaceListToStringList(tagsInterface)
			if err != nil {
				return nil, err
			}
			now := time.Now()
			userAccess := []map[string]interface{}{{
				"id":   userIDString,
				"type": editAccessLevel[0],
			}}
			projectData := bson.M{
				"updated": now.Unix(),
				"name":    name,
				"forms":   bson.A{},
				"access": bson.M{
					userIDString: bson.M{
						"type":       userAccess[0]["type"].(string),
						"categories": categories,
						"tags":       tags,
					},
				},
				"public": validAccessTypes[3],
			}
			projectCreateRes, err := projectCollection.InsertOne(ctxMongo, projectData)
			if err != nil {
				return nil, err
			}
			projectID := projectCreateRes.InsertedID.(primitive.ObjectID)
			projectIDString := projectID.Hex()
			if err = changeUserProjectAccess(projectID, userAccess); err != nil {
				return nil, err
			}
			projectData["created"] = now.Unix()
			_, err = elasticClient.Index().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				BodyJson(projectData).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			if formatDate {
				projectData["created"] = now.Format(dateFormat)
				projectData["updated"] = now.Format(dateFormat)
			}
			projectData["id"] = projectIDString
			delete(projectData["access"].(map[string]interface{}), userIDString)
			projectData["access"] = userAccess[0]
			projectData["tags"] = tags
			projectData["categories"] = categories
			return projectData, nil
		},
	},
	"updateProject": &graphql.Field{
		Type:        ProjectType,
		Description: "Update a Project",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"access": &graphql.ArgumentConfig{
				Type: graphql.NewList(AccessInputType),
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"tags": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"public": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := validateLoggedIn(accessToken)
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			if params.Args["id"] == nil {
				return nil, errors.New("project id not provided")
			}
			projectIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast project id to string")
			}
			projectID, err := primitive.ObjectIDFromHex(projectIDString)
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
			projectData, err := checkProjectAccess(projectID, accessToken, editAccessLevel, formatDate, true)
			if err != nil {
				return nil, err
			}
			var categories []string
			if params.Args["categories"] != nil {
				categoriesInterface, ok := params.Args["categories"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting categories to interface array")
				}
				categories, err = interfaceListToStringList(categoriesInterface)
				if err != nil {
					return nil, err
				}
			}
			var tags []string
			if params.Args["tags"] != nil {
				tagsInterface, ok := params.Args["tags"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting tags to interface array")
				}
				tags, err = interfaceListToStringList(tagsInterface)
				if err != nil {
					return nil, err
				}
			}
			var updateData bson.M
			var access []map[string]interface{}
			if params.Args["access"] != nil {
				accessInterface, ok := params.Args["access"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting access to interface array")
				}
				access, err = interfaceListToMapList(accessInterface)
				if err != nil {
					return nil, err
				}
				for _, accessUser := range access {
					if err := checkAccessObj(accessUser); err != nil {
						return nil, err
					}
				}
				_, err = interfaceListToMapList(accessInterface)
				if err != nil {
					return nil, err
				}
				updateData, err = changeUserAccessData(projectID, projectType, userIDString, categories, tags, access)
				if err != nil {
					return nil, err
				}
			} else {
				updateData = bson.M{}
			}
			oldCategories, oldTags, newAccess := getFormattedGQLData(projectData, access, userIDString)
			projectData["access"] = newAccess
			if params.Args["categories"] == nil {
				projectData["categories"] = oldCategories
			}
			if params.Args["tags"] == nil {
				projectData["tags"] = oldTags
			}
			if params.Args["title"] != nil {
				title, ok := params.Args["title"].(string)
				if !ok {
					return nil, errors.New("problem casting title to string")
				}
				updateData["$set"].(bson.M)["title"] = title
				projectData["title"] = title
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				updateData["$set"].(bson.M)["multiple"] = multiple
				projectData["multiple"] = multiple
			}
			if params.Args["categories"] != nil {
				categoriesInterface, ok := params.Args["categories"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting categories to interface array")
				}
				categories, err := interfaceListToStringList(categoriesInterface)
				if err != nil {
					return nil, err
				}
				updateData["$set"].(bson.M)["categories"] = categories
			}
			if params.Args["tags"] != nil {
				tagsInterface, ok := params.Args["tags"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting tags to interface array")
				}
				tags, err := interfaceListToStringList(tagsInterface)
				if err != nil {
					return nil, err
				}
				updateData["$set"].(bson.M)["tags"] = tags
			}
			if params.Args["public"] != nil {
				public, ok := params.Args["public"].(string)
				if !ok {
					return nil, errors.New("problem casting public to string")
				}
				if !findInArray(public, validAccessTypes) {
					return nil, errors.New("invalid public access level")
				}
				updateData["$set"].(bson.M)["public"] = public
				projectData["public"] = public
			}
			projectData["access"] = access
			script := elastic.NewScript(addRemoveAccessScript).Lang("painless").Params(map[string]interface{}{
				"access":     access,
				"tags":       tags,
				"categories": categories,
			})
			_, err = elasticClient.Update().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				Script(script).
				Doc(updateData).
				Do(ctxElastic)
			_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
				"_id": projectID,
			}, updateData)
			if err != nil {
				return nil, err
			}
			return projectData, nil
		},
	},
	"deleteProject": &graphql.Field{
		Type:        ProjectType,
		Description: "Delete a Project",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := validateLoggedIn(accessToken)
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			if params.Args["id"] != nil {
				return nil, errors.New("project id not provided")
			}
			projectIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast project id to string")
			}
			projectID, err := primitive.ObjectIDFromHex(projectIDString)
			if err != nil {
				return nil, err
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			projectData, err := checkProjectAccess(projectID, accessToken, editAccessLevel, formatDate, false)
			if err != nil {
				return nil, err
			}
			categories, tags, access := getFormattedGQLData(projectData, nil, userIDString)
			projectData["categories"] = categories
			projectData["tags"] = tags
			for _, accessUserIDString := range access {
				accessUserID, err := primitive.ObjectIDFromHex(accessUserIDString)
				if err != nil {
					return nil, err
				}
				_, err = userCollection.UpdateOne(ctxMongo, bson.M{
					"_id": accessUserID,
				}, bson.M{
					"$pull": bson.M{
						"projects.id": projectIDString,
					},
				})
				if err != nil {
					return nil, err
				}
			}
			formsPrimitive, ok := projectData["forms"].(primitive.A)
			if !ok {
				return nil, errors.New("problem casting forms to array")
			}
			for _, formIDInterface := range formsPrimitive {
				formIDString, ok := formIDInterface.(string)
				if !ok {
					return nil, errors.New("cannot cast form id to string")
				}
				formID, err := primitive.ObjectIDFromHex(formIDString)
				if err != nil {
					return nil, err
				}
				err = deleteForm(formID, nil, formatDate)
				if err != nil {
					return nil, err
				}
			}
			_, err = elasticClient.Delete().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			_, err = projectCollection.DeleteOne(ctxMongo, bson.M{
				"_id": projectID,
			})
			if err != nil {
				return nil, err
			}
			return projectData, nil
		},
	},
}
