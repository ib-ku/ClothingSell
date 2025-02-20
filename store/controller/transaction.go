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

func ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	var paymentRequest PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		log.Println("Failed to decode payment request:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	log.Println("Processing transaction for:", paymentRequest)

	// Отправляем запрос в платежный сервис
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
		log.Println("Transaction successful. Saving history...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Получаем данные о товаре
		var cartItem CartItem
		err = cartCollection.FindOne(ctx, bson.M{"user_id": paymentRequest.UserID, "product_id": paymentRequest.ProductID}).Decode(&cartItem)
		if err != nil {
			log.Println("Cart item not found:", err)
			http.Error(w, "Cart item not found", http.StatusNotFound)
			return
		}

		historyData := bson.M{
			"user_id":      paymentRequest.UserID,
			"product_id":   cartItem.ProductID,
			"product_name": cartItem.ProductName,
			"price":        cartItem.Price,
		}

		// Проверяем, есть ли поле даты
		if _, exists := historyData["purchase_date"]; !exists {
			historyData["purchase_date"] = time.Now().Format(time.RFC3339)
		}

		_, err = transactionCollection.InsertOne(ctx, historyData)
		if err != nil {
			log.Println("Failed to save purchase history:", err)
		}

		// Удаляем товар из корзины
		_, err = cartCollection.DeleteOne(ctx, bson.M{"user_id": paymentRequest.UserID, "product_id": paymentRequest.ProductID})
		if err != nil {
			log.Println("Failed to remove item from cart:", err)
		}

		// Генерация чека
		receiptPath := utils.GenerateReceipt(
			paymentRequest.TransactionID,
			paymentRequest.UserID,
			paymentRequest.Amount,
			cartItem.ProductID,
			cartItem.ProductName,
			paymentRequest.PaymentMethod, // Теперь берет из запроса
			paymentRequest.PurchaseDate,  // Берет дату покупки
			paymentRequest.Email,         // Email пользователя
			paymentRequest.Address,       // Адрес пользователя
			paymentRequest.Phone,         // Телефон пользователя
		)

		log.Println("Receipt generated:", receiptPath)

		// Отправка чека на email пользователя
		utils.SendReceipt(paymentRequest.UserID, receiptPath)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction processed successfully"))
	} else {
		log.Println("Transaction declined. No history will be saved.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction declined"))
	}
}
