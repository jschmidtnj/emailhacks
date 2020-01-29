package main

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
)

// ShortLinkType account type object for user accounts graphql
var ShortLinkType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ShortLink",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"link": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func deleteShortLink(id string) error {
	_, err := shortLinkCollection.DeleteOne(ctxMongo, bson.M{
		"_id": id,
	})
	if err != nil {
		return err
	}
	return nil
}

func generateShortLink(link string) (string, error) {
	guid := xid.New()
	id := guid.String()
	decodedLink, err := url.QueryUnescape(link)
	if err != nil {
		return "", err
	}
	shortLinkData := bson.M{
		"_id":  id,
		"link": decodedLink,
	}
	_, err = shortLinkCollection.InsertOne(ctxMongo, shortLinkData)
	if err != nil {
		return "", err
	}
	return id, nil
}

func getShortLink(idstring string) (string, error) {
	var idstringarr = []string{
		idstring,
	}
	idstrings, err := getShortLinks(idstringarr)
	if err != nil {
		return "", err
	}
	if len(idstrings) == 0 {
		return "", errors.New("no link found")
	}
	return idstrings[0], nil
}

func getShortLinks(ids []string) ([]string, error) {
	cursor, err := shortLinkCollection.Find(ctxMongo, bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	})
	defer cursor.Close(ctxMongo)
	if err != nil {
		return nil, errors.New("error: " + err.Error())
	}
	fullLinks := make([]string, 0)
	var foundstuff = false
	for cursor.Next(ctxMongo) {
		foundstuff = true
		shortLinkDataPrimitive := &bson.D{}
		err = cursor.Decode(shortLinkDataPrimitive)
		if err != nil {
			return nil, errors.New("problem decoding shortlink data: " + err.Error())
		}
		shortLinkData := shortLinkDataPrimitive.Map()
		fullLink, ok := shortLinkData["link"].(string)
		if !ok {
			return nil, errors.New("cannot cast link to string")
		}
		fullLinks = append(fullLinks, fullLink)
	}
	if !foundstuff {
		return fullLinks, nil
	}
	return fullLinks, nil
}

func shortLinkRedirect(c *gin.Context) {
	response := c.Writer
	request := c.Request
	id := request.URL.Query().Get("id")
	if len(id) != 20 {
		http.Redirect(response, request, websiteURL+"/404", 301)
		return
	}
	fullLink, err := getShortLink(id)
	if err != nil {
		http.Redirect(response, request, websiteURL+"/404", 301)
		return
	}
	http.Redirect(response, request, fullLink, 301)
}
