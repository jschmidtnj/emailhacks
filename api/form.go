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

// Form type
type Form struct {
	ID                 string      `json:"id"`
	Owner              string      `json:"owner"`
	Responses          int64       `json:"responses"`
	Created            int64       `json:"created"`
	Updated            int64       `json:"updated"`
	Project            string      `json:"project"`
	Name               string      `json:"name"`
	Items              []*FormItem `json:"items"`
	Multiple           bool        `json:"multiple"`
	Access             interface{} `json:"access"`
	Public             string      `json:"public"`
	Views              int64       `json:"Views"`
	Tags               []string    `json:"tags"`
	Categories         []string    `json:"categories"`
	Files              []*File     `json:"files"`
	UpdatesAccessToken string      `json:"updatesAccessToken"`
}

// FormType form type object for user forms graphql
var FormType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Form",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"owner": &graphql.Field{
			Type: graphql.String,
		},
		"responses": &graphql.Field{
			Type: graphql.Int,
		},
		"created": &graphql.Field{
			Type: graphql.Int,
		},
		"updated": &graphql.Field{
			Type: graphql.Int,
		},
		"project": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"items": &graphql.Field{
			Type: graphql.NewList(FormItemType),
		},
		"multiple": &graphql.Field{
			Type: graphql.Boolean,
		},
		"access": &graphql.Field{
			Type: graphql.NewList(AccessType),
		},
		"public": &graphql.Field{
			Type: graphql.String,
		},
		"views": &graphql.Field{
			Type: graphql.Int,
		},
		"tags": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"categories": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"files": &graphql.Field{
			Type: graphql.NewList(FileType),
		},
		"updatesAccessToken": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// FormUpdateType form update response type
var FormUpdateType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FormUpdate",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"items": &graphql.Field{
			Type: graphql.NewList(UpdateFormFormItemType),
		},
		"multiple": &graphql.Field{
			Type: graphql.Boolean,
		},
		"files": &graphql.Field{
			Type: graphql.NewList(UpdateFileType),
		},
	},
})

func getForm(formID primitive.ObjectID, updated bool) (*Form, error) {
	var form Form
	err := formCollection.FindOne(ctxMongo, bson.M{
		"_id": formID,
	}).Decode(&form)
	if err != nil {
		return nil, err
	}
	form.Access = form.Access.(bson.D).Map()
	access := make(map[string]primitive.M, len(form.Access.(bson.M)))
	for id, accessData := range form.Access.(bson.M) {
		primitiveAccessDoc, ok := accessData.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast access to primitive D")
		}
		access[id] = primitiveAccessDoc.Map()
	}
	form.Access = access
	form.Created = objectidTimestamp(formID).Unix()
	if updated {
		form.Updated = time.Now().Unix()
	}
	form.ID = formID.Hex()
	return &form, nil
}

func checkFormAccess(formID primitive.ObjectID, accessToken string, necessaryAccess []string, updated bool) (*Form, error) {
	form, err := getForm(formID, updated)
	if err != nil {
		return nil, err
	}
	// if public just break
	if findInArray(form.Public, necessaryAccess) {
		return form, nil
	}
	// next check if logged in
	claims, err := getTokenData(accessToken)
	if err != nil {
		return nil, err
	}
	// admin can do anything
	if claims["type"].(string) == adminType {
		return form, nil
	}
	var userIDString = claims["id"].(string)
	for currentUserID := range form.Access.(map[string]bson.M) {
		if currentUserID == userIDString {
			return form, nil
		}
	}
	// check if user has access to project directly
	projectID, err := primitive.ObjectIDFromHex(form.Project)
	if err != nil {
		return nil, err
	}
	_, _, err = checkProjectAccess(projectID, accessToken, "", necessaryAccess, false)
	if err == nil {
		return form, nil
	}
	return nil, errors.New("user not authorized to access form")
}

/**
 * @api {get} /countForms Count forms for search term
 * @apiVersion 0.0.1
 * @apiParam {String} searchterm Search term to count results
 * @apiSuccess {String} count Result count
 * @apiGroup misc
 */
func countForms(c *gin.Context) {
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
	project := request.URL.Query().Get("project")
	var foundProject = false
	if len(project) > 0 {
		foundProject = true
		_, err := primitive.ObjectIDFromHex(project)
		if err != nil {
			handleError("error getting project id value", http.StatusBadRequest, response)
			return
		}
	}
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
	mustQueries := make([]elastic.Query, numtags+len(categories)+1)
	if foundProject {
		mustQueries[0] = elastic.NewTermQuery("project", project)
	} else {
		// get all shared directly forms (not in a project)
		mustQueries[0] = elastic.NewTermsQuery(fmt.Sprintf("access.%s.type", userIDString), stringListToInterfaceList(viewAccessLevel)...)
	}
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
		mainquery := elastic.NewMultiMatchQuery(searchterm, formSearchFields...)
		query = query.Filter(mainquery)
	}
	count, err := elasticClient.Count().
		Type(formElasticType).
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
