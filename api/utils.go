package main

import (
	"encoding/binary"
	"errors"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func stringListToInterfaceList(stringList []string) []interface{} {
	result := make([]interface{}, len(stringList))
	for i, item := range stringList {
		result[i] = item
	}
	return result
}

func interfaceListToStringList(interfaceList []interface{}) ([]string, error) {
	result := make([]string, len(interfaceList))
	for i, item := range interfaceList {
		itemStr, ok := item.(string)
		if !ok {
			return nil, errors.New("item in list cannot be cast to string")
		}
		result[i] = itemStr
	}
	return result, nil
}

func interfaceListToMapList(interfaceList []interface{}) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(interfaceList))
	for i, item := range interfaceList {
		itemObj, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("item in list cannot be converted to map")
		}
		result[i] = itemObj
	}
	return result, nil
}

func handleError(message string, statuscode int, response http.ResponseWriter) {
	// logger.Error(message)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statuscode)
	response.Write([]byte(`{"message":"` + message + `"}`))
}

func objectidTimestamp(id primitive.ObjectID) time.Time {
	unixSecs := binary.BigEndian.Uint32(id[0:4])
	return time.Unix(int64(unixSecs), 0).UTC()
}

func intTimestamp(data int64) time.Time {
	return time.Unix(data, 0).UTC()
}

func findInArray(thetype string, arr []string) bool {
	for _, b := range arr {
		if b == thetype {
			return true
		}
	}
	return false
}

func getAuthToken(request *http.Request) string {
	authToken := request.Header.Get("Authorization")
	splitToken := strings.Split(authToken, "Bearer ")
	if splitToken != nil && len(splitToken) > 1 {
		authToken = splitToken[1]
	}
	return authToken
}

func isDebug() bool {
	return mode == "debug"
}
