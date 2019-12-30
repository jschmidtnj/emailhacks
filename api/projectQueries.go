package main

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var projectQueryFields = graphql.Fields{
	"projects": &graphql.Field{
		Type:        graphql.NewList(ProjectType),
		Description: "Get list of projects",
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
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// see this: https://github.com/olivere/elastic/issues/483
			// for potential fix to source issue (tried gave null pointer error)
			claims, err := validateLoggedIn(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
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
			var projects []map[string]interface{}
			if len(fields) > 0 {
				sourceContext := elastic.NewFetchSourceContext(true).Include(fields...)
				var numtags = len(tags)
				mustQueries := make([]elastic.Query, numtags+len(categories)+1)
				mustQueries[0] = elastic.NewTermsQuery(fmt.Sprintf("access.%s.type", userIDString), stringListToInterfaceList(viewAccessLevel)...)
				for i, tag := range tags {
					mustQueries[i+1] = elastic.NewTermQuery(fmt.Sprintf("access.%s.tags", userIDString), tag)
				}
				for i, category := range categories {
					mustQueries[i+numtags+1] = elastic.NewTermQuery(fmt.Sprintf("access.%s.categories", userIDString), category)
				}
				// add search for access here
				query := elastic.NewBoolQuery().Must(mustQueries...)
				if len(searchterm) > 0 {
					mainquery := elastic.NewMultiMatchQuery(searchterm, projectSearchFields...)
					query = query.Filter(mainquery)
				}
				searchResult, err := elasticClient.Search().
					Index(projectElasticIndex).
					Query(query).
					Sort(sort, ascending).
					From(page * perpage).Size(perpage).
					Pretty(isDebug()).
					FetchSourceContext(sourceContext).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
				projects = make([]map[string]interface{}, len(searchResult.Hits.Hits))
				for i, hit := range searchResult.Hits.Hits {
					if hit.Source == nil {
						return nil, errors.New("no hit source found")
					}
					var projectData map[string]interface{}
					err := json.Unmarshal(hit.Source, &projectData)
					if err != nil {
						return nil, err
					}
					id, err := primitive.ObjectIDFromHex(hit.Id)
					if err != nil {
						return nil, err
					}
					createdTimestamp := objectidTimestamp(id)
					if formatDate {
						projectData["created"] = createdTimestamp.Format(dateFormat)
					} else {
						projectData["created"] = createdTimestamp.Unix()
					}
					updatedInt, ok := projectData["updated"].(float64)
					if !ok {
						return nil, errors.New("cannot cast updated time to float")
					}
					updatedTimestamp := intTimestamp(int64(updatedInt))
					if formatDate {
						projectData["updated"] = updatedTimestamp.Format(dateFormat)
					} else {
						projectData["updated"] = updatedTimestamp.Unix()
					}
					projectData["id"] = id.Hex()
					delete(projectData, "_id")
					access, tags, categories, err := getFormattedGQLData(projectData, nil, userIDString)
					if err != nil {
						return nil, err
					}
					projectData["access"] = access
					projectData["categories"] = categories
					projectData["tags"] = tags
					projects[i] = projectData
				}
			}
			return projects, nil
		},
	},
	"project": &graphql.Field{
		Type:        ProjectType,
		Description: "Get a Project",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := validateLoggedIn(accessToken)
			userIDString := claims["id"].(string)
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			projectIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			projectID, err := primitive.ObjectIDFromHex(projectIDString)
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
			projectData, err := checkProjectAccess(projectID, accessToken, editAccessLevel, formatDate, false)
			access, tags, categories, err := getFormattedGQLData(projectData, nil, userIDString)
			if err != nil {
				return nil, err
			}
			projectData["access"] = access
			projectData["categories"] = categories
			projectData["tags"] = tags
			if err != nil {
				return nil, err
			}
			_, err = projectCollection.UpdateOne(ctxMongo, bson.M{
				"_id": projectID,
			}, bson.M{
				"$inc": bson.M{
					"views": 1,
				},
			})
			if err != nil {
				return nil, err
			}
			views, ok := projectData["views"].(int32)
			if !ok {
				return nil, errors.New("cannot convert views to int")
			}
			newViews := int(views) + 1
			_, err = elasticClient.Update().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				Doc(bson.M{
					"views": newViews,
				}).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			return projectData, nil
		},
	},
}
