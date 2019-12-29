package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ItemType graphql question object
var ItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ItemType",
	Fields: graphql.Fields{
		"question": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"required": &graphql.Field{
			Type: graphql.Boolean,
		},
		"file": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// ItemInputType - type of graphql input
var ItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "QuestionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"question": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"options": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"text": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"required": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
		"file": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

// FormType form type object for user forms graphql
var FormType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Form",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.String,
		},
		"updated": &graphql.Field{
			Type: graphql.String,
		},
		"project": &graphql.Field{
			Type: graphql.String,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"items": &graphql.Field{
			Type: graphql.NewList(ItemType),
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
	},
})

func processFormFromDB(formData bson.M, formatDate bool, updated bool) (bson.M, error) {
	id := formData["_id"].(primitive.ObjectID)
	if formatDate {
		formData["created"] = objectidTimestamp(id).Format(dateFormat)
	} else {
		formData["created"] = objectidTimestamp(id).Unix()
	}
	var updatedTimestamp time.Time
	if updated {
		updatedTimestamp = time.Now()
	} else {
		updatedInt, ok := formData["updated"].(int32)
		if !ok {
			return nil, errors.New("cannot cast updated time to int")
		}
		updatedTimestamp = intTimestamp(int64(updatedInt))
	}
	if formatDate {
		formData["updated"] = updatedTimestamp.Format(dateFormat)
	} else {
		formData["updated"] = updatedTimestamp.Unix()
	}
	formData["id"] = id.Hex()
	delete(formData, "_id")
	fileArray, ok := formData["files"].(primitive.A)
	if !ok {
		return nil, errors.New("cannot cast files to array")
	}
	for i, file := range fileArray {
		primativeFile, ok := file.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast file to primitive D")
		}
		fileArray[i] = primativeFile.Map()
	}
	formData["files"] = fileArray
	itemArray, ok := formData["items"].(primitive.A)
	if !ok {
		return nil, errors.New("cannot cast items to array")
	}
	for i, item := range itemArray {
		primativeItem, ok := item.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast file to primitive D")
		}
		itemArray[i] = primativeItem.Map()
	}
	formData["items"] = itemArray
	accessPrimitive, ok := formData["access"].(primitive.M)
	if !ok {
		return nil, errors.New("cannot cast access to map")
	}
	access := make(map[string]primitive.M, len(accessPrimitive))
	for id, accessData := range accessPrimitive {
		primativeAccessDoc, ok := accessData.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast access to primitive D")
		}
		access[id] = primativeAccessDoc.Map()
	}
	formData["access"] = access
	return formData, nil
}

func getForm(formID primitive.ObjectID, formatDate bool, updated bool) (map[string]interface{}, error) {
	formDataCursor, err := formCollection.Find(ctxMongo, bson.M{
		"_id": formID,
	})
	defer formDataCursor.Close(ctxMongo)
	if err != nil {
		return nil, err
	}
	var formData map[string]interface{}
	var foundForm = false
	for formDataCursor.Next(ctxMongo) {
		foundForm = true
		formPrimitive := &bson.D{}
		err = formDataCursor.Decode(formPrimitive)
		if err != nil {
			return nil, err
		}
		formData, err = processFormFromDB(formPrimitive.Map(), formatDate, updated)
		if err != nil {
			return nil, err
		}
		break
	}
	if !foundForm {
		return nil, errors.New("form not found with given id")
	}
	return formData, nil
}

func checkFormAccess(formID primitive.ObjectID, accessToken string, necessaryAccess []string, formatDate bool, updated bool) (map[string]interface{}, error) {
	formData, err := getForm(formID, formatDate, updated)
	if err != nil {
		return nil, err
	}
	// if public just break
	var publicAccess = formData["public"].(string)
	if findInArray(publicAccess, viewAccessLevel) {
		return formData, nil
	}
	// next check if logged in
	claims, err := validateLoggedIn(accessToken)
	// admin can do anything
	_, err = validateAdmin(accessToken)
	if err == nil {
		return formData, nil
	}
	var userIDString = claims["id"].(string)
	access := formData["access"].(map[string]primitive.M)
	for currentUserID, _ := range access {
		if currentUserID == userIDString {
			return formData, nil
		}
	}
	// check if user has access to project directly
	projectIDString, ok := formData["project"].(string)
	if !ok {
		return nil, errors.New("cannot cast project id to string")
	}
	projectID, err := primitive.ObjectIDFromHex(projectIDString)
	if err != nil {
		return nil, err
	}
	_, err = checkProjectAccess(projectID, accessToken, necessaryAccess, false, false)
	if err == nil {
		return formData, nil
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
	searchterm := request.URL.Query().Get("searchterm")
	project := request.URL.Query().Get("project")
	_, err := primitive.ObjectIDFromHex(project)
	if err != nil {
		handleError("error getting project id value", http.StatusBadRequest, response)
		return
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
	mustQueries[0] = elastic.NewTermQuery("project", project)
	for i, tag := range tags {
		mustQueries[i+1] = elastic.NewTermQuery("tags", tag)
	}
	for i, category := range categories {
		mustQueries[i+numtags] = elastic.NewTermQuery("categories", category)
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
