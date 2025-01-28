package middleware

import (
	"net/http"
	"store/services"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Извлечение токена из заголовка
		token := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := services.ValidateJWT(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		tokenString := r.Header.Get("Authorization")

		// Если токен отсутствует
		if tokenString == "" {
			http.Error(w, "Forbidden: No token provided", http.StatusForbidden)
			return
		}

		// Проверка токена и извлечение роли из токена
		role, err := services.GetUserRoleFromToken(tokenString) // Функция для получения роли из токена
		if err != nil || role != "admin" {
			http.Error(w, "Forbidden: User is not an admin", http.StatusForbidden)
			return
		}

		// Если роль admin, продолжаем выполнение
		next.ServeHTTP(w, r)
	})
}
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		role, err := services.GetUserRoleFromToken(cookie.Value)
		if err != nil || role != "admin" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
