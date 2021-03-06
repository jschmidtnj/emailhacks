package main

import (
	"errors"
	"net/url"
	"time"

	"github.com/graphql-go/graphql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var blogMutationFields = graphql.Fields{
	"addBlog": &graphql.Field{
		Type:        BlogType,
		Description: "Create a Blog",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"title": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"caption": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"content": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"author": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"color": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"tags": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"heroimage": &graphql.ArgumentConfig{
				Type: FileInputType,
			},
			"tileimage": &graphql.ArgumentConfig{
				Type: FileInputType,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			_, err := validateAdmin(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("id not provided")
			}
			if params.Args["title"] == nil {
				return nil, errors.New("title not provided")
			}
			if params.Args["author"] == nil {
				return nil, errors.New("author not provided")
			}
			if params.Args["heroimage"] == nil {
				return nil, errors.New("heroimage not provided")
			}
			if params.Args["content"] == nil {
				return nil, errors.New("content not provided")
			}
			if params.Args["files"] == nil {
				return nil, errors.New("files not provided")
			}
			if params.Args["caption"] == nil {
				return nil, errors.New("caption not provided")
			}
			if params.Args["color"] == nil {
				return nil, errors.New("color not provided")
			}
			if params.Args["tags"] == nil {
				return nil, errors.New("tags not provided")
			}
			if params.Args["categories"] == nil {
				return nil, errors.New("categories not provided")
			}
			if params.Args["tileimage"] == nil {
				return nil, errors.New("tileimage not provided")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idstring)
			if err != nil {
				return nil, err
			}
			title, ok := params.Args["title"].(string)
			if !ok {
				return nil, errors.New("problem casting title to string")
			}
			caption, ok := params.Args["caption"].(string)
			if !ok {
				return nil, errors.New("problem casting caption to string")
			}
			author, ok := params.Args["author"].(string)
			if !ok {
				return nil, errors.New("problem casting author to string")
			}
			content, ok := params.Args["content"].(string)
			if !ok {
				return nil, errors.New("problem casting content to string")
			}
			color, ok := params.Args["color"].(string)
			if !ok {
				return nil, errors.New("problem casting color to string")
			}
			decodedColor, err := url.QueryUnescape(color)
			if err != nil {
				return nil, err
			}
			if !validHexcode.MatchString(decodedColor) {
				return nil, errors.New("invalid hex code for color")
			}
			tagsInterface, ok := params.Args["tags"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting tags to interface array")
			}
			tags, err := interfaceListToStringList(tagsInterface)
			if err != nil {
				return nil, err
			}
			categoriesInterface, ok := params.Args["categories"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting categories to interface array")
			}
			categories, err := interfaceListToStringList(categoriesInterface)
			if err != nil {
				return nil, err
			}
			heroimage, ok := params.Args["heroimage"].(map[string]interface{})
			if !ok {
				return nil, errors.New("problem casting heroimage to map")
			}
			if err := checkFileObjCreate(heroimage); err != nil {
				heroimage = nil
			}
			tileimage, ok := params.Args["tileimage"].(map[string]interface{})
			if !ok {
				return nil, errors.New("problem casting tileimage to map")
			}
			if err := checkFileObjCreate(tileimage); err != nil {
				return nil, err
			}
			filesinterface, ok := params.Args["files"].([]interface{})
			if !ok {
				return nil, errors.New("problem casting files to interface array")
			}
			files, err := interfaceListToMapList(filesinterface)
			if err != nil {
				return nil, err
			}
			for _, file := range files {
				if err := checkFileObjCreate(file); err != nil {
					return nil, err
				}
			}
			shortlink, err := generateShortLink(websiteURL + "/blog/" + idstring)
			if err != nil {
				return nil, err
			}
			created := objectidTimestamp(id)
			blogData := bson.M{
				"_id":        id,
				"title":      title,
				"caption":    caption,
				"content":    content,
				"author":     author,
				"color":      color,
				"updated":    created.Unix(),
				"tags":       tags,
				"categories": categories,
				"views":      0,
				"heroimage":  heroimage,
				"tileimage":  tileimage,
				"files":      files,
				"shortlink":  shortlink,
			}
			_, err = blogCollection.InsertOne(ctxMongo, blogData)
			if err != nil {
				return nil, err
			}
			blogData["created"] = blogData["updated"]
			delete(blogData, "_id")
			_, err = elasticClient.Index().
				Index(blogElasticIndex).
				Type(blogElasticType).
				Id(idstring).
				BodyJson(blogData).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			blogData["id"] = idstring
			return blogData, nil
		},
	},
	"updateBlog": &graphql.Field{
		Type:        BlogType,
		Description: "Update a Blog",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"title": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"caption": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"content": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"author": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"color": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"tags": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"categories": &graphql.ArgumentConfig{
				Type: graphql.NewList(graphql.String),
			},
			"heroimage": &graphql.ArgumentConfig{
				Type: FileInputType,
			},
			"tileimage": &graphql.ArgumentConfig{
				Type: FileInputType,
			},
			"files": &graphql.ArgumentConfig{
				Type: graphql.NewList(FileInputType),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			_, err := validateAdmin(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("blog id not provided")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idstring)
			if err != nil {
				return nil, err
			}
			updateData := bson.M{}
			if params.Args["title"] != nil {
				title, ok := params.Args["title"].(string)
				if !ok {
					return nil, errors.New("problem casting title to string")
				}
				updateData["title"] = title
			}
			if params.Args["caption"] != nil {
				caption, ok := params.Args["caption"].(string)
				if !ok {
					return nil, errors.New("problem casting caption to string")
				}
				updateData["caption"] = caption
			}
			if params.Args["author"] != nil {
				author, ok := params.Args["author"].(string)
				if !ok {
					return nil, errors.New("problem casting author to string")
				}
				updateData["author"] = author
			}
			if params.Args["content"] != nil {
				content, ok := params.Args["content"].(string)
				if !ok {
					return nil, errors.New("problem casting content to string")
				}
				updateData["content"] = content
			}
			if params.Args["color"] != nil {
				color, ok := params.Args["color"].(string)
				if !ok {
					return nil, errors.New("problem casting color to string")
				}
				decodedColor, err := url.QueryUnescape(color)
				if err != nil {
					return nil, err
				}
				if !validHexcode.MatchString(decodedColor) {
					return nil, errors.New("invalid hex code for color")
				}
				updateData["color"] = color
			}
			if params.Args["tags"] != nil {
				tagsInterface, ok := params.Args["tags"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting tags to interface array")
				}
				tags, err := interfaceListToStringList(tagsInterface)
				if err != nil {
					return nil, err
				}
				updateData["tags"] = tags
			}
			if params.Args["categories"] != nil {
				categoriesInterface, ok := params.Args["categories"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting categories to interface array")
				}
				categories, err := interfaceListToStringList(categoriesInterface)
				if err != nil {
					return nil, err
				}
				updateData["categories"] = categories
			}
			if params.Args["heroimage"] != nil {
				heroimage, ok := params.Args["heroimage"].(map[string]interface{})
				if !ok {
					return nil, errors.New("problem casting heroimage to map")
				}
				if len(heroimage) > 0 {
					if err := checkFileObjUpdate(heroimage); err != nil {
						return nil, err
					}
					updateData["heroimage"] = heroimage
				}
			}
			if params.Args["tileimage"] != nil {
				tileimage, ok := params.Args["tileimage"].(map[string]interface{})
				if !ok {
					return nil, errors.New("problem casting tileimage to map")
				}
				if len(tileimage) > 0 {
					if err := checkFileObjUpdate(tileimage); err != nil {
						return nil, err
					}
					updateData["tileimage"] = tileimage
				}
			}
			if params.Args["files"] != nil {
				filesinterface, ok := params.Args["files"].([]interface{})
				if !ok {
					return nil, errors.New("problem casting files to interface array")
				}
				files, err := interfaceListToMapList(filesinterface)
				if err != nil {
					return nil, err
				}
				for _, file := range files {
					if err := checkFileObjUpdate(file); err != nil {
						return nil, err
					}
				}
				updateData["files"] = files
			}
			updateData["updated"] = time.Now().Unix()
			_, err = elasticClient.Update().
				Index(blogElasticIndex).
				Type(blogElasticType).
				Id(idstring).
				Doc(updateData).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			_, err = blogCollection.UpdateOne(ctxMongo, bson.M{
				"_id": id,
			}, bson.M{
				"$set": updateData,
			})
			if err != nil {
				return nil, err
			}
			blogData, err := getBlog(id, true)
			if err != nil {
				return nil, err
			}
			return blogData, nil
		},
	},
	"deleteBlog": &graphql.Field{
		Type:        BlogType,
		Description: "Delete a Blog",
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			_, err := validateAdmin(params.Context.Value(tokenKey).(string))
			if err != nil {
				return nil, err
			}
			if params.Args["id"] == nil {
				return nil, errors.New("blog id not provided")
			}
			idstring, ok := params.Args["id"].(string)
			if !ok {
				return nil, errors.New("cannot cast id to string")
			}
			id, err := primitive.ObjectIDFromHex(idstring)
			if err != nil {
				return nil, err
			}
			_, err = elasticClient.Delete().
				Index(blogElasticIndex).
				Type(blogElasticType).
				Id(idstring).
				Do(ctxElastic)
			if err != nil {
				return nil, err
			}
			blogData, err := getBlog(id, true)
			if err != nil {
				return nil, err
			}
			_, err = blogCollection.DeleteOne(ctxMongo, bson.M{
				"_id": id,
			})
			if err != nil {
				return nil, err
			}
			err = deleteShortLink(blogData.ShortLink)
			if err != nil {
				return nil, err
			}
			if blogData.HeroImage != nil {
				heroobjblur := storageBucket.Object(blogFileIndex + "/" + idstring + "/" + blogData.HeroImage.ID + blurPath)
				heroobjoriginal := storageBucket.Object(blogFileIndex + "/" + idstring + "/" + blogData.HeroImage.ID + originalPath)
				if err := heroobjblur.Delete(ctxStorage); err != nil {
					return nil, err
				}
				if err := heroobjoriginal.Delete(ctxStorage); err != nil {
					return nil, err
				}
			}
			if blogData.TileImage != nil {
				tileobjblur := storageBucket.Object(blogFileIndex + "/" + idstring + "/" + blogData.TileImage.ID + blurPath)
				tileobjoriginal := storageBucket.Object(blogFileIndex + "/" + idstring + "/" + blogData.TileImage.ID + originalPath)
				if err := tileobjblur.Delete(ctxStorage); err != nil {
					return nil, err
				}
				if err := tileobjoriginal.Delete(ctxStorage); err != nil {
					return nil, err
				}
			}
			for _, file := range blogData.Files {
				_, err = deleteFile(blogType, blogData.ID, file.ID)
				if err != nil {
					return nil, err
				}
			}
			if err != nil {
				return nil, err
			}
			return blogData, nil
		},
	},
}
