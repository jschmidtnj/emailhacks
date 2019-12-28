package main

import (
	"errors"

	"github.com/graphql-go/graphql"

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
			userAccess := []map[string]interface{}{{
				"id":   userIDString,
				"type": editAccessLevel[0],
			}}
			projectData := bson.M{
				"name":  name,
				"forms": bson.A{},
				"access": bson.A{
					userAccess,
				},
				"public":     validAccessTypes[3],
				"tags":       tags,
				"categories": categories,
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
			timestamp := objectidtimestamp(projectID)
			projectData["date"] = timestamp.Unix()
			/*
				_, err = elasticClient.Index().
					Index(projectElasticIndex).
					Type(projectElasticType).
					Id(projectIDString).
					BodyJson(projectData).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
			*/
			projectData["date"] = timestamp.Format(dateFormat)
			projectData["id"] = projectIDString
			return projectData, nil
		},
	},
	"updateProjectOrganization": &graphql.Field{
		Type:        ProjectType,
		Description: "Update Project Organization",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"type": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"value": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"add": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
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
			projectData, _, err := checkProjectAccess(projectID, accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			if params.Args["type"] == nil {
				return nil, errors.New("type not provided")
			}
			if params.Args["value"] == nil {
				return nil, errors.New("value not provided")
			}
			if params.Args["add"] == nil {
				return nil, errors.New("add not provided")
			}
			thetype, ok := params.Args["type"].(string)
			if !ok {
				return nil, errors.New("problem casting type to string")
			}
			if !findInArray(thetype, validOrganization) {
				return nil, errors.New("invalid type given")
			}
			value, ok := params.Args["value"].(string)
			if !ok {
				return nil, errors.New("problem casting value to string")
			}
			add, ok := params.Args["value"].(bool)
			if !ok {
				return nil, errors.New("problem casting add to boolean")
			}
			updateData := bson.M{}
			valueUpdate := bson.M{
				thetype: value,
			}
			if add {
				updateData["$set"] = valueUpdate
				updateData["$upsert"] = true
			} else {
				updateData["$pull"] = valueUpdate
			}
			_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
				"_id": projectID,
			}, updateData)
			if err != nil {
				return nil, err
			}
			timestamp := objectidtimestamp(projectID)
			projectData["date"] = timestamp.Format(dateFormat)
			projectData["id"] = projectIDString
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
			projectData, _, err := checkProjectAccess(projectID, accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			updateData := bson.M{}
			if params.Args["title"] != nil {
				title, ok := params.Args["title"].(string)
				if !ok {
					return nil, errors.New("problem casting title to string")
				}
				updateData["title"] = title
				projectData["title"] = title
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				updateData["multiple"] = multiple
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
				updateData["categories"] = categories
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
				updateData["tags"] = tags
			}
			if params.Args["public"] != nil {
				public, ok := params.Args["public"].(string)
				if !ok {
					return nil, errors.New("problem casting public to string")
				}
				if !findInArray(public, validAccessTypes) {
					return nil, errors.New("invalid public access level")
				}
				updateData["public"] = public
				projectData["public"] = public
			}
			var access []map[string]interface{}
			if params.Args["access"] != nil {
				accessInterface, ok := params.Args["access"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting access to interface array")
				}
				access, err := interfaceListToMapList(accessInterface)
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
				if err = changeUserFormAccess(projectID, access); err != nil {
					return nil, err
				}
				updateData["access"] = access
			}
			projectData["access"] = access
			_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
				"_id": projectID,
			}, bson.M{
				"$set": updateData,
			})
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
			projectData, access, err := checkProjectAccess(projectID, accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			for _, user := range access {
				accessUserIDString, ok := user["id"].(string)
				if !ok {
					return nil, errors.New("cannot cast user id in project to string")
				}
				accessUserID, err := primitive.ObjectIDFromHex(accessUserIDString)
				if err != nil {
					return nil, err
				}
				_, err = userCollection.UpdateOne(ctxMongo, bson.M{
					"_id": accessUserID,
				}, bson.M{
					"$pull": bson.M{
						"projects": bson.M{
							"id": projectIDString,
						},
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
				_, err = formCollection.DeleteOne(ctxMongo, bson.M{
					"_id": formID,
				})
				if err != nil {
					return nil, err
				}
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
