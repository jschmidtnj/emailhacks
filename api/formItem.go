package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// UpdateItemResponseType response for item update
var UpdateItemResponseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateItemResponse",
	Fields: graphql.Fields{
		"question": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"required": &graphql.Field{
			Type: graphql.Boolean,
		},
		"files": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
		"updateAction": &graphql.Field{
			Type: graphql.String,
		},
		"index": &graphql.Field{
			Type: graphql.Int,
		},
		"newIndex": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

// UpdateItemInputType - type of graphql input
var UpdateItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateItemInput",
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
		"question": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"options": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"text": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"required": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
		"files": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int), // maybe change to IntArrayInputType
		},
	},
})

// ItemType graphql question object
var ItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Item",
	Fields: graphql.Fields{
		"question": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"text": &graphql.Field{
			Type: graphql.String,
		},
		"required": &graphql.Field{
			Type: graphql.Boolean,
		},
		// file is the index of the file in the array
		"files": &graphql.Field{
			Type: graphql.NewList(graphql.Int),
		},
	},
})

// ItemInputType - type of graphql input
var ItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ItemInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"question": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"options": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"text": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"required": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
		"files": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int), // maybe change to IntArrayInputType
		},
	},
})

func checkItemObjCreate(itemObj map[string]interface{}) error {
	if itemObj["question"] == nil {
		return errors.New("no name field given")
	}
	if _, ok := itemObj["question"].(string); !ok {
		return errors.New("problem casting id to string")
	}
	if itemObj["type"] == nil {
		return errors.New("no type field given")
	}
	if _, ok := itemObj["type"].(string); !ok {
		return errors.New("problem casting name to string")
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
	if itemObj["text"] == nil {
		return errors.New("no text field given")
	}
	if _, ok := itemObj["text"].(string); !ok {
		return errors.New("problem casting text to string")
	}
	if itemObj["required"] == nil {
		return errors.New("no required field given")
	}
	if _, ok := itemObj["required"].(bool); !ok {
		return errors.New("problem casting required to boolean")
	}
	if itemObj["file"] == nil {
		return errors.New("no file field given")
	}
	if _, ok := itemObj["file"].(string); !ok {
		return errors.New("problem casting file to string")
	}
	return nil
}

func checkItemObjUpdate(itemObj map[string]interface{}) error {
	if itemObj["question"] != nil {
		if _, ok := itemObj["question"].(string); !ok {
			return errors.New("problem casting id to string")
		}
	}
	if itemObj["type"] != nil {
		if _, ok := itemObj["type"].(string); !ok {
			return errors.New("problem casting name to string")
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
	if itemObj["text"] != nil {
		_, ok := itemObj["text"].(string)
		if !ok {
			return errors.New("problem casting text to string")
		}
	}
	if itemObj["required"] != nil {
		if _, ok := itemObj["required"].(bool); !ok {
			return errors.New("problem casting required to boolean")
		}
	}
	if itemObj["file"] != nil {
		if _, ok := itemObj["file"].(string); !ok {
			return errors.New("problem casting file to string")
		}
	}
	return nil
}

func checkItemObjUpdatePart(itemObj map[string]interface{}) error {
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
	if itemObj["question"] != nil {
		if _, ok := itemObj["question"].(string); !ok {
			return errors.New("problem casting id to string")
		}
	}
	if itemObj["type"] != nil {
		if _, ok := itemObj["type"].(string); !ok {
			return errors.New("problem casting name to string")
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
	if itemObj["text"] != nil {
		_, ok := itemObj["text"].(string)
		if !ok {
			return errors.New("problem casting text to string")
		}
	}
	if itemObj["required"] != nil {
		if _, ok := itemObj["required"].(bool); !ok {
			return errors.New("problem casting required to boolean")
		}
	}
	if itemObj["file"] != nil {
		if _, ok := itemObj["file"].(string); !ok {
			return errors.New("problem casting file to string")
		}
	}
	return nil
}
