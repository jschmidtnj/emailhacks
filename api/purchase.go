package main

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/webhook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// see https://stripe.com/docs/billing/subscriptions/payment
// https://stripe.com/docs/payments/payment-intents/migration/automatic-confirmation
// https://stripe.com/docs/payments/handling-payment-events#create-webhook

func handleStripeHooks(c *gin.Context) {
	response := c.Writer
	request := c.Request
	const MaxBodyBytes = int64(65536)
	request.Body = http.MaxBytesReader(response, request.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(request.Body)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	event, err := webhook.ConstructEvent(payload, request.Header.Get("Stripe-Signature"),
		stripeWebhookSecret)
	if err != nil {
		handleError(err.Error(), http.StatusBadRequest, response)
		return
	}
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		userUpdateData := bson.M{
			"$addToSet": bson.M{},
		}
		userUpdateData["$addToSet"].(bson.M)["purchases"] = paymentIntent.Metadata["productID"]
		userID, err := primitive.ObjectIDFromHex(paymentIntent.Metadata["userID"])
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		_, err = userCollection.UpdateOne(ctxMongo, bson.M{
			"_id": userID,
		}, userUpdateData)
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
	case "customer.subscription.trial_will_end":
		logger.Info("trial ended")
	case "payment_method.attached":
		var paymentMethod stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			handleError(err.Error(), http.StatusBadRequest, response)
			return
		}
		logger.Info("PaymentMethod was attached to a Customer!")
	// ... handle other event types
	default:
		handleError("event type not handled", http.StatusBadRequest, response)
		return
	}
	response.WriteHeader(http.StatusOK)
}

// user purchase a product
func purchase(userID primitive.ObjectID, productID primitive.ObjectID, couponIDString string, couponAmount int64, couponPercent bool, interval string, cardToken string) (*string, error) {
	productData, err := getProduct(productID, !isDebug())
	if err != nil {
		return nil, err
	}
	account, err := getAccount(userID, true)
	if err != nil {
		return nil, err
	}
	var foundPlan = false
	var planIDString string
	var amount int64
	for _, plan := range productData.Plans {
		if plan.Interval == interval {
			foundPlan = true
			if interval != singlePurchase {
				var foundCurrency = false
				for i := range plan.Currencies {
					if plan.Currencies[i].Currency == account.Billing.Currency {
						foundCurrency = true
						planIDString = plan.Currencies[i].StripeID
						break
					}
				}
				if !foundCurrency {
					return nil, errors.New("could not find currency for plan")
				}
				break
			} else {
				amount = plan.Amount
			}
			break
		}
	}
	if !foundPlan {
		return nil, errors.New("could not find plan")
	}
	userUpdateData := bson.M{
		"$set": bson.M{},
	}
	userIDString := userID.Hex()
	if _, ok := account.StripeIDs[account.Billing.Currency]; ok {
		if len(account.StripeIDs[account.Billing.Currency].Payment) > 0 {
			if _, err := stripeClient.PaymentMethods.Detach(account.StripeIDs[account.Billing.Currency].Payment, nil); err != nil {
				return nil, err
			}
		}
		userUpdateData["$set"].(bson.M)["stripeids."+account.Billing.Currency] = bson.M{
			"payment": &cardToken,
		}
		account.StripeIDs[account.Billing.Currency].Payment = cardToken
	} else {
		customerParams := &stripe.CustomerParams{
			Email: &account.Email,
			Params: stripe.Params{
				Metadata: map[string]string{
					"id":       userIDString,
					"currency": account.Billing.Currency,
				},
			},
		}
		newCustomer, err := stripeClient.Customers.New(customerParams)
		if err != nil {
			return nil, err
		}
		newStripeID := newCustomer.ID
		account.StripeIDs[account.Billing.Currency] = &PaymentIDs{
			Customer: newStripeID,
			Payment:  cardToken,
		}
	}
	userUpdateData["$set"].(bson.M)["stripeids."+account.Billing.Currency] = bson.M{
		"customer": account.StripeIDs[account.Billing.Currency].Customer,
		"payment":  account.StripeIDs[account.Billing.Currency].Payment,
	}
	if _, err := stripeClient.PaymentMethods.Attach(account.StripeIDs[account.Billing.Currency].Payment, &stripe.PaymentMethodAttachParams{
		Customer: &account.StripeIDs[account.Billing.Currency].Customer,
	}); err != nil {
		return nil, err
	}
	if _, err = stripeClient.Customers.Update(account.StripeIDs[account.Billing.Currency].Customer, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: &account.StripeIDs[account.Billing.Currency].Payment,
		},
	}); err != nil {
		return nil, err
	}
	clientSecret := ""
	if interval != singlePurchase {
		if len(account.SubscriptionID) > 0 {
			if _, err = stripeClient.Subscriptions.Cancel(account.SubscriptionID, nil); err != nil {
				return nil, err
			}
		}
		subscriptionParams := &stripe.SubscriptionParams{
			Customer: stripe.String(account.StripeIDs[account.Billing.Currency].Customer),
			Items: []*stripe.SubscriptionItemsParams{&stripe.SubscriptionItemsParams{
				Plan: &planIDString,
			},
			},
		}
		if len(couponIDString) > 0 {
			subscriptionParams.Coupon = &couponIDString
		}
		stripeSubscription, err := stripeClient.Subscriptions.New(subscriptionParams)
		if err != nil {
			return nil, err
		}
		userUpdateData["$set"].(bson.M)["plan"] = productData.ID
		userUpdateData["$set"].(bson.M)["subscriptionid"] = stripeSubscription.ID
	} else {
		var newPrice int64
		if couponPercent {
			if couponAmount >= 100 {
				newPrice = 0
			} else if couponAmount <= 0 {
				newPrice = amount
			} else {
				newPrice = int64(float64(couponAmount) / 100 * float64(amount))
			}
		} else {
			if couponAmount >= 100 {
				newPrice = 0
			} else if couponAmount <= 0 {
				newPrice = amount
			} else {
				newPrice = amount - couponAmount
			}
		}
		exchangeRate, err := getExchangeRate(account.Billing.Currency, false)
		if err != nil {
			return nil, err
		}
		newPrice = int64(100 * float64(newPrice) * *exchangeRate)
		paymentIntentParams := &stripe.PaymentIntentParams{
			Amount:   &newPrice,
			Currency: &account.Billing.Currency,
			Params: stripe.Params{
				Metadata: map[string]string{
					"productID": productID.Hex(),
					"userID":    userIDString,
				},
			},
		}
		paymentIntent, err := paymentintent.New(paymentIntentParams)
		if err != nil {
			return nil, err
		}
		clientSecret = paymentIntent.ClientSecret
	}
	// update user
	_, err = userCollection.UpdateOne(ctxMongo, bson.M{
		"_id": userID,
	}, userUpdateData)
	if err != nil {
		return nil, err
	}
	return &clientSecret, nil
}
