package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// Claims — данные пользователя из JWT, кладутся в контекст запроса
type Claims struct {
	UserID   int
	ClinicID int
	Role     string
}

// ClaimsFromContext извлекает Claims из контекста запроса
func ClaimsFromContext(r *http.Request) *Claims {
	v := r.Context().Value(ClaimsKey)
	if v == nil {
		return nil
	}
	return v.(*Claims)
}

// Auth проверяет JWT токен и кладёт claims в контекст запроса
func Auth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "недействительный токен", http.StatusUnauthorized)
			return
		}

		mapClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "недействительный токен", http.StatusUnauthorized)
			return
		}

		claims := &Claims{
			UserID:   int(mapClaims["user_id"].(float64)),
			ClinicID: int(mapClaims["clinic_id"].(float64)),
			Role:     mapClaims["role"].(string),
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}
