package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"store/model"
	"store/services"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartItem struct {
	UserID      string  `json:"user_id" bson:"user_id"`
	ProductID   int     `json:"product_id" bson:"product_id"`
	ProductName string  `json:"product_name" bson:"product_name"`
	Quantity    int     `json:"quantity" bson:"quantity"`
	Price       float64 `json:"price" bson:"price"`
}

var cartCollection *mongo.Collection

func InitializeCart(client *mongo.Client) {
	cartCollection = client.Database("storeDB").Collection("cartItems")
}

func AddToCart(w http.ResponseWriter, r *http.Request) {
	var cartItem CartItem
	if err := json.NewDecoder(r.Body).Decode(&cartItem); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cookie, err := r.Cookie("Authorization")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(cookie.Value, "Bearer ")
	claims, err := services.ParseJWT(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	var product model.Product
	err = productCollection.FindOne(ctx, bson.M{"id": cartItem.ProductID}).Decode(&product)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	cartItem.ProductName = product.Name
	cartItem.UserID = claims.Email
	cartItem.Quantity = 1
	cartItem.Price = product.Price

	_, err = cartCollection.InsertOne(ctx, cartItem)
	if err != nil {
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item added to cart"})
}

func GetCartItems(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Authorization")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(cookie.Value, "Bearer ")
	claims, err := services.ParseJWT(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"user_id": claims.Email}

	searchQuery := r.URL.Query().Get("search")
	if searchQuery != "" {
		filter["product_name"] = bson.M{"$regex": searchQuery, "$options": "i"}
	}

	cursor, err := cartCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to retrieve cart items", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var cartItems []CartItem
	if err = cursor.All(ctx, &cartItems); err != nil {
		http.Error(w, "Failed to parse cart items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cartItems)
}

func RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		ProductID int `json:"product_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Println("Invalid JSON format:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestData.ProductID == 0 {
		log.Println("Missing product_id in request")
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("Authorization")
	if err != nil || cookie.Value == "" {
		log.Println("Unauthorized: Missing token")
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(cookie.Value, "Bearer ")
	claims, err := services.ParseJWT(token)
	if err != nil {
		log.Println("Invalid or expired token")
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": claims.Email, "product_id": requestData.ProductID}
	result, err := cartCollection.DeleteOne(ctx, filter)

	if err != nil {
		log.Println("Failed to remove item from cart:", err)
		http.Error(w, "Failed to remove item from cart", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		log.Println("No matching product found for deletion:", requestData.ProductID)
		http.Error(w, "Item not found in cart", http.StatusNotFound)
		return
	}

	log.Println("Successfully removed item:", requestData.ProductID)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item removed from cart"})
}

func CompletePurchase(w http.ResponseWriter, r *http.Request) {
	var transactionData struct {
		TransactionID string  `json:"transaction_id"`
		Amount        float64 `json:"amount"`
		UserID        string  `json:"user_id"`
		ProductID     int     `json:"product_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&transactionData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaction successful"})
}
