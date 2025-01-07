package email

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"gopkg.in/gomail.v2"
)

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func SendEmail(from string, password string, to string, subject string, body string, attachment string) error {
	if !isValidEmail(to) {
		return fmt.Errorf("Invalid email format for recipient: %s", to)
	}

	if from == "" || password == "" {
		log.Println("Email address or password is not set.")
		return fmt.Errorf("email address or password not set")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	if attachment != "" {
		m.Attach(attachment)
	}

	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %s: %s", to, err)
		return err
	}

	log.Printf("Email successfully sent to %s", to)
	return nil
}

func HandleEmailRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var emailData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&emailData)
	if err != nil {
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	to, ok := emailData["to"].(string)
	if !ok || to == "" {
		http.Error(w, "Recipient email is required", http.StatusBadRequest)
		return
	}

	subject, ok := emailData["subject"].(string)
	if !ok || subject == "" {
		http.Error(w, "Subject is required", http.StatusBadRequest)
		return
	}

	body, ok := emailData["body"].(string)
	if !ok || body == "" {
		http.Error(w, "Body is required", http.StatusBadRequest)
		return
	}

	attachment, _ := emailData["attachment"].(string)

	from := "ibragimtop1@gmail.com"
	password := "emlx wxgk ajik qiyl"
	err = SendEmail(from, password, to, subject, body, attachment)
	if err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "Email sent successfully!",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
