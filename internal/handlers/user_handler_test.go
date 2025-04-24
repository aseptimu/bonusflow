package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/internal/services"
)

type mockUserManager struct {
	registerFunc func(ctx context.Context, login, password string) (string, error)
	loginFunc    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockUserManager) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFunc(ctx, login, password)
}

func (m *mockUserManager) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFunc(ctx, login, password)
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		regResult    string
		regError     error
		wantStatus   int
		wantAuth     string
		wantContains string
	}{
		{
			name:         "invalid JSON",
			body:         "{invalid}",
			regResult:    "",
			regError:     nil,
			wantStatus:   http.StatusBadRequest,
			wantContains: "Error decoding json",
		},
		{
			name:         "user exists",
			body:         `{"login":"a","password":"b"}`,
			regResult:    "",
			regError:     services.ErrorUserAlreadyExists,
			wantStatus:   http.StatusConflict,
			wantContains: services.ErrorUserAlreadyExists.Error(),
		},
		{
			name:         "other error",
			body:         `{"login":"a","password":"b"}`,
			regResult:    "",
			regError:     errors.New("fail"),
			wantStatus:   http.StatusBadRequest,
			wantContains: "fail",
		},
		{
			name:       "success",
			body:       `{"login":"user","password":"pass"}`,
			regResult:  "tok123",
			regError:   nil,
			wantStatus: http.StatusOK,
			wantAuth:   "Bearer tok123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockUserManager{
				registerFunc: func(ctx context.Context, login, password string) (string, error) {
					return tc.regResult, tc.regError
				},
			}
			handler := NewUserHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			handler.RegisterUser(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("got status %d, want %d", res.StatusCode, tc.wantStatus)
			}

			if tc.wantAuth != "" {
				got := res.Header.Get("Authorization")
				if got != tc.wantAuth {
					t.Errorf("got Authorization %q, want %q", got, tc.wantAuth)
				}
			}

			if tc.wantContains != "" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(res.Body)
				if !strings.Contains(buf.String(), tc.wantContains) {
					t.Errorf("response body %q does not contain %q", buf.String(), tc.wantContains)
				}
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		loginResult  string
		loginError   error
		wantStatus   int
		wantAuth     string
		wantContains string
	}{
		{
			name:         "invalid JSON",
			body:         "{invalid}",
			loginResult:  "",
			loginError:   nil,
			wantStatus:   http.StatusBadRequest,
			wantContains: "Error decoding json",
		},
		{
			name:         "unauthorized",
			body:         `{"login":"u","password":"p"}`,
			loginResult:  "",
			loginError:   errors.New("bad creds"),
			wantStatus:   http.StatusUnauthorized,
			wantContains: "bad creds",
		},
		{
			name:        "success",
			body:        `{"login":"user","password":"pass"}`,
			loginResult: "tok456",
			loginError:  nil,
			wantStatus:  http.StatusOK,
			wantAuth:    "Bearer tok456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockUserManager{
				loginFunc: func(ctx context.Context, login, password string) (string, error) {
					return tc.loginResult, tc.loginError
				},
			}
			handler := NewUserHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			handler.LoginUser(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("got status %d, want %d", res.StatusCode, tc.wantStatus)
			}

			if tc.wantAuth != "" {
				got := res.Header.Get("Authorization")
				if got != tc.wantAuth {
					t.Errorf("got Authorization %q, want %q", got, tc.wantAuth)
				}
			}

			if tc.wantContains != "" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(res.Body)
				if !strings.Contains(buf.String(), tc.wantContains) {
					t.Errorf("response body %q does not contain %q", buf.String(), tc.wantContains)
				}
			}
		})
	}
}
