package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	json "github.com/json-iterator/go"
)

// ConnectionACKMessage message for getting connection data
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

// Subscriber websocket subscriber object
type Subscriber struct {
	FormID        string
	Conn          *websocket.Conn
	RequestString string
	OperationID   string
	ID            string
}

// Connection websocket connection
type Connection struct {
	Conn          *redis.PubSub
	Subscribers   map[string]Subscriber
	LastSave      time.Time
	AlreadySaving bool
}

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
			if err != nil {
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
					Context:       context.WithValue(context.Background(), getTokenKey, true),
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
				updatesAccessTokenString := payload.Data.(map[string]interface{})["formUpdates"].(map[string]interface{})["id"].(string)
				formIDString, _, connectionIDString, _ := getUpdateClaimsData(updatesAccessTokenString, editAccessLevel)
				payload = graphql.Do(graphql.Params{
					Schema:        schema,
					RequestString: msg.Payload.Query,
					Context:       context.WithValue(context.Background(), getConnectionIDKey, connectionIDString),
				})
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
				if payload.HasErrors() || err != nil {
					if err = conn.Close(); err != nil {
						logger.Error("cannot close connection: " + err.Error())
					}
					return
				}
				var subscriber = Subscriber{
					FormID:        formIDString,
					Conn:          conn,
					RequestString: msg.Payload.Query,
					OperationID:   msg.OperationID,
				}
				currentConnection, _ := connections.Load(formIDString)
				if currentConnection == nil {
					pubsub := redisClient.Subscribe(updateFormPath + formIDString)
					currentTime := time.Now()
					connections.Store(formIDString, Connection{
						pubsub,
						map[string]Subscriber{},
						currentTime,
						false,
					})
					logger.Info("start websocket sender")
					go websocketSubscriptionSender(formIDString)
				} else {
					logger.Info("current connection not nil...")
				}
				currentConnection, _ = connections.Load(formIDString)
				currentSub, ok := currentConnection.(Connection).Subscribers[connectionIDString]
				if ok {
					if err = currentSub.Conn.Close(); err != nil {
						logger.Error("cannot close connection: " + err.Error())
					}
					delete(currentConnection.(Connection).Subscribers, connectionIDString)
				}
				currentConnection.(Connection).Subscribers[connectionIDString] = subscriber
				logger.Info("saved new subscriber")
			}
		}
	}()
}

func websocketSubscriptionSender(formIDString string) {
	for {
		connection, ok := connections.Load(formIDString)
		if !ok {
			logger.Info("cannot find connection...")
			return
		}
		connectionData := connection.(Connection)
		for msg := range connectionData.Conn.Channel() {
			for userIDString, subscriber := range connectionData.Subscribers {
				payload := graphql.Do(graphql.Params{
					Schema:        schema,
					RequestString: subscriber.RequestString,
					Context:       context.WithValue(context.Background(), dataKey, msg.Payload),
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
