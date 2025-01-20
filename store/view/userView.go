package view

import (
	"encoding/json"
	"net/http"
)

func RenderUsers(w http.ResponseWriter, users interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
