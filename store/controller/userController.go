package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"store/model"
	"store/view"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/time/rate"
)

var userCollection *mongo.Collection
var limiter = rate.NewLimiter(1, 3)

var log = logrus.New()

func init() {
	logFile, err := os.OpenFile("logging.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}

	log.SetOutput(logFile)

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)

	log.WithFields(logrus.Fields{
		"action": "initialize_logger",
		"status": "success",
	}).Info("Logger initialized and writing to logging.txt")
}

func InitializeUser(mongoClient *mongo.Client) {
	userCollection = mongoClient.Database("storeDB").Collection("users")
	log.WithFields(logrus.Fields{
		"action": "initialize",
		"status": "success",
	}).Info("User collection initialized")
}

func jsonResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func AllUsers(w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		log.WithFields(logrus.Fields{
			"action": "rate_limit_exceeded",
			"status": "fail",
		}).Warn("Rate limit exceeded")
		return
	}

	log.WithFields(logrus.Fields{
		"action": "fetch_users",
		"status": "start",
	}).Info("Fetching all users")

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

	// sort
	sortField := r.URL.Query().Get("sort")
	var sortOrder int
	if sortField != "" {
		sortOrder = 1
		if sortField[0] == '-' {
			sortField = sortField[1:]
			sortOrder = -1
		}
	} else {
		sortField = "username"
		sortOrder = 1
	}

	// pagination
	page := r.URL.Query().Get("page")
	limit := 10
	skip := 0

	if p, err := strconv.Atoi(page); err == nil && p > 1 {
		skip = (p - 1) * limit
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
		log.WithFields(logrus.Fields{
			"action": "all_users",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to fetch users from database")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to fetch users from database"})
		return
	}
	defer cursor.Close(context.TODO())

	var users model.Users
	if err := cursor.All(context.TODO(), &users); err != nil {
		log.WithFields(logrus.Fields{
			"action": "all_users",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Error decoding user data")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Error decoding user data"})
		return
	}

	view.RenderUsers(w, users)

	log.WithFields(logrus.Fields{
		"action": "all_users",
		"status": "success",
		"count":  len(users),
	}).Info("Fetched users successfully")
}

func HandleUserPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.WithFields(logrus.Fields{
			"action": "handle_post_request",
			"status": "fail",
		}).Warn("Only POST methods are allowed!")
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
		log.WithFields(logrus.Fields{
			"action": "handle_post_request",
			"status": "fail",
		}).Warn("Invalid JSON format or missing fields")
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
		log.WithFields(logrus.Fields{
			"action": "handle_post_request",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to insert user into the database")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to insert user into the database"})
		return
	}

	log.WithFields(logrus.Fields{
		"action":   "handle_post_request",
		"status":   "success",
		"email":    requestUser.Email,
		"username": requestUser.Username,
	}).Info("Received user information and inserted into the database")

	jsonResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "User data successfully received"})
}

func DeleteUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		log.WithFields(logrus.Fields{
			"action": "handle_delete_request",
			"status": "fail",
		}).Warn("Only DELETE methods are allowed!")
		jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "fail", "message": "Only DELETE methods are allowed!"})
		return
	}

	var request struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Email == "" {
		log.WithFields(logrus.Fields{
			"action": "handle_delete_request",
			"status": "fail",
		}).Warn("Invalid JSON format or missing email")
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing email"})
		return
	}

	request.Email = strings.TrimSpace(request.Email)
	filter := bson.M{"email": bson.M{"$regex": request.Email, "$options": "i"}}
	deleteResult, err := userCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "handle_delete_request",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to delete user from database")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to delete user from database"})
		return
	}

	if deleteResult.DeletedCount == 0 {
		log.WithFields(logrus.Fields{
			"action": "handle_delete_request",
			"status": "fail",
		}).Warn("No user found with the provided email")
		jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "No user found with the provided email"})
		return
	}

	log.WithFields(logrus.Fields{
		"action": "handle_delete_request",
		"status": "success",
		"email":  request.Email,
	}).Info("User successfully deleted from the database")

	jsonResponse(w, http.StatusOK, map[string]string{"status": "success", "message": fmt.Sprintf("User with email %s successfully deleted", request.Email)})
}

func UpdateUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		log.WithFields(logrus.Fields{
			"action": "handle_put_request",
			"status": "fail",
		}).Warn("Only PUT methods are allowed!")
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
		log.WithFields(logrus.Fields{
			"action": "handle_put_request",
			"status": "fail",
		}).Warn("Invalid JSON format or missing email")
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
		log.WithFields(logrus.Fields{
			"action": "handle_put_request",
			"status": "fail",
		}).Warn("No fields to update")
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "No fields to update"})
		return
	}

	filter := bson.M{"email": request.Email}
	update := bson.M{"$set": updateFields}
	updateResult, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil || updateResult.MatchedCount == 0 {
		log.WithFields(logrus.Fields{
			"action": "handle_put_request",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to update user")
		jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "No user found with the provided email"})
		return
	}

	log.WithFields(logrus.Fields{
		"action":  "handle_put_request",
		"status":  "success",
		"email":   request.Email,
		"updated": updateFields,
	}).Info("User successfully updated")

	jsonResponse(w, http.StatusOK, map[string]interface{}{"status": "success", "message": "User updated successfully", "updated": updateFields})
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

	log.WithFields(logrus.Fields{
		"action":   "get_user_by_username",
		"status":   "success",
		"username": request.Username,
	}).Info("User fetched successfully")
}

func GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.Email == "" {
		log.WithFields(logrus.Fields{
			"action": "get_user_by_email",
			"status": "fail",
		}).Warn("Invalid JSON format or missing email")
		jsonResponse(w, http.StatusBadRequest, map[string]string{"status": "fail", "message": "Invalid JSON format or missing email"})
		return
	}

	var user model.User
	err = userCollection.FindOne(context.TODO(), bson.M{"email": request.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.WithFields(logrus.Fields{
				"action": "get_user_by_email",
				"status": "fail",
			}).Warn("User not found")
			jsonResponse(w, http.StatusNotFound, map[string]string{"status": "fail", "message": "User not found"})
		} else {
			log.WithFields(logrus.Fields{
				"action": "get_user_by_email",
				"status": "fail",
				"error":  err.Error(),
			}).Error("Error fetching user from database")
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Error fetching user from database"})
		}
		return
	}

	view.RenderUsers(w, user)

	log.WithFields(logrus.Fields{
		"action": "get_user_by_email",
		"status": "success",
		"email":  request.Email,
	}).Info("User fetched successfully")
}
