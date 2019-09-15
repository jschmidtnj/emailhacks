package main

import (
	"github.com/graphql-go/graphql"
)

// AccessInputType - type of graphql input
var AccessInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AccessInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

// AccessType - type of graphql input
var AccessType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Access",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
	},
})
