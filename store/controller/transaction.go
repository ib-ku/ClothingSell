package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"store/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var transactionCollection *mongo.Collection

func InitializeTransaction(client *mongo.Client) {
	transactionCollection = client.Database("storeDB").Collection("transactions")
}

type PaymentRequest struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	UserID        string  `json:"user_id"`
	ProductID     int     `json:"product_id"`
}

type PaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

func ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	var paymentRequest PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		log.Println("Failed to decode payment request:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Println("Processing transaction for:", paymentRequest)

	// Simulate sending a request to the payment microservice
	requestBody, _ := json.Marshal(paymentRequest)
	resp, err := http.Post("http://localhost:8081/processPayment", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Failed to contact payment microservice:", err)
		http.Error(w, "Payment service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	var paymentResponse PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		log.Println("Failed to decode payment response:", err)
		http.Error(w, "Invalid response from payment service", http.StatusInternalServerError)
		return
	}

	log.Println("Payment response:", paymentResponse)

	if paymentResponse.Status == "Paid" {
		log.Println("Transaction successful. Generating receipt...")

		receiptPath := utils.GenerateReceipt(paymentRequest.TransactionID, paymentRequest.UserID, paymentRequest.Amount)
		log.Println("Receipt generated:", receiptPath)

		utils.SendReceipt(paymentRequest.UserID, receiptPath)

		// Insert transaction record into MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		transactionData := bson.M{
			"transaction_id": paymentRequest.TransactionID,
			"user_id":        paymentRequest.UserID,
			"product_id":     paymentRequest.ProductID,
			"amount":         paymentRequest.Amount,
			"status":         "Completed",
			"created_at":     time.Now(),
		}

		_, err = transactionCollection.InsertOne(ctx, transactionData)
		if err != nil {
			log.Println("Failed to insert transaction into MongoDB:", err)
		} else {
			log.Println("Transaction successfully inserted into MongoDB.")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction processed successfully"))
	} else {
		log.Println("Transaction declined. No receipt will be sent.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction declined"))
	}
}
