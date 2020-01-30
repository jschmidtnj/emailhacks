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

// Response response object
type Response struct {
	ID      string          `json:"id"`
	Views   int64           `json:"views"`
	Owner   string          `json:"owner"`
	User    string          `json:"user"`
	Form    string          `json:"form"`
	Project string          `json:"project"`
	Created int64           `json:"created"`
	Updated int64           `json:"updated"`
	Items   []*ResponseItem `json:"items"`
	Files   []*File         `json:"files"`
}

// ResponseType response to form
var ResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Response",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"views": &graphql.Field{
			Type: graphql.Int,
		},
		"owner": &graphql.Field{
			Type: graphql.String,
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
			Type: graphql.Int,
		},
		"updated": &graphql.Field{
			Type: graphql.Int,
		},
		"items": &graphql.Field{
			Type: graphql.NewList(ResponseItemType),
		},
		"files": &graphql.Field{
			Type: graphql.NewList(FileType),
		},
		"editAccessToken": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func processResponseFromDB(responseData bson.M, updated bool) (bson.M, error) {
	id := responseData["_id"].(primitive.ObjectID)
	responseData["created"] = objectidTimestamp(id).Unix()
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
	responseData["updated"] = updatedTimestamp.Unix()
	responseData["id"] = id.Hex()
	delete(responseData, "_id")
	itemsArray, ok := responseData["items"].(primitive.A)
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

func getResponse(responseID primitive.ObjectID, updated bool) (*Response, error) {
	var response Response
	err := responseCollection.FindOne(ctxMongo, bson.M{
		"_id": responseID,
	}).Decode(&response)
	if err != nil {
		return nil, err
	}
	response.Created = objectidTimestamp(responseID).Unix()
	if updated {
		response.Updated = time.Now().Unix()
	}
	response.ID = responseID.Hex()
	return &response, nil
}

func checkResponseAccess(responseID primitive.ObjectID, accessToken string, necessaryAccess []string, updated bool) (*Response, error) {
	response, err := getResponse(responseID, updated)
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
		return response, nil
	}
	// check if user submitted the response
	var userIDString = claims["id"].(string)
	if response.User == userIDString {
		return response, nil
	} else if checkEquivalentStringArray(necessaryAccess, editAccessLevel) {
		return nil, errors.New("cannot edit another user's response")
	}
	// in this case you are viewing or adding a new response
	// check if user has access to form directly
	formID, err := primitive.ObjectIDFromHex(response.Form)
	if err != nil {
		return nil, err
	}
	_, err = checkFormAccess(formID, accessToken, "", necessaryAccess, false)
	if err == nil {
		return response, nil
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
	mustQueries := make([]elastic.Query, 1)
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
