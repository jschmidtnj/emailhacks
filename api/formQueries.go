package main

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type formAccessClaims struct {
	FormID       string `json:"formid"`
	UserID       string `json:"userid"`
	ConnectionID string `json:"connectionid"`
	Type         string `json:"type"`
	jwt.StandardClaims
}

var formQueryFields = graphql.Fields{
	"forms": &graphql.Field{
		Type:        graphql.NewList(FormType),
		Description: "Get list of forms",
		Args: graphql.FieldConfigArgument{
			"project": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"perpage": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"page": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"searchterm": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"sort": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"ascending": &graphql.ArgumentConfig{
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
			// see this: https://github.com/olivere/elastic/issues/483
			// for potential fix to source issue (tried gave null pointer error)
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := getTokenData(accessToken)
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			var foundProject = false
			var sharedDirect = false
			var project string
			if params.Args["project"] != nil {
				project, ok = params.Args["project"].(string)
				if !ok {
					return nil, errors.New("project could not be cast to string")
				}
				if len(project) > 0 {
					foundProject = true
					projectID, err := primitive.ObjectIDFromHex(project)
					if err != nil {
						return nil, err
					}
					_, accessType, err := checkProjectAccess(projectID, accessToken, "", viewAccessLevel, false)
					if err != nil {
						return nil, err
					}
					sharedDirect = accessType == sharedAccessLevel
				}
			}
			if params.Args["perpage"] == nil {
				return nil, errors.New("no perpage argument found")
			}
			perpage, ok := params.Args["perpage"].(int)
			if !ok {
				return nil, errors.New("perpage could not be cast to int")
			}
			if params.Args["page"] == nil {
				return nil, errors.New("no page argument found")
			}
			page, ok := params.Args["page"].(int)
			if !ok {
				return nil, errors.New("page could not be cast to int")
			}
			var searchterm string
			if params.Args["searchterm"] != nil {
				searchterm, ok = params.Args["searchterm"].(string)
				if !ok {
					return nil, errors.New("searchterm could not be cast to string")
				}
			}
			if params.Args["sort"] == nil {
				return nil, errors.New("sort is undefined")
			}
			sort, ok := params.Args["sort"].(string)
			if !ok {
				return nil, errors.New("sort could not be cast to string")
			}
			if params.Args["ascending"] == nil {
				return nil, errors.New("ascending is undefined")
			}
			ascending, ok := params.Args["ascending"].(bool)
			if !ok {
				return nil, errors.New("ascending could not be cast to boolean")
			}
			if params.Args["tags"] == nil {
				return nil, errors.New("no tags argument found")
			}
			tags, ok := params.Args["tags"].([]interface{})
			if !ok {
				return nil, errors.New("tags could not be cast to string array")
			}
			if params.Args["categories"] == nil {
				return nil, errors.New("no categories argument found")
			}
			categories, ok := params.Args["categories"].([]interface{})
			if !ok {
				return nil, errors.New("categories could not be cast to string array")
			}
			fieldarray := params.Info.FieldASTs
			fieldselections := fieldarray[0].SelectionSet.Selections
			fields := make([]string, len(fieldselections))
			for i, field := range fieldselections {
				fieldast, ok := field.(*ast.Field)
				if !ok {
					return nil, errors.New("field cannot be converted to *ast.FIeld")
				}
				fields[i] = fieldast.Name.Value
			}
			var forms []map[string]interface{}
			if len(fields) > 0 {
				sourceContext := elastic.NewFetchSourceContext(true).Include(fields...)
				var numtags = len(tags)
				var showEverything = claims["type"] == superAdminType
				var startIndex = 1
				if showEverything && !foundProject {
					startIndex = 0
				}
				mustQueries := make([]elastic.Query, numtags+len(categories)+startIndex)
				if foundProject && !sharedDirect {
					mustQueries[0] = elastic.NewTermQuery("project", project)
				} else if !showEverything {
					// get all forms user has shared access to
					mustQueries[0] = elastic.NewTermsQuery(fmt.Sprintf("access.%s.type", userIDString), stringListToInterfaceList(viewAccessLevel)...)
				}
				for i, tag := range tags {
					mustQueries[i+startIndex] = elastic.NewTermQuery(fmt.Sprintf("access.%s.tags", userIDString), tag)
				}
				for i, category := range categories {
					mustQueries[i+numtags+startIndex] = elastic.NewTermQuery(fmt.Sprintf("access.%s.categories", userIDString), category)
				}
				query := elastic.NewBoolQuery().Must(mustQueries...)
				if len(searchterm) > 0 {
					mainquery := elastic.NewMultiMatchQuery(searchterm, formSearchFields...)
					query = query.Filter(mainquery)
				}
				searchResult, err := elasticClient.Search().
					Index(formElasticIndex).
					Query(query).
					Sort(sort, ascending).
					From(page * perpage).Size(perpage).
					Pretty(isDebug()).
					FetchSourceContext(sourceContext).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
				forms = make([]map[string]interface{}, len(searchResult.Hits.Hits))
				for i, hit := range searchResult.Hits.Hits {
					if hit.Source == nil {
						return nil, errors.New("no hit source found")
					}
					var formData map[string]interface{}
					err := json.Unmarshal(hit.Source, &formData)
					if err != nil {
						return nil, err
					}
					id, err := primitive.ObjectIDFromHex(hit.Id)
					if err != nil {
						return nil, err
					}
					createdTimestamp := objectidTimestamp(id)
					formData["created"] = createdTimestamp.Unix()
					formData["id"] = id.Hex()
					delete(formData, "_id")
					access, tags, categories, err := getFormattedAccessGQLData(formData, nil, userIDString)
					if err != nil {
						return nil, err
					}
					formData["access"] = access
					formData["categories"] = categories
					formData["tags"] = tags
					forms[i] = formData
				}
			}
			return forms, nil
		},
	},
	"form": &graphql.Field{
		Type:        FormType,
		Description: "Get a Form",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"editAccessToken": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			}, // true if need edit access, false if need just view access
			"accessToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
			if err != nil {
				return nil, err
			}
			var getFormUpdateToken = false
			var getFileOriginal = false
			var getFileBlur = false
			var getFilePlaceholder = false
			fieldarray := params.Info.FieldASTs
			fieldselections := fieldarray[0].SelectionSet.Selections
			for _, field := range fieldselections {
				fieldast, ok := field.(*ast.Field)
				if !ok {
					return nil, errors.New("field cannot be converted to *ast.FIeld")
				}
				if fieldast.Name.Value == "updatesAccessToken" {
					getFormUpdateToken = true
					continue
				}
				if fieldast.Name.Value == "files" {
					fileSelections := fieldast.GetSelectionSet().Selections
					for _, fileField := range fileSelections {
						fileFieldAST, ok := fileField.(*ast.Field)
						if !ok {
							return nil, errors.New("file field cannot be converted to *ast.Field")
						}
						switch fileFieldAST.Name.Value {
						case "originalSrc":
							getFileOriginal = true
							break
						case "blurSrc":
							getFileBlur = true
							break
						case "placeholderSrc":
							getFilePlaceholder = true
							break
						default:
							break
						}
						continue
					}
				}
			}
			var necessaryAccessLevel = viewAccessLevel
			var needEditAccess = false
			accessToken := params.Context.Value(tokenKey).(string)
			var userIDString = ""
			var useAccessToken = false
			if params.Args["accessToken"] != nil {
				useAccessToken = true
				if getFormUpdateToken && params.Args["editAccessToken"] != nil {
					needEditAccess, ok = params.Args["editAccessToken"].(bool)
					if !ok {
						return nil, errors.New("cannot cast editAccessToken to bool")
					}
					if needEditAccess {
						necessaryAccessLevel = editAccessLevel
					}
				}
				var tokenFormIDString string
				tokenFormIDString, _, _, userIDString, err = getResponseAddTokenData(accessToken, necessaryAccessLevel)
				if err != nil {
					return nil, err
				}
				if formIDString != tokenFormIDString {
					return nil, errors.New("token form id does not match given form id")
				}
			} else {
				claims, err := getTokenData(accessToken)
				if err == nil {
					userIDString = claims["id"].(string)
				}
				if getFormUpdateToken && len(userIDString) != 0 && params.Args["editAccessToken"] != nil {
					needEditAccess, ok = params.Args["editAccessToken"].(bool)
					if !ok {
						return nil, errors.New("cannot cast editAccessToken to bool")
					}
					if needEditAccess {
						necessaryAccessLevel = editAccessLevel
					}
				}
			}
			var formData *Form
			if !useAccessToken {
				formData, err = checkFormAccess(formID, accessToken, necessaryAccessLevel, false)
				if err != nil {
					return nil, err
				}
			} else {
				formData, err = getForm(formID, false)
				if err != nil {
					return nil, err
				}
			}
			if len(userIDString) == 0 {
				formData.Access = map[string]interface{}{}
				formData.Categories = []string{}
				formData.Tags = []string{}
			} else {
				access, tags, categories, err := getFormattedAccessGQLData(formData, nil, userIDString)
				if err != nil {
					return nil, err
				}
				formData.Access = access
				formData.Categories = categories
				formData.Tags = tags
			}
			var deleteResponses = needEditAccess && formData.Responses > 0
			if deleteResponses {
				// delete all previous responses (if there are any)
				if err = deleteAllResponses(formID); err != nil {
					return nil, err
				}
				formData.Responses = 0
			}
			if getFileOriginal || getFileBlur || getFilePlaceholder {
				for i := range formData.Files {
					filetype := formData.Files[i].Type
					fileID := formData.Files[i].ID
					if getFileOriginal {
						filepath := formFileIndex + "/" + formIDString + "/" + fileID + originalPath
						formData.Files[i].OriginalSrc, err = getSignedURL(filepath, validAccessTypes[2])
						if err != nil {
							return nil, err
						}
					}
					if filetype == "image/png" || filetype == "image/jpeg" || filetype == "image/gif" {
						var addPlaceholderPath = ""
						if filetype == "image/gif" {
							addPlaceholderPath = placeholderPath
						}
						if getFileBlur {
							filepath := formFileIndex + "/" + formIDString + "/" + fileID + addPlaceholderPath + blurPath
							formData.Files[i].BlurSrc, err = getSignedURL(filepath, validAccessTypes[2])
							if err != nil {
								return nil, err
							}
						}
						if getFilePlaceholder {
							filepath := formFileIndex + "/" + formIDString + "/" + fileID + addPlaceholderPath + originalPath
							formData.Files[i].PlaceholderSrc, err = getSignedURL(filepath, validAccessTypes[2])
							if err != nil {
								return nil, err
							}
						}
					}
				}
			}
			if getFormUpdateToken {
				// return access token with current claims + project
				expirationTime := time.Now().Add(time.Duration(tokenExpiration) * time.Hour)
				uuid, err := uuid.NewRandom()
				if err != nil {
					return nil, err
				}
				connectionIDString := uuid.String()
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, formAccessClaims{
					formIDString,
					userIDString,
					connectionIDString,
					validAccessTypes[0],
					jwt.StandardClaims{
						ExpiresAt: expirationTime.Unix(),
						Issuer:    jwtIssuer,
					},
				})
				tokenString, err := token.SignedString(jwtSecret)
				if err != nil {
					return nil, err
				}
				formData.UpdatesAccessToken = tokenString
			}
			updateDBData := bson.M{
				"$inc": bson.M{
					"views": 1,
				},
			}
			var elasticUpdateScript string
			if deleteResponses {
				elasticUpdateScript = "ctx.views+=1; ctx.responses=0;"
				updateDBData["$set"] = bson.M{
					"responses": 0,
				}
			} else {
				elasticUpdateScript = "ctx.views+=1;"
			}
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, updateDBData)
			if err != nil {
				return nil, err
			}
			script := elastic.NewScriptInline(elasticUpdateScript).Lang("painless")
			_, err = elasticClient.Update().
				Index(formElasticIndex).
				Type(formElasticType).
				Id(formIDString).
				Script(script).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			return formData, nil
		},
	},
}

func deleteAllResponses(formID primitive.ObjectID) error {
	formIDString := formID.Hex()
	sourceContext := elastic.NewFetchSourceContext(true).Include("id")
	mustQueries := []elastic.Query{
		elastic.NewTermsQuery("form", formIDString),
	}
	query := elastic.NewBoolQuery().Must(mustQueries...)
	searchResult, err := elasticClient.Search().
		Index(formElasticIndex).
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
		if err = deleteResponse(responseID, nil); err != nil {
			return err
		}
	}
	_, err = elasticClient.Update().
		Index(formElasticIndex).
		Type(formElasticType).
		Id(formIDString).
		Doc(bson.M{
			"responses": 0,
		}).
		Do(ctxElastic)
	if err != nil {
		return err
	}
	_, err = formCollection.UpdateOne(ctxMongo, bson.M{
		"_id": formID,
	}, bson.M{
		"$set": bson.M{
			"responses": 0,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
