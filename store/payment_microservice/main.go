package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type PaymentRequest struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	UserID        string  `json:"user_id"`
	ProductID     int     `json:"product_id"`
	PaymentMethod string  `json:"payment_method"`
	Email         string  `json:"email"`
	Address       string  `json:"address"`
	Phone         string  `json:"phone"`
	PurchaseDate  string  `json:"purchase_date"`
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

	log.Printf("Received payment request: %+v\n", paymentRequest)

	// Всегда возвращаем "Paid"
	paymentResponse := PaymentResponse{
		TransactionID: paymentRequest.TransactionID,
		Status:        "Paid",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paymentResponse)
	log.Println("Processed payment for transaction:", paymentRequest.TransactionID, "Status: Paid")
}
