package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getUpdateClaimsData(accessToken string) (string, string, error) {
	claims, err := getTokenData(accessToken)
	if err != nil {
		return "", "", err
	}
	if claims["type"] == nil {
		return "", "", errors.New("cannot find claims type")
	}
	claimsType, ok := claims["type"].(string)
	if !ok {
		return "", "", errors.New("cannot cast type to string")
	}
	if !findInArray(claimsType, editAccessLevel) {
		return "", "", errors.New("no edit access level found for editing form")
	}
	if claims["userid"] == nil {
		return "", "", errors.New("cannot find user id")
	}
	userIDString, ok := claims["userid"].(string)
	if !ok {
		return "", "", errors.New("cannot cast user id to string")
	}
	_, err = primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return "", "", err
	}
	formIDString, ok := claims["formid"].(string)
	if !ok {
		return "", "", errors.New("cannot cast user id to string")
	}
	_, err = primitive.ObjectIDFromHex(formIDString)
	if err != nil {
		return "", "", err
	}
	return formIDString, userIDString, nil
}

var collaborationFields = graphql.Fields{
	"formUpdates": &graphql.Field{
		Type:        FormType,
		Description: "Subscribe to form updates",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"updateToken": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"formatDate": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			data := params.Context.Value(dataKey)
			if data != nil {
				var dataObj map[string]interface{}
				if err := json.Unmarshal(data.([]byte), &dataObj); err != nil {
					return nil, err
				}
				var formatDate = false
				if params.Args["formatDate"] != nil {
					var ok bool
					formatDate, ok = params.Args["formatDate"].(bool)
					if !ok {
						return nil, errors.New("problem casting format date to boolean")
					}
				}
				if dataObj["created"] != nil && formatDate {
					if dataObj["id"] == nil {
						return nil, errors.New("id not found")
					}
					idString, ok := dataObj["id"].(string)
					if !ok {
						return nil, errors.New("unable to cast id to string")
					}
					formID, err := primitive.ObjectIDFromHex(idString)
					if err != nil {
						return nil, errors.New("unable to create object id from string")
					}
					dataObj["created"] = objectidTimestamp(formID).Format(dateFormat)
				}
				if dataObj["updated"] != nil && formatDate {
					updatedInt, ok := dataObj["updated"].(int64)
					if !ok {
						return nil, errors.New("cannot cast updated time to int")
					}
					dataObj["updated"] = intTimestamp(updatedInt).Format(dateFormat)
				}
				return dataObj, nil
			}
			fieldarray := params.Info.FieldASTs
			fieldselections := fieldarray[0].SelectionSet.Selections
			var foundIDField = false
			for _, field := range fieldselections {
				fieldast, ok := field.(*ast.Field)
				if !ok {
					return nil, errors.New("field cannot be converted to *ast.FIeld")
				}
				if fieldast.Name.Value == "id" {
					foundIDField = true
					break
				}
			}
			if !foundIDField {
				return nil, errors.New("you must query id field, even if you don't use it")
			}
			if params.Args["updateToken"] == nil {
				return nil, errors.New("cannot find update token")
			}
			updateTokenString, ok := params.Args["updateToken"].(string)
			if !ok {
				return nil, errors.New("cannot cast token to string")
			}
			tokenFormIDString, _, err := getUpdateClaimsData(updateTokenString)
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("cannot find form id")
			}
			givenFormIDString, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("unable to cast form id to string")
			}
			_, err = primitive.ObjectIDFromHex(givenFormIDString)
			if err != nil {
				return nil, errors.New("unable to create object id from string")
			}
			if givenFormIDString != tokenFormIDString {
				return nil, errors.New("token form id does not match given form id")
			}
			formData := map[string]interface{}{
				"id": updateTokenString,
			}
			return formData, nil
		},
	},
}
