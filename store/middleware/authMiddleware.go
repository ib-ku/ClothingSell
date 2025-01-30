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
			cookie, err := r.Cookie("Authorization")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			authHeader = cookie.Value
		}

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
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			cookie, err := r.Cookie("Authorization")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Forbidden: No token provided", http.StatusForbidden)
				return
			}
			authHeader = cookie.Value
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		role, err := services.GetUserRoleFromToken(token)
		if err != nil {
			http.Error(w, "Forbidden: User is not an admin", http.StatusForbidden)
			return
		}
		if role != "admin" {
			http.Error(w, "Forbidden: User is not an admin", http.StatusForbidden)
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
