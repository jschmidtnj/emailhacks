package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

func checkArrayInputObj(accessObj map[string]interface{}) error {
	if accessObj["index"] == nil {
		return errors.New("no index field given")
	}
	_, ok := accessObj["index"].(int)
	if !ok {
		return errors.New("cannot cast index to string")
	}
	if accessObj["updateAction"] == nil {
		return errors.New("no update action provided")
	}
	action, ok := accessObj["updateAction"].(string)
	if !ok {
		return errors.New("update action cannot be cast to string")
	}
	if !findInArray(action, validUpdateArrayActions) {
		return errors.New("invalid action given")
	}
	if accessObj["value"] == nil {
		return errors.New("value not found")
	}
	return nil
}

// StringArrayInputType - type of graphql input
var StringArrayInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "StringArrayInputType",
		Fields: graphql.InputObjectConfigFieldMap{
			"value": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"index": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"action": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	},
)

// IntArrayInputType - type of graphql input
var IntArrayInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "IntArrayInputType",
		Fields: graphql.InputObjectConfigFieldMap{
			"value": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"index": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"action": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	},
)
