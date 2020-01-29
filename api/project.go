package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Project object
type Project struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Owner      string      `json:"owner"`
	Created    int64       `json:"created"`
	Updated    int64       `json:"updated"`
	Forms      int64       `json:"forms"`
	Access     interface{} `json:"access"`
	LinkAccess *LinkAccess `json:"linkaccess"`
	Public     string      `json:"public"`
	Tags       []string    `json:"tags"`
	Categories []string    `json:"categories"`
	Views      int64       `json:"Views"`
}

// ProjectType overarching project
var ProjectType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Project",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"owner": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.Int,
		},
		"updated": &graphql.Field{
			Type: graphql.Int,
		},
		"forms": &graphql.Field{
			Type: graphql.Int,
		},
		"access": &graphql.Field{
			Type: graphql.NewList(AccessType),
		},
		"linkaccess": &graphql.Field{
			Type: LinkAccessType,
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
		"views": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

func getProject(projectID primitive.ObjectID, updated bool) (*Project, error) {
	var project Project
	err := projectCollection.FindOne(ctxMongo, bson.M{
		"_id": projectID,
	}).Decode(&project)
	if err != nil {
		return nil, err
	}
	project.Access = project.Access.(bson.D).Map()
	access := make(map[string]primitive.M, len(project.Access.(bson.M)))
	for id, accessData := range project.Access.(bson.M) {
		primitiveAccessDoc, ok := accessData.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast access to primitive D")
		}
		access[id] = primitiveAccessDoc.Map()
	}
	project.Access = access
	project.Created = objectidTimestamp(projectID).Unix()
	if updated {
		project.Updated = time.Now().Unix()
	}
	project.ID = projectID.Hex()
	return &project, nil
}

func checkProjectAccess(projectID primitive.ObjectID, accessToken string, accessKey string, necessaryAccess []string, updated bool) (*Project, string, error) {
	project, err := getProject(projectID, updated)
	if err != nil {
		return nil, "", err
	}
	// if public just break
	if findInArray(project.Public, necessaryAccess) {
		return project, project.Public, nil
	}
	// next check if logged in
	claims, err := getTokenData(accessToken)
	if err != nil {
		return nil, "", err
	}
	// admin can do anything
	if claims["type"].(string) == adminType {
		return project, editAccessLevel[0], nil
	}
	// check for valid key
	if len(accessKey) > 0 {
		if project.LinkAccess.Secret != accessKey {
			return nil, "", errors.New("invalid access key")
		}
		if !findInArray(project.LinkAccess.Type, necessaryAccess) {
			return nil, "", errors.New("you do not have the necessary access")
		}
		return project, "", nil
	}
	var userIDString = claims["id"].(string)
	for currentUserID, accessVal := range project.Access.(map[string]bson.M) {
		if currentUserID == userIDString {
			return project, accessVal["type"].(string), nil
		}
	}
	return nil, "", errors.New("user not authorized to access project")
}

/**
 * @api {get} /countProjects Count projects for search term
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
	claims, err := getTokenData(getAuthToken(request))
	if err != nil {
		handleError("user not logged in", http.StatusBadRequest, response)
		return
	}
	userIDString, ok := claims["id"].(string)
	if !ok {
		handleError("cannot cast user id to string", http.StatusBadRequest, response)
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
	var onlyShared = false
	onlyshared := request.URL.Query().Get("onlyshared")
	if len(onlyshared) > 0 {
		onlyShared = onlyshared == "true"
	}
	tags = removeEmptyStrings(tags)
	var numtags = len(tags)
	mustQueries := make([]elastic.Query, numtags+len(categories)+1)
	necessaryAccessLevel := viewAccessLevel
	if onlyShared {
		necessaryAccessLevel = []string{sharedAccessLevel}
	}
	mustQueries[0] = elastic.NewTermsQuery(fmt.Sprintf("access.%s.type", userIDString), stringListToInterfaceList(necessaryAccessLevel)...)
	for i, tag := range tags {
		mustQueries[i+1] = elastic.NewTermQuery(fmt.Sprintf("access.%s.tags", userIDString), tag)
	}
	for i, category := range categories {
		mustQueries[i+numtags+1] = elastic.NewTermQuery(fmt.Sprintf("access.%s.categories", userIDString), category)
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
