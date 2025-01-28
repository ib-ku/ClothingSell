package controller

import (
	"encoding/json"
	"net/http"
	"store/email"
)

func SendPromotionalEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	to := requestBody["to"]
	subject := requestBody["subject"]
	body := requestBody["body"]

	if to == "" || subject == "" || body == "" {
		http.Error(w, "Missing required fields: to, subject, or body", http.StatusBadRequest)
		return
	}

	err := email.SendEmail("ibragimtop1@gmail.com", "emlx wxgk ajik qiyl", to, subject, body, "")
	if err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}
