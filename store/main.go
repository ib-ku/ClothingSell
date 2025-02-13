package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"store/controller"
	"store/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func connectMongoDB() *mongo.Client {
	mongoURI := "mongodb://storeUser:securePassword@localhost:27017/storeDB"
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("Successfully connected to MongoDB!")
	testCollection := client.Database("storeDB").Collection("test")
	_, err = testCollection.InsertOne(context.TODO(), map[string]string{"test": "connection"})
	if err != nil {
		log.Fatalf("Test insertion failed: %v", err)
	} else {
		fmt.Println("Test document inserted successfully")
	}
	return client
}

func message(w http.ResponseWriter, r *http.Request) {
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

	_, exists := reqData["message"]
	if !exists {
		response := map[string]string{
			"status":  "fail",
			"message": "key message is absent",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	_, ok := reqData["message"].(string)
	if !ok {
		response := map[string]string{
			"status":  "fail",
			"message": "Message field must be a string",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "Hello,This is postman ",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the protected route!")
}

func AdminPanelHandler(w http.ResponseWriter, r *http.Request) {
	adminData := map[string]string{
		"title":   "Admin Dashboard",
		"message": "Welcome to the admin panel!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(adminData)
}

func handleRequests() {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
	controller.InitializeProduct(client)
	controller.InitializeUser(client)
	controller.InitializeCart(client)

	http.HandleFunc("/home", message)

	http.Handle("/admin", middleware.IsAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/admin.html")
	})))

	http.Handle("/addToCart", middleware.AuthMiddleware(http.HandlerFunc(controller.AddToCart)))
	http.Handle("/getCartItems", middleware.AuthMiddleware(http.HandlerFunc(controller.GetCartItems)))

	http.HandleFunc("/allProducts", controller.AllProducts)
	http.HandleFunc("/allUsers", controller.AllUsers)

	http.HandleFunc("/postUser", controller.HandleUserPostRequest)
	http.HandleFunc("/postProduct", controller.HandleProductPostRequest)

	http.HandleFunc("/deleteProductById", controller.DeleteProductByID)
	http.HandleFunc("/deleteUserByEmail", controller.DeleteUserByEmail)

	http.HandleFunc("/updateProductById", controller.UpdateProductByID)
	http.HandleFunc("/updateUserByEmail", controller.UpdateUserByEmail)

	http.HandleFunc("/getUserEmail", controller.GetUserByEmail)
	http.HandleFunc("/getUsername", controller.GetUserByUsername)

	http.HandleFunc("/getProductByID", controller.GetProductByID)
	http.HandleFunc("/getProductByName", controller.GetProductByName)
	http.HandleFunc("/getUser", controller.GetUser)

	http.HandleFunc("/logout", controller.Logout)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Регистрация и подтверждение email
	http.HandleFunc("/signup", controller.SignUp)
	http.HandleFunc("/confirm", controller.Confirm)

	// Вход, получение данных о пользователе
	http.HandleFunc("/login", controller.Login)

	// Маршруты для безопасности (JWT + роли)
	http.Handle("/protected", middleware.AuthMiddleware(http.HandlerFunc(ProtectedHandler)))
	http.Handle("/assign-role", middleware.IsAdmin(http.HandlerFunc(controller.AssignRole)))

	// Двухфакторная аутентификация (OTP)
	http.HandleFunc("/send-otp", controller.SendOTP)
	http.HandleFunc("/verify-otp", controller.VerifyOTP)

	http.HandleFunc("/sendEmail", controller.SendPromotionalEmail)

	server := &http.Server{Addr: ":8085", Handler: nil}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Println("Server is shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		log.Println("Server exiting")
	}()

	log.Fatal(server.ListenAndServe())
}

func main() {
	client = connectMongoDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
		fmt.Println("Disconnected from MongoDB")
	}()
	fmt.Println("http://localhost:8085")
	handleRequests()
}
