package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// Plan object
type Plan struct {
	Interval string `json:"interval"`
	Amount   int64  `json:"amount"`
	StripeID string `json:"stripeid"`
}

// PlanType provides pricing and interval for product
var PlanType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Plan",
	Fields: graphql.Fields{
		"interval": &graphql.Field{
			Type: graphql.String,
		},
		"amount": &graphql.Field{
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
		"amount": &graphql.InputObjectFieldConfig{
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
	if itemObj["amount"] == nil {
		return errors.New("no amount field given")
	}
	if _, ok := itemObj["amount"].(int); !ok {
		return errors.New("problem casting amount to int")
	}
	return nil
}
