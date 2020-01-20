package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/stripe/stripe-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var productMutationFields = graphql.Fields{
	"addProduct": &graphql.Field{
		Type:        ProductType,
		Description: "Create a Product",
		Args: graphql.FieldConfigArgument{
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"plans": &graphql.ArgumentConfig{
				Type: graphql.NewList(PlanInputType),
			},
			"maxprojects": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "maximum number of projects a user can make",
			},
			"maxforms": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "maximum number of forms a user can make",
			},
			"maxstorage": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "maximum amount of file storage in Gb",
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["name"] == nil {
				return nil, errors.New("name was not provided")
			}
			name, ok := params.Args["name"].(string)
			if !ok {
				return nil, errors.New("problem casting name to string")
			}
			if params.Args["plans"] == nil {
				return nil, errors.New("plans was not provided")
			}
			plansInterface, ok := params.Args["files"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting plans to interface array")
			}
			plans, err := interfaceListToMapList(plansInterface)
			if err != nil {
				return nil, err
			}
			for _, plan := range plans {
				if err := checkPlanItemObj(plan); err != nil {
					return nil, err
				}
			}
			if params.Args["maxprojects"] == nil {
				return nil, errors.New("max projects was not provided")
			}
			maxProjects, ok := params.Args["maxprojects"].(int)
			if !ok {
				return nil, errors.New("problem casting max projects to int")
			}
			if maxProjects <= 0 && maxProjects != -1 {
				return nil, errors.New("invalid max projects given")
			}
			if params.Args["maxforms"] == nil {
				return nil, errors.New("max forms was not provided")
			}
			maxForms, ok := params.Args["maxforms"].(int)
			if !ok {
				return nil, errors.New("problem casting max forms to int")
			}
			if maxForms <= 0 && maxForms != -1 {
				return nil, errors.New("invalid max forms given")
			}
			if params.Args["maxstorage"] == nil {
				return nil, errors.New("max storage was not provided")
			}
			maxStorage, ok := params.Args["maxstorage"].(int)
			if !ok {
				return nil, errors.New("problem casting max storage to int")
			}
			if maxStorage <= 0 && maxStorage != -1 {
				return nil, errors.New("invalid max storage given")
			}
			stripeProduct, err := stripeClient.Products.New(&stripe.ProductParams{
				Active: stripe.Bool(true),
				Name:   &name,
				Type:   stripe.String("service"),
			})
			if err != nil {
				return nil, err
			}
			stripeProductIDString := stripeProduct.ID
			returnPlans := make([]map[string]interface{}, len(plans))
			for i := range plans {
				interval := plans[i]["interval"].(string)
				for key, val := range plans[i] {
					returnPlans[i][key] = val
				}
				if interval != singlePurchase {
					planParams := &stripe.PlanParams{
						ProductID:     &stripeProductIDString,
						BillingScheme: stripe.String("per_unit"),
						UsageType:     stripe.String("licensed"),
						Interval:      &interval,
						Amount:        stripe.Int64(int64(plans[i]["amount"].(int))),
						Currency:      &defaultCurrency,
					}
					stripePlan, err := stripeClient.Plans.New(planParams)
					if err != nil {
						return nil, err
					}
					plans[i]["stripeid"] = stripePlan.ID
				} else {
					plans[i]["stripeid"] = ""
				}
			}
			productData := bson.M{
				"name":        name,
				"stripeid":    stripeProductIDString,
				"plans":       plans,
				"maxprojects": int64(maxProjects),
				"maxforms":    int64(maxForms),
				"maxstorage":  int64(maxStorage),
			}
			productCreateRes, err := productCollection.InsertOne(ctxMongo, productData)
			if err != nil {
				return nil, err
			}
			productID := productCreateRes.InsertedID.(primitive.ObjectID)
			productIDString := productID.Hex()
			productData["id"] = productIDString
			productData["plans"] = returnPlans
			return productData, nil
		},
	},
	"updateProduct": &graphql.Field{
		Type:        ProductType,
		Description: "Update a Product",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"plans": &graphql.ArgumentConfig{
				Type: graphql.NewList(PlanInputType),
			},
			"maxprojects": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "maximum number of projects a user can make",
			},
			"maxforms": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "maximum number of forms a user can make",
			},
			"maxstorage": &graphql.ArgumentConfig{
				Type:        graphql.Int,
				Description: "maximum amount of file storage in Gb",
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("name was not provided")
			}
			productIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("problem casting id to string")
			}
			productID, err := primitive.ObjectIDFromHex(productIDString)
			if err != nil {
				return nil, err
			}
			productData, err := getProduct(productID, false)
			if err != nil {
				return nil, err
			}
			stripeProductIDString := productData.StripeID
			updateDataDB := bson.M{
				"$set": bson.M{},
			}
			stripeProductParams := &stripe.ProductParams{
				Active: stripe.Bool(true),
				Type:   stripe.String("service"),
			}
			if params.Args["name"] == nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				updateDataDB["$set"].(bson.M)["name"] = name
				stripeProductParams.Name = &name
			}
			if params.Args["plans"] != nil {
				plansInterface, ok := params.Args["files"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting plans to interface array")
				}
				plans, err := interfaceListToMapList(plansInterface)
				if err != nil {
					return nil, err
				}
				for _, plan := range plans {
					if err := checkPlanItemObj(plan); err != nil {
						return nil, err
					}
				}
				if err = deleteAllPlans(productData); err != nil {
					return nil, err
				}
				returnPlans := make([]*Plan, len(plans))
				for i := range plans {
					if err = mapstructure.Decode(plans[i], &returnPlans[i]); err != nil {
						return nil, err
					}
					interval := plans[i]["interval"].(string)
					if interval != singlePurchase {
						planParams := &stripe.PlanParams{
							ProductID:     &stripeProductIDString,
							BillingScheme: stripe.String("per_unit"),
							UsageType:     stripe.String("licensed"),
							Interval:      &interval,
							Amount:        stripe.Int64(int64(plans[i]["amount"].(int))),
							Currency:      &defaultCurrency,
						}
						stripePlan, err := stripeClient.Plans.New(planParams)
						if err != nil {
							return nil, err
						}
						plans[i]["stripeid"] = stripePlan.ID
					} else {
						plans[i]["stripeid"] = ""
					}
				}
				updateDataDB["$set"].(bson.M)["plans"] = plans
				productData.Plans = returnPlans
			}
			if params.Args["maxprojects"] != nil {
				maxProjects, ok := params.Args["maxprojects"].(int)
				if !ok {
					return nil, errors.New("problem casting max projects to int")
				}
				if maxProjects <= 0 && maxProjects != -1 {
					return nil, errors.New("invalid max projects given")
				}
				updateDataDB["$set"].(bson.M)["maxprojects"] = maxProjects
			}
			if params.Args["maxforms"] != nil {
				maxForms, ok := params.Args["maxforms"].(int)
				if !ok {
					return nil, errors.New("problem casting max forms to int")
				}
				if maxForms <= 0 && maxForms != -1 {
					return nil, errors.New("invalid max forms given")
				}
				updateDataDB["$set"].(bson.M)["maxforms"] = maxForms
			}
			if params.Args["maxstorage"] != nil {
				maxStorage, ok := params.Args["maxstorage"].(int)
				if !ok {
					return nil, errors.New("problem casting max storage to int")
				}
				if maxStorage <= 0 && maxStorage != -1 {
					return nil, errors.New("invalid max storage given")
				}
				updateDataDB["$set"].(bson.M)["maxstorage"] = maxStorage
			}
			_, err = stripeClient.Products.Update(stripeProductIDString, stripeProductParams)
			if err != nil {
				return nil, err
			}
			_, err = productCollection.UpdateOne(ctxMongo, bson.M{
				"_id": productID,
			}, updateDataDB)
			if err != nil {
				return nil, err
			}
			return productData, nil
		},
	},
	"deleteProduct": &graphql.Field{
		Type:        ProductType,
		Description: "Delete a Product",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("product id not provided")
			}
			productIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast product id to string")
			}
			productID, err := primitive.ObjectIDFromHex(productIDString)
			if err != nil {
				return nil, err
			}
			productData, err := getProduct(productID, false)
			if err != nil {
				return nil, err
			}
			if err = deleteAllPlans(productData); err != nil {
				return nil, err
			}
			if _, err := stripeClient.Products.Del(productData.StripeID, nil); err != nil {
				return nil, err
			}
			_, err = productCollection.DeleteOne(ctxMongo, bson.M{
				"_id": productID,
			})
			if err != nil {
				return nil, err
			}
			return productData, nil
		},
	},
}

func deleteAllPlans(productData *Product) error {
	for _, currentPlan := range productData.Plans {
		if _, err := stripeClient.Plans.Del(currentPlan.StripeID, nil); err != nil {
			return err
		}
	}
	return nil
}
