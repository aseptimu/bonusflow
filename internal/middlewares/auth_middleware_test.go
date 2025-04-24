package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuthMiddleware(t *testing.T) {
	secret := "mysecret"
	mw := JWTAuthMiddleware(secret)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.Context().Value(UserIDKey)
		if uid == nil {
			http.Error(w, "no user", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "user:%v", uid)
	})
	handler := mw(next)

	tests := []struct {
		name       string
		authHeader string
		wantCode   int
		wantBody   string
	}{
		{"no header", "", http.StatusUnauthorized, "Unauthorized"},
		{"bad format", "BearerToken", http.StatusUnauthorized, "Unauthorized"},
		{"wrong scheme", "Basic abc.def", http.StatusUnauthorized, "Unauthorized"},
		{"malformed token", "Bearer abc", http.StatusUnauthorized, "Unauthorized"},
		{"invalid signature", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 1})
			s, _ := tok.SignedString([]byte("wrong"))
			return "Bearer " + s
		}(), http.StatusUnauthorized, "Unauthorized"},
		{"missing user_id", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"foo": "bar"})
			s, _ := tok.SignedString([]byte(secret))
			return "Bearer " + s
		}(), http.StatusUnauthorized, "Unauthorized"},
		{"non-numeric user_id", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": "notnum"})
			s, _ := tok.SignedString([]byte(secret))
			return "Bearer " + s
		}(), http.StatusUnauthorized, "Unauthorized"},
		{"valid token", func() string {
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 42})
			s, _ := tok.SignedString([]byte(secret))
			return "Bearer " + s
		}(), http.StatusOK, "user:42"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantCode {
				t.Errorf("%s: got status %d, want %d", tc.name, res.StatusCode, tc.wantCode)
			}
			body := strings.TrimSpace(w.Body.String())
			if !strings.Contains(body, tc.wantBody) {
				t.Errorf("%s: body %q does not contain %q", tc.name, body, tc.wantBody)
			}
		})
	}
}
