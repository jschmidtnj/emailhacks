package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var responseQueryFields = graphql.Fields{
	"responses": &graphql.Field{
		Type:        graphql.NewList(ResponseType),
		Description: "Get responses",
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
			"form": &graphql.ArgumentConfig{
				Type: graphql.String,
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
			var foundForm = false
			var form string
			if params.Args["form"] != nil {
				form, ok = params.Args["form"].(string)
				if !ok {
					return nil, errors.New("form could not be cast to string")
				}
				if len(form) > 0 {
					foundForm = true
					formID, err := primitive.ObjectIDFromHex(form)
					if err != nil {
						return nil, err
					}
					_, err = checkFormAccess(formID, accessToken, viewAccessLevel, false)
					if err != nil {
						return nil, err
					}
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
			var responses []map[string]interface{}
			if len(fields) > 0 {
				sourceContext := elastic.NewFetchSourceContext(true).Include(fields...)
				var showEverything = claims["type"] == superAdminType
				var numMustQueries = 0
				if !showEverything && !foundForm {
					numMustQueries = 1
				}
				mustQueries := make([]elastic.Query, numMustQueries)
				if foundForm {
					mustQueries[0] = elastic.NewTermQuery("form", form)
				} else if !showEverything {
					// get all responses user has access to
					mustQueries[0] = elastic.NewTermsQuery("user", userIDString)
				}
				query := elastic.NewBoolQuery().Must(mustQueries...)
				if len(searchterm) > 0 {
					mainquery := elastic.NewMultiMatchQuery(searchterm, responseSearchFields...)
					query = query.Filter(mainquery)
				}
				searchResult, err := elasticClient.Search().
					Index(responseElasticIndex).
					Query(query).
					Sort(sort, ascending).
					From(page * perpage).Size(perpage).
					Pretty(isDebug()).
					FetchSourceContext(sourceContext).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
				responses = make([]map[string]interface{}, len(searchResult.Hits.Hits))
				for i, hit := range searchResult.Hits.Hits {
					if hit.Source == nil {
						return nil, errors.New("no hit source found")
					}
					var responseData map[string]interface{}
					err := json.Unmarshal(hit.Source, &responseData)
					if err != nil {
						return nil, err
					}
					id, err := primitive.ObjectIDFromHex(hit.Id)
					if err != nil {
						return nil, err
					}
					createdTimestamp := objectidTimestamp(id)
					responseData["created"] = createdTimestamp.Unix()
					responseData["id"] = id.Hex()
					delete(responseData, "_id")
					responses[i] = responseData
				}
			}
			return responses, nil
		},
	},
	"response": &graphql.Field{
		Type:        ResponseType,
		Description: "Get a Response",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			responseIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			responseID, err := primitive.ObjectIDFromHex(responseIDString)
			if err != nil {
				return nil, err
			}
			responseData, err := checkResponseAccess(responseID, accessToken, editAccessLevel, false)
			if err != nil {
				return nil, err
			}
			_, err = responseCollection.UpdateOne(ctxMongo, bson.M{
				"_id": responseID,
			}, bson.M{
				"$inc": bson.M{
					"views": 1,
				},
			})
			if err != nil {
				return nil, err
			}
			script := elastic.NewScriptInline("ctx.views+=1").Lang("painless")
			_, err = elasticClient.Update().
				Index(responseElasticIndex).
				Type(responseElasticType).
				Id(responseIDString).
				Script(script).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			return responseData, nil
		},
	},
}
