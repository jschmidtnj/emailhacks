package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"

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
			numProjectsAlready, err := elasticClient.Count().
				Type(projectElasticType).
				Query(query).
				Pretty(false).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			if numProjectsAlready >= int64(productData.MaxProjects) {
				return nil, errors.New("you reached the maximum amount of projects")
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
			uuidKey, err := uuid.NewRandom()
			if err != nil {
				return nil, err
			}
			key := uuidKey.String()
			projectID := primitive.NewObjectID()
			projectIDString := projectID.Hex()
			shortlink, err := generateShortLink(websiteURL + "?id=" + projectIDString + "?key=" + key)
			if err != nil {
				return nil, err
			}
			now := time.Now()
			userAccess := []map[string]interface{}{{
				"id":   userIDString,
				"type": editAccessLevel[0],
			}}
			projectData := bson.M{
				"_id":     projectID,
				"updated": now.Unix(),
				"name":    name,
				"forms":   int64(0),
				"views":   int64(0),
				"owner":   userIDString,
				"linkaccess": bson.M{
					"shortlink": shortlink,
					"secret":    key,
					"type":      noAccessLevel,
				},
				"access": bson.M{
					userIDString: bson.M{
						"type":       userAccess[0]["type"].(string),
						"categories": categories,
						"tags":       tags,
					},
				},
				"public": noAccessLevel,
			}
			_, err = projectCollection.InsertOne(ctxMongo, projectData)
			if err != nil {
				return nil, err
			}
			delete(projectData, "_id")
			projectData["created"] = now.Unix()
			_, err = elasticClient.Index().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				BodyJson(projectData).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			projectData["id"] = projectIDString
			delete(projectData["access"].(bson.M), userIDString)
			projectData["access"] = userAccess
			projectData["tags"] = tags
			projectData["categories"] = categories
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
			"linkaccess": &graphql.ArgumentConfig{
				Type: graphql.String,
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
			"accessKey": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sharable link key for project",
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
			var accessKey = ""
			if params.Args["accessKey"] != nil {
				accessKey, ok = params.Args["accessKey"].(string)
				if !ok {
					return nil, errors.New("cannot cast access key to string")
				}
			}
			projectData, _, err := checkProjectAccess(projectID, accessToken, accessKey, editAccessLevel, true)
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
				updateDataDB, err = changeUserAccessData(projectID, projectType, userIDString, categories, tags, access)
				if err != nil {
					return nil, err
				}
			} else {
				updateDataDB = bson.M{}
			}
			if updateDataDB["$set"] == nil {
				updateDataDB["$set"] = bson.M{}
			}
			newAccess, oldTags, oldCategories, currentAccessType, err := getFormattedAccessGQLData(projectData.Access, access, userIDString)
			if err != nil {
				return nil, err
			}
			if !findInArray(currentAccessType, editAccessLevel) && currentAccessType != "" {
				projectData.LinkAccess = nil
				projectData.Access = nil
			} else {
				projectData.Access = newAccess
			}
			if params.Args["categories"] == nil {
				projectData.Categories = oldCategories
			}
			if params.Args["tags"] == nil {
				projectData.Tags = oldTags
			}
			updateDataElastic := bson.M{}
			if params.Args["name"] != nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				updateDataDB["$set"].(bson.M)["name"] = name
				projectData.Name = name
				updateDataElastic["name"] = name
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
				projectData.Public = public
				updateDataElastic["public"] = public
			}
			if params.Args["linkaccess"] != nil {
				linkaccess, ok := params.Args["linkaccess"].(string)
				if !ok {
					return nil, errors.New("problem casting public to string")
				}
				if !findInArray(linkaccess, validAccessTypes) {
					return nil, errors.New("invalid link access level")
				}
				updateDataDB["$set"].(bson.M)["linkaccess"] = bson.M{
					"type": linkaccess,
				}
				projectData.LinkAccess.Type = linkaccess
				updateDataElastic["linkaccess"] = bson.M{
					"type": linkaccess,
				}
			}
			if len(access) > 0 {
				script := elastic.NewScript(addRemoveAccessScript).Params(map[string]interface{}{
					"access":     access,
					"tags":       tags,
					"categories": categories,
				})
				_, err = elasticClient.Update().
					Index(projectElasticIndex).
					Type(projectElasticType).
					Id(projectIDString).
					Script(script).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
			}
			updateDataElastic["updated"] = projectData.Updated
			_, err = elasticClient.Update().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				Doc(updateDataElastic).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
				"_id": projectID,
			}, updateDataDB)
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
			"accessKey": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sharable link key for project",
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
			var accessKey = ""
			if params.Args["accessKey"] != nil {
				accessKey, ok = params.Args["accessKey"].(string)
				if !ok {
					return nil, errors.New("cannot cast access key to string")
				}
			}
			projectData, _, err := checkProjectAccess(projectID, accessToken, accessKey, editAccessLevel, false)
			if err != nil {
				return nil, err
			}
			access, tags, categories, currentAccessType, err := getFormattedAccessGQLData(projectData.Access, nil, userIDString)
			if err != nil {
				return nil, err
			}
			err = deleteShortLink(projectData.LinkAccess.ShortLink)
			if err != nil {
				return nil, err
			}
			if !findInArray(currentAccessType, editAccessLevel) && currentAccessType != "" {
				projectData.LinkAccess = nil
				projectData.Access = nil
			} else {
				projectData.Access = access
			}
			projectData.Categories = categories
			projectData.Tags = tags
			if err = deleteProject(projectID); err != nil {
				return nil, err
			}
			return projectData, nil
		},
	},
}

func deleteProject(projectID primitive.ObjectID) error {
	sourceContext := elastic.NewFetchSourceContext(true).Include("id", "project")
	projectIDString := projectID.Hex()
	mustQueries := []elastic.Query{
		elastic.NewTermsQuery("project", projectIDString),
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
		formIDString := hit.Id
		formID, err := primitive.ObjectIDFromHex(formIDString)
		if err != nil {
			return err
		}
		if hit.Source == nil {
			return errors.New("no hit source found")
		}
		var formData map[string]interface{}
		err = json.Unmarshal(hit.Source, &formData)
		if err != nil {
			return err
		}
		project, ok := formData["project"].(string)
		if !ok {
			return errors.New("cannot cast project to string")
		}
		if err = changeFormProject(formIDString, project, "", "", ""); err != nil {
			return err
		}
		if _, err = deleteForm(formID, nil, ""); err != nil {
			return err
		}
	}
	_, err = elasticClient.Delete().
		Index(projectElasticIndex).
		Type(projectElasticType).
		Id(projectIDString).
		Do(ctxElastic)
	if err != nil {
		return err
	}
	_, err = projectCollection.DeleteOne(ctxMongo, bson.M{
		"_id": projectID,
	})
	if err != nil {
		return err
	}
	return nil
}
