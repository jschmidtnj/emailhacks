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

// ResponseType response to form
var ResponseType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Response",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"views": &graphql.Field{
			Type: graphql.Int,
		},
		"user": &graphql.Field{
			Type: graphql.String,
		},
		"form": &graphql.Field{
			Type: graphql.String,
		},
		"project": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.String,
		},
		"updated": &graphql.Field{
			Type: graphql.String,
		},
		"items": &graphql.Field{
			Type: graphql.NewList(ResponseItemType),
		},
		"files": &graphql.Field{
			Type: graphql.NewList(FileType),
		},
	},
})

func processResponseFromDB(responseData bson.M, formatDate bool, updated bool) (bson.M, error) {
	id := responseData["_id"].(primitive.ObjectID)
	if formatDate {
		responseData["created"] = objectidTimestamp(id).Format(dateFormat)
	} else {
		responseData["created"] = objectidTimestamp(id).Unix()
	}
	var updatedTimestamp time.Time
	if updated {
		updatedTimestamp = time.Now()
	} else {
		updatedInt, ok := responseData["updated"].(int64)
		if !ok {
			return nil, errors.New("cannot cast updated time to int")
		}
		updatedTimestamp = intTimestamp(updatedInt)
	}
	if formatDate {
		responseData["updated"] = updatedTimestamp.Format(dateFormat)
	} else {
		responseData["updated"] = updatedTimestamp.Unix()
	}
	responseData["id"] = id.Hex()
	delete(responseData, "_id")
	itemsArray, ok := responseData["item"].(primitive.A)
	if !ok {
		return nil, errors.New("cannot cast items to array")
	}
	for i, item := range itemsArray {
		primitiveItem, ok := item.(primitive.D)
		if !ok {
			return nil, errors.New("cannot cast file to primitive D")
		}
		itemsArray[i] = primitiveItem.Map()
	}
	responseData["items"] = itemsArray
	return responseData, nil
}

func getResponse(responseID primitive.ObjectID, formatDate bool, updated bool) (map[string]interface{}, error) {
	responseDataCursor, err := responseCollection.Find(ctxMongo, bson.M{
		"_id": responseID,
	})
	defer responseDataCursor.Close(ctxMongo)
	if err != nil {
		return nil, err
	}
	var responseData map[string]interface{}
	var foundResponse = false
	for responseDataCursor.Next(ctxMongo) {
		foundResponse = true
		responsePrimitive := &bson.D{}
		err = responseDataCursor.Decode(responsePrimitive)
		if err != nil {
			return nil, err
		}
		responseData, err = processResponseFromDB(responsePrimitive.Map(), formatDate, updated)
		if err != nil {
			return nil, err
		}
		break
	}
	if !foundResponse {
		return nil, errors.New("response not found with given id")
	}
	return responseData, nil
}

func checkResponseAccess(responseID primitive.ObjectID, accessToken string, necessaryAccess []string, formatDate bool, updated bool) (map[string]interface{}, error) {
	responseData, err := getResponse(responseID, formatDate, updated)
	if err != nil {
		return nil, err
	}
	// check if logged in
	claims, err := getTokenData(accessToken)
	if err != nil {
		return nil, err
	}
	// admin can do anything
	if claims["type"] == adminType {
		return responseData, nil
	}
	// check if user submitted the response
	var userIDString = claims["id"].(string)
	if responseData["user"].(string) == userIDString {
		return responseData, nil
	} else if updated {
		return nil, errors.New("cannot edit another user's response")
	}
	// in this case you are viewing or adding a new response
	// check if user has access to form directly
	formIDString, ok := responseData["form"].(string)
	if !ok {
		return nil, errors.New("cannot cast project id to string")
	}
	formID, err := primitive.ObjectIDFromHex(formIDString)
	if err != nil {
		return nil, err
	}
	_, err = checkFormAccess(formID, accessToken, necessaryAccess, false, false)
	if err == nil {
		return responseData, nil
	}
	return nil, errors.New("user not authorized to access form")
}

/**
 * @api {get} /countResponses Count responses for search term
 * @apiVersion 0.0.1
 * @apiParam {String} searchterm Search term to count results
 * @apiSuccess {String} count Result count
 * @apiGroup misc
 */
func countResponses(c *gin.Context) {
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
	form := request.URL.Query().Get("form")
	var foundForm = false
	if len(form) > 0 {
		foundForm = true
		_, err := primitive.ObjectIDFromHex(form)
		if err != nil {
			handleError("error getting form id value", http.StatusBadRequest, response)
			return
		}
	}
	request.ParseForm()
	var numMustQueries = 0
	if foundForm {
		numMustQueries = 1
	}
	mustQueries := make([]elastic.Query, numMustQueries)
	if foundForm {
		mustQueries[0] = elastic.NewTermsQuery("form", form)
	} else {
		// otherwise get all responses for forms (not just one form)
		mustQueries[0] = elastic.NewTermQuery("user", userIDString)
	}
	query := elastic.NewBoolQuery()
	if len(mustQueries) > 0 {
		query = query.Must(mustQueries...)
	}
	if len(searchterm) > 0 {
		mainquery := elastic.NewMultiMatchQuery(searchterm, responseSearchFields...)
		query = query.Filter(mainquery)
	}
	count, err := elasticClient.Count().
		Type(responseElasticType).
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
