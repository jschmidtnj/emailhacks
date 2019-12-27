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

func checkItemObjCreate(itemObj map[string]interface{}) error {
	if itemObj["question"] == nil {
		return errors.New("no name field given")
	}
	if itemObj["type"] == nil {
		return errors.New("no type field given")
	}
	if itemObj["options"] == nil {
		return errors.New("no options field given")
	}
	if itemObj["text"] == nil {
		return errors.New("no text field given")
	}
	if itemObj["required"] == nil {
		return errors.New("no required field given")
	}
	if itemObj["file"] == nil {
		return errors.New("no file field given")
	}
	return checkItemObjUpdate(itemObj)
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
		textArray, ok := itemObj["text"].([]interface{})
		if !ok {
			return errors.New("problem casting text to interface array")
		}
		if _, err := interfaceListToMapList(textArray); err != nil {
			return errors.New("problem casting text to map array")
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

var formMutationFields = graphql.Fields{
	"addForm": &graphql.Field{
		Type:        FormType,
		Description: "Create a Form",
		Args: graphql.FieldConfigArgument{
			"subject": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"recipient": &graphql.ArgumentConfig{
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
			claims, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			userID, err := primitive.ObjectIDFromHex(userIDString)
			if err != nil {
				return nil, err
			}
			if params.Args["subject"] == nil {
				return nil, errors.New("subject not provided")
			}
			if params.Args["recipient"] == nil {
				return nil, errors.New("recipient not provided")
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
			subject, ok := params.Args["subject"].(string)
			if !ok {
				return nil, errors.New("problem casting subject to string")
			}
			recipient, ok := params.Args["recipient"].(string)
			if !ok {
				return nil, errors.New("problem casting recipient to string")
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
			formData := bson.M{
				"subject":    subject,
				"recipient":  recipient,
				"items":      items,
				"multiple":   multiple,
				"tags":       tags,
				"categories": categories,
				"views":      0,
				"access": bson.A{
					bson.M{
						"id":     userIDString,
						"access": editAccessLevel[0],
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
			if err = addUserFormAccess(userID, "owner", formIDString); err != nil {
				return nil, err
			}
			timestamp := objectidtimestamp(formID)
			formData["date"] = timestamp.Format(dateFormat)
			formData["id"] = formIDString
			return formData, nil
		},
	},
	"updateFormOrganization": &graphql.Field{
		Type:        FormType,
		Description: "Update Form Organization",
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
			formData, _, err := checkFormAccess(formID, accessToken, editAccessLevel)
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
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, updateData)
			if err != nil {
				return nil, err
			}
			timestamp := objectidtimestamp(formID)
			formData["date"] = timestamp.Format(dateFormat)
			formData["id"] = formIDString
			return formData, nil
		},
	},
	"updateFormAccess": &graphql.Field{
		Type:        FormType,
		Description: "Update Form Access",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
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
			formData, _, err := checkFormAccess(formID, accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			if params.Args["value"] == nil {
				return nil, errors.New("value not provided")
			}
			if params.Args["add"] == nil {
				return nil, errors.New("add not provided")
			}
			accessType, ok := params.Args["value"].(string)
			if !ok {
				return nil, errors.New("problem casting access type to string")
			}
			if !findInArray(accessType, validAccessTypes) {
				return nil, errors.New("invalid access type given")
			}
			add, ok := params.Args["value"].(bool)
			if !ok {
				return nil, errors.New("problem casting add to boolean")
			}
			updateData := bson.M{}
			valueUpdate := bson.M{
				"access": accessType,
			}
			if add {
				updateData["$set"] = valueUpdate
				updateData["$upsert"] = true
			} else {
				updateData["$pull"] = valueUpdate
			}
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, updateData)
			if err != nil {
				return nil, err
			}
			timestamp := objectidtimestamp(formID)
			formData["date"] = timestamp.Format(dateFormat)
			formData["id"] = formIDString
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
			"subject": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"recipient": &graphql.ArgumentConfig{
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
			formData, _, err := checkFormAccess(formID, accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			updateData := bson.M{}
			if params.Args["subject"] != nil {
				subject, ok := params.Args["subject"].(string)
				if !ok {
					return nil, errors.New("problem casting subject to string")
				}
				updateData["subject"] = subject
				formData["subject"] = subject
			}
			if params.Args["recipient"] != nil {
				recipient, ok := params.Args["recipient"].(string)
				if !ok {
					return nil, errors.New("problem casting description to string")
				}
				updateData["recipient"] = recipient
				formData["recipient"] = recipient
			}
			if params.Args["multiple"] != nil {
				multiple, ok := params.Args["multiple"].(bool)
				if !ok {
					return nil, errors.New("problem casting multple to bool")
				}
				updateData["multiple"] = multiple
				formData["multiple"] = multiple
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
				updateData["items"] = items
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
				updateData["public"] = public
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
				updateData["files"] = files
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
				updateData["access"] = access
			}
			for _, accessUser := range access {
				newAccessUserIDString, ok := accessUser["id"].(string)
				if !ok {
					return nil, errors.New("cannot convert access user id to string")
				}
				newAccessUserID, err := primitive.ObjectIDFromHex(newAccessUserIDString)
				if err != nil {
					return nil, err
				}
				newAccessType, ok := accessUser["type"].(string)
				if !ok {
					return nil, errors.New("cannot convert access user id to string")
				}
				if err = addUserFormAccess(newAccessUserID, newAccessType, formIDString); err != nil {
					return nil, err
				}
				var foundUser = false
				for i, currentAccess := range access {
					currentAccessID, ok := currentAccess["id"].(string)
					if !ok {
						return nil, errors.New("cannot convert current access user id to string")
					}
					if currentAccessID == newAccessUserIDString {
						foundUser = true
						access[i] = accessUser
						break
					}
				}
				if !foundUser {
					access = append(access, accessUser)
				}
			}
			formData["access"] = access
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, bson.M{
				"$set": updateData,
			})
			if err != nil {
				return nil, err
			}
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
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			if params.Args["id"] != nil {
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
			formData, access, err := checkFormAccess(formID, accessToken, editAccessLevel)
			if err != nil {
				return nil, err
			}
			formDataCursor, err := formCollection.Find(ctxMongo, bson.M{
				"_id": formID,
			})
			defer formDataCursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			for _, user := range access {
				accessUserIDString, ok := user["id"].(string)
				if !ok {
					return nil, errors.New("cannot cast user id in form to string")
				}
				accessUserID, err := primitive.ObjectIDFromHex(accessUserIDString)
				if err != nil {
					return nil, err
				}
				_, err = userCollection.UpdateOne(ctxMongo, bson.M{
					"_id": accessUserID,
				}, bson.M{
					"$pull": bson.M{
						"forms": bson.M{
							"id": formIDString,
						},
					},
				})
				if err != nil {
					return nil, err
				}
			}
			_, err = formCollection.DeleteOne(ctxMongo, bson.M{
				"_id": formID,
			})
			if err != nil {
				return nil, err
			}
			primativefiles, ok := formData["files"].(primitive.A)
			if !ok {
				return nil, errors.New("cannot convert files to primitive")
			}
			for _, primativefile := range primativefiles {
				filedatadoc, ok := primativefile.(primitive.D)
				if !ok {
					return nil, errors.New("cannot convert file to primitive doc")
				}
				filedata := filedatadoc.Map()
				fileid, ok := filedata["id"].(string)
				if !ok {
					return nil, errors.New("cannot convert file id to string")
				}
				filetype, ok := filedata["type"].(string)
				if !ok {
					return nil, errors.New("cannot convert file type to string")
				}
				fileobj := storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + originalPath)
				if err := fileobj.Delete(ctxStorage); err != nil {
					return nil, err
				}
				if filetype == "image/gif" {
					fileobj = storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + placeholderPath + originalPath)
					blurobj := storageBucket.Object(formFileIndex + "/" + formIDString + "/" + fileid + placeholderPath + blurPath)
					if err := fileobj.Delete(ctxStorage); err != nil {
						return nil, err
					}
					if err := blurobj.Delete(ctxStorage); err != nil {
						return nil, err
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
							return nil, err
						}
					}
				}
			}
			if err != nil {
				return nil, err
			}
			return formData, nil
		},
	},
}
