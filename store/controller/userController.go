package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"store/model"
	"store/view"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var userCollection *mongo.Collection

func InitializeUser(mongoClient *mongo.Client) {
	userCollection = mongoClient.Database("storeDB").Collection("users")
	fmt.Println("User collection initialized")
}

func jsonResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func AllUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: All Users")

	// filter
	filterEmail := r.URL.Query().Get("email")
	filterUsername := r.URL.Query().Get("username")
	filter := bson.M{}

	if filterEmail != "" {
		filter["email"] = bson.M{"$regex": filterEmail, "$options": "i"}
	}
	if filterUsername != "" {
		filter["username"] = bson.M{"$regex": filterUsername, "$options": "i"}
	}

	//sort
	sortField := r.URL.Query().Get("sort")
	var sortOrder int 
	if sortField != ""{
		sortOrder = 1
		if sortField[0] == '-'{
			sortField = sortField[1:]
			sortOrder = -1
		}
	} else {
        sortField = "username" 
        sortOrder = 1
    }

	//pagination
	page := r.URL.Query().Get("page")
	limit := 10
	skip := 0
	
	if p, err := strconv.Atoi(page); err == nil && p > 1{
		skip = (p-1) * limit
	} else {
		page = "1"
	}

	cursor, err := userCollection.Find(
        context.TODO(),
        filter,
        options.Find().
            SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
            SetLimit(int64(limit)).
            SetSkip(int64(skip)),
    )

	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to fetch users from database"})
		return
	}
	defer cursor.Close(context.TODO())

	var users model.Users
	if err := cursor.All(context.TODO(), &users); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Error decoding user data"})
		return
	}

	view.RenderUsers(w, users)
}


func HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "fail", "message": "Only POST methods are allowed!"})
		return
	}

	var requestUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestUser)
	if err != nil || requestUser.Email == "" || requestUser.Password == "" || requestUser.Username == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing fields"})
		return
	}

	newUser := model.User{
		Email:    requestUser.Email,
		Password: requestUser.Password,
		Username: requestUser.Username,
	}

	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to insert user into the database"})
		return
	}

	fmt.Println("Received User Information:", "Email:", requestUser.Email, "Password:", requestUser.Password, "Username:", requestUser.Username)

	jsonResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "User data successfully received"})
}

func DeleteUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "fail", "message": "Only DELETE methods are allowed!"})
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Email == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing email"})
		return
	}

	request.Email = strings.TrimSpace(request.Email)
	filter := bson.M{"email": bson.M{"$regex": request.Email, "$options": "i"}}
	deleteResult, err := userCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to delete user from database"})
		return
	}

	if deleteResult.DeletedCount == 0 {
		jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "No user found with the provided email"})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("User with email %s successfully deleted", request.Email)})
}

func UpdateUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "fail", "message": "Only PUT methods are allowed!"})
		return
	}

	var request struct {
		Email    string  `json:"email"`
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Email == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing email"})
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
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "No fields to update"})
		return
	}

	filter := bson.M{"email": request.Email}
	update := bson.M{"$set": updateFields}
	updateResult, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil || updateResult.MatchedCount == 0 {
		jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "No user found with the provided email"})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "User updated successfully", "updated": updateFields})
}

func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Email == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing email"})
		return
	}

	var user model.User
	err = userCollection.FindOne(context.TODO(), bson.M{"email": request.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "User not found"})
		} else {
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Error fetching user from database"})
		}
		return
	}

	view.RenderUsers(w, user)
}

func GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Username == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing username"})
		return
	}

	var user model.User
	err = userCollection.FindOne(context.TODO(), bson.M{"username": request.Username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "User not found"})
		} else {
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Error fetching user from database"})
		}
		return
	}

	view.RenderUsers(w, user)
}

