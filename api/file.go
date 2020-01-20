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
	if fileobj["index"] == nil {
		return errors.New("no index given")
	}
	if _, ok := fileobj["index"].(int); !ok {
		return errors.New("cannot cast index to int")
	}
	if fileobj["fileIndex"] == nil {
		return errors.New("no file index given")
	}
	if _, ok := fileobj["fileIndex"].(int); !ok {
		return errors.New("cannot cast file index to int")
	}
	if fileobj["itemIndex"] == nil {
		return errors.New("no item index given")
	}
	if _, ok := fileobj["itemIndex"].(int); !ok {
		return errors.New("cannot cast item index to int")
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

// UpdateFileType graphql file object
var UpdateFileType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateFile",
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
		"updateAction": &graphql.Field{
			Type: graphql.String,
		},
		"fileIndex": &graphql.Field{
			Type: graphql.Int,
		},
		"itemIndex": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

// UpdateFileInputType - type of graphql input
var UpdateFileInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "UpdateFileInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"updateAction": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"index": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"fileIndex": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
			},
			"itemIndex": &graphql.InputObjectFieldConfig{
				Type: graphql.Int,
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

// File object for giving user access
type File struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Width          int64  `json:"width"`
	Height         int64  `json:"height"`
	Type           string `json:"type"`
	OriginalSrc    string `json:"originalSrc"`
	BlurSrc        string `json:"blurSrc"`
	PlaceholderSrc string `json:"placeholderSrc"`
}

// FileType graphql image object
var FileType = graphql.NewObject(graphql.ObjectConfig{
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
		"originalSrc": &graphql.Field{
			Type: graphql.String,
		},
		"blurSrc": &graphql.Field{
			Type: graphql.String,
		},
		"placeholderSrc": &graphql.Field{
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
