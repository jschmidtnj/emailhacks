package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// PlanType provides pricing and interval for product
var PlanType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Plan",
	Fields: graphql.Fields{
		"interval": &graphql.Field{
			Type: graphql.String,
		},
		"price": &graphql.Field{
			Type: graphql.Int,
		},
		"stripeid": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// PlanInputType provides pricing and interval for product
var PlanInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "PlanInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"interval": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"price": &graphql.InputObjectFieldConfig{
			Type:        graphql.Int,
			Description: "amount in cents for plan",
		},
	},
})

func checkPlanItemObj(itemObj map[string]interface{}) error {
	if itemObj["interval"] == nil {
		return errors.New("no form index field given")
	}
	interval, ok := itemObj["interval"].(string)
	if !ok {
		return errors.New("cannot cast interval to string")
	}
	if !findInArray(interval, validIntervals) {
		return errors.New("invalid interval given")
	}
	if itemObj["price"] == nil {
		return errors.New("no price field given")
	}
	if _, ok := itemObj["price"].(int); !ok {
		return errors.New("problem casting price to int")
	}
	return nil
}
