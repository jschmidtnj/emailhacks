package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// UpdateFormFormItemType response for item update
var UpdateFormFormItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateFormItem",
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

// UpdateFormItemInputType - type of graphql input
var UpdateFormItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateFormItemInput",
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

// FormItem form struct
type FormItem struct {
	Question string   `json:"question"`
	Type     string   `json:"type"`
	Options  []string `json:"options"`
	Text     string   `json:"text"`
	Required bool     `json:"required"`
	Files    []int    `json:"files"`
}

// FormItemType graphql question object
var FormItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FormItem",
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

// FormItemInputType - type of graphql input
var FormItemInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "FormItemInput",
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

func checkFormItemObjCreate(itemObj map[string]interface{}) error {
	if itemObj["question"] == nil {
		return errors.New("no question field given")
	}
	if _, ok := itemObj["question"].(string); !ok {
		return errors.New("problem casting question to string")
	}
	if itemObj["type"] == nil {
		return errors.New("no type field given")
	}
	itemType, ok := itemObj["type"].(string)
	if !ok {
		return errors.New("problem casting type to string")
	}
	if !findInArray(itemType, validFormItemTypes) {
		return errors.New("invalid type for form item found")
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
	if itemObj["files"] == nil {
		return errors.New("no files field given")
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

func checkFormItemObjUpdate(itemObj map[string]interface{}) error {
	if itemObj["question"] != nil {
		if _, ok := itemObj["question"].(string); !ok {
			return errors.New("problem casting question to string")
		}
	}
	if itemObj["type"] != nil {
		itemType, ok := itemObj["type"].(string)
		if !ok {
			return errors.New("problem casting type to string")
		}
		if !findInArray(itemType, validFormItemTypes) {
			return errors.New("invalid type for form item found")
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

func checkFormItemObjUpdatePart(itemObj map[string]interface{}) error {
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
		itemType, ok := itemObj["type"].(string)
		if !ok {
			return errors.New("problem casting type to string")
		}
		if !findInArray(itemType, validFormItemTypes) {
			return errors.New("invalid type for form item found")
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
