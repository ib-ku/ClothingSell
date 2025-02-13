package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type PaymentRequest struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	UserID        string  `json:"user_id"`
	ProductID     int     `json:"product_id"` // Added ProductID
}

type PaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

func main() {
	http.HandleFunc("/processPayment", processPaymentHandler)
	log.Println("Payment Microservice running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func processPaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var paymentRequest PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		log.Println("Failed to decode request:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Printf("Received payment request: %+v\n", paymentRequest) // Log the full request to debug

	// Simulate payment processing
	status := "Paid"
	if time.Now().Second()%2 == 0 {
		status = "Declined"
	}

	paymentResponse := PaymentResponse{
		TransactionID: paymentRequest.TransactionID,
		Status:        status,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paymentResponse)
	log.Println("Processed payment for transaction:", paymentRequest.TransactionID, "Status:", status)
}
