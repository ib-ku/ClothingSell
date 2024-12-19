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

var productCollection *mongo.Collection

func InitializeProduct(mongoClient *mongo.Client) {
	productCollection = mongoClient.Database("storeDB").Collection("products")
	fmt.Println("Product collection initialized")
}

func AllProducts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: All Products Hit")

	cursor, err := productCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch products from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var products model.Products
	if err := cursor.All(context.TODO(), &products); err != nil {
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
		http.Error(w, "Failed to insert product into MongoDB", http.StatusInternalServerError)
		return
	}

	fmt.Println("Received Product Information:", "ID:", *requestProduct.ID, "Name:", requestProduct.Name, "Price:", *requestProduct.Price)

	response := map[string]string{
		"status":  "success",
		"message": "Product data successfully received",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetProductByID(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID int `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.ID == 0 {
		http.Error(w, "Invalid JSON format or missing product ID", http.StatusBadRequest)
		return
	}

	var product model.Product
	err = productCollection.FindOne(context.TODO(), bson.M{"id": request.ID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching product from MongoDB", http.StatusInternalServerError)
		}
		return
	}

	view.RenderProducts(w, product)
}

func GetProductByName(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Name == "" {
		http.Error(w, "Invalid JSON format or missing product name", http.StatusBadRequest)
		return
	}

	var product model.Product

	err = productCollection.FindOne(context.TODO(), bson.M{"name": request.Name}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching product from MongoDB", http.StatusInternalServerError)
		}
		return
	}

	view.RenderProducts(w, product)
}
