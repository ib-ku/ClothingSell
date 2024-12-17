package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"store/controller"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func connectMongoDB() *mongo.Client {
	mongoURI := "mongodb://storeUser:securePassword@localhost:27017/storeDB"
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("Successfully connected to MongoDB!")
	testCollection := client.Database("storeDB").Collection("test")
	_, err = testCollection.InsertOne(context.TODO(), map[string]string{"test": "connection"})
	if err != nil {
		log.Fatalf("Test insertion failed: %v", err)
	} else {
		fmt.Println("Test document inserted successfully")
	}
	return client
}

func handleRequests() {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
	controller.InitializeProduct(client)
	controller.InitializeUser(client)
	http.HandleFunc("/allProducts", controller.AllProducts)
	http.HandleFunc("/allUsers", controller.AllUsers)

	http.HandleFunc("/postUser", controller.HandleUserPostRequest)
	http.HandleFunc("/postProduct", controller.HandleProductPostRequest)

	http.HandleFunc("/deleteProductById", controller.DeleteProductByID)
	http.HandleFunc("/deleteUserByEmail", controller.DeleteUserByEmail)

	http.HandleFunc("/updateProductById", controller.UpdateProductByID)
	http.HandleFunc("/updateUserByEmail", controller.UpdateUserByEmail)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	client = connectMongoDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
		fmt.Println("Disconnected from MongoDB")
	}()
	fmt.Println("http://localhost:8080")
	handleRequests()
}
