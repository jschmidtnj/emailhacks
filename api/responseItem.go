package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// UpdateResponseItemInputType response item type
var UpdateResponseItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateResponseItemInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"updateAction": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"index": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"newIndex": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"formIndex": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"text": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"options": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int),
		},
		"files": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int),
		},
	},
})

// ResponseItemType response item type
var ResponseItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ResponseItem",
	Fields: graphql.Fields{
		"formIndex": &graphql.Field{
			Type: graphql.Int,
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
		"files": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
	},
})

// ResponseItemInputType response item type
var ResponseItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ResponseItemInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"formIndex": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"text": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"options": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int),
		},
		"files": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int),
		},
	},
})

func checkResponseItemObjCreate(itemObj map[string]interface{}) error {
	if itemObj["formIndex"] == nil {
		return errors.New("no form index field given")
	}
	if _, ok := itemObj["formIndex"].(int); !ok {
		return errors.New("cannot cast form index to int")
	}
	if itemObj["text"] == nil {
		return errors.New("no text field given")
	}
	if _, ok := itemObj["text"].(string); !ok {
		return errors.New("problem casting text to string")
	}
	if itemObj["options"] == nil {
		return errors.New("no options field given")
	}
	optionsArray, ok := itemObj["options"].([]interface{})
	if !ok {
		return errors.New("problem casting options to interface array")
	}
	if _, err := interfaceListToStringList(optionsArray); err != nil {
		return errors.New("problem casting options to string array")
	}
	filesArray, ok := itemObj["files"].([]interface{})
	if !ok {
		return errors.New("problem casting files to interface array")
	}
	if _, err := interfaceListToIntList(filesArray); err != nil {
		return errors.New("problem casting files to int array")
	}
	return nil
}

func checkResponseItemObjUpdatePart(itemObj map[string]interface{}) error {
	if itemObj["updateAction"] == nil {
		return errors.New("no update action given")
	}
	action, ok := itemObj["updateAction"].(string)
	if !ok {
		return errors.New("update action cannot be cast to string")
	}
	if !findInArray(action, validUpdateArrayActions) {
		return errors.New("invalid action given")
	}
	if action != validUpdateArrayActions[0] {
		if itemObj["index"] == nil {
			return errors.New("no index given")
		}
		if _, ok := itemObj["index"].(int); !ok {
			return errors.New("cannot cast index to int")
		}
	}
	if action == validUpdateArrayActions[2] {
		if itemObj["newIndex"] == nil {
			return errors.New("no new index given")
		}
		if _, ok := itemObj["newIndex"].(int); !ok {
			return errors.New("cannot cast to index to int")
		}
	}
	if itemObj["formIndex"] != nil {
		_, ok := itemObj["formIndex"].(int)
		if !ok {
			return errors.New("cannot cast form index to int")
		}
	}
	if itemObj["text"] != nil {
		_, ok := itemObj["text"].(string)
		if !ok {
			return errors.New("problem casting text to string")
		}
	}
	if itemObj["options"] != nil {
		optionsArray, ok := itemObj["options"].([]interface{})
		if !ok {
			return errors.New("problem casting options to interface array")
		}
		if _, err := interfaceListToStringList(optionsArray); err != nil {
			return errors.New("problem casting options to string array")
		}
	}
	if itemObj["files"] != nil {
		filesArray, ok := itemObj["files"].([]interface{})
		if !ok {
			return errors.New("problem casting files to interface array")
		}
		if _, err := interfaceListToIntList(filesArray); err != nil {
			return errors.New("problem casting files to int array")
		}
	}
	return nil
}
