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

var userCollection *mongo.Collection

func InitializeUser(mongoClient *mongo.Client) {
	userCollection = mongoClient.Database("storeDB").Collection("users")
	fmt.Println("User collection initialized")
}

func AllUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: All Users Hit")

	cursor, err := userCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch users from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var users model.Users
	if err := cursor.All(context.TODO(), &users); err != nil {
		http.Error(w, "Error decoding user data", http.StatusInternalServerError)
		return
	}

	view.RenderUsers(w, users)
}

func HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var requestUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestUser)
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

	if requestUser.Email == "" || requestUser.Password == "" || requestUser.Username == "" {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid user data: Email, Password, and Username are required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	newUser := model.User{
		Email:    requestUser.Email,
		Password: requestUser.Password,
		Username: requestUser.Username,
	}

	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		http.Error(w, "Failed to insert user into MongoDB", http.StatusInternalServerError)
		return
	}

	fmt.Println("Received User Information:", "Email:", requestUser.Email, "Password:", requestUser.Password, "Username:", requestUser.Username)

	response := map[string]string{
		"status":  "success",
		"message": "User data successfully received",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Email == "" {
		http.Error(w, "Invalid json format or missing email", http.StatusBadRequest)
		return
	}

	var user model.User

	err = userCollection.FindOne(context.TODO(), bson.M{"email": request.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user from DB", http.StatusInternalServerError)
		}
	}

	view.RenderUsers(w, user)
}

func GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Username == "" {
		http.Error(w, "Invalid json format or missing username", http.StatusBadRequest)
		return
	}

	var user model.User
	fmt.Println("Received Username:", request.Username)

	err = userCollection.FindOne(context.TODO(), bson.M{"username": request.Username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user from DB", http.StatusInternalServerError)
		}
	}

	view.RenderUsers(w, user)
}
