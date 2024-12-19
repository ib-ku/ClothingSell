package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"store/model"
	"store/view"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection

func InitializeUser(mongoClient *mongo.Client) {
	userCollection = client.Database("storeDB").Collection("users")
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

func DeleteUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Email string `json:"email"`
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

	if request.Email == "" {
		response := map[string]string{
			"status":  "fail",
			"message": "Email is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	request.Email = strings.TrimSpace(request.Email)
	fmt.Printf("Deleting user with email: %s\n", request.Email)

	filter := bson.M{"email": bson.M{"$regex": request.Email, "$options": "i"}}

	deleteResult, err := userCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Failed to delete user from the database", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "No user found with the provided Email",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("User with email %s successfully deleted", request.Email),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}



func UpdateUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Email    string  `json:"email"`
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
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

	if request.Email == "" {
		response := map[string]string{
			"status":  "fail",
			"message": "Email is required to update a user",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	updateFields := bson.M{}
	if request.Username != nil {
		updateFields["username"] = *request.Username
	}
	if request.Password != nil {
		updateFields["password"] = *request.Password
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

	filter := bson.M{"email": request.Email}
	update := bson.M{"$set": updateFields}

	updateResult, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Failed to update user in the database", http.StatusInternalServerError)
		return
	}

	if updateResult.MatchedCount == 0 {
		response := map[string]string{
			"status":  "fail",
			"message": "No user found with the provided Email",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "User updated successfully",
		"updated": updateFields,
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