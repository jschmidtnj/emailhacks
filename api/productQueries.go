package main

import (
	"errors"

	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var productQueryFields = graphql.Fields{
	"products": &graphql.Field{
		Type:        graphql.NewList(ProductType),
		Description: "Get list of products",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			products, err := getProducts(true)
			if err != nil {
				return nil, err
			}
			return *products, nil
		},
	},
	"currencyOptions": &graphql.Field{
		Type:        graphql.NewList(graphql.String),
		Description: "Get list of currency options",
		Args:        graphql.FieldConfigArgument{},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			accessToken := params.Context.Value(tokenKey).(string)
			_, err := validateAdmin(accessToken)
			if err != nil {
				return nil, err
			}
			countryData, err := stripeClient.CountrySpec.Get(defaultCountry, nil)
			if err != nil {
				return nil, err
			}
			currencies := []string{}
			for _, currency := range countryData.SupportedPaymentCurrencies {
				currencies = append(currencies, string(currency))
			}
			return currencies, nil
		},
	},
	"product": &graphql.Field{
		Type:        ProductType,
		Description: "Get a Product",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			if params.Args["id"] == nil {
				return nil, errors.New("no id argument found")
			}
			productIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			productID, err := primitive.ObjectIDFromHex(productIDString)
			if err != nil {
				return nil, err
			}
			productData, err := getProduct(productID, !isDebug())
			if err != nil {
				return nil, err
			}
			return productData, nil
		},
	},
}

func getProducts(useCache bool) (*[]Product, error) {
	pathMap := map[string]string{
		"path": "products",
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
			var products []Product
			json.UnmarshalFromString(cachedres, &products)
			return &products, nil
		}
	}
	cursor, err := productCollection.Find(ctxMongo, bson.M{}, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctxMongo)
	var products = []Product{}
	for cursor.Next(ctxMongo) {
		productDataPrimitive := &bson.D{}
		if err = cursor.Decode(productDataPrimitive); err != nil {
			return nil, err
		}
		productData := productDataPrimitive.Map()
		productID := productData["_id"].(primitive.ObjectID)
		var currentProduct Product
		if err = cursor.Decode(&currentProduct); err != nil {
			return nil, err
		}
		currentProduct.ID = productID.Hex()
		products = append(products, currentProduct)
	}
	productsResBytes, err := json.Marshal(products)
	if err != nil {
		return nil, err
	}
	err = redisClient.Set(cachepath, string(productsResBytes), cacheTime).Err()
	if err != nil {
		return nil, err
	}
	return &products, nil
}
