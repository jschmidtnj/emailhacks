package main

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func changeFormProject(formIDString string, oldProjectIDString string, newProjectIDString string) error {
	if len(oldProjectIDString) > 0 {
		projectID, err := primitive.ObjectIDFromHex(oldProjectIDString)
		if err != nil {
			return err
		}
		_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
			"_id": projectID,
		}, bson.M{
			"$pull": bson.M{
				"forms": formIDString,
			},
		})
		if err != nil {
			return err
		}
	}
	if len(newProjectIDString) > 0 {
		projectID, err := primitive.ObjectIDFromHex(newProjectIDString)
		if err != nil {
			return err
		}
		_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
			"_id": projectID,
		}, bson.M{
			"$addToSet": bson.M{
				"forms": formIDString,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

var updateFormPath = "update-form-"

var formMutationFields = graphql.Fields{
	"addForm": &graphql.Field{
		Type:        FormType,
		Description: "Create a Form",
		Args: graphql.FieldConfigArgument{
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"project": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(ItemInputType),
			},
			"multiple": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"tags": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
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
			if params.Args["project"] == nil {
				return nil, errors.New("project not provided")
			}
			if params.Args["name"] == nil {
				return nil, errors.New("name not provided")
			}
			if params.Args["items"] == nil {
				return nil, errors.New("items not provided")
			}
			if params.Args["tags"] == nil {
				return nil, errors.New("tags not provided")
			}
			if params.Args["categories"] == nil {
				return nil, errors.New("categories not provided")
			}
			if params.Args["files"] == nil {
				return nil, errors.New("files not provided")
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			project, ok := params.Args["project"].(string)
			if !ok {
				return nil, errors.New("problem casting project id to string")
			}
			name, ok := params.Args["name"].(string)
			if !ok {
				return nil, errors.New("problem casting name to string")
			}
			itemsInterface, ok := params.Args["items"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting items to interface array")
			}
			items, err := interfaceListToMapList(itemsInterface)
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				if err := checkItemObjCreate(item); err != nil {
					return nil, err
				}
			}
			multiple, ok := params.Args["multiple"].(bool)
			if !ok {
				return nil, errors.New("problem casting multiple to boolean")
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
			filesinterface, ok := params.Args["files"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting files to interface array")
			}
			files, err := interfaceListToMapList(filesinterface)
			if err != nil {
				return nil, err
			}
			for _, file := range files {
				if err := checkFileObjCreate(file); err != nil {
					return nil, err
				}
			}
			now := time.Now()
			userAccess := []map[string]interface{}{{
				"id":   userIDString,
				"type": editAccessLevel[0],
			}}
			formData := bson.M{
				"updated":  now.Unix(),
				"project":  project,
				"name":     name,
				"items":    items,
				"multiple": multiple,
				"views":    0,
				"access": bson.M{
					userIDString: bson.M{
						"type":       userAccess[0]["type"].(string),
						"categories": categories,
						"tags":       tags,
					},
				},
				"files":  files,
				"public": validAccessTypes[3],
			}
			formCreateRes, err := formCollection.InsertOne(ctxMongo, formData)
			if err != nil {
				return nil, err
			}
			formID := formCreateRes.InsertedID.(primitive.ObjectID)
			formIDString := formID.Hex()
			if _, err = changeUserAccessData(formID, formType, userIDString, categories, tags, userAccess); err != nil {
				return nil, err
			}
			if err = changeFormProject(formIDString, "", project); err != nil {
				return nil, err
			}
			formData["created"] = now.Unix()
			_, err = elasticClient.Index().
				Index(formElasticIndex).
				Type(formElasticType).
				Id(formIDString).
				BodyJson(formData).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			if formatDate {
				formData["created"] = now.Format(dateFormat)
				formData["updated"] = now.Format(dateFormat)
			}
			formData["id"] = formIDString
			delete(formData["access"].(bson.M), userIDString)
			formData["access"] = userAccess[0]
			formData["tags"] = tags
			formData["categories"] = categories
			return formData, nil
		},
	},
	"updateForm": &graphql.Field{
		Type:        FormType,
		Description: "Update a Form",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"project": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(ItemInputType),
			},
			"multiple": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"access": &graphql.ArgumentConfig{
				Type: graphql.NewList(AccessInputType),
			},
			"tags": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"public": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			}, // eventually get files and tags and categories updating piece by piece
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := getTokenData(accessToken)
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			if params.Args["id"] == nil {
				return nil, errors.New("form id not provided")
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast form id to string")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
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
			formData, err := checkFormAccess(formID, accessToken, editAccessLevel, formatDate, true)
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
			var updateDataDB bson.M
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
				updateDataDB, err = changeUserAccessData(formID, formType, userIDString, categories, tags, access)
				if err != nil {
					return nil, err
				}
			} else {
				updateDataDB = bson.M{}
			}
			if updateDataDB["$set"] == nil {
				updateDataDB["$set"] = bson.M{}
			}
			newAccess, oldTags, oldCategories, err := getFormattedGQLData(formData, access, userIDString)
			if err != nil {
				return nil, err
			}
			formData["access"] = newAccess
			if params.Args["categories"] == nil {
				formData["categories"] = oldCategories
			}
			if params.Args["tags"] == nil {
				formData["tags"] = oldTags
			}
			updateDataElastic := bson.M{}
			var updateProject = false
			var newProject string
			var oldProject string
			if params.Args["project"] != nil {
				updateProject = true
				newProject, ok = params.Args["project"].(string)
				if !ok {
					return nil, errors.New("problem casting new project to string")
				}
				oldProject, ok = formData["project"].(string)
				if !ok {
					return nil, errors.New("problem casting old project to string")
				}
				updateDataDB["$set"].(bson.M)["project"] = newProject
				formData["project"] = newProject
				updateDataElastic["project"] = newProject
			}
			if params.Args["name"] != nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				updateDataDB["$set"].(bson.M)["name"] = name
				formData["name"] = name
				updateDataElastic["name"] = name
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				updateDataDB["$set"].(bson.M)["multiple"] = multiple
				formData["multiple"] = multiple
				updateDataElastic["multiple"] = multiple
			}
			if params.Args["items"] != nil {
				itemsInterface, ok := params.Args["items"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting items to interface array")
				}
				items, err := interfaceListToMapList(itemsInterface)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if err := checkItemObjUpdate(item); err != nil {
						return nil, err
					}
				}
				updateDataDB["$set"].(bson.M)["items"] = items
				formData["items"] = items
				updateDataElastic["items"] = items
			}
			if params.Args["public"] != nil {
				public, ok := params.Args["public"].(string)
				if !ok {
					return nil, errors.New("problem casting public to string")
				}
				if !findInArray(public, validAccessTypes) {
					return nil, errors.New("invalid public access level")
				}
				updateDataDB["$set"].(bson.M)["public"] = public
				formData["public"] = public
				updateDataElastic["public"] = public
			}
			if params.Args["files"] != nil {
				filesinterface, ok := params.Args["files"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting files to interface array")
				}
				files, err := interfaceListToMapList(filesinterface)
				if err != nil {
					return nil, err
				}
				for _, file := range files {
					if err := checkFileObjUpdate(file); err != nil {
						return nil, err
					}
				}
				updateDataDB["$set"].(bson.M)["files"] = files
				updateDataElastic["files"] = files
			}
			formData["access"] = access
			if len(access) > 0 {
				script := elastic.NewScriptInline(addRemoveAccessScript).Lang("painless").Params(map[string]interface{}{
					"access":     access,
					"tags":       tags,
					"categories": categories,
				})
				_, err = elasticClient.Update().
					Index(formElasticIndex).
					Type(formElasticType).
					Id(formIDString).
					Script(script).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
			}
			_, err = elasticClient.Update().
				Index(formElasticIndex).
				Type(formElasticType).
				Id(formIDString).
				Doc(updateDataElastic).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, updateDataDB)
			if err != nil {
				return nil, err
			}
			if updateProject {
				err = changeFormProject(formIDString, oldProject, newProject)
				if err != nil {
					return nil, err
				}
			}
			return formData, nil
		},
	},
	"updateFormPart": &graphql.Field{
		Type:        FormUpdateType,
		Description: "Update a Form",
		Args: graphql.FieldConfigArgument{
			"updatesAccessToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(UpdateItemInputType),
			},
			"multiple": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"public": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(UpdateFileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// need to have input of either updatesAccessToken or just params
			// save to redis database the updated parts
			// update access and tags immediately - no need to send to everyone
			// output just what was updated
			if params.Args["id"] == nil {
				return nil, errors.New("form id not provided")
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast form id to string")
			}
			_, err := primitive.ObjectIDFromHex(formIDString)
			if err != nil {
				return nil, err
			}
			accessToken, ok := params.Args["updatesAccessToken"].(string)
			if !ok {
				return nil, errors.New("cannot cast update token to string")
			}
			tokenFormIDString, _, connectionIDString, err := getUpdateClaimsData(accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			if tokenFormIDString != formIDString {
				return nil, errors.New("form id in token does not match given form id")
			}
			var updateData map[string]interface{}
			newUpdateData := map[string]interface{}{
				"id": connectionIDString,
			}
			savedUpdateData, err := redisClient.Get(updateFormPath + formIDString).Result()
			if err != nil {
				updateData = map[string]interface{}{}
			} else {
				err = json.UnmarshalFromString(savedUpdateData, &updateData)
				if err != nil {
					return nil, err
				}
			}
			if params.Args["name"] != nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				updateData["name"] = name
				newUpdateData["name"] = name
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				updateData["multiple"] = multiple
				newUpdateData["multiple"] = multiple
			}
			if params.Args["items"] != nil {
				itemsInterface, ok := params.Args["items"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting items to interface array")
				}
				items, err := interfaceListToMapList(itemsInterface)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if err := checkItemObjUpdatePart(item); err != nil {
						return nil, err
					}
				}
				newUpdateData["items"] = items
				if updateData["items"] != nil {
					currentItemUpdates, err := interfaceListToMapList(updateData["items"].([]interface{}))
					if err != nil {
						return nil, err
					}
					items = append(currentItemUpdates, items...)
				}
				updateData["items"] = items
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
				newUpdateData["public"] = public
			}
			if params.Args["files"] != nil {
				filesinterface, ok := params.Args["files"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting files to interface array")
				}
				files, err := interfaceListToMapList(filesinterface)
				if err != nil {
					return nil, err
				}
				for _, file := range files {
					if err := checkFileObjUpdatePart(file); err != nil {
						return nil, err
					}
				}
				newUpdateData["files"] = files
				if updateData["files"] != nil {
					currentFileUpdates, err := interfaceListToMapList(updateData["files"].([]interface{}))
					if err != nil {
						return nil, err
					}
					files = append(currentFileUpdates, files...)
				}
				updateData["files"] = files
			}
			updateDataJSON, err := json.Marshal(updateData)
			if err != nil {
				return nil, err
			}
			err = redisClient.Set(updateFormPath+formIDString, updateDataJSON, time.Second*time.Duration(autosaveTime*2)).Err()
			if err != nil {
				return nil, err
			}
			// save update task
			msg := saveFormTask.WithArgs(ctxMessageQueue, formIDString).OnceInPeriod(time.Duration(autosaveTime) * time.Second)
			msg.Delay = time.Duration(autosaveTime) * time.Second
			if err = messageQueue.Add(msg); err != nil {
				logger.Info("update already saved: " + err.Error())
			}
			newUpdateDataJSON, err := json.Marshal(newUpdateData)
			if err != nil {
				return nil, err
			}
			// send to other clients
			err = redisClient.Publish(updateFormPath+formIDString, newUpdateDataJSON).Err()
			if err != nil {
				return nil, err
			}
			return updateData, nil
		},
	},
	"deleteForm": &graphql.Field{
		Type:        FormType,
		Description: "Delete a Form",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := getTokenData(accessToken)
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			if params.Args["id"] == nil {
				return nil, errors.New("form id not provided")
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast form id to string")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
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
			var formData map[string]interface{}
			if justDeleteElastic {
				formData = nil
			} else {
				formData, err := checkFormAccess(formID, accessToken, editAccessLevel, formatDate, false)
				if err != nil {
					return nil, err
				}
				oldProject, ok := formData["project"].(string)
				if !ok {
					return nil, errors.New("problem casting old project to string")
				}
				if err = changeFormProject(formIDString, oldProject, ""); err != nil {
					return nil, err
				}
			}
			formData, err = deleteForm(formID, formData, formatDate, userIDString)
			if err != nil {
				return nil, err
			}
			return formData, nil
		},
	},
}
