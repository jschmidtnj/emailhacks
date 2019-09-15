package main

import (
	"github.com/graphql-go/graphql"
)

// QuestionType graphql question object
var QuestionType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "QuestionType",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"required": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

// QuestionInputType - type of graphql input
var QuestionInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "QuestionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"name": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"options": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"required": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
	},
})

// FormType form type object for user forms graphql
var FormType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Form",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"description": &graphql.Field{
			Type: graphql.String,
		},
		"questions": &graphql.Field{
			Type: graphql.NewList(QuestionType),
		},
		"access": &graphql.Field{
			Type: graphql.NewList(AccessType),
		},
	},
})
