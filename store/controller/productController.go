package controller

import (
	"fmt"
	"net/http"
	"store/model"
	"store/view"
)

func AllProducts(w http.ResponseWriter, r *http.Request) {
	products := model.Products{
		model.Product{ID: 1, Name: "T-Shirt", Price: 52.52},
	}

	fmt.Println("Endpoint Hit: All Products Endpoint")
	view.RenderProducts(w, products)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage Endpoint Hit")
}
