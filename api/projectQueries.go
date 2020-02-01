package main

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
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
			"onlyshared": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// see this: https://github.com/olivere/elastic/issues/483
			// for potential fix to source issue (tried gave null pointer error)
			claims, err := getTokenData(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
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
			var onlyShared = false
			if params.Args["onlyshared"] != nil {
				onlyShared, ok = params.Args["onlyshared"].(bool)
				if !ok {
					return nil, errors.New("only shared could not be cast to bool")
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
			var projects []*Project
			if len(fields) > 0 {
				sourceContext := elastic.NewFetchSourceContext(true).Include(fields...)
				var numtags = len(tags)
				var showEverything = claims["type"] == superAdminType
				var startIndex = 0
				if !showEverything {
					startIndex = 1
				}
				mustQueries := make([]elastic.Query, numtags+len(categories)+startIndex)
				if !showEverything {
					necessaryAccessLevel := viewAccessLevel
					if onlyShared {
						necessaryAccessLevel = []string{sharedAccessLevel}
					}
					mustQueries[0] = elastic.NewTermsQuery(fmt.Sprintf("access.%s.type", userIDString), stringListToInterfaceList(necessaryAccessLevel)...)
				}
				for i, tag := range tags {
					mustQueries[i+startIndex] = elastic.NewTermQuery(fmt.Sprintf("access.%s.tags", userIDString), tag)
				}
				for i, category := range categories {
					mustQueries[i+numtags+startIndex] = elastic.NewTermQuery(fmt.Sprintf("access.%s.categories", userIDString), category)
				}
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
				projects = make([]*Project, len(searchResult.Hits.Hits))
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
					var currentProject Project
					if err = mapstructure.Decode(projectData, &currentProject); err != nil {
						return nil, err
					}
					currentProject.ID = id.Hex()
					currentProject.Created = objectidTimestamp(id).Unix()
					access, tags, categories, currentAccessType, err := getFormattedAccessGQLData(projectData["access"], nil, userIDString)
					if err != nil {
						return nil, err
					}
					var linkAccessData LinkAccess
					if err = mapstructure.Decode(projectData["linkaccess"], &linkAccessData); err != nil {
						return nil, err
					}
					currentProject.LinkAccess = &linkAccessData
					needEditAccessLevelForLink := findInArray(currentProject.LinkAccess.Type, editAccessLevel)
					necessaryAccessLevelForLink := viewAccessLevel
					if needEditAccessLevelForLink {
						necessaryAccessLevelForLink = editAccessLevel
					}
					logger.Info("testb")
					if !findInArray(currentAccessType, necessaryAccessLevelForLink) {
						delete(projectData, "linkaccess")
						delete(projectData, "access")
					} else {
						projectData["access"] = access
					}
					projectData["categories"] = categories
					projectData["tags"] = tags
					projects[i] = &currentProject
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
			"accessKey": &graphql.ArgumentConfig{
				Type:        graphql.String,
				Description: "sharable link key for project",
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := getTokenData(accessToken)
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
			var accessKey = ""
			if params.Args["accessKey"] != nil {
				accessKey, ok = params.Args["accessKey"].(string)
				if !ok {
					return nil, errors.New("cannot cast access key to string")
				}
			}
			project, _, err := checkProjectAccess(projectID, accessToken, accessKey, editAccessLevel, false)
			access, tags, categories, currentAccessType, err := getFormattedAccessGQLData(project.Access, nil, userIDString)
			if err != nil {
				return nil, err
			}
			needEditAccessLevelForLink := findInArray(project.LinkAccess.Type, editAccessLevel)
			necessaryAccessLevelForLink := viewAccessLevel
			if needEditAccessLevelForLink {
				necessaryAccessLevelForLink = editAccessLevel
			}
			if !findInArray(currentAccessType, necessaryAccessLevelForLink) {
				project.LinkAccess = nil
				project.Access = nil
			} else {
				project.Access = access
			}
			project.Categories = categories
			project.Tags = tags
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
			script := elastic.NewScriptInline("ctx._source.views+=1")
			_, err = elasticClient.Update().
				Index(projectElasticIndex).
				Type(projectElasticType).
				Id(projectIDString).
				Script(script).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			return project, nil
		},
	},
}
