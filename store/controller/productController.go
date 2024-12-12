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

func SetProduct(w http.ResponseWriter, r *http.Request) {
	var newProduct model.Product

	err := json.NewDecoder(r.Body).Decode(&newProduct)

	if err != nil {
		http.Error(w, "Invalid Product data", http.StatusBadRequest)
		return
	}

	products = append(products, newProduct)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)

}
