package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllProducts(t *testing.T) {
	req, err := http.NewRequest("GET", "/products?page=1&sort=price", nil)
	if err != nil {
		t.Fatal(err)
	}

	InitializeProduct(client)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AllProducts)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("Expected success status, got %v", response["status"])
	}
}
