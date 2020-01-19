package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type responseEditClaims struct {
	ResponseID  string `json:"responseid"`
	UserID      string `json:"userid"`
	Type        string `json:"type"` // access type for response (view or edit)
	FormOwnerID string `json:"formownerid"`
	jwt.StandardClaims
}

func getResponseEditTokenData(accessToken string, accessLevel []string) (string, string, string, error) {
	claims, err := getTokenData(accessToken)
	if err != nil {
		return "", "", "", err
	}
	if claims["type"] == nil {
		return "", "", "", errors.New("cannot find claims type")
	}
	claimsType, ok := claims["type"].(string)
	if !ok {
		return "", "", "", errors.New("cannot cast type to string")
	}
	if !findInArray(claimsType, accessLevel) {
		return "", "", "", errors.New("invalid access level found for response")
	}
	if claims["userid"] == nil {
		return "", "", "", errors.New("cannot find user id")
	}
	userIDString, ok := claims["userid"].(string)
	if !ok {
		return "", "", "", errors.New("cannot cast user id to string")
	}
	_, err = primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return "", "", "", err
	}
	responseIDString, ok := claims["responseid"].(string)
	if !ok {
		return "", "", "", errors.New("cannot cast response id to string")
	}
	_, err = primitive.ObjectIDFromHex(responseIDString)
	if err != nil {
		return "", "", "", err
	}
	if claims["formownerid"] == nil {
		return "", "", "", errors.New("cannot find user id")
	}
	formOwnerIDString, ok := claims["formownerid"].(string)
	if !ok {
		return "", "", "", errors.New("cannot cast form owner id to string")
	}
	_, err = primitive.ObjectIDFromHex(formOwnerIDString)
	if err != nil {
		return "", "", "", err
	}
	return responseIDString, formOwnerIDString, userIDString, nil
}

func getResponseAddTokenData(accessToken string, accessLevel []string) (string, string, string, string, error) {
	claims, err := getTokenData(accessToken)
	if err != nil {
		return "", "", "", "", err
	}
	if claims["type"] == nil {
		return "", "", "", "", errors.New("cannot find claims type")
	}
	claimsType, ok := claims["type"].(string)
	if !ok {
		return "", "", "", "", errors.New("cannot cast type to string")
	}
	if !findInArray(claimsType, accessLevel) {
		return "", "", "", "", errors.New("invalid access level found for response")
	}
	if claims["userid"] == nil {
		return "", "", "", "", errors.New("cannot find user id")
	}
	userIDString, ok := claims["userid"].(string)
	if !ok {
		return "", "", "", "", errors.New("cannot cast user id to string")
	}
	_, err = primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return "", "", "", "", err
	}
	if claims["formid"] == nil {
		return "", "", "", "", errors.New("cannot find form id")
	}
	formIDString, ok := claims["formid"].(string)
	if !ok {
		return "", "", "", "", errors.New("cannot cast user id to string")
	}
	_, err = primitive.ObjectIDFromHex(formIDString)
	if err != nil {
		return "", "", "", "", err
	}
	if claims["projectid"] == nil {
		return "", "", "", "", errors.New("cannot find project id")
	}
	projectIDString, ok := claims["projectid"].(string)
	if !ok {
		return "", "", "", "", errors.New("cannot cast project id to string")
	}
	_, err = primitive.ObjectIDFromHex(projectIDString)
	if err != nil {
		return "", "", "", "", err
	}
	if claims["owner"] == nil {
		return "", "", "", "", errors.New("cannot find owner")
	}
	ownerIDString, ok := claims["owner"].(string)
	if !ok {
		return "", "", "", "", errors.New("cannot cast project id to string")
	}
	_, err = primitive.ObjectIDFromHex(ownerIDString)
	if err != nil {
		return "", "", "", "", err
	}
	return formIDString, projectIDString, ownerIDString, userIDString, nil
}

var responseMutationFields = graphql.Fields{
	"addResponse": &graphql.Field{
		Type:        ResponseType,
		Description: "Create a Response",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"accessToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(ResponseItemInputType),
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("unable to cast form id to string")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
			if err != nil {
				return nil, errors.New("unable to create object id from string")
			}
			var useAccessToken = false
			if params.Args["accessToken"] != nil {
				useAccessToken = true
			}
			var projectID primitive.ObjectID
			var userID primitive.ObjectID
			var accessToken string
			var userIDString string
			var ownerIDString string
			if useAccessToken {
				accessToken, ok = params.Args["accessToken"].(string)
				if !ok {
					return nil, errors.New("cannot cast access token to string")
				}
				var tokenFormIDString string
				var projectIDString string
				tokenFormIDString, projectIDString, ownerIDString, userIDString, err = getResponseAddTokenData(accessToken, viewAccessLevel)
				if err != nil {
					return nil, err
				}
				projectID, err = primitive.ObjectIDFromHex(projectIDString)
				if err != nil {
					return nil, err
				}
				if formIDString != tokenFormIDString {
					return nil, errors.New("token form id does not match given form id")
				}
				userID, _ = primitive.ObjectIDFromHex(userIDString)
			} else {
				accessToken = params.Context.Value(tokenKey).(string)
				claims, err := getTokenData(accessToken)
				if err != nil {
					return nil, err
				}
				userIDString, ok = claims["id"].(string)
				if !ok {
					return nil, errors.New("cannot cast user id to string")
				}
				userID, err = primitive.ObjectIDFromHex(userIDString)
				if err != nil {
					return nil, err
				}
				formData, err := checkFormAccess(formID, accessToken, viewAccessLevel, false, false)
				if err != nil {
					return nil, err
				}
				ownerIDString, ok = formData["owner"].(string)
				if !ok {
					return nil, errors.New("cannot cast owner to string")
				}
				projectID, err = primitive.ObjectIDFromHex(formData["project"].(string))
				if err != nil {
					return nil, err
				}
			}
			ownerID, err := primitive.ObjectIDFromHex(ownerIDString)
			if err != nil {
				return nil, err
			}
			if params.Args["items"] == nil {
				return nil, errors.New("items was not provided")
			}
			itemsInterface, ok := params.Args["items"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting items to interface array")
			}
			if params.Args["files"] == nil {
				return nil, errors.New("files was not provided")
			}
			filesInterface, ok := params.Args["files"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting files to interface array")
			}
			var getResponseEditToken = false
			fieldarray := params.Info.FieldASTs
			fieldselections := fieldarray[0].SelectionSet.Selections
			for _, field := range fieldselections {
				fieldast, ok := field.(*ast.Field)
				if !ok {
					return nil, errors.New("field cannot be converted to *ast.FIeld")
				}
				if fieldast.Name.Value == "editAccessToken" {
					getResponseEditToken = true
					continue
				}
			}
			responseData, err := addResponse(itemsInterface, filesInterface, formID, projectID, ownerID, userID, formatDate)
			if err != nil {
				return nil, err
			}
			if getResponseEditToken {
				// return access token with current claims + project
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, responseEditClaims{
					responseData["id"].(string),
					userIDString,
					validAccessTypes[0],
					ownerIDString,
					jwt.StandardClaims{
						Issuer: jwtIssuer,
					},
				})
				tokenString, err := token.SignedString(jwtSecret)
				if err != nil {
					return nil, err
				}
				responseData["editAccessToken"] = tokenString
			}
			return responseData, nil
		},
	},
	"updateResponse": &graphql.Field{
		Type:        ResponseType,
		Description: "Update a Response",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"items": &graphql.ArgumentConfig{
				Type: graphql.NewList(UpdateResponseItemInputType),
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
			"accessToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if params.Args["id"] == nil {
				return nil, errors.New("response id not provided")
			}
			responseIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast response id to string")
			}
			responseID, err := primitive.ObjectIDFromHex(responseIDString)
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
			var useAccessToken = false
			if params.Args["accessToken"] != nil {
				useAccessToken = true
			}
			var accessToken string
			var responseData map[string]interface{}
			if useAccessToken {
				accessToken, ok = params.Args["accessToken"].(string)
				if !ok {
					return nil, errors.New("cannot cast access token to string")
				}
				tokenResponseIDString, _, _, err := getResponseEditTokenData(accessToken, editAccessLevel)
				if err != nil {
					return nil, err
				}
				_, err = primitive.ObjectIDFromHex(tokenResponseIDString)
				if err != nil {
					return nil, err
				}
				if tokenResponseIDString != responseIDString {
					return nil, errors.New("token response id does not match given form id")
				}
				responseData, err = getResponse(responseID, formatDate, true)
				if err != nil {
					return nil, err
				}
			} else {
				accessToken = params.Context.Value(tokenKey).(string)
				responseData, err = checkResponseAccess(responseID, accessToken, editAccessLevel, formatDate, true)
				if err != nil {
					return nil, err
				}
			}
			updateDataDB := bson.M{}
			updateDataElastic := bson.M{}
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
			formIDString, ok := responseData["form"].(string)
			if !ok {
				return nil, errors.New("unable to cast form id to string")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
			if err != nil {
				return nil, errors.New("unable to create object id from string")
			}
			if params.Args["items"] != nil {
				itemsInterface, ok := params.Args["items"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting items to interface array")
				}
				itemsUpdate, err := interfaceListToMapList(itemsInterface)
				if err != nil {
					return nil, err
				}
				if err = validateResponseItems(formID, itemsUpdate); err != nil {
					return nil, err
				}
				items, ok := responseData["items"].(primitive.A)
				if !ok {
					return nil, errors.New("cannot cast items to array")
				}
				for _, itemUpdate := range itemsUpdate {
					action := itemUpdate["updateAction"].(string)
					delete(itemUpdate, "updateAction")
					if action == validUpdateArrayActions[0] {
						// add
						items = append(items, itemUpdate)
					} else {
						index := int(itemUpdate["index"].(float64))
						delete(itemUpdate, "index")
						if index >= len(items) || index < 0 {
							continue
						}
						if action == validUpdateArrayActions[1] {
							// remove
							items = append(items[:index], items[index+1:]...)
						} else if action == validUpdateArrayActions[2] {
							// move to new index
							newIndex := int(itemUpdate["newIndex"].(float64))
							delete(itemUpdate, "newIndex")
							err = moveArray(items, index, newIndex)
							if err != nil {
								logger.Info(err.Error())
							}
						} else if action == validUpdateArrayActions[3] {
							// set index to value
							items[index] = itemUpdate
						}
					}
				}
				updateDataDB["$set"].(bson.M)["items"] = items
				updateDataElastic["items"] = items
			}
			_, err = elasticClient.Update().
				Index(responseElasticIndex).
				Type(responseElasticType).
				Id(responseIDString).
				Doc(updateDataElastic).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			_, err = responseCollection.UpdateOne(ctxMongo, bson.M{
				"_id": responseID,
			}, updateDataDB)
			if err != nil {
				return nil, err
			}
			return responseData, nil
		},
	},
	"deleteResponse": &graphql.Field{
		Type:        ResponseType,
		Description: "Delete a Response",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			if params.Args["id"] == nil {
				return nil, errors.New("response id not provided")
			}
			responseIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast response id to string")
			}
			responseID, err := primitive.ObjectIDFromHex(responseIDString)
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
			responseData, err := checkResponseAccess(responseID, accessToken, editAccessLevel, formatDate, false)
			if err != nil {
				return nil, err
			}
			if err = deleteResponse(responseID, responseData); err != nil {
				return nil, err
			}
			formIDString := responseData["form"].(string)
			formID, err := primitive.ObjectIDFromHex(formIDString)
			if err != nil {
				return nil, err
			}
			script := elastic.NewScriptInline("ctx.responses-=1").Lang("painless")
			_, err = elasticClient.Update().
				Index(formElasticIndex).
				Type(formElasticType).
				Id(formIDString).
				Script(script).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, bson.M{
				"$dec": bson.M{
					"responses": 1,
				},
			})
			if err != nil {
				return nil, err
			}
			return responseData, nil
		},
	},
}

/**
 * @api {put} /addResponse Add response
 * @apiVersion 0.0.1
 * @apiParam {String} id Form id for response
 * @apiParam {String} accessToken Token for authentication
 * @apiParam {Array} items Item objects
 * @apiParam {Array} files File objects
 * @apiParam {Boolean} formatDate Format the date with output
 * @apiSuccess {Object} data Response data
 */
func addResponseHandler(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != http.MethodPost {
		handleError("add response http method not POST", http.StatusBadRequest, response)
		return
	}
	var responsedata map[string]interface{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		handleError("error getting request body: "+err.Error(), http.StatusBadRequest, response)
		return
	}
	err = json.Unmarshal(body, &responsedata)
	if err != nil {
		handleError("error parsing request body: "+err.Error(), http.StatusBadRequest, response)
		return
	}
	if responsedata["formatDate"] == nil {
		handleError("no format date provided", http.StatusBadRequest, response)
		return
	}
	formatDate, ok := responsedata["formatDate"].(bool)
	if !ok {
		handleError("format date cannot be cast to bool", http.StatusBadRequest, response)
		return
	}
	if responsedata["id"] == nil {
		handleError("no form id provided", http.StatusBadRequest, response)
		return
	}
	formIDString, ok := responsedata["id"].(string)
	if !ok {
		handleError("cannot cast form id to string", http.StatusBadRequest, response)
		return
	}
	formID, err := primitive.ObjectIDFromHex(formIDString)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	if responsedata["accessToken"] == nil {
		handleError("no access token provided", http.StatusBadRequest, response)
		return
	}
	accessToken, ok := responsedata["accessToken"].(string)
	if !ok {
		handleError("cannot cast access token to string", http.StatusBadRequest, response)
		return
	}
	tokenFormIDString, projectIDString, ownerIDString, userIDString, err := getResponseAddTokenData(accessToken, viewAccessLevel)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	ownerID, err := primitive.ObjectIDFromHex(ownerIDString)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	projectID, err := primitive.ObjectIDFromHex(projectIDString)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	if formIDString != tokenFormIDString {
		handleError("token form id does not match given form id", http.StatusBadRequest, response)
		return
	}
	userID, _ := primitive.ObjectIDFromHex(userIDString)
	if responsedata["items"] == nil {
		handleError("items was not provided", http.StatusBadRequest, response)
		return
	}
	itemsInterface, ok := responsedata["items"].([]interface{})
	if !ok {
		handleError("problem casting items to interface array", http.StatusBadRequest, response)
		return
	}
	if responsedata["files"] == nil {
		handleError("files was not provided", http.StatusBadRequest, response)
		return
	}
	filesInterface, ok := responsedata["files"].([]interface{})
	if !ok {
		handleError("problem casting files to interface array", http.StatusBadRequest, response)
		return
	}
	responseData, err := addResponse(itemsInterface, filesInterface, formID, projectID, ownerID, userID, formatDate)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	responseDataBytes, err := json.Marshal(responseData)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	response.Write(responseDataBytes)
}

func addResponse(itemsInterface []interface{}, filesInterface []interface{}, formID primitive.ObjectID, projectID primitive.ObjectID, ownerID primitive.ObjectID, userID primitive.ObjectID, formatDate bool) (map[string]interface{}, error) {
	items, err := interfaceListToMapList(itemsInterface)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if err := checkResponseItemObjCreate(item); err != nil {
			return nil, err
		}
	}
	files, err := interfaceListToMapList(filesInterface)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if err := checkFileObjCreate(file); err != nil {
			return nil, err
		}
	}
	if err = validateResponseItems(formID, items); err != nil {
		return nil, err
	}
	now := time.Now()
	responseData := bson.M{
		"project": projectID.Hex(),
		"updated": now.Unix(),
		"user":    userID.Hex(),
		"form":    formID.Hex(),
		"items":   items,
		"files":   files,
		"views":   0,
		"owner":   ownerID.Hex(),
	}
	responseCreateRes, err := responseCollection.InsertOne(ctxMongo, responseData)
	if err != nil {
		return nil, err
	}
	responseID := responseCreateRes.InsertedID.(primitive.ObjectID)
	responseIDString := responseID.Hex()
	responseData["created"] = now.Unix()
	_, err = elasticClient.Index().
		Index(responseElasticIndex).
		Type(responseElasticType).
		Id(responseIDString).
		BodyJson(responseData).
		Do(ctxElastic)
	if err != nil {
		return nil, err
	}
	script := elastic.NewScriptInline("ctx.responses+=1").Lang("painless")
	_, err = elasticClient.Update().
		Index(formElasticIndex).
		Type(formElasticType).
		Id(formID.Hex()).
		Script(script).
		Do(ctxElastic)
	if err != nil {
		return nil, err
	}
	_, err = formCollection.UpdateOne(ctxMongo, bson.M{
		"_id": formID,
	}, bson.M{
		"$inc": bson.M{
			"responses": 1,
		},
	})
	if err != nil {
		return nil, err
	}
	if formatDate {
		responseData["created"] = now.Format(dateFormat)
		responseData["updated"] = now.Format(dateFormat)
	}
	responseData["id"] = responseIDString
	return responseData, nil
}

func deleteResponse(responseID primitive.ObjectID, responseData bson.M) error {
	if !justDeleteElastic {
		var err error
		if responseData == nil {
			responseData, err = getForm(responseID, false, false)
			if err != nil {
				return err
			}
		}
	}
	responseIDString := responseID.Hex()
	_, err := elasticClient.Delete().
		Index(responseElasticIndex).
		Type(responseElasticType).
		Id(responseIDString).
		Do(ctxElastic)
	if err != nil {
		return err
	}
	if !justDeleteElastic {
		_, err = responseCollection.DeleteOne(ctxMongo, bson.M{
			"_id": responseID,
		})
		if err != nil {
			return err
		}
		primitivefiles, ok := responseData["files"].(primitive.A)
		if !ok {
			return errors.New("cannot convert files to primitive")
		}
		for _, primitivefile := range primitivefiles {
			filedatadoc, ok := primitivefile.(primitive.D)
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
			fileobj := storageBucket.Object(responseFileIndex + "/" + responseIDString + "/" + fileid + originalPath)
			if err := fileobj.Delete(ctxStorage); err != nil {
				return err
			}
			if filetype == "image/gif" {
				fileobj = storageBucket.Object(responseFileIndex + "/" + responseIDString + "/" + fileid + placeholderPath + originalPath)
				blurobj := storageBucket.Object(responseFileIndex + "/" + responseIDString + "/" + fileid + placeholderPath + blurPath)
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
					fileobj = storageBucket.Object(responseFileIndex + "/" + responseIDString + "/" + fileid + blurPath)
					if err := fileobj.Delete(ctxStorage); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func validateResponseItems(formID primitive.ObjectID, responseItems []map[string]interface{}) error {
	formData, err := getForm(formID, false, false)
	if err != nil {
		return err
	}
	formItems := formData["items"].(primitive.A)
	responseItemIndexes := map[int]bool{}
	for _, responseItem := range responseItems {
		formIndex, _ := responseItem["formIndex"].(int)
		if formIndex >= len(formItems) || formIndex < 0 {
			return errors.New("response index outside length of form")
		}
		if _, ok := responseItemIndexes[formIndex]; ok {
			return errors.New("cannot have duplicate form index")
		}
		formItemObj := formItems[formIndex].(primitive.M)
		questionType := formItemObj["type"].(string)
		if !findInArray(questionType, validResponseItemTypes) {
			return errors.New("invalid type for response item found")
		}
		questionRequired := formItemObj["required"].(bool)
		if findInArray(questionType, itemTypesRequireOptions) {
			selectedOptions, err := interfaceListToStringList(responseItem["options"].([]interface{}))
			if err != nil {
				return errors.New("problem casting selected options to int array")
			}
			if !findInArray(questionType, itemTypesAllowMultipleOptions) && len(selectedOptions) > 0 {
				return errors.New("cannot select multiple options")
			}
			questionOptions, err := interfaceListToStringList(formItemObj["options"].([]interface{}))
			if err != nil {
				return errors.New("problem casting question options to string array")
			}
			var foundOption = false
			for _, option := range selectedOptions {
				if !findInArray(option, questionOptions) {
					return errors.New("cannot find given option in question options")
				}
				foundOption = true
			}
			if questionRequired && !foundOption {
				return errors.New("cannot find a valid selected option")
			}
		} else if findInArray(questionType, itemTypesText) {
			// text input
			if questionRequired && len(responseItem["text"].(string)) == 0 {
				return errors.New("cannot find any text for response item")
			}
		} else if findInArray(questionType, itemTypesFile) {
			// file input
			if questionRequired && len(responseItem["files"].([]interface{})) == 0 {
				return errors.New("cannot find any files for response item")
			}
		}
		responseItemIndexes[formIndex] = true
	}
	for i, formItem := range formItems {
		if formItem.(primitive.M)["required"].(bool) {
			if _, ok := responseItemIndexes[i]; !ok {
				return errors.New("required item not found")
			}
		}
	}
	return nil
}
