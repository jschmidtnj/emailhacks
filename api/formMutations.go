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

func checkItemObjCreate(itemObj map[string]interface{}) error {
	if itemObj["question"] == nil {
		return errors.New("no name field given")
	}
	if _, ok := itemObj["question"].(string); !ok {
		return errors.New("problem casting id to string")
	}
	if itemObj["type"] == nil {
		return errors.New("no type field given")
	}
	if _, ok := itemObj["type"].(string); !ok {
		return errors.New("problem casting name to string")
	}
	if itemObj["options"] == nil {
		return errors.New("no options field given")
	}
	optionsArray, ok := itemObj["options"].([]interface{})
	if !ok {
		return errors.New("problem casting options to interface array")
	}
	if _, err := interfaceListToStringList(optionsArray); err != nil {
		return errors.New("problem casting options to string array")
	}
	if itemObj["text"] == nil {
		return errors.New("no text field given")
	}
	if _, ok := itemObj["text"].(string); !ok {
		return errors.New("problem casting text to string")
	}
	if itemObj["required"] == nil {
		return errors.New("no required field given")
	}
	if _, ok := itemObj["required"].(bool); !ok {
		return errors.New("problem casting required to boolean")
	}
	if itemObj["file"] == nil {
		return errors.New("no file field given")
	}
	if _, ok := itemObj["file"].(string); !ok {
		return errors.New("problem casting file to string")
	}
	return nil
}

func checkItemObjUpdate(itemObj map[string]interface{}) error {
	if itemObj["question"] != nil {
		if _, ok := itemObj["question"].(string); !ok {
			return errors.New("problem casting id to string")
		}
	}
	if itemObj["type"] != nil {
		if _, ok := itemObj["type"].(string); !ok {
			return errors.New("problem casting name to string")
		}
	}
	if itemObj["options"] != nil {
		optionsArray, ok := itemObj["options"].([]interface{})
		if !ok {
			return errors.New("problem casting options to interface array")
		}
		if _, err := interfaceListToStringList(optionsArray); err != nil {
			return errors.New("problem casting options to string array")
		}
	}
	if itemObj["text"] != nil {
		_, ok := itemObj["text"].(string)
		if !ok {
			return errors.New("problem casting text to string")
		}
	}
	if itemObj["required"] != nil {
		if _, ok := itemObj["required"].(bool); !ok {
			return errors.New("problem casting required to boolean")
		}
	}
	if itemObj["file"] != nil {
		if _, ok := itemObj["file"].(string); !ok {
			return errors.New("problem casting file to string")
		}
	}
	return nil
}

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
			formData["userId"] = userIDString
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
			"updateToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
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
				Type: graphql.NewList(StringArrayInputType),
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(StringArrayInputType),
			},
			"public": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// need to have input of either updateToken or just params
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
			savedUpdateData, err := redisClient.Get("update-form-" + formIDString).Result()
			if err != nil {
				return nil, err
			}
			var savedUpdateDataObj map[string]interface{}
			err = json.UnmarshalFromString(savedUpdateData, &savedUpdateDataObj)
			if err != nil {
				return nil, err
			}
			logger.Info(savedUpdateData)
			argsData, err := json.Marshal(params.Args)
			if err != nil {
				return nil, err
			}
			redisClient.Set("update-form-"+formIDString, argsData, 0)
			var userIDString string
			var formData bson.M
			var foundUpdateToken = params.Args["updateToken"] != nil
			if foundUpdateToken {
				accessToken, ok := params.Args["updateToken"].(string)
				if !ok {
					return nil, errors.New("cannot cast update token to string")
				}
				var tokenFormIDString string
				tokenFormIDString, userIDString, err = getUpdateClaimsData(accessToken)
				if err != nil {
					return nil, err
				}
				if tokenFormIDString != formIDString {
					return nil, errors.New("form id in token does not match given form id")
				}
				formData = bson.M{}
			} else {
				accessToken := params.Context.Value(tokenKey).(string)
				claims, err := getTokenData(accessToken)
				if err != nil {
					return nil, err
				}
				userIDString, ok = claims["id"].(string)
				if !ok {
					return nil, errors.New("cannot cast user id to string")
				}
				formData, err = checkFormAccess(formID, accessToken, editAccessLevel, formatDate, true)
				if err != nil {
					return nil, err
				}
			}
			// updateData := bson.M{}
			if params.Args["categories"] != nil {
				categoriesInterface, ok := params.Args["categories"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting categories to interface array")
				}
				categories, err := interfaceListToMapList(categoriesInterface)
				if err != nil {
					return nil, err
				}
				for _, category := range categories {
					if err := checkArrayInputObj(category); err != nil {
						return nil, err
					}
				}
				formData["categories"] = categories
			}
			if params.Args["tags"] != nil {
				tagsInterface, ok := params.Args["tags"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting tags to interface array")
				}
				tags, err := interfaceListToMapList(tagsInterface)
				if err != nil {
					return nil, err
				}
				for _, tag := range tags {
					if err := checkArrayInputObj(tag); err != nil {
						return nil, err
					}
				}
				formData["tags"] = tags
			}
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
				if err != nil {
					return nil, err
				}
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
			var newProject string
			if params.Args["project"] != nil {
				newProject, ok = params.Args["project"].(string)
				if !ok {
					return nil, errors.New("problem casting new project to string")
				}
				formData["project"] = newProject
			}
			if params.Args["name"] != nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				formData["name"] = name
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				formData["multiple"] = multiple
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
				formData["items"] = items
			}
			if params.Args["public"] != nil {
				public, ok := params.Args["public"].(string)
				if !ok {
					return nil, errors.New("problem casting public to string")
				}
				if !findInArray(public, validAccessTypes) {
					return nil, errors.New("invalid public access level")
				}
				formData["public"] = public
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
			}
			formData["access"] = access
			return formData, nil
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
			formData["userId"] = userIDString
			return formData, nil
		},
	},
}

func updateForm(formIDString string) error {
	// get update data from redis, then update elastic and mongodb from it
	savedUpdateData, err := redisClient.Get("update-form-" + formIDString).Result()
	if err != nil {
		return err
	}
	var savedUpdateDataObj map[string]interface{}
	err = json.UnmarshalFromString(savedUpdateData, &savedUpdateDataObj)
	if err != nil {
		return err
	}
	logger.Info(savedUpdateData)
	formID, err := primitive.ObjectIDFromHex(formIDString)
	if err != nil {
		return err
	}
	var categories []string
	if savedUpdateDataObj["categories"] != nil {
		categoriesInterface, ok := savedUpdateDataObj["categories"].([]interface{})
		if !ok {
			return errors.New("problem casting categories to interface array")
		}
		categories, err = interfaceListToStringList(categoriesInterface)
		if err != nil {
			return err
		}
	}
	var tags []string
	if savedUpdateDataObj["tags"] != nil {
		tagsInterface, ok := savedUpdateDataObj["tags"].([]interface{})
		if !ok {
			return errors.New("problem casting tags to interface array")
		}
		tags, err = interfaceListToStringList(tagsInterface)
		if err != nil {
			return err
		}
	}
	var updateDataDB bson.M
	var access []map[string]interface{}
	if savedUpdateDataObj["access"] != nil {
		accessInterface, ok := savedUpdateDataObj["access"].([]interface{})
		if !ok {
			return errors.New("problem casting access to interface array")
		}
		access, err = interfaceListToMapList(accessInterface)
		if err != nil {
			return err
		}
		for _, accessUser := range access {
			if err := checkAccessObj(accessUser); err != nil {
				return err
			}
		}
		_, err = interfaceListToMapList(accessInterface)
		if err != nil {
			return err
		}
		updateDataDB, err = changeUserAccessData(formID, formType, "", categories, tags, access)
		if err != nil {
			return err
		}
	} else {
		updateDataDB = bson.M{}
	}
	if updateDataDB["$set"] == nil {
		updateDataDB["$set"] = bson.M{}
	}
	updateDataElastic := bson.M{}
	var updateProject = false
	var newProject string
	var oldProject string
	if savedUpdateDataObj["project"] != nil {
		updateProject = true
		var ok bool
		newProject, ok = savedUpdateDataObj["project"].(string)
		if !ok {
			return errors.New("problem casting new project to string")
		}
		oldProject, ok = savedUpdateDataObj["project"].(string)
		if !ok {
			return errors.New("problem casting old project to string")
		}
		updateDataDB["$set"].(bson.M)["project"] = newProject
		updateDataElastic["project"] = newProject
	}
	if savedUpdateDataObj["name"] != nil {
		name, ok := savedUpdateDataObj["name"].(string)
		if !ok {
			return errors.New("problem casting name to string")
		}
		updateDataDB["$set"].(bson.M)["name"] = name
		updateDataElastic["name"] = name
	}
	if savedUpdateDataObj["multiple"] != nil {
		multiple, ok := savedUpdateDataObj["multiple"].(bool)
		if !ok {
			return errors.New("problem casting multple to bool")
		}
		updateDataDB["$set"].(bson.M)["multiple"] = multiple
		updateDataElastic["multiple"] = multiple
	}
	if savedUpdateDataObj["items"] != nil {
		itemsInterface, ok := savedUpdateDataObj["items"].([]interface{})
		if !ok {
			return errors.New("problem casting items to interface array")
		}
		items, err := interfaceListToMapList(itemsInterface)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := checkItemObjUpdate(item); err != nil {
				return err
			}
		}
		updateDataDB["$set"].(bson.M)["items"] = items
		updateDataElastic["items"] = items
	}
	if savedUpdateDataObj["public"] != nil {
		public, ok := savedUpdateDataObj["public"].(string)
		if !ok {
			return errors.New("problem casting public to string")
		}
		if !findInArray(public, validAccessTypes) {
			return errors.New("invalid public access level")
		}
		updateDataDB["$set"].(bson.M)["public"] = public
		updateDataElastic["public"] = public
	}
	if savedUpdateDataObj["files"] != nil {
		filesinterface, ok := savedUpdateDataObj["files"].([]interface{})
		if !ok {
			return errors.New("problem casting files to interface array")
		}
		files, err := interfaceListToMapList(filesinterface)
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := checkFileObjUpdate(file); err != nil {
				return err
			}
		}
		updateDataDB["$set"].(bson.M)["files"] = files
		updateDataElastic["files"] = files
	}
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
			return err
		}
	}
	_, err = elasticClient.Update().
		Index(formElasticIndex).
		Type(formElasticType).
		Id(formIDString).
		Doc(updateDataElastic).
		Do(ctxElastic)
	if err != nil {
		return err
	}
	_, err = formCollection.UpdateOne(ctxMongo, bson.M{
		"_id": formID,
	}, updateDataDB)
	if err != nil {
		return err
	}
	if updateProject {
		err = changeFormProject(formIDString, oldProject, newProject)
		if err != nil {
			return err
		}
	}
	err = redisClient.Del("update-form-" + formIDString).Err()
	if err != nil {
		return err
	}
	return nil
}
