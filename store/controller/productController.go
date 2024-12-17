package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"store/model"
	"store/view"
)

var products = model.Products{
	model.Product{ID: 1, Name: "T-Shirt", Price: 52.52},
}

func AllProducts(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Endpoint Hit: All Products Endpoint")
	view.RenderProducts(w, products)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint Hit")
}

func HandleProductPostRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
	}

	var requestProduct struct {
		ID    *int     `json:"id"`
		Name  string   `json:"name"`
		Price *float64 `json:"price"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestProduct)
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

	if requestProduct.ID == nil || requestProduct.Name == "" || requestProduct.Price == nil {
		response := map[string]string{
			"status":  "fail",
			"message": "Invalid product data: Id, Name, and Price are required",
		}
		w.Header().Set("Content-Type", "applocation/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	newProduct := model.Product{
		ID:    *requestProduct.ID,
		Name:  requestProduct.Name,
		Price: *requestProduct.Price,
	}
	SetProductAfterHandling(newProduct)

	fmt.Println("Received information:", "ID: \n", requestProduct.ID, "Name: \n", requestProduct.Name, "Price: \n", requestProduct.Price)

	response := map[string]string{
		"status":  "success",
		"message": "Product data successfully received",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func SetProductAfterHandling(product model.Product) {
	products = append(products, product)
}
