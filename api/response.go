package main

import (
	"github.com/graphql-go/graphql"
)

// ResponseItemInputType response item type
var ResponseItemInputType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "ResponseItem",
	Fields: graphql.Fields{
		"index": &graphql.Field{
			Type: graphql.Int,
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
		"files": &graphql.Field{
			Type: graphql.NewList(FileInputType),
		},
	},
})

// ResponseItemType response item type
var ResponseItemType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "ResponseItem",
	Fields: graphql.Fields{
		"index": &graphql.Field{
			Type: graphql.Int,
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
		"files": &graphql.Field{
			Type: graphql.NewList(FileType),
		},
	},
})

// ResponseType response to form
var ResponseType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Response",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"user": &graphql.Field{
			Type: graphql.String,
		},
		"form": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.String,
		},
		"updated": &graphql.Field{
			Type: graphql.String,
		},
		"data": &graphql.Field{
			Type: graphql.NewList(ResponseItemType),
		},
	},
})
