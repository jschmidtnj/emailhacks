package main

import (
	"github.com/graphql-go/graphql"
)

func rootQuery() *graphql.Object {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type:        graphql.String,
			Description: "Say Hi",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "Hello World!", nil
			},
		},
	}
	for key := range couponQueryFields {
		fields[key] = couponQueryFields[key]
	}
	for key := range productQueryFields {
		fields[key] = productQueryFields[key]
	}
	for key := range userQueryFields {
		fields[key] = userQueryFields[key]
	}
	for key := range responseQueryFields {
		fields[key] = responseQueryFields[key]
	}
	for key := range formQueryFields {
		fields[key] = formQueryFields[key]
	}
	for key := range projectQueryFields {
		fields[key] = projectQueryFields[key]
	}
	for key := range purchaseQueryFields {
		fields[key] = purchaseQueryFields[key]
	}
	for key := range blogQueryFields {
		fields[key] = blogQueryFields[key]
	}
	return graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: fields,
	})
}
