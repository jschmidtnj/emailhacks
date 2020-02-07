package main

import (
	"time"

	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Organize object for tags, categories, etc.
type Organize struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// OrganizeType type for a name and count of that object
var OrganizeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Organize",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"color": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Billing object for billing settings
type Billing struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Company   string `json:"company"`
	Address1  string `json:"address1"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Currency  string `json:"currency"`
}

// BillingType billing type object for user billing settings graphql
var BillingType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Billing",
	Fields: graphql.Fields{
		"firstname": &graphql.Field{
			Type: graphql.String,
		},
		"lastname": &graphql.Field{
			Type: graphql.String,
		},
		"company": &graphql.Field{
			Type: graphql.String,
		},
		"address1": &graphql.Field{
			Type: graphql.String,
		},
		"city": &graphql.Field{
			Type: graphql.String,
		},
		"state": &graphql.Field{
			Type: graphql.String,
		},
		"zip": &graphql.Field{
			Type: graphql.String,
		},
		"country": &graphql.Field{
			Type: graphql.String,
		},
		"phone": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
		"currency": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Account object for users
type Account struct {
	ID             string            `json:"id"`
	Email          string            `json:"email"`
	Password       string            `json:"password"`
	Created        int64             `json:"created"`
	Updated        int64             `json:"updated"`
	EmailVerified  bool              `json:"emailverified"`
	Type           string            `json:"type"`
	Categories     []*Organize       `json:"categories"`
	Tags           []*Organize       `json:"tags"`
	Plan           string            `json:"plan"`
	Purchases      []string          `json:"purchases"`
	SubscriptionID string            `json:"subscriptionid"`
	StripeIDs      map[string]string `json:"stripids"`
	Storage        int64             `json:"storage"`
	Billing        *Billing          `json:"billing"`
}

// AccountType account type object for user accounts graphql
var AccountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Account",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.Int,
		},
		"updated": &graphql.Field{
			Type: graphql.Int,
		},
		"emailverified": &graphql.Field{
			Type: graphql.Boolean,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"categories": &graphql.Field{
			Type: graphql.NewList(OrganizeType),
		},
		"tags": &graphql.Field{
			Type: graphql.NewList(OrganizeType),
		},
		"plan": &graphql.Field{
			Type:        graphql.String,
			Description: "current plan",
		},
		"purchases": &graphql.Field{
			Type:        graphql.NewList(graphql.String),
			Description: "one-time purchases",
		},
		"billing": &graphql.Field{
			Type: BillingType,
		},
	},
})

// PublicAccountType data publically available
var PublicAccountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PublicAccount",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func getAccount(accountID primitive.ObjectID, updated bool) (*Account, error) {
	var account Account
	err := userCollection.FindOne(ctxMongo, bson.M{
		"_id": accountID,
	}).Decode(&account)
	if err != nil {
		return nil, err
	}
	account.Created = objectidTimestamp(accountID).Unix()
	if updated {
		account.Updated = time.Now().Unix()
	}
	account.ID = accountID.Hex()
	return &account, nil
}
