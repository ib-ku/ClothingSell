package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username   string `json:"username" bson:"username"`
	Email      string `json:"email" bson:"email"`
	Password   string `json:"password" bson:"password"`
	Role       string `json:"role" bson:"role"`
	IsVerified bool   `json:"is_verified" bson:"is_verified"` // Добавлен флаг подтверждения email
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func sendConfirmationEmail(email string, confirmationLink string) error {
	from := "abylaymoldakhmet@gmail.com"
	password := "byub ifbs izua qezn"
	to := email
	subject := "Account Creation Confirmation"
	body := fmt.Sprintf(`
    <html>
    <body>
      <h2>Welcome to Our Service!</h2>
      <p>Your account has been successfully created. Click the link below to confirm your registration:</p>
      <a href="%s" style="display: inline-block; padding: 10px 15px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px;">
        Confirm Registration
      </a>
      <p>If you did not request this, you can ignore this email.</p>
    </body>
    </html>
  `, confirmationLink)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Println("Error sending email:", err)
		return err
	}
	return nil
}

func saveUserToDB(user User) error {
	_, err := userCollection.InsertOne(context.TODO(), user)
	return err
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			http.Error(w, "Passwords do not match!", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		user := User{
			Username:   username,
			Email:      email,
			Password:   string(hashedPassword),
			Role:       "user",
			IsVerified: false, // Новый пользователь не подтвержден
		}

		_, err = userCollection.InsertOne(context.TODO(), user)
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}

		// Отправляем email с подтверждением
		confirmationLink := fmt.Sprintf("http://localhost:8085/confirm?email=%s", email)
		sendConfirmationEmail(email, confirmationLink)

		http.Redirect(w, r, "/verify.html", http.StatusSeeOther)
	}
}

// Controller to handle account confirmation
func Confirm(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Invalid confirmation link", http.StatusBadRequest)
		return
	}

	// Find the user in the database and update their status
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"is_verified": true}}

	_, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Error confirming account", http.StatusInternalServerError)
		return
	}

	// Redirect the user to the main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
