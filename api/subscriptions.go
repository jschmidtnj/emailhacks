package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
)

type ConnectionACKMessage struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query string `json:"query"`
	} `json:"payload,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"graphql-ws"},
}

type Subscriber struct {
	FormID        string
	Conn          *websocket.Conn
	RequestString string
	OperationID   string
}

type Connection struct {
	Conn        *redis.PubSub
	Subscribers map[string]Subscriber
}

var connections sync.Map

func rootSubscription() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Subscription",
		Fields: graphql.Fields{
			"formUpdates": collaborationFields["formUpdates"],
		},
	})
}

func subscriptionsHandler(c *gin.Context) {
	response := c.Writer
	request := c.Request
	conn, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		message := "failed to do websocket upgrade: " + err.Error()
		logger.Error(message)
		handleError(message, http.StatusBadRequest, response)
		return
	}
	connectionACK, err := json.Marshal(map[string]string{
		"type": "connection_ack",
	})
	if err != nil {
		message := "failed to marshal ws connection ack: " + err.Error()
		logger.Error(message)
	}
	if err := conn.WriteMessage(websocket.TextMessage, connectionACK); err != nil {
		message := "failed to write to ws connection: " + err.Error()
		logger.Error(message)
		return
	}
	go func() {
		for {
			_, p, err := conn.ReadMessage()
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				return
			}
			if err != nil {
				message := "failed to read websocket message: " + err.Error()
				logger.Error(message)
				return
			}
			var msg ConnectionACKMessage
			if err := json.Unmarshal(p, &msg); err != nil {
				message := "failed to unmarshal: " + err.Error()
				logger.Error(message)
				return
			}
			if msg.Type == "start" {
				payload := graphql.Do(graphql.Params{
					Schema:        schema,
					RequestString: msg.Payload.Query,
				})
				if payload.HasErrors() {
					message, err := json.Marshal(map[string]interface{}{
						"type":    "data",
						"id":      msg.OperationID,
						"payload": payload,
					})
					if err != nil {
						logger.Info("failed to marshal message: " + err.Error())
					} else if err = conn.WriteMessage(websocket.TextMessage, message); err != nil {
						logger.Error("failed to write to ws connection: " + err.Error())
					}
					if err = conn.Close(); err != nil {
						logger.Error("cannot close connection: " + err.Error())
					}
					return
				}
				updateTokenString := payload.Data.(map[string]interface{})["formUpdates"].(map[string]interface{})["id"].(string)
				formIDString, userIDString, _ := getUpdateClaimsData(updateTokenString)
				var subscriber = Subscriber{
					FormID:        formIDString,
					Conn:          conn,
					RequestString: msg.Payload.Query,
					OperationID:   msg.OperationID,
				}
				currentConnection, ok := connections.Load(formIDString)
				if !ok {
					pubsub := redisClient.Subscribe("form-" + formIDString)
					connections.Store(formIDString, Connection{
						pubsub,
						map[string]Subscriber{},
					})
					go websocketSubscriptionSender(formIDString)
				}
				currentConnection, _ = connections.Load(formIDString)
				currentSub, ok := currentConnection.(Connection).Subscribers[userIDString]
				if ok {
					if err = currentSub.Conn.Close(); err != nil {
						logger.Error("cannot close connection: " + err.Error())
					}
					delete(currentConnection.(Connection).Subscribers, userIDString)
				}
				currentConnection.(Connection).Subscribers[userIDString] = subscriber
			}
		}
	}()
}

func websocketSubscriptionSender(formIDString string) {
	for {
		connection, ok := connections.Load(formIDString)
		if !ok {
			return
		}
		connectionData := connection.(Connection)
		for msg := range connectionData.Conn.Channel() {
			for userIDString, subscriber := range connectionData.Subscribers {
				payload := graphql.Do(graphql.Params{
					Schema:        schema,
					RequestString: subscriber.RequestString,
					Context:       context.WithValue(context.Background(), dataKey, msg),
				})
				message, err := json.Marshal(map[string]interface{}{
					"type":    "data",
					"id":      subscriber.OperationID,
					"payload": payload,
				})
				if err != nil {
					logger.Error("error marshalling data: " + err.Error())
					continue
				}
				if err := subscriber.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					if err == websocket.ErrCloseSent {
						delete(connectionData.Subscribers, userIDString)
					} else {
						logger.Error("failed to write to ws connection: " + err.Error())
					}
				}
			}
		}
	}
}

func sender() {
	go func() {
		for {
			/*
				connections.Range(func(key, value interface{}) bool {
					connection, ok := value.(*Connection)
					if !ok {
						return true
					}
					connection.Conn
					return true
				})
				ch := pubsub.Channel()
				// Consume messages.
				for msg := range ch {
					fmt.Println(msg.Channel, msg.Payload)
				}
				_ = pubsub.Close()
					time.Sleep(1 * time.Second)
					for _, post := range posts {
						post.Likes = post.Likes + 1
					}
					subscribers.Range(func(key, value interface{}) bool {
						subscriber, ok := value.(*Subscriber)
						if !ok {
							return true
						}
						// pass into context update data
						// maybe look into this: https://stackoverflow.com/a/40380147
						params := graphql.Params{
							Schema:         *schema,
							RequestString:  opts.Query,
							VariableValues: opts.Variables,
							OperationName:  opts.OperationName,
							Context: context.WithValue(context.Background(), "x-auth-id", "12345"),
						}
						graphql.Do(params)
						payload := graphql.Do(graphql.Params{
							Schema:        schema,
							RequestString: subscriber.RequestString,
						})
						message, err := json.Marshal(map[string]interface{}{
							"type":    "data",
							"id":      subscriber.OperationID,
							"payload": payload,
						})
						if err != nil {
							log.Printf("failed to marshal message: %v", err)
							return true
						}
						if err := subscriber.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
							if err == websocket.ErrCloseSent {
								subscribers.Delete(key)
								return true
							}
							log.Printf("failed to write to ws connection: %v", err)
							return true
						}
						return true
					})
			*/
		}
	}()
}

func executeJobs() {
	// update routine to check redis db for jobs to do. if there is a job, assign a unique id to it
	// and read again to make sure no one else took it. then perform task and sleep, do next task next
	// https://github.com/graphql-go/graphql/issues/49

}
