package controller

import (
	"testing"
)

func TestValidateProductFields(t *testing.T) {
	tests := []struct {
		name     string
		reqData  map[string]interface{}
		expected map[string]string
		valid    bool
	}{
		{
			name: "Valid data",
			reqData: map[string]interface{}{
				"id":    1,
				"name":  "Product1",
				"price": 10.5,
			},
			expected: nil, 
			valid:    true,
		},
		{
			name: "Invalid id",
			reqData: map[string]interface{}{
				"id":    -1,
				"name":  "Product1",
				"price": 10.5,
			},
			expected: map[string]string{
				"status":  "fail",
				"message": "'id' must be a positive number",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, valid := validateProductFields(tt.reqData, []string{"id", "name", "price"})
			if valid != tt.valid {
				t.Errorf("expected valid = %v, got %v", tt.valid, valid)
			}
			if resp != nil {
				if resp["status"] != tt.expected["status"] {
					t.Errorf("expected status = %v, got %v", tt.expected["status"], resp["status"])
				}
				if resp["message"] != tt.expected["message"] {
					t.Errorf("expected message = %v, got %v", tt.expected["message"], resp["message"])
				}
			}
		})
	}
}
