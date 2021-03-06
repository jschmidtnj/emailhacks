package main

import (
	"errors"
	json "github.com/json-iterator/go"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var blogQueryFields = graphql.Fields{
	"blogs": &graphql.Field{
		Type:        graphql.NewList(BlogType),
		Description: "Get list of blogs",
		Args: graphql.FieldConfigArgument{
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
			"cache": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// see this: https://github.com/olivere/elastic/issues/483
			// for potential fix to source issue (tried gave null pointer error)
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
			if params.Args["cache"] == nil {
				return nil, errors.New("no cache argument found")
			}
			cache, ok := params.Args["cache"].(bool)
			if !ok {
				return nil, errors.New("cache could not be cast to bool")
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
			params.Args["cache"] = true
			pathMap := map[string]interface{}{
				"path":   "blogs",
				"args":   params.Args,
				"fields": fields,
			}
			cachepathBytes, err := json.Marshal(pathMap)
			if err != nil {
				return nil, err
			}
			cachepath := string(cachepathBytes)
			if cache && !isDebug() {
				cachedresStr, err := redisClient.Get(cachepath).Result()
				if err != nil {
					if err != redis.Nil {
						return nil, err
					}
				} else {
					if len(cachedresStr) > 0 {
						var cachedres []map[string]interface{}
						err = json.Unmarshal([]byte(cachedresStr), &cachedres)
						if err != nil {
							return nil, err
						}
						return cachedres, nil
					}
				}
			}
			var blogs []map[string]interface{}
			if len(fields) > 0 {
				sourceContext := elastic.NewFetchSourceContext(true).Include(fields...)
				var numtags = len(tags)
				mustQueries := make([]elastic.Query, numtags+len(categories))
				for i, tag := range tags {
					mustQueries[i] = elastic.NewTermQuery("tags", tag)
				}
				for i, category := range categories {
					mustQueries[i+numtags] = elastic.NewTermQuery("categories", category)
				}
				query := elastic.NewBoolQuery().Must(mustQueries...)
				if len(searchterm) > 0 {
					mainquery := elastic.NewMultiMatchQuery(searchterm, blogSearchFields...)
					query = query.Filter(mainquery)
				}
				searchResult, err := elasticClient.Search().
					Index(blogElasticIndex).
					Query(query).
					Sort(sort, ascending).
					From(page * perpage).Size(perpage).
					Pretty(false).
					FetchSourceContext(sourceContext).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
				blogs = make([]map[string]interface{}, len(searchResult.Hits.Hits))
				for i, hit := range searchResult.Hits.Hits {
					if hit.Source == nil {
						return nil, errors.New("no hit source found")
					}
					var blogData map[string]interface{}
					err := json.Unmarshal(hit.Source, &blogData)
					if err != nil {
						return nil, err
					}
					id, err := primitive.ObjectIDFromHex(hit.Id)
					if err != nil {
						return nil, err
					}
					blogData["created"] = objectidTimestamp(id).Unix()
					blogData["id"] = id.Hex()
					delete(blogData, "_id")
					blogs[i] = blogData
				}
			}
			blogsBytes, err := json.Marshal(blogs)
			if err != nil {
				return nil, err
			}
			err = redisClient.Set(cachepath, string(blogsBytes), cacheTime).Err()
			if err != nil {
				return nil, err
			}
			return blogs, nil
		},
	},
	"blog": &graphql.Field{
		Type:        BlogType,
		Description: "Get a Blog",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"cache": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if params.Args["cache"] == nil {
				return nil, errors.New("no cache argument found")
			}
			cache, ok := params.Args["cache"].(bool)
			if !ok {
				return nil, errors.New("cache could not be cast to bool")
			}
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idstring)
			if err != nil {
				return nil, err
			}
			_, err = blogCollection.UpdateOne(ctxMongo, bson.M{
				"_id": id,
			}, bson.M{
				"$inc": bson.M{
					"views": 1,
				},
			})
			if err != nil {
				return nil, err
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
			params.Args["cache"] = true
			pathMap := map[string]interface{}{
				"path":   "blog",
				"args":   params.Args,
				"fields": fields,
			}
			cachepathBytes, err := json.Marshal(pathMap)
			if err != nil {
				return nil, err
			}
			cachepath := string(cachepathBytes)
			if cache && !isDebug() {
				cachedresStr, err := redisClient.Get(cachepath).Result()
				if err != nil {
					if err != redis.Nil {
						return nil, err
					}
				} else {
					if len(cachedresStr) > 0 {
						var cachedres map[string]interface{}
						err = json.Unmarshal([]byte(cachedresStr), &cachedres)
						if err != nil {
							return nil, err
						}
						return cachedres, nil
					}
				}
			}
			blog, err := getBlog(id, false)
			if err != nil {
				return nil, err
			}
			_, err = elasticClient.Update().
				Index(blogElasticIndex).
				Type(blogElasticType).
				Id(idstring).
				Doc(bson.M{
					"views": blog.Views,
				}).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			blogBytes, err := json.Marshal(blog)
			if err != nil {
				return nil, err
			}
			err = redisClient.Set(cachepath, string(blogBytes), cacheTime).Err()
			if err != nil {
				return nil, err
			}
			return blog, nil
		},
	},
}
