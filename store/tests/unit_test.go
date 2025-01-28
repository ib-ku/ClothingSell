package tests

import (
	"testing"
)

// Example function to calculate product discount
func CalculateDiscount(price int, discountPercent int) int {
	return price - (price * discountPercent / 100)
}

// Test for CalculateDiscount function
func TestCalculateDiscount(t *testing.T) {
	price := 100
	discountPercent := 10
	discountedPrice := CalculateDiscount(price, discountPercent)
	expectedPrice := 90

	if discountedPrice != expectedPrice {
		t.Errorf("Incorrect discounted price. Expected: %d, Got: %d", expectedPrice, discountedPrice)
	}
}
