package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/graphql-go/graphql"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	// medium "github.com/medium/medium-sdk-go"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var jwtSecret []byte

var tokenExpiration int

var sendgridAPIKey string

var websiteURL string

var apiURL string

var mongoClient *mongo.Client

var ctxMongo context.Context

var mainDatabase = "website"

var userCollection *mongo.Collection

var userMongoName = "users"

var ctxStorage context.Context

var storageClient *storage.Client

var storageBucket *storage.BucketHandle

var logger *zap.Logger

type tokenKeyType string

var tokenKey tokenKeyType

var redisClient *redis.Client

var cacheTime time.Duration

var mainRecaptchaSecret string

var serviceEmail string

var jwtIssuer string

var mode string

/**
 * @api {get} /hello Test rest request
 * @apiVersion 0.0.1
 * @apiSuccess {String} message Hello message
 * @apiGroup misc
 */
func hello(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(`{"message":"Hello!"}`))
}

func main() {
	// "./logs"
	loggerconfig := []byte(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"initialFields": {},
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
  }`)
	var zapconfig zap.Config
	if err := json.Unmarshal(loggerconfig, &zapconfig); err != nil {
		panic(err)
	}
	var err error
	logger, err = zapconfig.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("logger created")
	err = godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}
	jwtSecret = []byte(os.Getenv("SECRET"))
	tokenExpiration, err = strconv.Atoi(os.Getenv("TOKENEXPIRATION"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	sendgridAPIKey = os.Getenv("SENDGRIDAPIKEY")
	serviceEmail = os.Getenv("SERVICEEMAIL")
	jwtIssuer = os.Getenv("JWTISSUER")
	mode = os.Getenv("MODE")
	websiteURL = os.Getenv("WEBSITEURL")
	apiURL = os.Getenv("APIURL")
	ctxMongo, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cancel()
	mongouri := os.Getenv("MONGOURI")
	mongoClient, err = mongo.Connect(ctxMongo, options.Client().ApplyURI(mongouri))
	if err != nil {
		logger.Fatal(err.Error())
	}
	userCollection = mongoClient.Database(mainDatabase).Collection(userMongoName)
	var storageconfigstr = os.Getenv("STORAGECONFIG")
	var storageconfigjson map[string]interface{}
	json.Unmarshal([]byte(storageconfigstr), &storageconfigjson)
	ctxStorage = context.Background()
	storageconfigjsonbytes, err := json.Marshal(storageconfigjson)
	if err != nil {
		logger.Fatal(err.Error())
	}
	storageClient, err = storage.NewClient(ctxStorage, option.WithCredentialsJSON([]byte(storageconfigjsonbytes)))
	if err != nil {
		logger.Fatal(err.Error())
	}
	bucketName := os.Getenv("STORAGEBUCKETNAME")
	storageBucket = storageClient.Bucket(bucketName)
	gcpprojectid, ok := storageconfigjson["project_id"].(string)
	if !ok {
		logger.Fatal("could not cast gcp project id to string")
	}
	if err := storageBucket.Create(ctxStorage, gcpprojectid, nil); err != nil {
		logger.Info(err.Error())
	}
	cacheSeconds, err := strconv.Atoi(os.Getenv("CACHETIME"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	cacheTime = time.Duration(cacheSeconds) * time.Second
	pong, err := redisClient.Ping().Result()
	if err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("connected to redis cache: " + pong)
	}
	mainRecaptchaSecret = os.Getenv("MAINRECAPTCHASECRET")
	port := ":" + os.Getenv("PORT")
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery(),
		Mutation: rootMutation(),
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(response http.ResponseWriter, request *http.Request) {
		tokenKey = tokenKeyType("token")
		var query = ""
		queryString := request.URL.Query().Get("query")
		if queryString != "" {
			query = queryString
		} else if request.Method == http.MethodPost || request.Method == http.MethodPut {
			logger.Info("got put or post")
			var querydata map[string]interface{}
			err = nil
			querybody, err := ioutil.ReadAll(request.Body)
			if err == nil {
				err = json.Unmarshal(querybody, &querydata)
				if err == nil {
					queryfromjson, ok := querydata["query"].(string)
					if ok {
						query = queryfromjson
					}
				}
			}
		}
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: query,
			Context:       context.WithValue(context.Background(), tokenKey, getAuthToken(request)),
		})
		response.Header().Set("Content-Type", "application/json")
		json.NewEncoder(response).Encode(result)
	})
	mux.HandleFunc("/sendTestEmail", sendTestEmail)
	mux.HandleFunc("/loginEmailPassword", loginEmailPassword)
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/verify", verifyEmail)
	mux.HandleFunc("/sendResetEmail", sendPasswordResetEmail)
	mux.HandleFunc("/reset", resetPassword)
	mux.HandleFunc("/hello", hello)
	var allowedOrigins []string
	if mode == "debug" {
		allowedOrigins = []string{
			"*",
		}
	} else {
		allowedOrigins = []string{
			websiteURL,
		}
	}
	thecors := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		OptionsPassthrough: false,
		Debug:              mode == "debug",
	})
	handler := thecors.Handler(mux)
	http.ListenAndServe(port, handler)
	logger.Info("Starting the application at " + port + " 🚀")
}

func getAuthToken(request *http.Request) string {
	authToken := request.Header.Get("Authorization")
	splitToken := strings.Split(authToken, "Bearer ")
	if splitToken != nil && len(splitToken) > 1 {
		authToken = splitToken[1]
	}
	return authToken
}
