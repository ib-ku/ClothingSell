package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"store/model"
	"store/view"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	//filter
	filterName := r.URL.Query().Get("name")
	filter := bson.M{}

	if filterName != "" {
		filter["name"] = bson.M{"$regex": filterName, "$options": "i"}
	}

	//sort
	sortField := r.URL.Query().Get("sort")
	var sortOrder int
	if sortField != "" {
		sortOrder = 1
		if sortField[0] == '-' {
			sortField = sortField[1:]
			sortOrder = -1
		}
	} else {
		sortField = "price"
		sortOrder = 1
	}

	//pagination
	page := r.URL.Query().Get("page")
	limit := 10
	skip := 0

	if p, err := strconv.Atoi(page); err == nil && p > 1 {
		skip = (p-1) * limit
	}else {
		page = "1"
	}

	cursor, err := productCollection.Find(
		context.TODO(),
		filter, 
		options.Find().SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetLimit(int64(limit)).
		SetSkip(int64(skip)),
	)
	
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

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate fields
	id, idExists := reqData["id"]
	name, nameExists := reqData["name"]
	price, priceExists := reqData["price"]

	if !idExists || !nameExists || !priceExists {
		response := map[string]string{
			"status":  "fail",
			"message": "Fields 'id', 'name', and 'price' are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := id.(float64); !ok { // JSON numbers are decoded as float64
		response := map[string]string{
			"status":  "fail",
			"message": "'id' must be a number",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := name.(string); !ok {
		response := map[string]string{
			"status":  "fail",
			"message": "'name' must be a string",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := price.(float64); !ok {
		response := map[string]string{
			"status":  "fail",
			"message": "'price' must be a number",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert and insert the product
	newProduct := model.Product{
		ID:    int(id.(float64)),
		Name:  name.(string),
		Price: price.(float64),
	}

	_, err = productCollection.InsertOne(context.TODO(), newProduct)
	if err != nil {
		http.Error(w, "Error inserting product into database", http.StatusInternalServerError)
		return
	}

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

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, idExists := reqData["id"]
	if !idExists {
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'id' is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	idFloat, ok := id.(float64)
	if !ok || idFloat <= 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "'id' must be a positive number",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	filter := bson.M{"id": int(idFloat)}
	deleteResult, err := productCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Failed to delete product from the database", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": fmt.Sprintf("No product found with ID %d", int(idFloat)),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Product with ID %d successfully deleted", int(idFloat)),
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

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, idExists := reqData["id"]
	if !idExists {
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'id' is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	idFloat, ok := id.(float64)
	if !ok || idFloat <= 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "'id' must be a positive number",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	updateFields := bson.M{}
	if name, nameExists := reqData["name"]; nameExists {
		if nameStr, ok := name.(string); ok {
			updateFields["name"] = nameStr
		} else {
			response := map[string]string{
				"status":  "fail",
				"message": "'name' must be a string",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	if price, priceExists := reqData["price"]; priceExists {
		if priceFloat, ok := price.(float64); ok {
			updateFields["price"] = priceFloat
		} else {
			response := map[string]string{
				"status":  "fail",
				"message": "'price' must be a number",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
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

	filter := bson.M{"id": int(idFloat)}
	update := bson.M{"$set": updateFields}
	updateResult, err := productCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Failed to update product in the database", http.StatusInternalServerError)
		return
	}

	if updateResult.MatchedCount == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": fmt.Sprintf("No product found with ID %d", int(idFloat)),
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

func GetProductByID(w http.ResponseWriter, r *http.Request) {
	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, idExists := reqData["id"]
	if !idExists {
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'id' is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	idFloat, ok := id.(float64)
	if !ok || idFloat <= 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "'id' must be a positive number",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	var product model.Product
	err = productCollection.FindOne(context.TODO(), bson.M{"id": int(idFloat)}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching product from DB", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"product": product,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetProductByName(w http.ResponseWriter, r *http.Request) {
	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	name, nameExists := reqData["name"]
	if !nameExists {
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'name' is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	nameStr, ok := name.(string)
	if !ok || nameStr == "" {
		response := map[string]string{
			"status":  "fail",
			"message": "'name' must be a non-empty string",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	var product model.Product
	err = productCollection.FindOne(context.TODO(), bson.M{"name": nameStr}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching product from DB", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"product": product,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
