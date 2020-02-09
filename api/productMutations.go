package main

import (
	"errors"
	"strings"

	"github.com/graphql-go/graphql"
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
			plansInterface, ok := params.Args["plans"].([]interface{})
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
			var currencies *[]string
			if name != defaultPlanName {
				currencies, err = getCurrencies(false)
				if err != nil {
					return nil, err
				}
			} else {
				theCurrencies := []string{
					defaultCurrency,
				}
				currencies = &theCurrencies
			}
			allPlans := make([]*Plan, len(plans))
			for i := range plans {
				interval := plans[i]["interval"].(string)
				amount := int64(plans[i]["amount"].(int))
				if name == defaultPlanName && amount != 0 {
					return nil, errors.New("default plan amount is not 0")
				}
				var lenPlanCurrencies int
				if interval == singlePurchase || name == defaultPlanName {
					lenPlanCurrencies = 1
				} else {
					lenPlanCurrencies = len(*currencies)
				}
				allPlans[i] = &Plan{
					Interval:   interval,
					Amount:     amount,
					Currencies: make([]*PlanCurrency, lenPlanCurrencies),
				}
			}
			productData, err := addProduct(name, int64(maxProjects), int64(maxForms), int64(maxStorage), allPlans, currencies)
			return *productData, nil
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
			var updateStripeProduct = false
			stripeProductParams := &stripe.ProductParams{}
			if params.Args["name"] != nil {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("problem casting name to string")
				}
				updateDataDB["$set"].(bson.M)["name"] = name
				productData.Name = name
				stripeProductParams.Name = &name
				updateStripeProduct = true
			}
			if params.Args["plans"] != nil {
				plansInterface, ok := params.Args["plans"].([]interface{})
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
				var currencies *[]string
				if productData.Name != defaultPlanName {
					currencies, err = getCurrencies(false)
					if err != nil {
						return nil, err
					}
				} else {
					theCurrencies := []string{
						defaultCurrency,
					}
					currencies = &theCurrencies
				}
				allPlans := make([]*Plan, len(plans))
				for i := range plans {
					interval := plans[i]["interval"].(string)
					amount := int64(plans[i]["amount"].(int))
					if productData.Name == defaultPlanName && amount != 0 {
						return nil, errors.New("default plan amount is not 0")
					}
					var lenPlanCurrencies int
					if interval == singlePurchase || productData.Name == defaultPlanName {
						lenPlanCurrencies = 1
					} else {
						lenPlanCurrencies = len(*currencies)
					}
					allPlans[i] = &Plan{
						Interval:   interval,
						Amount:     amount,
						Currencies: make([]*PlanCurrency, lenPlanCurrencies),
					}
				}
				allPlans, err = addPlans(allPlans, currencies, stripeProductIDString, productData.Name)
				if err != nil {
					return nil, err
				}
				updateDataDB["$set"].(bson.M)["plans"] = allPlans
				productData.Plans = allPlans
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
			if updateStripeProduct {
				_, err = stripeClient.Products.Update(stripeProductIDString, stripeProductParams)
				if err != nil {
					return nil, err
				}
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
	"addCurrency": &graphql.Field{
		Type:        graphql.String,
		Description: "Add currency",
		Args: graphql.FieldConfigArgument{
			"currency": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["currency"] == nil {
				return nil, errors.New("name was not provided")
			}
			currency, ok := params.Args["currency"].(string)
			if !ok {
				return nil, errors.New("problem casting currency to string")
			}
			countryData, err := stripeClient.CountrySpec.Get(defaultCountry, nil)
			if err != nil {
				return nil, err
			}
			currency = strings.ToLower(currency)
			exists, err := currencyExists(currency)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.New("currency already added")
			}
			var exchangeRate float64
			if currency != defaultCurrency {
				currencyObj := stripe.Currency(currency)
				var foundCurrency = false
				for _, supportedCurrency := range countryData.SupportedPaymentCurrencies {
					if supportedCurrency == currencyObj {
						foundCurrency = true
						break
					}
				}
				if !foundCurrency {
					return nil, errors.New("currency not available in default country")
				}
				actualExchangeRate, err := getActualExchangeRate(currency)
				if err != nil {
					return nil, err
				}
				exchangeRate = *actualExchangeRate
			} else {
				exchangeRate = 1
			}
			currencyData := Currency{
				Name:         currency,
				ExchangeRate: exchangeRate,
			}
			_, err = currencyCollection.InsertOne(ctxMongo, currencyData)
			if err != nil {
				return nil, err
			}
			return currency, nil
		},
	},
	"deleteCurrency": &graphql.Field{
		Type:        graphql.String,
		Description: "Delete a Currency",
		Args: graphql.FieldConfigArgument{
			"currency": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			if params.Args["currency"] == nil {
				return nil, errors.New("currency name not provided")
			}
			currency, ok := params.Args["currency"].(string)
			if !ok {
				return nil, errors.New("cannot cast currency to string")
			}
			currency = strings.ToLower(currency)
			if currency == defaultCurrency {
				return nil, errors.New("cannot delete default currency")
			}
			exists, err := currencyExists(currency)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, errors.New("currency doesn't exist")
			}
			products, err := getProducts(false)
			if err != nil {
				return nil, err
			}
			for _, product := range *products {
				for _, plan := range product.Plans {
					for _, currencyData := range plan.Currencies {
						if currencyData.Currency == currency {
							if _, err := stripeClient.Plans.Del(currencyData.StripeID, nil); err != nil {
								return nil, err
							}
							break
						}
					}
				}
			}
			_, err = productCollection.UpdateMany(ctxMongo, bson.M{}, bson.M{
				"$unset": bson.M{
					"plans": bson.M{
						"currencies": bson.M{
							"currency": currency,
						},
					},
				},
			})
			if err != nil {
				return nil, err
			}
			accounts, err := getAccounts()
			if err != nil {
				return nil, err
			}
			for _, currentAccout := range *accounts {
				if stripeIDs, ok := currentAccout.StripeIDs[currency]; ok {
					if len(stripeIDs.Payment) > 0 {
						if _, err := stripeClient.PaymentMethods.Detach(stripeIDs.Payment, nil); err != nil {
							return nil, err
						}
					}
					if _, err := stripeClient.Customers.Del(stripeIDs.Customer, nil); err != nil {
						return nil, err
					}
				}
			}
			_, err = userCollection.UpdateMany(ctxMongo, bson.M{}, bson.M{
				"$unset": bson.M{
					"stripeids": currency,
				},
			})
			if err != nil {
				return nil, err
			}
			_, err = currencyCollection.DeleteOne(ctxMongo, bson.M{
				"name": currency,
			})
			if err != nil {
				return nil, err
			}
			return currency, nil
		},
	},
}

func deleteAllPlans(productData *Product) error {
	for _, currentPlan := range productData.Plans {
		if currentPlan.Interval != singlePurchase {
			for currencyIndex := range currentPlan.Currencies {
				if _, err := stripeClient.Plans.Del(currentPlan.Currencies[currencyIndex].StripeID, nil); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func addPlans(allPlans []*Plan, currencies *[]string, stripeProductIDString string, name string) ([]*Plan, error) {
	exchangeRates := map[string]float64{}
	for i, currency := range *currencies {
		currency = strings.ToLower(currency)
		currentExchangeRate, err := getActualExchangeRate(currency)
		if err != nil {
			return nil, err
		}
		exchangeRates[currency] = *currentExchangeRate
		(*currencies)[i] = currency
	}
	for i := range allPlans {
		if allPlans[i].Interval != singlePurchase {
			allPlans[i].Currencies = make([]*PlanCurrency, len(*currencies))
			for j := range *currencies {
				currentAmount := int64(100 * float64(allPlans[i].Amount) * exchangeRates[(*currencies)[j]])
				planParams := &stripe.PlanParams{
					ProductID:     &stripeProductIDString,
					BillingScheme: stripe.String("per_unit"),
					UsageType:     stripe.String("licensed"),
					Interval:      &allPlans[i].Interval,
					Currency:      &(*currencies)[j],
					Amount:        &currentAmount,
				}
				if name == defaultPlanName {
					// trial lasts forever
					planParams.TrialPeriodDays = stripe.Int64(720) // max days in trial
				}
				stripePlan, err := stripeClient.Plans.New(planParams)
				if err != nil {
					return nil, err
				}
				allPlans[i].Currencies[j] = &PlanCurrency{
					Currency: (*currencies)[j],
					StripeID: stripePlan.ID,
				}
				if name == defaultPlanName {
					break
				}
			}
		} else {
			// don't need plans for single purchase
			allPlans[i].Currencies[0] = &PlanCurrency{
				Currency: defaultCurrency,
				StripeID: "",
			}
		}
	}
	return allPlans, nil
}

func addProduct(name string, maxProjects int64, maxForms int64, maxStorage int64, allPlans []*Plan, currencies *[]string) (*Product, error) {
	stripeProduct, err := stripeClient.Products.New(&stripe.ProductParams{
		Active: stripe.Bool(true),
		Name:   &name,
		Type:   stripe.String("service"),
	})
	if err != nil {
		return nil, err
	}
	stripeProductIDString := stripeProduct.ID
	allPlans, err = addPlans(allPlans, currencies, stripeProductIDString, name)
	if err != nil {
		return nil, err
	}
	productData := Product{
		Name:        name,
		StripeID:    stripeProductIDString,
		Plans:       allPlans,
		MaxProjects: maxProjects,
		MaxForms:    maxForms,
		MaxStorage:  maxStorage,
	}
	productCreateRes, err := productCollection.InsertOne(ctxMongo, bson.M{
		"name":        name,
		"stripeid":    stripeProductIDString,
		"plans":       allPlans,
		"maxprojects": maxProjects,
		"maxforms":    maxForms,
		"maxstorage":  maxStorage,
	})
	if err != nil {
		return nil, err
	}
	productID := productCreateRes.InsertedID.(primitive.ObjectID)
	productIDString := productID.Hex()
	productData.ID = productIDString
	return &productData, nil
}
