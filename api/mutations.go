package main

import (
	"errors"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func interfaceListToStringList(interfaceList []interface{}) ([]string, error) {
	result := make([]string, len(interfaceList))
	for i, item := range interfaceList {
		itemStr, ok := item.(string)
		if !ok {
			return nil, errors.New("item in list cannot be cast to string")
		}
		result[i] = itemStr
	}
	return result, nil
}

func interfaceListToMapList(interfaceList []interface{}) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(interfaceList))
	for i, item := range interfaceList {
		itemObj, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("item in list cannot be converted to map")
		}
		result[i] = itemObj
	}
	return result, nil
}

func checkAccessObj(accessObj map[string]interface{}) error {
	if accessObj["id"] == nil {
		return errors.New("no id field given")
	}
	if accessObj["type"] == nil {
		return errors.New("no type field given")
	}
	return nil
}

func checkItemObjCreate(itemObj map[string]interface{}) error {
	if itemObj["name"] == nil {
		return errors.New("no name field given")
	}
	if itemObj["type"] == nil {
		return errors.New("no type field given")
	}
	if itemObj["options"] == nil {
		return errors.New("no options field given")
	}
	if itemObj["required"] == nil {
		return errors.New("no required field given")
	}
	return checkItemObjUpdate(itemObj)
}

func checkItemObjUpdate(itemObj map[string]interface{}) error {
	if itemObj["name"] != nil {
		if _, ok := itemObj["name"].(string); !ok {
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
	if itemObj["required"] != nil {
		if _, ok := itemObj["required"].(bool); !ok {
			return errors.New("problem casting required to boolean")
		}
	}
	return nil
}

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

func deleteAccount(idstring string) (interface{}, error) {
	id, err := primitive.ObjectIDFromHex(idstring)
	if err != nil {
		return nil, err
	}
	cursor, err := userCollection.Find(ctxMongo, bson.M{
		"_id": id,
	})
	defer cursor.Close(ctxMongo)
	if err != nil {
		return nil, err
	}
	var userData map[string]interface{}
	var foundstuff = false
	for cursor.Next(ctxMongo) {
		userPrimitive := &bson.D{}
		err = cursor.Decode(userPrimitive)
		if err != nil {
			return nil, err
		}
		userData = userPrimitive.Map()
		id := userData["_id"].(primitive.ObjectID)
		userData["date"] = objectidtimestamp(id).Format(dateFormat)
		userData["id"] = id.Hex()
		delete(userData, "_id")
		foundstuff = true
		break
	}
	if !foundstuff {
		return nil, errors.New("user not found with given id")
	}
	_, err = userCollection.DeleteOne(ctxMongo, bson.M{
		"_id": id,
	})
	if err != nil {
		return nil, err
	}
	return userData, nil
}

func rootMutation() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addForm": &graphql.Field{
				Type:        FormType,
				Description: "Create a Form",
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"description": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"items": &graphql.ArgumentConfig{
						Type: graphql.NewList(ItemInputType),
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
					if params.Args["name"] == nil {
						return nil, errors.New("name not provided")
					}
					if params.Args["description"] == nil {
						return nil, errors.New("description not provided")
					}
					if params.Args["items"] == nil {
						return nil, errors.New("items not provided")
					}
					name, ok := params.Args["name"].(string)
					if !ok {
						return nil, errors.New("problem casting name to string")
					}
					description, ok := params.Args["description"].(string)
					if !ok {
						return nil, errors.New("problem casting description to string")
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
					formData := bson.M{
						"name":        name,
						"description": description,
						"items":       items,
						"access": bson.A{
							bson.M{
								"id":     userIDString,
								"access": "owner",
							},
						},
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
			"updateForm": &graphql.Field{
				Type:        FormType,
				Description: "Update a Form",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"description": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"items": &graphql.ArgumentConfig{
						Type: graphql.NewList(ItemInputType),
					},
					"access": &graphql.ArgumentConfig{
						Type: graphql.NewList(AccessInputType),
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
					formDataCursor, err := formCollection.Find(ctxMongo, bson.M{
						"_id": formID,
					})
					defer formDataCursor.Close(ctxMongo)
					if err != nil {
						return nil, err
					}
					var allFormData map[string]interface{}
					var allAccess []map[string]interface{}
					var foundForm = false
					for formDataCursor.Next(ctxMongo) {
						foundForm = true
						formPrimitive := &bson.D{}
						err = formDataCursor.Decode(formPrimitive)
						if err != nil {
							return nil, err
						}
						allFormData = formPrimitive.Map()
						id := allFormData["_id"].(primitive.ObjectID)
						allFormData["date"] = objectidtimestamp(id).Format(dateFormat)
						allFormData["id"] = id.Hex()
						delete(allFormData, "_id")
						usersInterface, ok := allFormData["access"].([]interface{})
						if !ok {
							return nil, errors.New("problem casting user access to interface array")
						}
						users, err := interfaceListToMapList(usersInterface)
						if err != nil {
							return nil, err
						}
						allAccess = users
						allFormData["access"] = allAccess
						var foundUser = false
						for _, user := range users {
							accessUserIDString, ok := user["id"].(string)
							if !ok {
								return nil, errors.New("cannot cast user id in form to string")
							}
							if accessUserIDString == userIDString {
								foundUser = true
								accessType, ok := user["type"].(string)
								if !ok {
									return nil, errors.New("cannot cast access type to string")
								}
								if !(accessType == "owner" || accessType == "edit") {
									return nil, errors.New("user not authorized to edit form")
								}
								break
							}
						}
						if !foundUser {
							return nil, errors.New("user not authorized to edit form")
						}
						break
					}
					if !foundForm {
						return nil, errors.New("form not found with given id")
					}
					updateData := bson.M{}
					updateSet := bson.M{}
					updateAdd := bson.M{}
					if params.Args["name"] != nil {
						name, ok := params.Args["name"].(string)
						if !ok {
							return nil, errors.New("problem casting name to string")
						}
						updateSet["name"] = name
						allFormData["name"] = name
					}
					if params.Args["description"] != nil {
						description, ok := params.Args["description"].(string)
						if !ok {
							return nil, errors.New("problem casting description to string")
						}
						updateSet["description"] = description
						allFormData["description"] = description
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
						updateSet["items"] = items
						allFormData["items"] = items
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
						updateAdd["access"] = access
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
						for i, currentAccess := range allAccess {
							currentAccessID, ok := currentAccess["id"].(string)
							if !ok {
								return nil, errors.New("cannot convert current access user id to string")
							}
							if currentAccessID == newAccessUserIDString {
								foundUser = true
								allAccess[i] = accessUser
								break
							}
						}
						if !foundUser {
							allAccess = append(allAccess, accessUser)
						}
					}
					allFormData["access"] = allAccess
					updateData["$set"] = updateSet
					updateData["$addToSet"] = updateAdd
					_, err = formCollection.UpdateOne(ctxMongo, bson.M{
						"_id": formID,
					}, updateData)
					if err != nil {
						return nil, err
					}
					return allFormData, nil
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
					claims, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
					if err != nil {
						return nil, err
					}
					userIDString, ok := claims["id"].(string)
					if !ok {
						return nil, errors.New("cannot cast user id to string")
					}
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
					formDataCursor, err := formCollection.Find(ctxMongo, bson.M{
						"_id": formID,
					})
					defer formDataCursor.Close(ctxMongo)
					if err != nil {
						return nil, err
					}
					var formData map[string]interface{}
					var access []map[string]interface{}
					var foundForm = false
					for formDataCursor.Next(ctxMongo) {
						foundForm = true
						formPrimitive := &bson.D{}
						err = formDataCursor.Decode(formPrimitive)
						if err != nil {
							return nil, err
						}
						formData = formPrimitive.Map()
						id := formData["_id"].(primitive.ObjectID)
						formData["date"] = objectidtimestamp(id).Format(dateFormat)
						formData["id"] = id.Hex()
						delete(formData, "_id")
						usersInterface, ok := formData["access"].([]interface{})
						if !ok {
							return nil, errors.New("problem casting user access to interface array")
						}
						users, err := interfaceListToMapList(usersInterface)
						if err != nil {
							return nil, err
						}
						access = users
						formData["access"] = access
						var foundUser = false
						for _, user := range users {
							accessUserIDString, ok := user["id"].(string)
							if !ok {
								return nil, errors.New("cannot cast user id in form to string")
							}
							if accessUserIDString == userIDString {
								foundUser = true
								accessType, ok := user["type"].(string)
								if !ok {
									return nil, errors.New("cannot cast access type to string")
								}
								if !(accessType == "owner" || accessType == "edit") {
									return nil, errors.New("user not authorized to edit form")
								}
								break
							}
						}
						if !foundUser {
							return nil, errors.New("user not authorized to edit form")
						}
						break
					}
					if !foundForm {
						return nil, errors.New("form not found with given id")
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
					return formData, nil
				},
			},
			"deleteUser": &graphql.Field{
				Type:        AccountType,
				Description: "Delete a User",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					_, err := validateAdmin(params.Context.Value(tokenKey).(string))
					if err != nil {
						return nil, err
					}
					if params.Args["id"] == nil {
						return nil, errors.New("user id not provided")
					}
					idstring, ok := params.Args["id"].(string)
					if !ok {
						return nil, errors.New("cannot cast id to string")
					}
					return deleteAccount(idstring)
				},
			},
			"deleteAccount": &graphql.Field{
				Type:        AccountType,
				Description: "Delete a User",
				Args:        graphql.FieldConfigArgument{},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					claims, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
					if err != nil {
						return nil, err
					}
					idstring, ok := claims["id"].(string)
					if !ok {
						return nil, errors.New("cannot cast id to string")
					}
					return deleteAccount(idstring)
				},
			},
		},
	})
}
