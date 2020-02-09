package main

import (
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/olivere/elastic/v7"
)

// Blog object
type Blog struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Caption    string   `json:"caption"`
	Content    string   `json:"content"`
	Author     string   `json:"author"`
	Color      string   `json:"color"`
	Tags       []string `json:"tags"`
	Categories []string `json:"categories"`
	Views      int64    `json:"views"`
	Created    int64    `json:"created"`
	Updated    int64    `json:"updated"`
	HeroImage  *File    `json:"heroimage"`
	TileImage  *File    `json:"tileimage"`
	Files      []*File  `json:"files"`
	ShortLink  string   `json:"shortlink"`
}

// BlogType graphql blog object
var BlogType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Post",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"caption": &graphql.Field{
			Type: graphql.String,
		},
		"content": &graphql.Field{
			Type: graphql.String,
		},
		"author": &graphql.Field{
			Type: graphql.String,
		},
		"color": &graphql.Field{
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
		"created": &graphql.Field{
			Type: graphql.Int,
		},
		"updated": &graphql.Field{
			Type: graphql.Int,
		},
		"heroimage": &graphql.Field{
			Type: FileType,
		},
		"tileimage": &graphql.Field{
			Type: FileType,
		},
		"files": &graphql.Field{
			Type: graphql.NewList(FileType),
		},
		"shortlink": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func getBlog(blogID primitive.ObjectID, updated bool) (*Blog, error) {
	var blog Blog
	err := blogCollection.FindOne(ctxMongo, bson.M{
		"_id": blogID,
	}).Decode(&blog)
	if err != nil {
		return nil, err
	}
	blog.Created = objectidTimestamp(blogID).Unix()
	if updated {
		blog.Updated = time.Now().Unix()
	}
	blog.ID = blogID.Hex()
	return &blog, nil
}

/**
 * @api {get} /countBlogs Count posts for search term
 * @apiVersion 0.0.1
 * @apiParam {String} searchterm Search term to count results
 * @apiSuccess {String} count Result count
 * @apiGroup misc
 */
func countBlogs(c *gin.Context) {
	response := c.Writer
	request := c.Request
	if request.Method != http.MethodGet {
		handleError("register http method not Get", http.StatusBadRequest, response)
		return
	}
	searchterm := request.URL.Query().Get("searchterm")
	request.ParseForm()
	categoriesStr := request.URL.Query().Get("categories")
	categories := request.Form["categories"]
	if categories == nil {
		handleError("error getting categories string array from query", http.StatusBadRequest, response)
		return
	}
	categories = removeEmptyStrings(categories)
	tagsStr := request.URL.Query().Get("tags")
	tags := request.Form["tags"]
	if tags == nil {
		handleError("error getting tags string array from query", http.StatusBadRequest, response)
		return
	}
	tags = removeEmptyStrings(tags)
	getcache := request.URL.Query().Get("cache")
	response.Header().Set("Content-Type", "application/json")
	pathMap := map[string]string{
		"path":       "count",
		"searchterm": searchterm,
		"tags":       tagsStr,
		"categories": categoriesStr,
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	cachepath := string(cachepathBytes)
	if getcache != "false" && !isDebug() {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				handleError(err.Error(), http.StatusBadRequest, response)
				return
			}
		} else {
			response.Write([]byte(cachedres))
			return
		}
	}
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
		mainquery := elastic.NewMultiMatchQuery(searchterm, blogSearchFields...)
		query = query.Filter(mainquery)
	}
	count, err := elasticClient.Count().
		Type(blogElasticType).
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
	err = redisClient.Set(cachepath, string(countResBytes), cacheTime).Err()
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	response.Write(countResBytes)
}
