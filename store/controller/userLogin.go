package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/smtp"
	"store/services"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var otpStore = make(map[string]string) // email → OTP

func sendEmail(to string, body string) error {
	from := "your_email@gmail.com"
	password := "your_password" // Используй переменные окружения в продакшене!
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)
	msg := []byte("Subject: Verification Code\n\n" + body)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return err
	}
	fmt.Println("Email sent successfully to:", to)
	return nil
}

func SendOTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	otp := fmt.Sprintf("%06d", rand.Intn(1000000)) // Генерируем 6-значный код
	otpStore[email] = otp

	sendEmail(email, fmt.Sprintf("Your OTP code is: %s", otp))
	fmt.Fprintln(w, "OTP sent to your email")
}

func VerifyOTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	otp := r.FormValue("otp")

	if otpStore[email] == otp {
		delete(otpStore, email)
		fmt.Fprintln(w, "OTP verified, login successful")
	} else {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
	}
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var user User
		err := userCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
		if err != nil || !checkPasswordHash(password, user.Password) {
			http.Redirect(w, r, "/login.html?error=Invalid email or password", http.StatusSeeOther)
			return
		}

		if !user.IsVerified { // Проверяем, подтвержден ли email
			http.Redirect(w, r, "/login.html?error=Email not verified", http.StatusSeeOther)
			return
		}

		token, err := services.GenerateJWT(email, user.Role)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + token,
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: GetUser called")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		cookie, err := r.Cookie("Authorization")
		if err != nil {
			fmt.Println("DEBUG: No token in headers or cookies")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		authHeader = cookie.Value
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := services.ParseJWT(token)
	if err != nil {
		fmt.Println("DEBUG: Invalid token:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"email": claims.Email,
		"role":  claims.Role, // ✅ Теперь передаем роль
	}

	fmt.Println("DEBUG: Returning user data =", response) // Проверка в консоли сервера
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AssignRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	role := r.FormValue("role")

	_, err := userCollection.UpdateOne(
		context.TODO(),
		bson.M{"email": email},
		bson.M{"$set": bson.M{"role": role}},
	)

	if err != nil {
		http.Error(w, "Error updating role", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User %s now has role %s", email, role)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Сразу удаляем
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}
