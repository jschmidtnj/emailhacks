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
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
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
			"cache": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// see this: https://github.com/olivere/elastic/issues/483
			// for potential fix to source issue (tried gave null pointer error)
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
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
					if formatDate {
						blogData["created"] = objectidTimestamp(id).Format(dateFormat)
					} else {
						blogData["created"] = objectidTimestamp(id).Unix()
					}
					updatedInt, ok := blogData["updated"].(int64)
					if !ok {
						return nil, errors.New("cannot cast updated time to int")
					}
					if formatDate {
						blogData["updated"] = intTimestamp(updatedInt).Format(dateFormat)
					} else {
						blogData["updated"] = intTimestamp(updatedInt).Unix()
					}
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
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
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
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
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
			cursor, err := blogCollection.Find(ctxMongo, bson.M{
				"_id": id,
			})
			defer cursor.Close(ctxMongo)
			if err != nil {
				return nil, err
			}
			var blogData map[string]interface{}
			var foundstuff = false
			for cursor.Next(ctxMongo) {
				blogPrimitive := &bson.D{}
				err = cursor.Decode(blogPrimitive)
				if err != nil {
					return nil, err
				}
				blogData = blogPrimitive.Map()
				if formatDate {
					blogData["created"] = objectidTimestamp(id).Format(dateFormat)
				} else {
					blogData["created"] = objectidTimestamp(id).Unix()
				}
				updatedInt, ok := blogData["updated"].(int64)
				if !ok {
					return nil, errors.New("cannot cast updated time to int")
				}
				if formatDate {
					blogData["updated"] = intTimestamp(updatedInt).Format(dateFormat)
				} else {
					blogData["updated"] = intTimestamp(updatedInt).Unix()
				}
				blogData["id"] = idstring
				if blogData["tileimage"] != nil {
					primitiveTileImage, ok := blogData["tileimage"].(primitive.D)
					if !ok {
						return nil, errors.New("cannot cast tile image to primitive D")
					}
					blogData["tileimage"] = primitiveTileImage.Map()
				}
				if blogData["heroimage"] != nil {
					primitiveHeroImage, ok := blogData["heroimage"].(primitive.D)
					if !ok {
						return nil, errors.New("cannot cast hero image to primitive D")
					}
					blogData["heroimage"] = primitiveHeroImage.Map()
				}
				fileArray, ok := blogData["files"].(primitive.A)
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
				blogData["files"] = fileArray
				delete(blogData, "_id")
				foundstuff = true
				break
			}
			if !foundstuff {
				return nil, errors.New("blog not found with given id")
			}
			_, err = elasticClient.Update().
				Index(blogElasticIndex).
				Type(blogElasticType).
				Id(idstring).
				Doc(bson.M{
					"views": int(blogData["views"].(int32)),
				}).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			blogBytes, err := json.Marshal(blogData)
			if err != nil {
				return nil, err
			}
			err = redisClient.Set(cachepath, string(blogBytes), cacheTime).Err()
			if err != nil {
				return nil, err
			}
			return blogData, nil
		},
	},
}
