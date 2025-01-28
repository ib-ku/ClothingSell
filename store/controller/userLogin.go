package controller

import (
	"context"
	"net/http"
	"store/services"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "admin@gmail.com" && password == "admin" {
			http.Redirect(w, r, "/admin.html", http.StatusSeeOther)
			return
		}

		if email == "" || password == "" {
			http.Redirect(w, r, "/login?error=Email and password are required!", http.StatusSeeOther)
			return
		}

		var user User
		err := userCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
		if err != nil {
			http.Redirect(w, r, "/login?error=Invalid email or password", http.StatusSeeOther)
			return
		}

		if !checkPasswordHash(password, user.Password) {
			http.Redirect(w, r, "/login?error=Invalid email or password", http.StatusSeeOther)
			return
		}

		token, err := services.GenerateJWT(email, "user")
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

		if user.Role == "admin" {
			http.Redirect(w, r, "/admin.html", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return

	}
	http.ServeFile(w, r, "static/login.html")
}
