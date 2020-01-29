package main

import (
	"bytes"
	"html/template"

	"github.com/graphql-go/graphql"
)

// FormEmailType form email data
var FormEmailType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FormEmail",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"data": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var formEmailTemplate = template.Must(template.ParseFiles("templates/formEmail.html"))

// SendEmailData send email object
type SendEmailData struct {
	Form  *Form  `json:"form"`
	Email string `json:"email"`
}

func getFormEmailData(emailData *SendEmailData) (string, error) {
	var templateData bytes.Buffer
	err := formEmailTemplate.Execute(&templateData, emailData)
	if err != nil {
		return "", err
	}
	minifiedData, err := minifier.String("text/html", templateData.String())
	if err != nil {
		return "", err
	}
	return minifiedData, nil
}
