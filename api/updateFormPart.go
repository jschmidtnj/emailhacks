package main

import (
	"errors"

	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func updateForm(formIDString string) error {
	// get update data from redis, then update elastic and mongodb from it
	savedUpdateData, err := redisClient.Get(updateFormPath + formIDString).Result()
	if err != nil {
		// no update data found
		return nil
	}
	var savedUpdateDataObj map[string]interface{}
	err = json.UnmarshalFromString(savedUpdateData, &savedUpdateDataObj)
	if err != nil {
		return err
	}
	formID, err := primitive.ObjectIDFromHex(formIDString)
	if err != nil {
		return err
	}
	formData, err := getForm(formID, false, false)
	if err != nil {
		return err
	}
	updateDataDB := bson.M{
		"$set": bson.M{},
	}
	updateDataElastic := bson.M{}
	if savedUpdateDataObj["name"] != nil {
		name, ok := savedUpdateDataObj["name"].(string)
		if !ok {
			return errors.New("problem casting name to string")
		}
		updateDataDB["$set"].(bson.M)["name"] = name
		updateDataElastic["name"] = name
	}
	if savedUpdateDataObj["multiple"] != nil {
		multiple, ok := savedUpdateDataObj["multiple"].(bool)
		if !ok {
			return errors.New("problem casting multple to bool")
		}
		updateDataDB["$set"].(bson.M)["multiple"] = multiple
		updateDataElastic["multiple"] = multiple
	}
	if savedUpdateDataObj["items"] != nil {
		itemsUpdateInterface, ok := savedUpdateDataObj["items"].([]interface{})
		if !ok {
			return errors.New("problem casting items to interface array")
		}
		itemsUpdate, err := interfaceListToMapList(itemsUpdateInterface)
		if err != nil {
			return err
		}
		items, ok := formData["items"].(primitive.A)
		if !ok {
			return errors.New("cannot cast items to array")
		}
		for _, itemUpdate := range itemsUpdate {
			action := itemUpdate["updateAction"].(string)
			delete(itemUpdate, "updateAction")
			if action == validUpdateArrayActions[0] {
				// add
				items = append(items, itemUpdate)
			} else {
				index := int(itemUpdate["index"].(float64))
				delete(itemUpdate, "index")
				if action == validUpdateArrayActions[1] {
					// remove
					items = append(items[:index], items[index+1:]...)
				} else if action == validUpdateArrayActions[2] {
					// move to new index
					newIndex := int(itemUpdate["newIndex"].(float64))
					delete(itemUpdate, "newIndex")
					temp := items[newIndex]
					items[newIndex] = items[index]
					items[index] = temp
				} else if action == validUpdateArrayActions[3] {
					// set index to value
					items[index] = itemUpdate
				}
			}
		}
		updateDataDB["$set"].(bson.M)["items"] = items
		updateDataElastic["items"] = items
	}
	if savedUpdateDataObj["public"] != nil {
		public, ok := savedUpdateDataObj["public"].(string)
		if !ok {
			return errors.New("problem casting public to string")
		}
		updateDataDB["$set"].(bson.M)["public"] = public
		updateDataElastic["public"] = public
	}
	if savedUpdateDataObj["files"] != nil {
		filesUpdateInterface, ok := savedUpdateDataObj["files"].([]interface{})
		if !ok {
			return errors.New("problem casting files to interface array")
		}
		filesUpdate, err := interfaceListToMapList(filesUpdateInterface)
		if err != nil {
			return err
		}
		files, ok := formData["files"].(primitive.A)
		if !ok {
			return errors.New("cannot cast files to array")
		}
		for _, fileUpdate := range filesUpdate {
			action := fileUpdate["updateAction"].(string)
			delete(fileUpdate, "updateAction")
			if action == validUpdateMapActions[0] {
				// add
				files = append(files, fileUpdate)
			} else {
				var index = -1
				for i, fileItem := range files {
					if fileItem.(primitive.M)["id"].(string) == fileUpdate["id"].(string) {
						index = i
						break
					}
				}
				if index < 0 {
					continue
				}
				if action == validUpdateMapActions[1] {
					// remove
					files = append(files[:index], files[index+1:]...)
				} else if action == validUpdateMapActions[2] {
					// set to value
					files[index] = fileUpdate
				}
			}
		}
		updateDataDB["$set"].(bson.M)["files"] = files
		updateDataElastic["files"] = files
	}
	_, err = elasticClient.Update().
		Index(formElasticIndex).
		Type(formElasticType).
		Id(formIDString).
		Doc(updateDataElastic).
		Do(ctxElastic)
	if err != nil {
		return err
	}
	_, err = formCollection.UpdateOne(ctxMongo, bson.M{
		"_id": formID,
	}, updateDataDB)
	if err != nil {
		return err
	}
	err = redisClient.Del(updateFormPath + formIDString).Err()
	if err != nil {
		logger.Error(err.Error())
	}
	return nil
}
