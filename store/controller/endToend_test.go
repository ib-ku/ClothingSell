package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndDeleteProduct(t *testing.T) {
	product := map[string]interface{}{
		"id":    101,
		"name":  "New Product",
		"price": 99.99,
		"image": "product_image.png",
	}

	data, _ := json.Marshal(product)

	req, err := http.NewRequest("POST", "/products", bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandleProductPostRequest)
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

	deleteReq, err := http.NewRequest("DELETE", "/products", bytes.NewBuffer([]byte(`{"id": 101}`)))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	deleteHandler := http.HandlerFunc(DeleteProductByID)
	deleteHandler.ServeHTTP(rr, deleteReq)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, status)
	}
}
