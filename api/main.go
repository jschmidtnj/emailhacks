package main

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	json "github.com/json-iterator/go"

	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

var formCollection *mongo.Collection

var formMongoName = "forms"

var ctxStorage context.Context

var storageClient *storage.Client

var storageBucket *storage.BucketHandle

var logger *zap.Logger

var tokenKey = "token"

var cacheTime time.Duration

var mainRecaptchaSecret string

var serviceEmail string

var jwtIssuer string

var mode string

var graphiQL = false

var graphqlPlayground = true

/**
 * @api {get} /hello Test rest request
 * @apiVersion 0.0.1
 * @apiSuccess {String} message Hello message
 * @apiGroup misc
 */
func hello(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write([]byte(`{"message":"Hello!"}`))
}

func getAuthToken(request *http.Request) string {
	authToken := request.Header.Get("Authorization")
	splitToken := strings.Split(authToken, "Bearer ")
	if splitToken != nil && len(splitToken) > 1 {
		authToken = splitToken[1]
	} else {
		return ""
	}
	return authToken
}

func graphqlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request
		// set auth
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), tokenKey, getAuthToken(c.Request)))
		c.Next()
		// after request
	}
}

func graphqlHandler(schema graphql.Schema) gin.HandlerFunc {
	handler := handler.New(
		&handler.Config{
			Schema:     &schema,
			Pretty:     isDebug(),
			GraphiQL:   graphiQL,
			Playground: graphqlPlayground,
		})
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func isDebug() bool {
	return mode == "debug"
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
	formCollection = mongoClient.Database(mainDatabase).Collection(formMongoName)
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
	mainRecaptchaSecret = os.Getenv("MAINRECAPTCHASECRET")
	port := ":" + os.Getenv("PORT")
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery(),
		Mutation: rootMutation(),
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	if isDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	if isDebug() {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{
			websiteURL,
		}
	}
	corsConfig.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
	}
	corsConfig.AllowHeaders = []string{
		"Authorization",
		"Content-Type",
	}
	router.Use(cors.New(corsConfig))
	graphqlGroup := router.Group("/graphql")
	graphqlGroup.Use(graphqlMiddleware())
	{
		graphqlGroup.GET("", graphqlHandler(schema))
		graphqlGroup.POST("", graphqlHandler(schema))
		graphqlGroup.PUT("", graphqlHandler(schema))
		graphqlGroup.DELETE("", graphqlHandler(schema))
	}
	router.POST("/sendTestEmail", sendTestEmail)
	router.PUT("/loginEmailPassword", loginEmailPassword)
	router.POST("/register", register)
	router.POST("/verify", verifyEmail)
	router.PUT("/sendResetEmail", sendPasswordResetEmail)
	router.POST("/reset", resetPassword)
	router.GET("/hello", hello)
	router.Run()
	logger.Info("Starting the application at " + port + " 🚀")
}
