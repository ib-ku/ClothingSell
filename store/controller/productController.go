package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"store/model"
	"store/view"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client
var productCollection *mongo.Collection

func InitializeProduct(mongoClient *mongo.Client) {
	client = mongoClient
	productCollection = client.Database("storeDB").Collection("products")
	fmt.Println("Product collection initialized")
}

func AllProducts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: All Products Endpoint")

	cursor, err := productCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Error fetching products from database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var products model.Products
	if err = cursor.All(context.TODO(), &products); err != nil {
		http.Error(w, "Error decoding product data", http.StatusInternalServerError)
		return
	}

	view.RenderProducts(w, products)
}

func HandleProductPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var requestProduct struct {
		ID    *int     `json:"id"`
		Name  string   `json:"name"`
		Price *float64 `json:"price"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestProduct)
	if err != nil {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid JSON format",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if requestProduct.ID == nil || requestProduct.Name == "" || requestProduct.Price == nil {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid product data: ID, Name, and Price are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	newProduct := model.Product{
		ID:    *requestProduct.ID,
		Name:  requestProduct.Name,
		Price: *requestProduct.Price,
	}

	_, err = productCollection.InsertOne(context.TODO(), newProduct)
	if err != nil {
		http.Error(w, "Error inserting product into database", http.StatusInternalServerError)
		return
	}

	fmt.Println("Received information:", "ID:", *requestProduct.ID, "Name:", requestProduct.Name, "Price:", *requestProduct.Price)

	response := map[string]string{
		"status":  "success",
		"message": "Product data successfully received",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
