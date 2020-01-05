package main

import (
	"errors"

	"github.com/graphql-go/graphql"
)

func checkFileObjCreate(fileobj map[string]interface{}) error {
	if fileobj["id"] == nil {
		return errors.New("no file id given")
	}
	if _, ok := fileobj["id"].(string); !ok {
		return errors.New("problem casting id to string")
	}
	if fileobj["name"] == nil {
		return errors.New("no file name given")
	}
	if _, ok := fileobj["name"].(string); !ok {
		return errors.New("problem casting name to string")
	}
	if fileobj["width"] == nil {
		return errors.New("no file width given")
	}
	if _, ok := fileobj["width"].(int); !ok {
		return errors.New("problem casting width to int")
	}
	if fileobj["height"] == nil {
		return errors.New("no file height given")
	}
	if _, ok := fileobj["height"].(int); !ok {
		return errors.New("problem casting height to int")
	}
	if fileobj["type"] == nil {
		return errors.New("no file type given")
	}
	if _, ok := fileobj["type"].(string); !ok {
		return errors.New("problem casting type to string")
	}
	return nil
}

func checkFileObjUpdate(fileobj map[string]interface{}) error {
	if fileobj["id"] != nil {
		if _, ok := fileobj["id"].(string); !ok {
			return errors.New("problem casting id to string")
		}
	}
	if fileobj["name"] != nil {
		if _, ok := fileobj["name"].(string); !ok {
			return errors.New("problem casting name to string")
		}
	}
	if fileobj["width"] != nil {
		if _, ok := fileobj["width"].(int); !ok {
			return errors.New("problem casting width to int")
		}
	}
	if fileobj["height"] != nil {
		if _, ok := fileobj["height"].(int); !ok {
			return errors.New("problem casting height to int")
		}
	}
	if fileobj["type"] != nil {
		if _, ok := fileobj["type"].(string); !ok {
			return errors.New("problem casting type to string")
		}
	}
	return nil
}

func checkFileObjUpdatePart(fileobj map[string]interface{}) error {
	if fileobj["id"] == nil {
		return errors.New("no file id given")
	}
	if _, ok := fileobj["id"].(string); !ok {
		return errors.New("problem casting id to string")
	}
	if fileobj["updateAction"] == nil {
		return errors.New("no update action given")
	}
	action, ok := fileobj["updateAction"].(string)
	if !ok {
		return errors.New("update action cannot be cast to string")
	}
	if !findInArray(action, validUpdateMapActions) {
		return errors.New("invalid action given")
	}
	if fileobj["name"] != nil {
		if _, ok := fileobj["name"].(string); !ok {
			return errors.New("problem casting name to string")
		}
	}
	if fileobj["width"] != nil {
		if _, ok := fileobj["width"].(int); !ok {
			return errors.New("problem casting width to int")
		}
	}
	if fileobj["height"] != nil {
		if _, ok := fileobj["height"].(int); !ok {
			return errors.New("problem casting height to int")
		}
	}
	if fileobj["type"] != nil {
		if _, ok := fileobj["type"].(string); !ok {
			return errors.New("problem casting type to string")
		}
	}
	return nil
}

// FileType graphql image object
var FileType *graphql.Object = graphql.NewObject(graphql.ObjectConfig{
	Name: "File",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"width": &graphql.Field{
			Type: graphql.Int,
		},
		"height": &graphql.Field{
			Type: graphql.Int,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// FileInputType - type of graphql input
var FileInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "FileInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"updateAction": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"name": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"height": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"width": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"type": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	},
)
