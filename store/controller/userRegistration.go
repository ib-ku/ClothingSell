package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"

	"golang.org/x/crypto/bcrypt"
)

// User модель для хранения информации о пользователе
type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Функция хеширования пароля
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Функция для отправки email
func sendConfirmationEmail(email string) error {
	from := "abylaymoldakhmet@gmail.com"
	password := "byub ifbs izua qezn"
	to := email
	subject := "Account Creation Confirmation"
	body := "Your account has been successfully created. Please log in to access your account."

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Println("Error sending email:", err)
		return err
	}
	return nil
}

// Сохранение пользователя в MongoDB
func saveUserToDB(user User) error {
	// Используем уже существующую коллекцию UserCollection
	_, err := UserCollection.InsertOne(context.TODO(), user)
	return err
}

// Контроллер для обработки регистрации
func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Получаем данные из формы
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Валидация: пароли должны совпадать
		if password != confirmPassword {
			http.Error(w, "Passwords do not match!", http.StatusBadRequest)
			return
		}

		// Хешируем пароль
		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Создаем нового пользователя
		user := User{
			Name:     name,
			Email:    email,
			Password: hashedPassword,
		}

		// Сохраняем пользователя в базе данных
		err = saveUserToDB(user)
		if err != nil {
			http.Error(w, "Error saving user to database", http.StatusInternalServerError)
			return
		}

		// Отправляем email с подтверждением
		err = sendConfirmationEmail(email)
		if err != nil {
			http.Error(w, "Error sending confirmation email", http.StatusInternalServerError)
			return
		}

		// Уведомляем пользователя об успешной регистрации
		fmt.Fprintf(w, "Registration successful! Please check your email to verify your account.")
		return
	}

	// Если не POST-запрос, рендерим форму
	http.ServeFile(w, r, "public/signup.html")
}
