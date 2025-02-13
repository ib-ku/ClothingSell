package services

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Secret key for signing tokens
var jwtSecret = []byte("your_secret_key")

// Claims structure
type Claims struct {
    Email string `json:"email"`
    Role  string `json:"role"`
    jwt.StandardClaims
}


// GenerateJWT generates a new token for a user
func GenerateJWT(email string, role string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // Token valid for 7 days

	// Добавляем email и role в токен
	claims := jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates the token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		log.Println("Token validation failed in ValidateJWT:", err)
		return nil, err
	}

	return claims, nil
}

// Extracts the role from a JWT token
func GetUserRoleFromToken(tokenString string) (string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("failed to parse claims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return "", errors.New("role not found in token")
	}

	return role, nil
}

// ParseJWT extracts email from a JWT token
func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
