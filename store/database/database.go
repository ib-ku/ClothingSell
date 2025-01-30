package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var ProductCollection *mongo.Collection

// Подключение к MongoDB
func connectMongoDB() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Ошибка при проверке соединения с MongoDB:", err)
	}

	fmt.Println("✅ Успешное подключение к MongoDB")
	ProductCollection = client.Database("store_db").Collection("products")

	return client
}

// InitDB инициализирует соединение
func InitDB() {
	if Client != nil {
		fmt.Println("🔄 База уже подключена, повторное подключение не требуется")
		return
	}
	Client = connectMongoDB()
}

// CloseDB закрывает соединение
func CloseDB() {
	if Client != nil {
		err := Client.Disconnect(context.TODO())
		if err != nil {
			log.Fatalf("Ошибка при отключении от базы данных: %v", err)
		}
		fmt.Println("❌ Отключение от MongoDB")
		Client = nil
	}
}
