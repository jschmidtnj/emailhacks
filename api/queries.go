package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// TODO - DELETE THIS - not being used
func getAccessDataQuery(userIDString string, formData map[string]interface{}) ([]map[string]interface{}, []string, []string, error) {
	if len(userIDString) == 0 {
		return []map[string]interface{}{}, []string{}, []string{}, nil
	}
	allAccessData, ok := formData["access"].(map[string]interface{})
	if !ok {
		return nil, nil, nil, errors.New("cannot cast access to array")
	}
	userAccessData, ok := allAccessData[userIDString].(map[string]interface{})
	if !ok {
		return nil, nil, nil, errors.New("cannot cast user access data to map")
	}
	categoriesArray, ok := userAccessData["categories"].([]interface{})
	if !ok {
		return nil, nil, nil, errors.New("cannot cast user categories to array")
	}
	categories, err := interfaceListToStringList(categoriesArray)
	if err != nil {
		return nil, nil, nil, err
	}
	tagsArray, ok := userAccessData["tags"].([]interface{})
	if !ok {
		return nil, nil, nil, errors.New("cannot cast user categories to array")
	}
	tags, err := interfaceListToStringList(tagsArray)
	if err != nil {
		return nil, nil, nil, err
	}
	finalAccessData := make([]map[string]interface{}, len(allAccessData))
	var j = 0
	for id, accessVal := range allAccessData {
		accessElem, ok := accessVal.(map[string]interface{})
		if !ok {
			return nil, nil, nil, errors.New("cannot cast access Elem to map")
		}
		accessType, ok := accessElem["type"].(string)
		if !ok {
			return nil, nil, nil, errors.New("cannot cast access type to string")
		}
		accessData := map[string]interface{}{
			"id":   id,
			"type": accessType,
		}
		finalAccessData[j] = accessData
	}
	return finalAccessData, categories, tags, nil
}

func rootQuery() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"hello": &graphql.Field{
				Type:        graphql.String,
				Description: "Say Hi",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "Hello World!", nil
				},
			},
			"account":    userQueryFields["account"],
			"user":       userQueryFields["user"],
			"userPublic": userQueryFields["userPublic"],
			"forms":      formQueryFields["forms"],
			"form":       formQueryFields["form"],
			"projects":   projectQueryFields["projects"],
			"project":    projectQueryFields["project"],
			"blogs":      blogQueryFields["blogs"],
			"blog":       blogQueryFields["blog"],
		},
	})
}
