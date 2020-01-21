package main

import (
	"errors"
	"time"

	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
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
	formData, err := getForm(formID, false)
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
		for _, itemUpdate := range itemsUpdate {
			action := itemUpdate["updateAction"].(string)
			delete(itemUpdate, "updateAction")
			var itemObj *FormItem
			if err = mapstructure.Decode(itemUpdate, &itemObj); err != nil {
				return err
			}
			if action == validUpdateArrayActions[0] {
				// add
				delete(itemUpdate, "index")
				formData.Items = append(formData.Items, itemObj)
			} else {
				index := int(itemUpdate["index"].(float64))
				delete(itemUpdate, "index")
				if index >= len(formData.Items) || index < 0 {
					continue
				}
				if action == validUpdateArrayActions[1] {
					// remove
					formData.Items = append(formData.Items[:index], formData.Items[index+1:]...)
				} else if action == validUpdateArrayActions[2] {
					// move to new index
					newIndex := int(itemUpdate["newIndex"].(float64))
					delete(itemUpdate, "newIndex")
					err = moveSliceFormItems(formData.Items, index, newIndex)
					if err != nil {
						return err
					}
				} else if action == validUpdateArrayActions[3] {
					// set index to value
					formData.Items[index] = itemObj
				}
			}
		}
		updateDataDB["$set"].(bson.M)["items"] = formData.Items
		updateDataElastic["items"] = formData.Items
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
		for _, fileUpdate := range filesUpdate {
			index := int(fileUpdate["index"].(float64))
			delete(fileUpdate, "index")
			delete(fileUpdate, "fileIndex")
			delete(fileUpdate, "itemIndex")
			action := fileUpdate["updateAction"].(string)
			delete(fileUpdate, "updateAction")
			var fileObj *File
			if err = mapstructure.Decode(fileUpdate, &fileObj); err != nil {
				return err
			}
			if action == validUpdateMapActions[0] {
				// add
				formData.Files = append(formData.Files, fileObj)
			} else {
				if action == validUpdateMapActions[1] {
					// remove
					if index >= 0 && index < len(formData.Files) {
						formData.Files = append(formData.Files[:index], formData.Files[index+1:]...)
					}
				} else if action == validUpdateMapActions[2] {
					// set to value
					formData.Files[index] = fileObj
				}
			}
		}
		fileData := make([]*FileDB, len(formData.Files))
		for i := range formData.Files {
			if err = mapstructure.Decode(formData.Files[i], &fileData[i]); err != nil {
				return err
			}
		}
		updateDataDB["$set"].(bson.M)["files"] = fileData
		updateDataElastic["files"] = fileData
	}
	if formData.Responses > 0 {
		// delete all previous responses (if there are any)
		bytesRemoved, err := deleteAllResponses(formID)
		if err != nil {
			return err
		}
		ownerID, err := primitive.ObjectIDFromHex(formData.Owner)
		if err != nil {
			return err
		}
		if err = changeUserStorage(ownerID, -1*bytesRemoved); err != nil {
			return err
		}
		formData.Responses = 0
		updateDataElastic["responses"] = int64(0)
		updateDataDB["$set"].(bson.M)["responses"] = int64(0)
	}
	updateDataElastic["updated"] = time.Now().Unix()
	delete(updateDataElastic, "created")
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
