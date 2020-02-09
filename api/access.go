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

// Access object for giving user access
type Access struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

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

// LinkAccess object for link user access
type LinkAccess struct {
	ShortLink string `json:"shortlink"`
	Secret    string `json:"secret"`
	Type      string `json:"type"`
}

// LinkAccessType - type of graphql input
var LinkAccessType = graphql.NewObject(graphql.ObjectConfig{
	Name: "LinkAccess",
	Fields: graphql.Fields{
		"shortlink": &graphql.Field{
			Type: graphql.String,
		},
		"secret": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
	},
})
