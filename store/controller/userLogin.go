package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Функция для проверки пароля
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Контроллер для авторизации (Login)
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Получаем данные из формы
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Валидация данных
		if email == "" || password == "" {
			http.Error(w, "Email and password are required!", http.StatusBadRequest)
			return
		}

		// Поиск пользователя в базе данных
		var user User
		err := UserCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Проверка пароля
		if !checkPasswordHash(password, user.Password) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
<<<<<<< HEAD
=======

>>>>>>> 314048de93309a1a2d46aa857dde6c102def36ee
		// Успешный вход
		response := map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Welcome back, %s!", user.Name),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Если не POST-запрос, рендерим форму
	http.ServeFile(w, r, "public/login.html")
}
