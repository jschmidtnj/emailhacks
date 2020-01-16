package main

import (
	"github.com/graphql-go/graphql"
)

// NameCountType type for a name and count of that object
var NameCountType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "NameCount",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"count": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

// AccountType account type object for user accounts graphql
var AccountType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "Account",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
		"password": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.String,
		},
		"updated": &graphql.Field{
			Type: graphql.String,
		},
		"emailverified": &graphql.Field{
			Type: graphql.Boolean,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"categories": &graphql.Field{
			Type: graphql.NewList(NameCountType),
		},
		"tags": &graphql.Field{
			Type: graphql.NewList(NameCountType),
		},
	},
})

// PublicAccountType data publically available
var PublicAccountType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "PublicAccount",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
	},
})
