package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"store/services"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var chatCollection *mongo.Collection

type ChatMessage struct {
	ChatID    string    `json:"chatId" bson:"chatId"`
	UserID    string    `json:"userId" bson:"userId"`
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
	// Upgrade initial GET request to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	// Validate user token and get user information
	cookie, err := r.Cookie("Authorization")
	if err != nil || cookie.Value == "" {
		return
	}

	// Remove "Bearer " prefix from the token if it exists
	token := strings.TrimPrefix(cookie.Value, "Bearer ")

	claims, err := services.ParseJWT(token)
	if err != nil {
		log.Println("Token parsing failed:", err)
		return
	}

	// Listen for new chat messages
	for {
		var msg ChatMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Add user identification and timestamp to the message
		msg.UserID = claims.Email
		msg.Timestamp = time.Now()

		// Insert message into MongoDB
		_, err = chatCollection.InsertOne(context.TODO(), msg)
		if err != nil {
			log.Printf("Error inserting message into MongoDB: %v", err)
		} else {
			log.Println("Message inserted successfully into MongoDB")
		}

		// Broadcast the message to the WebSocket
		msgBytes, _ := json.Marshal(msg)
		if err = conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			log.Printf("Error writing message to WebSocket: %v", err)
			break
		}
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

func GetChatHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	var messages []ChatMessage
	filter := bson.M{"userId": userID}

	cursor, err := chatCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Println("Failed to retrieve chat history:", err)
		http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var msg ChatMessage
		if err := cursor.Decode(&msg); err != nil {
			log.Println("Failed to decode message:", err)
			continue
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func GetUserChatHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	messages, err := GetChatHistoryByUserID(userID)
	if err != nil {
		log.Println("Failed to retrieve chat history:", err)
		http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func GetChatHistoryByUserID(userID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	filter := bson.M{"userId": userID}

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
