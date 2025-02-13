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
		log.Println("Failed to decode JSON:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cookie, err := r.Cookie("Authorization")
	if err != nil || cookie.Value == "" {
		log.Println("Missing or invalid token:", err)
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(cookie.Value, "Bearer ")
	claims, err := services.ParseJWT(token)
	if err != nil {
		log.Println("Token parsing failed:", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	log.Println("User ID (Email) from token:", claims.Email)

	var product model.Product
	err = productCollection.FindOne(ctx, bson.M{"id": cartItem.ProductID}).Decode(&product)
	if err != nil {
		log.Println("Product not found:", err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	cartItem.ProductName = product.Name
	cartItem.UserID = claims.Email
	cartItem.Quantity = 1
	cartItem.Price = product.Price

	_, err = cartCollection.InsertOne(ctx, cartItem)
	if err != nil {
		log.Println("Failed to add item to cart:", err)
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	log.Println("Item added to cart:", cartItem)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Item added to cart"})
}

func GetCartItems(w http.ResponseWriter, r *http.Request) {
	log.Println("GetCartItems handler reached")
	cookie, err := r.Cookie("Authorization")
	if err != nil || cookie.Value == "" {
		log.Println("Missing or invalid token:", err)
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(cookie.Value, "Bearer ")
	claims, err := services.ParseJWT(token)
	if err != nil {
		log.Println("Token parsing failed:", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	log.Println("User:", claims.Email)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"user_id": claims.Email}
	cursor, err := cartCollection.Find(ctx, filter)
	if err != nil {
		log.Println("Failed to retrieve cart items:", err)
		http.Error(w, "Failed to retrieve cart items", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var cartItems []CartItem
	if err = cursor.All(ctx, &cartItems); err != nil {
		log.Println("Failed to parse cart items:", err)
		http.Error(w, "Failed to parse cart items", http.StatusInternalServerError)
		return
	}

	log.Println("Cart items retrieved:", cartItems)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cartItems)
}
