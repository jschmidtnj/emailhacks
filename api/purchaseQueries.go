package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
)

var purchaseQueryFields = graphql.Fields{
	"countries": &graphql.Field{
		Type:        graphql.NewList(graphql.String),
		Description: "Get list of possible countries",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			countryData, err := getCountries()
			if err != nil {
				return nil, err
			}
			return *countryData, nil
		},
	},
	"currencies": &graphql.Field{
		Type:        graphql.NewList(graphql.String),
		Description: "Get currencies",
		Args: graphql.FieldConfigArgument{
			"useCache": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			var useCache = true
			if params.Args["useCache"] != nil {
				useCache = params.Args["useCache"].(bool)
			}
			currencyData, err := getCurrencies(useCache)
			if err != nil {
				return nil, err
			}
			return *currencyData, nil
		},
	},
	"exchangerate": &graphql.Field{
		Type:        graphql.Float,
		Description: "Get exchange rate",
		Args: graphql.FieldConfigArgument{
			"currency": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if params.Args["currency"] == nil {
				return nil, errors.New("no currency argument found")
			}
			currency, ok := params.Args["currency"].(string)
			if !ok {
				return nil, errors.New("cannot cast currency to string")
			}
			exchangeRate, err := getExchangeRate(currency, true)
			if err != nil {
				return nil, err
			}
			return *exchangeRate, nil
		},
	},
}

func getActualExchangeRate(currency string) (*float64, error) {
	if currency == defaultCurrency {
		res := float64(1)
		return &res, nil
	}
	currency = strings.ToUpper(currency)
	request := fasthttp.AcquireRequest()
	request.SetRequestURI(foreignExchangeURL)
	request.URI().QueryArgs().Add("base", strings.ToUpper(defaultCurrency))
	request.URI().QueryArgs().Add("symbols", currency)
	response := fasthttp.AcquireResponse()
	httpClient.Do(request, response)
	body := response.Body()
	var exchangeData map[string]interface{}
	err := json.Unmarshal([]byte(body), &exchangeData)
	if err != nil {
		return nil, err
	}
	logger.Info(string(body))
	if exchangeData["error"] != nil {
		return nil, errors.New(exchangeData["error"].(string))
	}
	if exchangeData["rates"] == nil {
		return nil, errors.New("cannot find exchange rates array")
	}
	ratesArray, ok := exchangeData["rates"].(map[string]interface{})
	if !ok {
		return nil, errors.New("cannot convert rates to map")
	}
	if ratesArray[currency] == nil {
		return nil, errors.New("cannot find currency in response")
	}
	rate, ok := ratesArray[currency].(float64)
	if !ok {
		return nil, errors.New("cannot convert rate to float64")
	}
	return &rate, nil
}

func getExchangeRate(currency string, useCache bool) (*float64, error) {
	currency = strings.ToUpper(currency)
	pathMap := map[string]string{
		"path":     "forex",
		"currency": currency,
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	cachepath := string(cachepathBytes)
	if useCache {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			rate, err := strconv.ParseFloat(cachedres, 64)
			if err != nil {
				return nil, err
			}
			return &rate, nil
		}
	}
	var currencyObj Currency
	err = currencyCollection.FindOne(ctxMongo, bson.M{
		"name": currency,
	}).Decode(&currencyObj)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, fmt.Sprintf("%f", currencyObj.ExchangeRate), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return &currencyObj.ExchangeRate, nil
}

func getCountries() (*[]string, error) {
	pathMap := map[string]string{
		"path": "countries",
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	cachepath := string(cachepathBytes)
	if !isDebug() {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			var countries []string
			json.UnmarshalFromString(cachedres, &countries)
			return &countries, nil
		}
	}
	cursor := stripeClient.CountrySpec.List(nil)
	countries := []string{}
	for cursor.Next() {
		currentCountry := cursor.CountrySpec()
		countries = append(countries, currentCountry.ID)
	}
	countriesResBytes, err := json.Marshal(countries)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(countriesResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return &countries, nil
}

func currencyExists(currency string) (bool, error) {
	num, err := currencyCollection.CountDocuments(ctxMongo, bson.M{
		"name": currency,
	}, nil)
	if err != nil {
		return false, err
	}
	return num > 0, nil
}

func getCurrencies(useCache bool) (*[]string, error) {
	pathMap := map[string]string{
		"path": "currencies",
	}
	cachepathBytes, err := json.Marshal(pathMap)
	if err != nil {
		return nil, err
	}
	cachepath := string(cachepathBytes)
	if useCache && !isDebug() {
		cachedres, err := redisClient.Get(cachepath).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
		} else {
			var currencies []string
			json.UnmarshalFromString(cachedres, &currencies)
			return &currencies, nil
		}
	}
	cursor, err := currencyCollection.Find(ctxMongo, bson.M{}, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxMongo)
	currencies := []string{}
	for cursor.Next(ctxMongo) {
		currencyData := PlanCurrency{}
		if err = cursor.Decode(&currencyData); err != nil {
			return nil, err
		}
		currencies = append(currencies, currencyData.Currency)
	}
	currenciesResBytes, err := json.Marshal(currencies)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(currenciesResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return &currencies, nil
}
