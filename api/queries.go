package main

import (
	"github.com/graphql-go/graphql"
)

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
			"responses":  responseQueryFields["responses"],
			"response":   responseQueryFields["response"],
			"forms":      formQueryFields["forms"],
			"form":       formQueryFields["form"],
			"projects":   projectQueryFields["projects"],
			"project":    projectQueryFields["project"],
			"blogs":      blogQueryFields["blogs"],
			"blog":       blogQueryFields["blog"],
		},
	})
}
