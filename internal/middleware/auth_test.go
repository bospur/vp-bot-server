package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-server/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

// makeToken создаёт JWT токен для тестов
func makeToken(secret string, expired bool) string {
	exp := time.Now().Add(24 * time.Hour)
	if expired {
		exp = time.Now().Add(-1 * time.Hour)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   float64(1),
		"clinic_id": float64(1),
		"role":      "admin",
		"exp":       exp.Unix(),
	})

	str, _ := token.SignedString([]byte(secret))
	return str
}

// okHandler — заглушка следующего хендлера
func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAuth_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	middleware.Auth(testSecret, okHandler)(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ожидали 401, получили %d", rr.Code)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidToken")
	rr := httptest.NewRecorder()

	middleware.Auth(testSecret, okHandler)(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ожидали 401, получили %d", rr.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	token := makeToken(testSecret, false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware.Auth(testSecret, okHandler)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ожидали 200, получили %d", rr.Code)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	token := makeToken(testSecret, true)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware.Auth(testSecret, okHandler)(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ожидали 401, получили %d", rr.Code)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	token := makeToken("other-secret", false)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware.Auth(testSecret, okHandler)(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("ожидали 401, получили %d", rr.Code)
	}
}
