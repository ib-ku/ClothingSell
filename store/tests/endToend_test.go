package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"store/controller"
	"testing"
)

func TestEndToEndProductSearch(t *testing.T) {
	product := map[string]interface{}{
		"id":    1,
		"name":  "Test Product",
		"price": 99.99,
	}
	productData, _ := json.Marshal(product)

	reqAdd := httptest.NewRequest(http.MethodPost, "/product", bytes.NewBuffer(productData))
	reqAdd.Header.Set("Content-Type", "application/json")
	recAdd := httptest.NewRecorder()

	controller.HandleProductPostRequest(recAdd, reqAdd)
	if recAdd.Code != http.StatusOK {
		t.Fatalf("Failed to add product. Got status %v, response: %s", recAdd.Code, recAdd.Body.String())
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/allproducts", nil)
	recGet := httptest.NewRecorder()

	controller.AllProducts(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Fatalf("Failed to fetch products. Got status %v, response: %s", recGet.Code, recGet.Body.String())
	}

	var products []map[string]interface{}
	err := json.Unmarshal(recGet.Body.Bytes(), &products)
	if err != nil {
		t.Fatalf("Failed to parse response JSON. Error: %v", err)
	}

	found := false
	for _, p := range products {
		if p["id"] == float64(product["id"].(int)) &&
			p["name"] == product["name"] &&
			p["price"] == product["price"] {
			found = true
			break
		}
	}

	if !found {
		t.Error("Added product not found in the fetched products")
	}
}
