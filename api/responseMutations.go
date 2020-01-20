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
	"github.com/mitchellh/mapstructure"
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
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
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
			logger.Info("add the response")
			var projectID primitive.ObjectID
			var userID primitive.ObjectID
			var accessToken string
			var userIDString string
			var ownerIDString string
			if useAccessToken {
				logger.Info("use access token")
				accessToken, ok = params.Args["accessToken"].(string)
				if !ok {
					return nil, errors.New("cannot cast access token to string")
				}
				logger.Info("access token: " + accessToken)
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
				formData, err := checkFormAccess(formID, accessToken, viewAccessLevel, false)
				if err != nil {
					return nil, err
				}
				ownerIDString = formData.Owner
				projectID, err = primitive.ObjectIDFromHex(formData.Project)
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
			responseData, err := addResponse(itemsInterface, filesInterface, formID, projectID, ownerID, userID)
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
				Type: graphql.NewList(UpdateFileInputType),
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
			var useAccessToken = false
			if params.Args["accessToken"] != nil {
				useAccessToken = true
			}
			var accessToken string
			var userIDString string
			var responseData *Response
			if useAccessToken {
				accessToken, ok = params.Args["accessToken"].(string)
				if !ok {
					return nil, errors.New("cannot cast access token to string")
				}
				var tokenResponseIDString string
				tokenResponseIDString, _, userIDString, err = getResponseEditTokenData(accessToken, editAccessLevel)
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
				responseData, err = getResponse(responseID, true)
				if err != nil {
					return nil, err
				}
			} else {
				accessToken = params.Context.Value(tokenKey).(string)
				responseData, err = checkResponseAccess(responseID, accessToken, editAccessLevel, true)
				if err != nil {
					return nil, err
				}
				claims, err := getTokenData(accessToken)
				if err != nil {
					return nil, err
				}
				userIDString = claims["id"].(string)
			}
			userID, err := primitive.ObjectIDFromHex(userIDString)
			if err != nil {
				return nil, err
			}
			updateDataDB := bson.M{
				"$set": bson.M{},
			}
			updateDataElastic := bson.M{}
			if params.Args["files"] != nil {
				filesUpdateInterface, ok := params.Args["files"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting files to interface array")
				}
				filesUpdate, err := interfaceListToMapList(filesUpdateInterface)
				if err != nil {
					return nil, err
				}
				for _, file := range filesUpdate {
					if err := checkFileObjUpdate(file); err != nil {
						return nil, err
					}
				}
				for _, fileUpdate := range filesUpdate {
					index := fileUpdate["index"].(int)
					delete(fileUpdate, "index")
					delete(fileUpdate, "fileIndex")
					delete(fileUpdate, "itemIndex")
					action := fileUpdate["updateAction"].(string)
					delete(fileUpdate, "updateAction")
					var fileObj *File
					if err = mapstructure.Decode(fileUpdate, &fileObj); err != nil {
						return nil, err
					}
					if action == validUpdateMapActions[0] {
						// add
						responseData.Files = append(responseData.Files, fileObj)
					} else {
						if action == validUpdateMapActions[1] {
							// remove
							if index >= 0 && index < len(responseData.Files) {
								responseData.Files = append(responseData.Files[:index], responseData.Files[index+1:]...)
							}
						} else if action == validUpdateMapActions[2] {
							// set to value
							responseData.Files[index] = fileObj
						}
					}
				}
				updateDataDB["$set"].(bson.M)["files"] = responseData.Files
				updateDataElastic["files"] = responseData.Files
			}
			formID, err := primitive.ObjectIDFromHex(responseData.Form)
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
				if err = validateResponseItems(formID, userID, true, &itemsUpdate); err != nil {
					return nil, err
				}
				for _, itemUpdate := range itemsUpdate {
					action := itemUpdate["updateAction"].(string)
					delete(itemUpdate, "updateAction")
					var itemObj *ResponseItem
					if err = mapstructure.Decode(itemUpdate, &itemObj); err != nil {
						return nil, err
					}
					if action == validUpdateArrayActions[0] {
						// add
						delete(itemUpdate, "index")
						responseData.Items = append(responseData.Items, itemObj)
					} else {
						index := itemUpdate["index"].(int)
						delete(itemUpdate, "index")
						if index >= len(responseData.Items) || index < 0 {
							continue
						}
						if action == validUpdateArrayActions[1] {
							// remove
							responseData.Items = append(responseData.Items[:index], responseData.Items[index+1:]...)
						} else if action == validUpdateArrayActions[2] {
							// move to new index
							newIndex := itemUpdate["newIndex"].(int)
							delete(itemUpdate, "newIndex")
							err = moveArray(responseData.Items, index, newIndex)
							if err != nil {
								logger.Info(err.Error())
							}
						} else if action == validUpdateArrayActions[3] {
							// set index to value
							responseData.Items[index] = itemObj
						}
					}
				}
				updateDataDB["$set"].(bson.M)["items"] = responseData.Items
				updateDataElastic["items"] = responseData.Items
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
			var response *Response
			if !justDeleteElastic {
				response, err = checkResponseAccess(responseID, accessToken, editAccessLevel, false)
				if err != nil {
					return nil, err
				}
			}
			bytesRemoved, err := deleteResponse(responseID, response)
			if err != nil {
				return nil, err
			}
			ownerID, err := primitive.ObjectIDFromHex(response.Owner)
			if err != nil {
				return nil, err
			}
			if err = changeUserStorage(ownerID, -1*bytesRemoved); err != nil {
				return nil, err
			}
			if !justDeleteElastic {
				formID, err := primitive.ObjectIDFromHex(response.ID)
				if err != nil {
					return nil, err
				}
				script := elastic.NewScriptInline("ctx._source.responses-=1")
				_, err = elasticClient.Update().
					Index(formElasticIndex).
					Type(formElasticType).
					Id(response.Form).
					Script(script).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
				_, err = formCollection.UpdateOne(ctxMongo, bson.M{
					"_id": formID,
				}, bson.M{
					"$inc": bson.M{
						"responses": -1,
					},
				})
				if err != nil {
					return nil, err
				}
			}
			return response, nil
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
	responseData, err := addResponse(itemsInterface, filesInterface, formID, projectID, ownerID, userID)
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

func addResponse(itemsInterface []interface{}, filesInterface []interface{}, formID primitive.ObjectID, projectID primitive.ObjectID, ownerID primitive.ObjectID, userID primitive.ObjectID) (map[string]interface{}, error) {
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
	if err = validateResponseItems(formID, userID, false, &items); err != nil {
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
	script := elastic.NewScriptInline("ctx._source.responses+=1")
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
	responseData["id"] = responseIDString
	return responseData, nil
}

func deleteResponse(responseID primitive.ObjectID, response *Response) (int64, error) {
	if !justDeleteElastic {
		var err error
		if response == nil {
			response, err = getResponse(responseID, false)
			if err != nil {
				return 0, err
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
		return 0, err
	}
	var bytesRemoved int64 = 0
	if !justDeleteElastic {
		_, err = responseCollection.DeleteOne(ctxMongo, bson.M{
			"_id": responseID,
		})
		if err != nil {
			return 0, err
		}
		for _, file := range response.Files {
			newBytesRemoved, err := deleteFile(responseType, response.ID, file.ID)
			if err != nil {
				return 0, err
			}
			bytesRemoved += newBytesRemoved
		}
	}
	return bytesRemoved, nil
}

func validateResponseItems(formID primitive.ObjectID, userID primitive.ObjectID, updating bool, responseItems *[]map[string]interface{}) error {
	formData, err := getForm(formID, false)
	if err != nil {
		return err
	}
	if !formData.Multiple && !updating {
		mustQueries := make([]elastic.Query, 2)
		mustQueries[0] = elastic.NewTermsQuery("form", formID.Hex())
		mustQueries[1] = elastic.NewTermQuery("user", userID.Hex())
		query := elastic.NewBoolQuery()
		if len(mustQueries) > 0 {
			query = query.Must(mustQueries...)
		}
		count, err := elasticClient.Count().
			Type(responseElasticType).
			Query(query).
			Pretty(false).
			Do(ctxElastic)
		if err != nil {
			return err
		}
		if count != 0 {
			return errors.New("cannot submit multiple responses")
		}
	}
	formItems := formData.Items
	responseItemIndexes := map[int]bool{}
	for i, responseItem := range *responseItems {
		formIndex, _ := responseItem["formIndex"].(int)
		if formIndex >= len(formItems) || formIndex < 0 {
			return errors.New("response index outside length of form")
		}
		if _, ok := responseItemIndexes[formIndex]; ok {
			return errors.New("cannot have duplicate form index")
		}
		formItemObj := formItems[formIndex]
		questionType := formItemObj.Type
		if !findInArray(questionType, validResponseItemTypes) {
			return errors.New("invalid type for response item found")
		}
		questionRequired := formItemObj.Required
		if findInArray(questionType, itemTypesRequireOptions) {
			selectedOptions, err := interfaceListToStringList(responseItem["options"].([]interface{}))
			if err != nil {
				return errors.New("problem casting selected options to int array")
			}
			if !findInArray(questionType, itemTypesAllowMultipleOptions) && len(selectedOptions) > 1 {
				return errors.New("cannot select multiple options")
			}
			questionOptions := formItemObj.Options
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
		} else {
			(*responseItems)[i]["options"] = bson.A{}
		}
		if findInArray(questionType, itemTypesText) {
			// text input
			if questionRequired && len(responseItem["text"].(string)) == 0 {
				return errors.New("cannot find any text for response item")
			}
		} else {
			(*responseItems)[i]["text"] = ""
		}
		if findInArray(questionType, itemTypesFile) {
			// file input
			if questionRequired && len(responseItem["files"].([]interface{})) == 0 {
				return errors.New("cannot find any files for response item")
			}
		} else {
			(*responseItems)[i]["files"] = bson.A{}
		}
		responseItemIndexes[formIndex] = true
	}
	for i, formItem := range formItems {
		if formItem.Required {
			if _, ok := responseItemIndexes[i]; !ok {
				return errors.New("required item not found")
			}
		}
	}
	return nil
}
