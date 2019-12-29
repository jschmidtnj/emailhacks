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

var formQueryFields = graphql.Fields{
	"forms": &graphql.Field{
		Type:        graphql.NewList(FormType),
		Description: "Get list of forms",
		Args: graphql.FieldConfigArgument{
			"project": &graphql.ArgumentConfig{
				Type: graphql.String,
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
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			// see this: https://github.com/olivere/elastic/issues/483
			// for potential fix to source issue (tried gave null pointer error)
			accessToken := params.Context.Value(tokenKey).(string)
			claims, err := validateLoggedIn(accessToken)
			if err != nil {
				return nil, err
			}
			userIDString, ok := claims["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast user id to string")
			}
			var foundProject = false
			var project string
			if params.Args["project"] != nil {
				foundProject = true
				project, ok = params.Args["project"].(string)
				if !ok {
					return nil, errors.New("project could not be cast to string")
				}
				projectID, err := primitive.ObjectIDFromHex(project)
				if err != nil {
					return nil, err
				}
				_, err = checkProjectAccess(projectID, accessToken, viewAccessLevel, false, false)
				if err != nil {
					return nil, err
				}
			}
			var formatDate = false
			if params.Args["formatDate"] != nil {
				var ok bool
				formatDate, ok = params.Args["formatDate"].(bool)
				if !ok {
					return nil, errors.New("problem casting format date to boolean")
				}
			}
			_, err = primitive.ObjectIDFromHex(project)
			if err != nil {
				return nil, err
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
			var forms []map[string]interface{}
			if len(fields) > 0 {
				sourceContext := elastic.NewFetchSourceContext(true).Include(fields...)
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
				query := elastic.NewBoolQuery().Must(mustQueries...)
				if len(searchterm) > 0 {
					mainquery := elastic.NewMultiMatchQuery(searchterm, formSearchFields...)
					query = query.Filter(mainquery)
				}
				searchResult, err := elasticClient.Search().
					Index(formElasticIndex).
					Query(query).
					Sort(sort, ascending).
					From(page * perpage).Size(perpage).
					Pretty(isDebug()).
					FetchSourceContext(sourceContext).
					Do(ctxElastic)
				if err != nil {
					return nil, err
				}
				forms = make([]map[string]interface{}, len(searchResult.Hits.Hits))
				for i, hit := range searchResult.Hits.Hits {
					if hit.Source == nil {
						return nil, errors.New("no hit source found")
					}
					var formData map[string]interface{}
					err := json.Unmarshal(hit.Source, &formData)
					if err != nil {
						return nil, err
					}
					id, err := primitive.ObjectIDFromHex(hit.Id)
					if err != nil {
						return nil, err
					}
					createdTimestamp := objectidTimestamp(id)
					if formatDate {
						formData["created"] = createdTimestamp.Format(dateFormat)
					} else {
						formData["created"] = createdTimestamp.Unix()
					}
					updatedInt, ok := formData["updated"].(int32)
					if !ok {
						return nil, errors.New("cannot cast updated time to int")
					}
					updatedTimestamp := intTimestamp(int64(updatedInt))
					if formatDate {
						formData["updated"] = updatedTimestamp.Format(dateFormat)
					} else {
						formData["updated"] = updatedTimestamp.Unix()
					}
					formData["id"] = id.Hex()
					delete(formData, "_id")
					access, categories, tags := getFormattedGQLData(formData, nil, userIDString)
					formData["access"] = access
					formData["categories"] = categories
					formData["tags"] = tags
					forms[i] = formData
				}
			}
			return forms, nil
		},
	},
	"form": &graphql.Field{
		Type:        FormType,
		Description: "Get a Form",
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
			var userIDString = ""
			claims, err := validateLoggedIn(accessToken)
			if err == nil {
				userIDString = claims["id"].(string)
			}
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			formIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			formID, err := primitive.ObjectIDFromHex(formIDString)
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
			formData, err := checkFormAccess(formID, accessToken, editAccessLevel, formatDate, false)
			if err != nil {
				return nil, err
			}
			if len(userIDString) == 0 {
				formData["access"] = map[string]interface{}{}
				formData["categories"] = []string{}
				formData["tags"] = []string{}
			} else {
				access, categories, tags := getFormattedGQLData(formData, nil, userIDString)
				formData["access"] = access
				formData["categories"] = categories
				formData["tags"] = tags
			}
			_, err = formCollection.UpdateOne(ctxMongo, bson.M{
				"_id": formID,
			}, bson.M{
				"$inc": bson.M{
					"views": 1,
				},
			})
			if err != nil {
				return nil, err
			}
			views, ok := formData["views"].(int32)
			if !ok {
				return nil, errors.New("cannot convert views to int")
			}
			newViews := int(views) + 1
			_, err = elasticClient.Update().
				Index(formElasticIndex).
				Type(formElasticType).
				Id(formIDString).
				Doc(bson.M{
					"views": newViews,
				}).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			return formData, nil
		},
	},
}
