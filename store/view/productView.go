package view

import (
	"encoding/json"
	"net/http"
)

func RenderProducts(w http.ResponseWriter, products interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}
