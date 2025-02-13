package utils

import (
	"log"

	"gopkg.in/gomail.v2"
)

func SendReceipt(email, filePath string) {
	log.Println("Preparing to send receipt to:", email)
	log.Println("Attachment file path:", filePath)

	m := gomail.NewMessage()
	m.SetHeader("From", "your_email@example.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Payment Receipt")
	m.SetBody("text/plain", "Thank you for your payment. Please find the receipt attached.")
	m.Attach(filePath)

	d := gomail.NewDialer("smtp.gmail.com", 587, "hog123von@gmail.com", "byrvzckrqigqpfii")

	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
	} else {
		log.Println("Email sent successfully to:", email)
	}
}
