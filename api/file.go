package main

import (
	"github.com/graphql-go/graphql"
)

// FileType graphql image object
var FileType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "File",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"width": &graphql.Field{
			Type: graphql.Int,
		},
		"height": &graphql.Field{
			Type: graphql.Int,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// FileInputType - type of graphql input
var FileInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "FileInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"updateAction": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"name": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"height": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"width": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"type": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	},
)
