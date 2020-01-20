package main

import (
	"errors"
	"time"

	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic/v7"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func changeFormProject(formIDString string, oldProjectIDString string, newProjectIDString string, accessToken string) error {
	if len(oldProjectIDString) > 0 {
		projectID, err := primitive.ObjectIDFromHex(oldProjectIDString)
		if err != nil {
			return err
		}
		script := elastic.NewScriptInline("ctx._source.forms-=1")
		_, err = elasticClient.Update().
			Index(projectElasticIndex).
			Type(projectElasticType).
			Id(oldProjectIDString).
			Script(script).
			Do(ctxElastic)
		if err != nil {
			return err
		}
		_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
			"_id": projectID,
		}, bson.M{
			"$inc": bson.M{
				"forms": -1,
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
		if _, _, err := checkProjectAccess(projectID, accessToken, "", editAccessLevel, false); err != nil {
			return err
		}
		script := elastic.NewScriptInline("ctx._source.forms+=1")
		_, err = elasticClient.Update().
			Index(projectElasticIndex).
			Type(projectElasticType).
			Id(newProjectIDString).
			Script(script).
			Do(ctxElastic)
		if err != nil {
			return err
		}
		_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
			"_id": projectID,
		}, bson.M{
			"$inc": bson.M{
				"forms": 1,
			},
		})
		if err != nil {
			return err
		}
	}
	if len(oldProjectIDString) > 0 && len(newProjectIDString) > 0 {
		sourceContext := elastic.NewFetchSourceContext(true).Include("id")
		mustQueries := []elastic.Query{
			elastic.NewTermsQuery("form", formIDString),
		}
		query := elastic.NewBoolQuery().Must(mustQueries...)
		searchResult, err := elasticClient.Search().
			Index(responseElasticIndex).
			Query(query).
			Pretty(isDebug()).
			FetchSourceContext(sourceContext).
			Do(ctxElastic)
		if err != nil {
			return err
		}
		for _, hit := range searchResult.Hits.Hits {
			responseIDString := hit.Id
			responseID, err := primitive.ObjectIDFromHex(responseIDString)
			if err != nil {
				return err
			}
			_, err = responseCollection.UpdateOne(ctxMongo, bson.M{
				"_id": responseID,
			}, bson.M{
				"$set": bson.M{
					"project": newProjectIDString,
				},
			})
			if err != nil {
				return err
			}
			_, err = elasticClient.Update().
				Index(responseElasticIndex).
				Type(responseElasticType).
				Id(responseIDString).
				Doc(bson.M{
					"project": newProjectIDString,
				}).
				Do(ctxElastic)
			if err != nil {
				return err
			}
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
			"project": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(FormItemInputType),
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
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := getTokenData(accessToken)
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
			planIDString, ok := claims["plan"].(string)
			if !ok {
				return nil, errors.New("cannot convert plan to string")
			}
			var planID primitive.ObjectID
			if len(planIDString) > 0 {
				planID, err = primitive.ObjectIDFromHex(planIDString)
				if err != nil {
					return nil, err
				}
			} else {
				planID = primitive.NilObjectID
			}
			productData, err := getProduct(planID, !isDebug())
			if err != nil {
				return nil, err
			}
			mustQueries := []elastic.Query{
				elastic.NewTermsQuery("owner", userIDString),
			}
			query := elastic.NewBoolQuery().Must(mustQueries...)
			numFormsAlready, err := elasticClient.Count().
				Type(formElasticType).
				Query(query).
				Pretty(false).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			if numFormsAlready >= int64(productData.MaxForms) {
				return nil, errors.New("you reached the maximum amount of forms")
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
				if err := checkFormItemObjCreate(item); err != nil {
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
				"updated":   now.Unix(),
				"project":   project,
				"name":      name,
				"items":     items,
				"multiple":  multiple,
				"views":     0,
				"responses": 0,
				"owner":     userIDString,
				"access": bson.M{
					userIDString: bson.M{
						"type":       userAccess[0]["type"].(string),
						"categories": categories,
						"tags":       tags,
					},
				},
				"files":  files,
				"public": noAccessLevel,
			}
			formCreateRes, err := formCollection.InsertOne(ctxMongo, formData)
			if err != nil {
				return nil, err
			}
			formID := formCreateRes.InsertedID.(primitive.ObjectID)
			formIDString := formID.Hex()
			if err = changeFormProject(formIDString, "", project, accessToken); err != nil {
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
			"project": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(FormItemInputType),
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
			formData, err := checkFormAccess(formID, accessToken, editAccessLevel, true)
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
				projectID, err := primitive.ObjectIDFromHex(formData.Project)
				if err != nil {
					return nil, err
				}
				for _, accessUser := range access {
					if accessUser["type"] != nil {
						var changeProjectAccess = false
						var newAccessType = accessUser["type"].(string)
						_, currentAccessType, err := checkProjectAccess(projectID, accessToken, "", viewAccessLevel, false)
						if err != nil {
							changeProjectAccess = true
						}
						if currentAccessType == newAccessType {
							continue
						}
						if currentAccessType == sharedAccessLevel && newAccessType == noAccessLevel {
							changeProjectAccess = true
						}
						if changeProjectAccess {
							projectAccess := []map[string]interface{}{accessUser}
							projectUpdateDataDB, err := changeUserAccessData(projectID, projectType, userIDString, nil, nil, projectAccess)
							if err != nil {
								return nil, err
							}
							script := elastic.NewScript(addRemoveAccessScript).Params(map[string]interface{}{
								"access":     projectAccess,
								"tags":       nil,
								"categories": nil,
							})
							_, err = elasticClient.Update().
								Index(projectElasticIndex).
								Type(projectElasticType).
								Id(projectID.Hex()).
								Script(script).
								Do(ctxElastic)
							if err != nil {
								return nil, err
							}
							_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
								"_id": projectID,
							}, projectUpdateDataDB)
							if err != nil {
								return nil, err
							}
						}
					}
				}
			} else {
				updateDataDB = bson.M{}
			}
			if updateDataDB["$set"] == nil {
				updateDataDB["$set"] = bson.M{}
			}
			newAccess, oldTags, oldCategories, err := getFormattedAccessGQLData(formData.Access, access, userIDString)
			if err != nil {
				return nil, err
			}
			formData.Access = newAccess
			if params.Args["categories"] == nil {
				formData.Categories = oldCategories
			}
			if params.Args["tags"] == nil {
				formData.Tags = oldTags
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
				oldProject = formData.Project
				updateDataDB["$set"].(bson.M)["project"] = newProject
				formData.Project = newProject
				updateDataElastic["project"] = newProject
			}
			if params.Args["name"] != nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				updateDataDB["$set"].(bson.M)["name"] = name
				formData.Name = name
				updateDataElastic["name"] = name
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				updateDataDB["$set"].(bson.M)["multiple"] = multiple
				formData.Multiple = multiple
				updateDataElastic["multiple"] = multiple
			}
			if params.Args["items"] != nil {
				itemsInterface, ok := params.Args["items"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting items to interface array")
				}
				itemsMap, err := interfaceListToMapList(itemsInterface)
				if err != nil {
					return nil, err
				}
				for _, item := range itemsMap {
					if err := checkFormItemObjUpdate(item); err != nil {
						return nil, err
					}
				}
				items := make([]*FormItem, len(itemsMap))
				for i, item := range itemsMap {
					if err = mapstructure.Decode(item, &items[i]); err != nil {
						return nil, err
					}
				}
				updateDataDB["$set"].(bson.M)["items"] = items
				formData.Items = items
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
				formData.Public = public
				updateDataElastic["public"] = public
			}
			if params.Args["files"] != nil {
				filesinterface, ok := params.Args["files"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting files to interface array")
				}
				filesMap, err := interfaceListToMapList(filesinterface)
				if err != nil {
					return nil, err
				}
				for _, file := range filesMap {
					if err := checkFileObjUpdate(file); err != nil {
						return nil, err
					}
				}
				files := make([]*File, len(filesMap))
				for i, file := range filesMap {
					if err = mapstructure.Decode(file, &files[i]); err != nil {
						return nil, err
					}
				}
				formData.Files = files
				updateDataDB["$set"].(bson.M)["files"] = files
				updateDataElastic["files"] = files
			}
			formData.Access = access
			if len(access) > 0 {
				script := elastic.NewScriptInline(addRemoveAccessScript).Params(map[string]interface{}{
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
			updateDataElastic["updated"] = formData.Updated
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
				err = changeFormProject(formIDString, oldProject, newProject, accessToken)
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
				Type: graphql.NewList(UpdateFormItemInputType),
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
			if params.Args["updatesAccessToken"] == nil {
				return nil, errors.New("update token not found")
			}
			accessToken, ok := params.Args["updatesAccessToken"].(string)
			if !ok {
				return nil, errors.New("cannot cast update token to string")
			}
			tokenFormIDString, _, connectionIDString, err := getFormUpdateClaimsData(accessToken, editAccessLevel)
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
					if err := checkFormItemObjUpdatePart(item); err != nil {
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
			var form *Form
			if !justDeleteElastic {
				form, err = checkFormAccess(formID, accessToken, editAccessLevel, false)
				if err != nil {
					return nil, err
				}
				if err = changeFormProject(formIDString, form.Project, "", ""); err != nil {
					return nil, err
				}
			}
			form, err = deleteForm(formID, form, userIDString)
			if err != nil {
				return nil, err
			}
			return form, nil
		},
	},
}

func deleteForm(formID primitive.ObjectID, form *Form, userIDString string) (*Form, error) {
	formIDString := formID.Hex()
	if !justDeleteElastic {
		var err error
		if form == nil {
			form, err = getForm(formID, false)
			if err != nil {
				return nil, err
			}
		}
		if err = deleteAllResponses(formID); err != nil {
			return nil, err
		}
		access, tags, categories, err := getFormattedAccessGQLData(form.Access, nil, userIDString)
		if err != nil {
			return nil, err
		}
		form.Responses = 0
		form.Access = access
		form.Categories = categories
		form.Tags = tags
	}
	_, err := elasticClient.Delete().
		Index(formElasticIndex).
		Type(formElasticType).
		Id(formIDString).
		Do(ctxElastic)
	if err != nil {
		return nil, err
	}
	if !justDeleteElastic {
		_, err = formCollection.DeleteOne(ctxMongo, bson.M{
			"_id": formID,
		})
		if err != nil {
			return nil, err
		}
		for _, file := range form.Files {
			fileobj := storageBucket.Object(formFileIndex + "/" + formIDString + "/" + file.ID + originalPath)
			if err := fileobj.Delete(ctxStorage); err != nil {
				return nil, err
			}
			if file.Type == "image/gif" {
				fileobj = storageBucket.Object(formFileIndex + "/" + formIDString + "/" + file.ID + placeholderPath + originalPath)
				blurobj := storageBucket.Object(formFileIndex + "/" + formIDString + "/" + file.ID + placeholderPath + blurPath)
				if err := fileobj.Delete(ctxStorage); err != nil {
					return nil, err
				}
				if err := blurobj.Delete(ctxStorage); err != nil {
					return nil, err
				}
			} else {
				var hasblur = false
				for _, blurtype := range haveblur {
					if blurtype == file.Type {
						hasblur = true
						break
					}
				}
				if hasblur {
					fileobj = storageBucket.Object(formFileIndex + "/" + formIDString + "/" + file.ID + blurPath)
					if err := fileobj.Delete(ctxStorage); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return form, nil
}
