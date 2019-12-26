package main

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

func parseLiteral(astValue ast.Value) interface{} {
	kind := astValue.GetKind()

	switch kind {
	case kinds.StringValue:
		return astValue.GetValue()
	case kinds.BooleanValue:
		return astValue.GetValue()
	case kinds.IntValue:
		return astValue.GetValue()
	case kinds.FloatValue:
		return astValue.GetValue()
	case kinds.ObjectValue:
		obj := make(map[string]interface{})
		for _, v := range astValue.GetValue().([]*ast.ObjectField) {
			obj[v.Name.Value] = parseLiteral(v.Value)
		}
		return obj
	case kinds.ListValue:
		astValueList := astValue.GetValue().([]ast.Value)
		list := make([]interface{}, len(astValueList))
		for i, v := range astValueList {
			list[i] = parseLiteral(v)
		}
		return list
	default:
		return nil
	}
}

// JSON json type
var jsonType = graphql.NewScalar(
	graphql.ScalarConfig{
		Name:        "JSON",
		Description: "The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf)",
		Serialize: func(value interface{}) interface{} {
			return value
		},
		ParseValue: func(value interface{}) interface{} {
			return value
		},
		ParseLiteral: parseLiteral,
	},
)

// ItemType graphql question object
var ItemType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "ItemType",
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

// ItemInputType - type of graphql input
var ItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
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
			Type: graphql.NewList(ItemType),
		},
		"access": &graphql.Field{
			Type: graphql.NewList(AccessType),
		},
	},
})
