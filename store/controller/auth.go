package controller

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Claims structure для JWT
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"` // Добавляем роль
	jwt.StandardClaims
}

// Функция для генерации токена с ролью
func GenerateToken(user User) (string, error) {
	// Время истечения токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: user.Email,
		Role:  user.Role, // Роль пользователя (admin, user и т.д.)
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Создаем новый токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your_secret_key"))
}
