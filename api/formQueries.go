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
	"github.com/mitchellh/mapstructure"
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
	"formEmail": &graphql.Field{
		Type:        FormEmailType,
		Description: "Get form email data",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"accessKey": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sharable link key",
			},
			"email": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "email to send to",
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := getTokenData(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("id could not be cast to int")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
			if err != nil {
				return nil, err
			}
			if params.Args["email"] == nil {
				return nil, errors.New("no email argument provided")
			}
			email, ok := params.Args["email"].(string)
			if !ok {
				return nil, errors.New("cannot cast email to string")
			}
			var accessKey = ""
			if params.Args["accessKey"] != nil {
				accessKey, ok = params.Args["accessKey"].(string)
				if !ok {
					return nil, errors.New("cannot cast access key to string")
				}
			}
			form, err := checkFormAccess(formID, accessToken, accessKey, viewAccessLevel, false)
			if err != nil {
				return nil, err
			}
			if err = getFileURLs(form, true, true, true); err != nil {
				return nil, err
			}
			formEmail, err := getFormEmailData(&SendEmailData{
				Form:  form,
				Email: email,
			})
			if err != nil {
				return nil, err
			}
			formEmailData := map[string]string{
				"id":   formIDString,
				"data": formEmail,
			}
			return formEmailData, nil
		},
	},
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
			"accessKey": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sharable link key for project",
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
					var accessKey = ""
					if params.Args["accessKey"] != nil {
						accessKey, ok = params.Args["accessKey"].(string)
						if !ok {
							return nil, errors.New("cannot cast access key to string")
						}
					}
					_, accessType, err := checkProjectAccess(projectID, accessToken, accessKey, viewAccessLevel, false)
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
			var forms []*Form
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
				forms = make([]*Form, len(searchResult.Hits.Hits))
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
					var currentForm Form
					if err = mapstructure.Decode(formData, &currentForm); err != nil {
						return nil, err
					}
					currentForm.ID = id.Hex()
					currentForm.Created = objectidTimestamp(id).Unix()
					access, tags, categories, currentAccessType, err := getFormattedAccessGQLData(currentForm.Access, nil, userIDString)
					if err != nil {
						return nil, err
					}
					needEditAccessLevelForLink := findInArray(currentForm.LinkAccess.Type, editAccessLevel)
					necessaryAccessLevel := viewAccessLevel
					if needEditAccessLevelForLink {
						necessaryAccessLevel = editAccessLevel
					}
					if !findInArray(currentAccessType, necessaryAccessLevel) {
						currentForm.LinkAccess = nil
						currentForm.Access = nil
					} else {
						currentForm.Access = access
					}
					currentForm.Categories = categories
					currentForm.Tags = tags
					forms[i] = &currentForm
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
			"accessKey": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sharable link key",
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
			var accessKey = ""
			if params.Args["accessKey"] != nil {
				accessKey, ok = params.Args["accessKey"].(string)
				if !ok {
					return nil, errors.New("cannot cast access key to string")
				}
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
			var form *Form
			if !useAccessToken {
				form, err = checkFormAccess(formID, accessToken, accessKey, necessaryAccessLevel, false)
				if err != nil {
					return nil, err
				}
			} else {
				form, err = getForm(formID, false)
				if err != nil {
					return nil, err
				}
			}
			if len(userIDString) == 0 {
				form.Access = map[string]interface{}{}
				form.Categories = []string{}
				form.Tags = []string{}
			} else {
				access, tags, categories, currentAccessType, err := getFormattedAccessGQLData(form.Access, nil, userIDString)
				if err != nil {
					return nil, err
				}
				needEditAccessLevelForLink := findInArray(form.LinkAccess.Type, editAccessLevel)
				necessaryAccessLevelForLink := viewAccessLevel
				if needEditAccessLevelForLink {
					necessaryAccessLevelForLink = editAccessLevel
				}
				if !findInArray(currentAccessType, necessaryAccessLevelForLink) {
					form.LinkAccess = nil
					form.Access = nil
				} else {
					form.Access = access
				}
				form.Categories = categories
				form.Tags = tags
			}
			if err = getFileURLs(form, getFileOriginal, getFileBlur, getFilePlaceholder); err != nil {
				return nil, err
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
				form.UpdatesAccessToken = tokenString
			}
			updateDBData := bson.M{
				"$inc": bson.M{
					"views": 1,
				},
			}
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, updateDBData)
			if err != nil {
				return nil, err
			}
			script := elastic.NewScriptInline("ctx._source.views+=1;")
			_, err = elasticClient.Update().
				Index(formElasticIndex).
				Type(formElasticType).
				Id(formIDString).
				Script(script).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			return form, nil
		},
	},
}

func getFileURLs(form *Form, getFileOriginal bool, getFileBlur bool, getFilePlaceholder bool) error {
	var err error
	formIDString := form.ID
	if getFileOriginal || getFileBlur || getFilePlaceholder {
		for i := range form.Files {
			filetype := form.Files[i].Type
			fileID := form.Files[i].ID
			if getFileOriginal {
				filepath := formFileIndex + "/" + formIDString + "/" + fileID + originalPath
				form.Files[i].OriginalSrc, err = getSignedURL(filepath, validAccessTypes[2])
				if err != nil {
					return err
				}
			}
			if filetype == "image/png" || filetype == "image/jpeg" || filetype == "image/gif" {
				var addPlaceholderPath = ""
				if filetype == "image/gif" {
					addPlaceholderPath = placeholderPath
				}
				if getFileBlur {
					filepath := formFileIndex + "/" + formIDString + "/" + fileID + addPlaceholderPath + blurPath
					form.Files[i].BlurSrc, err = getSignedURL(filepath, validAccessTypes[2])
					if err != nil {
						return err
					}
				}
				if getFilePlaceholder {
					filepath := formFileIndex + "/" + formIDString + "/" + fileID + addPlaceholderPath + originalPath
					form.Files[i].PlaceholderSrc, err = getSignedURL(filepath, validAccessTypes[2])
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func deleteAllResponses(formID primitive.ObjectID) (int64, error) {
	logger.Info("delete all responses")
	formIDString := formID.Hex()
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
		return 0, err
	}
	var bytesRemoved int64 = 0
	for _, hit := range searchResult.Hits.Hits {
		responseIDString := hit.Id
		responseID, err := primitive.ObjectIDFromHex(responseIDString)
		if err != nil {
			return 0, err
		}
		newBytesRemoved, err := deleteResponse(responseID, nil)
		if err != nil {
			return 0, err
		}
		bytesRemoved += newBytesRemoved
	}
	return bytesRemoved, nil
}
