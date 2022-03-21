module github.com/jschmidtnj/emailhacks/api

go 1.12

require (
	cloud.google.com/go/storage v1.5.0
	github.com/DataDog/zstd v1.4.4 // indirect
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/andybalholm/cascadia v1.1.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.0
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-redis/redis/v7 v7.0.0-beta.4
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/graphql-go/graphql v0.7.8
	github.com/graphql-go/handler v0.2.3
	github.com/ikeikeikeike/go-sitemap-generator/v2 v2.0.2
	github.com/joho/godotenv v1.3.0
	github.com/json-iterator/go v1.1.9
	github.com/mitchellh/mapstructure v1.1.2
	github.com/olivere/elastic/v7 v7.0.9
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/rs/xid v1.2.1
	github.com/sendgrid/rest v2.4.1+incompatible
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/stripe/stripe-go v68.11.0+incompatible
	github.com/tdewolff/minify/v2 v2.7.2
	github.com/tidwall/pretty v1.0.0 // indirect
	github.com/unrolled/secure v1.0.7
	github.com/valyala/fasthttp v1.34.0
	github.com/vmihailenco/taskq/v2 v2.2.3
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.2.0
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	google.golang.org/api v0.15.0
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
