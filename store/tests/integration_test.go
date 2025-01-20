package tests

import (
	"net/http"
	"net/http/httptest"
	"store/controller"
	"testing"
)

func TestAllProducts(t *testing.T) {
	req, err := http.NewRequest("GET", "/allproducts", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.AllProducts)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rec.Code)
	}
}
