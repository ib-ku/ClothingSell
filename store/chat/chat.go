package chat

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var chatCollection *mongo.Collection

type ChatMessage struct {
	ChatID    string    `json:"chatId" bson:"chatId"`
	Sender    string    `json:"sender" bson:"sender"`
	Message   string    `json:"message" bson:"message"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	IsActive  bool      `json:"isActive" bson:"isActive"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan ChatMessage)

func InitializeChatCollection(client *mongo.Client) {
	chatCollection = client.Database("storeDB").Collection("chats")
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade WebSocket:", err)
		return
	}
	defer ws.Close()

	clients[ws] = ""

	for {
		var msg ChatMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			delete(clients, ws)
			break
		}
		msg.Timestamp = time.Now()

		log.Printf("Received message from %s: %s", msg.Sender, msg.Message)

		_, err = chatCollection.InsertOne(context.TODO(), msg)
		if err != nil {
			log.Printf("Error inserting message into MongoDB: %v", err)
		} else {
			log.Println("Message inserted successfully")
		}

		checkForAutoResponse(msg)

		broadcast <- msg
	}
}

func checkForAutoResponse(msg ChatMessage) {
	userMessage := strings.ToLower(msg.Message)
	keywordMap := map[string][]string{
		"You can buy them here: http://localhost:8085/":                                       {"buy", "product", "purchase"},
		"You can reach us on Telegram: https://t.me/bvbl1 or by email: ibragimtop1@gmail.com": {"contact", "reach", "support"},
		"Place an order here: `http://localhost:8085/cart.html` ":                             {"order", "checkout", "purchase"},
		"Hello! How can I help you?":                                                          {"hello", "hi", "whatsup"},
		"How can I assist you?":                                                               {"help", "assist", "support"},
	}

	for response, keywords := range keywordMap {
		for _, keyword := range keywords {
			if strings.Contains(userMessage, keyword) {
				autoResponse := ChatMessage{
					ChatID:    msg.ChatID,
					Sender:    "Bot",
					Message:   response,
					Timestamp: time.Now(),
					IsActive:  true,
				}

				_, err := chatCollection.InsertOne(context.TODO(), autoResponse)
				if err != nil {
					log.Printf("Error inserting auto-response into MongoDB: %v", err)
				}

				broadcast <- autoResponse
				return
			}
		}
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error broadcasting message: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func GetChatHistory(chatID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	filter := bson.M{"chatId": chatID}

	cur, err := chatCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var msg ChatMessage
		if err := cur.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
