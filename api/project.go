package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProjectType overarching project
var ProjectType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Project",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"date": &graphql.Field{
			Type: graphql.String,
		},
		"forms": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"access": &graphql.Field{
			Type: graphql.NewList(AccessType),
		},
		"public": &graphql.Field{
			Type: graphql.String,
		},
		"tags": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"categories": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
	},
})

func checkProjectAccess(projectID primitive.ObjectID, accessToken string, necessaryAccess []string) (map[string]interface{}, []map[string]interface{}, error) {
	projectDataCursor, err := projectCollection.Find(ctxMongo, bson.M{
		"_id": projectID,
	})
	defer projectDataCursor.Close(ctxMongo)
	if err != nil {
		return nil, nil, err
	}
	var projectData map[string]interface{}
	var access []map[string]interface{}
	var foundProject = false
	for projectDataCursor.Next(ctxMongo) {
		foundProject = true
		projectPrimitive := &bson.D{}
		err = projectDataCursor.Decode(projectPrimitive)
		if err != nil {
			return nil, nil, err
		}
		projectData = projectPrimitive.Map()
		id := projectData["_id"].(primitive.ObjectID)
		projectData["date"] = objectidtimestamp(id).Format(dateFormat)
		projectData["id"] = id.Hex()
		delete(projectData, "_id")
		accessPrimitive, ok := projectData["access"].(primitive.A)
		if !ok {
			return nil, nil, errors.New("cannot cast access to array")
		}
		access = make([]map[string]interface{}, len(accessPrimitive))
		for i, accessData := range accessPrimitive {
			primativeAccessDoc, ok := accessData.(primitive.D)
			if !ok {
				return nil, nil, errors.New("cannot cast access to primitive D")
			}
			access[i] = primativeAccessDoc.Map()
		}
		projectData["access"] = access
		// if public just break
		publicAccess, ok := projectData["public"].(string)
		if findInArray(publicAccess, viewAccessLevel) {
			break
		}
		// next check if logged in
		claims, err := validateLoggedIn(accessToken)
		// admin can do anything
		_, err = validateAdmin(accessToken)
		if err == nil {
			break
		}
		userIDString, ok := claims["id"].(string)
		var foundUser = false
		for _, user := range access {
			accessUserIDString, ok := user["id"].(string)
			if !ok {
				return nil, nil, errors.New("cannot cast user id in form to string")
			}
			if accessUserIDString == userIDString {
				foundUser = true
				accessType, ok := user["access"].(string)
				if !ok {
					return nil, nil, errors.New("cannot cast access type to string")
				}
				if !findInArray(accessType, necessaryAccess) {
					return nil, nil, errors.New("user not authorized to access form")
				}
				break
			}
		}
		if !foundUser {
			// check if user has access to project directly
			return nil, nil, errors.New("user not authorized to access form")
		}
		break
	}
	if !foundProject {
		return nil, nil, errors.New("form not found with given id")
	}
	return projectData, access, nil
}

/**
 * @api {get} /countForms Count projects for search term
 * @apiVersion 0.0.1
 * @apiParam {String} searchterm Search term to count results
 * @apiSuccess {String} count Result count
 * @apiGroup misc
 */
func countProjects(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != http.MethodGet {
		handleError("register http method not Get", http.StatusBadRequest, response)
		return
	}
	searchterm := request.URL.Query().Get("searchterm")
	request.ParseForm()
	categories := request.Form["categories"]
	if categories == nil {
		handleError("error getting categories string array from query", http.StatusBadRequest, response)
		return
	}
	categories = removeEmptyStrings(categories)
	tags := request.Form["tags"]
	if tags == nil {
		handleError("error getting tags string array from query", http.StatusBadRequest, response)
		return
	}
	tags = removeEmptyStrings(tags)
	var numtags = len(tags)
	mustQueries := make([]elastic.Query, numtags+len(categories))
	for i, tag := range tags {
		mustQueries[i] = elastic.NewTermQuery("tags", tag)
	}
	for i, category := range categories {
		mustQueries[i+numtags] = elastic.NewTermQuery("categories", category)
	}
	query := elastic.NewBoolQuery()
	if len(mustQueries) > 0 {
		query = query.Must(mustQueries...)
	}
	if len(searchterm) > 0 {
		mainquery := elastic.NewMultiMatchQuery(searchterm, projectSearchFields...)
		query = query.Filter(mainquery)
	}
	count, err := elasticClient.Count().
		Type(projectElasticType).
		Query(query).
		Pretty(false).
		Do(ctxElastic)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	countMap := map[string]int64{
		"count": count,
	}
	countResBytes, err := json.Marshal(countMap)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	response.Write(countResBytes)
}
