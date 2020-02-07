package main

import (
	"time"

	"github.com/stripe/stripe-go"
	"github.com/vmihailenco/taskq/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var saveFormTask *taskq.Task

var updateForexTask *taskq.Task

func initDefaultPlan() error {
	_, err := getProduct(primitive.NilObjectID, false)
	if err != nil {
		logger.Info("create default plan")
		planObj := Plan{
			Interval:   validIntervals[0],
			Currencies: make([]*PlanCurrency, 1),
			Amount:     0,
		}
		allPlansArr := []*Plan{
			&planObj,
		}
		theCurrencies := []string{
			defaultCurrency,
		}
		currencies := &theCurrencies
		_, err = addProduct(defaultPlanName, 0, 0, 0, allPlansArr, currencies)
		if err != nil {
			return err
		}
	}
	return nil
}

func initQueue() {
	saveFormTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "saveForm",
		Handler: func(formIDString string) error {
			return updateForm(formIDString)
		},
	})
	updateForexTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "updateForex",
		Handler: func() error {
			return updateForex()
		},
	})
	scheduleNextUpdateForex()
}

func scheduleNextUpdateForex() {
	msg := updateForexTask.WithArgs(ctxMessageQueue).OnceInPeriod(time.Duration(forexUpdateTime) * time.Hour)
	msg.Delay = time.Duration(forexUpdateTime) * time.Hour
	if err := messageQueue.Add(msg); err != nil {
		logger.Info("update already saved: " + err.Error())
	}
}

func updateForex() error {
	exchangeRates := map[string]float64{}
	currencies, err := getCurrencies(false)
	if err != nil {
		return err
	}
	for _, currency := range *currencies {
		currentExchangeRate, err := getActualExchangeRate(currency)
		if err != nil {
			return err
		}
		exchangeRates[currency] = *currentExchangeRate
		if _, err = currencyCollection.UpdateOne(ctxMongo, bson.M{
			"currency": currency,
		}, bson.M{
			"$set": bson.M{
				"exchangerate": currentExchangeRate,
			},
		}); err != nil {
			return err
		}
	}
	products, err := getProducts(false)
	if err != nil {
		return err
	}
	for _, product := range *products {
		currencyData := bson.A{}
		for i := range product.Plans {
			interval := product.Plans[i].Interval
			if interval != singlePurchase {
				amount := int64(product.Plans[i].Amount)
				for _, currencyVal := range product.Plans[i].Currencies {
					currentAmount := int64(float64(amount) * exchangeRates[currencyVal.Currency])
					planParams := &stripe.PlanParams{
						ProductID:     &product.StripeID,
						BillingScheme: stripe.String("per_unit"),
						UsageType:     stripe.String("licensed"),
						Interval:      &interval,
						Amount:        &currentAmount,
						Currency:      &currencyVal.Currency,
					}
					stripePlan, err := stripeClient.Plans.New(planParams)
					if err != nil {
						return err
					}
					currencyData = append(currencyData, bson.M{
						"currency": currencyVal.Currency,
						"stripeid": stripePlan.ID,
					})
				}
			}
		}
		_, err = productCollection.UpdateOne(ctxMongo, bson.M{
			"_id": product.ID,
		}, bson.M{
			"$set": bson.M{
				"plans": bson.M{
					"currencies": currencyData,
				},
			},
		})
		if err != nil {
			return err
		}
	}
	scheduleNextUpdateForex()
	return nil
}
