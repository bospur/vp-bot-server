package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Auth проверяет JWT токен в заголовке Authorization
// Использование: http.HandleFunc("/api/admin/...", middleware.Auth(secret, handler))
func Auth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ожидаем заголовок вида: Authorization: Bearer <token>
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "требуется авторизация", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "неверный формат токена", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Проверяем и парсим токен
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			// Проверяем что алгоритм подписи — HMAC (не допускаем подмены)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "недействительный токен", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
