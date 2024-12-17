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

func DeleteProductByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ID int `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
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

	if request.ID == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid product ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	filter := bson.M{"id": request.ID}
	deleteResult, err := productCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Failed to delete product from the database", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "No product found with the provided ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Product with ID %d successfully deleted", request.ID),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UpdateProductByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		ID    int      `json:"id"`
		Name  *string  `json:"name,omitempty"`
		Price *float64 `json:"price,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
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

	if request.ID == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "ID is required to update a product",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	updateFields := bson.M{}
	if request.Name != nil {
		updateFields["name"] = *request.Name
	}
	if request.Price != nil {
		updateFields["price"] = *request.Price
	}

	if len(updateFields) == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "No fields to update",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	filter := bson.M{"id": request.ID}
	update := bson.M{"$set": updateFields}

	updateResult, err := productCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Failed to update product in the database", http.StatusInternalServerError)
		return
	}

	if updateResult.MatchedCount == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "No product found with the provided ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Product updated successfully",
		"updated": updateFields,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
