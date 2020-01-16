package main

import (
	"context"
	"regexp"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	json "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"github.com/vmihailenco/taskq/v2"
	"github.com/vmihailenco/taskq/v2/redisq"

	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var jwtSecret []byte

var sendgridAPIKey string

var websiteURL string

var apiURL string

var mongoClient *mongo.Client

var ctxMongo context.Context

var userCollection *mongo.Collection

var formCollection *mongo.Collection

var projectCollection *mongo.Collection

var blogCollection *mongo.Collection

var shortLinkCollection *mongo.Collection

var elasticClient *elastic.Client

var ctxElastic context.Context

var ctxStorage context.Context

var storageClient *storage.Client

var storageBucket *storage.BucketHandle

var storageAccessID string

var storagePrivateKey []byte

var logger *zap.Logger

var redisClient *redis.Client

var validHexcode *regexp.Regexp

var mainRecaptchaSecret string

var shortlinkRecaptchaSecret string

var shortlinkURL string

var serviceEmail string

var jwtIssuer string

var mode string

var schema graphql.Schema

var queueFactory taskq.Factory

var messageQueue taskq.Queue

var ctxMessageQueue context.Context

var saveFormTask *taskq.Task

var connections sync.Map

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

func graphqlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request
		// set auth
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), tokenKey, getAuthToken(c.Request)))
		c.Next()
		// after request
	}
}

func graphqlHandler() gin.HandlerFunc {
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

func setupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Info("closing gracefully...")
		if err := queueFactory.Close(); err != nil {
			logger.Error("problem closing message queue factory: " + err.Error())
		}
		connections.Range(func(key, value interface{}) bool {
			if err := value.(Connection).Conn.Close(); err != nil {
				logger.Error("problem closing redis pubsub connection: " + err.Error())
			}
			return true
		})
		if err := redisClient.Close(); err != nil {
			logger.Error("problem closing redis client: " + err.Error())
		}
		elasticClient.Stop()
		if err := storageClient.Close(); err != nil {
			logger.Error("error closing storage connection: " + err.Error())
		}
		if err := mongoClient.Disconnect(ctxMongo); err != nil {
			logger.Error("problem closing mongodb connection: " + err.Error())
		}
		os.Exit(0)
	}()
}

func main() {
	initAddRemoveAccessScript()
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
	projectCollection = mongoClient.Database(mainDatabase).Collection(projectMongoName)
	blogCollection = mongoClient.Database(mainDatabase).Collection(blogMongoName)
	shortLinkCollection = mongoClient.Database(mainDatabase).Collection(shortLinkMongoName)
	elasticuri := os.Getenv("ELASTICURI")
	elasticClient, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(elasticuri))
	if err != nil {
		logger.Fatal(err.Error())
	}
	ctxElastic = context.Background()
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
	storageBucket = storageClient.Bucket(storageBucketName)
	gcpprojectid, ok := storageconfigjson["project_id"].(string)
	if !ok {
		logger.Fatal("could not cast gcp project id to string")
	}
	if err := storageBucket.Create(ctxStorage, gcpprojectid, nil); err != nil {
		logger.Info("error creating storage bucket: " + err.Error())
	}
	storagePrivateKeyString, ok := storageconfigjson["private_key"].(string)
	if !ok {
		logger.Fatal("cannot cast private key to string")
	}
	storagePrivateKey = []byte(storagePrivateKeyString)
	storageAccessID, ok = storageconfigjson["client_email"].(string)
	if !ok {
		logger.Fatal("cannot cast storage access id to string")
	}
	redisAddress := os.Getenv("REDISADDRESS")
	redisPassword := os.Getenv("REDISPASSWORD")
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // use default DB
	})
	pong, err := redisClient.Ping().Result()
	if err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("connected to redis cache: " + pong)
	}
	queueFactory = redisq.NewFactory()
	messageQueue = queueFactory.RegisterQueue(&taskq.QueueOptions{
		Name:  "api-task-worker",
		Redis: redisClient,
	})
	ctxMessageQueue = context.Background()
	saveFormTask = taskq.RegisterTask(&taskq.TaskOptions{
		Name: "saveForm",
		Handler: func(formIDString string) error {
			return updateForm(formIDString)
		},
	})
	setupCloseHandler()
	// setup message queue worker
	err = queueFactory.StartConsumers(ctxMessageQueue)
	if err != nil {
		logger.Fatal("cannot start queue consumer: " + err.Error())
	}
	validHexcode, err = regexp.Compile(hexRegex)
	if err != nil {
		logger.Fatal(err.Error())
	}
	mainRecaptchaSecret = os.Getenv("MAINRECAPTCHASECRET")
	shortlinkRecaptchaSecret = os.Getenv("SHORTLINKRECAPTCHASECRET")
	shortlinkURL = os.Getenv("SHORTLINKURL")
	port := ":" + os.Getenv("PORT")
	schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query:        rootQuery(),
		Mutation:     rootMutation(),
		Subscription: rootSubscription(),
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
		graphqlGroup.GET("", graphqlHandler())
		graphqlGroup.POST("", graphqlHandler())
		graphqlGroup.PUT("", graphqlHandler())
		graphqlGroup.DELETE("", graphqlHandler())
	}
	router.GET("/subscriptions", subscriptionsHandler)
	router.POST("/sendTestEmail", sendTestEmail)
	router.PUT("/loginEmailPassword", loginEmailPassword)
	router.POST("/register", register)
	router.POST("/verify", verifyEmail)
	router.PUT("/sendResetEmail", sendPasswordResetEmail)
	router.POST("/reset", resetPassword)
	router.GET("/getFile", getFile)
	router.PUT("/writeFile", writeFile)
	router.DELETE("/deleteFiles", deleteFiles)
	router.GET("/countForms", countForms)
	router.GET("/countProjects", countProjects)
	router.GET("/countBlogs", countBlogs)
	router.Any("/shortLink", shortLinkRedirect)
	router.GET("/hello", hello)
	router.Run()
	logger.Info("Starting the application at " + port + " 🚀")
}
