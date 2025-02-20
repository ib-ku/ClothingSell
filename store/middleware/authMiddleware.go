package middleware

import (
	"log"
	"net/http"
	"store/services"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		if err != nil || cookie.Value == "" {
			log.Println("Missing or invalid token:", err)
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(cookie.Value, "Bearer ")

		claims, err := services.ValidateJWT(token)
		if err != nil {
			log.Println("Token validation failed:", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Example role check:
		if claims.Role != "user" && claims.Role != "admin" {
			log.Println("Access denied: unauthorized role")
			http.Error(w, "Forbidden: unauthorized role", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Authorization")
		if err != nil || cookie.Value == "" {
			log.Println("Admin access denied: No token")
			http.NotFound(w, r)
			return
		}

		// Логируем сам токен
		log.Println("Admin access token:", cookie.Value)

		role, err := services.GetUserRoleFromToken(cookie.Value)
		if err != nil {
			log.Println("Failed to extract role from token:", err)
			http.NotFound(w, r)
			return
		}

		// Логируем, какая роль была найдена
		log.Println("User role extracted from token:", role)

		if role != "admin" {
			log.Println("Access denied: User is not an admin")
			http.NotFound(w, r)
			return
		}

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
