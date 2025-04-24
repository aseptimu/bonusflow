package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/internal/middlewares"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/services"
)

type mockBalanceManager struct {
	getBalanceFunc      func(ctx context.Context, userID int) (*models.Balance, error)
	withdrawFunc        func(ctx context.Context, userID int, orderNumber string, sum float64) error
	listWithdrawalsFunc func(ctx context.Context, userID int) ([]*models.Withdrawal, error)
}

func (m *mockBalanceManager) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	return m.getBalanceFunc(ctx, userID)
}

func (m *mockBalanceManager) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	return m.withdrawFunc(ctx, userID, orderNumber, sum)
}

func (m *mockBalanceManager) ListWithdrawals(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
	return m.listWithdrawalsFunc(ctx, userID)
}

func TestGetUserBalance(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		balance      *models.Balance
		err          error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "unauthenticated",
			ctx:          context.Background(),
			wantStatus:   http.StatusUnauthorized,
			wantContains: "Пользователь не аутентифицирован",
		},
		{
			name:         "invalid userID type",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, "x"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Не удаётся определить пользователя",
		},
		{
			name:         "service error",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			err:          errors.New("fail"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Внутренняя ошибка сервера",
		},
		{
			name:       "success",
			ctx:        context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			balance:    &models.Balance{Current: 10.5, Withdrawn: 2},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockBalanceManager{
				getBalanceFunc: func(ctx context.Context, userID int) (*models.Balance, error) {
					return tc.balance, tc.err
				},
			}
			handler := NewBalanceHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/balance", nil)
			req = req.WithContext(tc.ctx)
			w := httptest.NewRecorder()

			handler.GetUserBalance(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("%s: got status %d, want %d", tc.name, res.StatusCode, tc.wantStatus)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			bodyStr := buf.String()
			if tc.wantContains != "" && !strings.Contains(bodyStr, tc.wantContains) {
				t.Errorf("%s: body %q does not contain %q", tc.name, bodyStr, tc.wantContains)
			}
			if tc.name == "success" {
				expected, _ := json.Marshal(tc.balance)
				expectedStr := string(expected) + "\n"
				if bodyStr != expectedStr {
					t.Errorf("success: body %q, want %q", bodyStr, expectedStr)
				}
			}
		})
	}
}

func TestWithdrawBalance(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		body         string
		withdrawErr  error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "unauthenticated",
			ctx:          context.Background(),
			body:         `{"order":"1","sum":5}`,
			wantStatus:   http.StatusUnauthorized,
			wantContains: "Пользователь не аутентифицирован",
		},
		{
			name:         "invalid userID type",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, "x"),
			body:         `{"order":"1","sum":5}`,
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Не удаётся определить пользователя",
		},
		{
			name:         "bad JSON",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         `{`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "Неверный формат запроса",
		},
		{
			name:         "invalid order number",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         `{"order":"abc","sum":5}`,
			withdrawErr:  services.ErrInvalidOrderNumber,
			wantStatus:   http.StatusUnprocessableEntity,
			wantContains: services.ErrInvalidOrderNumber.Error(),
		},
		{
			name:         "insufficient funds",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         `{"order":"1","sum":1000}`,
			withdrawErr:  services.ErrInsufficientFunds,
			wantStatus:   http.StatusPaymentRequired,
			wantContains: services.ErrInsufficientFunds.Error(),
		},
		{
			name:         "service error",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         `{"order":"1","sum":5}`,
			withdrawErr:  errors.New("fail"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Внутренняя ошибка сервера",
		},
		{
			name:       "success",
			ctx:        context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:       `{"order":"1","sum":5}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockBalanceManager{
				withdrawFunc: func(ctx context.Context, userID int, orderNumber string, sum float64) error {
					return tc.withdrawErr
				},
			}
			handler := NewBalanceHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/withdraw", strings.NewReader(tc.body))
			req = req.WithContext(tc.ctx)
			w := httptest.NewRecorder()

			handler.WithdrawBalance(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("%s: got status %d, want %d", tc.name, res.StatusCode, tc.wantStatus)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			bodyStr := buf.String()
			if tc.wantContains != "" && !strings.Contains(bodyStr, tc.wantContains) {
				t.Errorf("%s: body %q does not contain %q", tc.name, bodyStr, tc.wantContains)
			}
		})
	}
}

func TestGetUserWithdrawals(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		list         []*models.Withdrawal
		err          error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "unauthenticated",
			ctx:          context.Background(),
			wantStatus:   http.StatusUnauthorized,
			wantContains: "Пользователь не аутентифицирован",
		},
		{
			name:         "invalid userID type",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, "x"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Не удаётся определить пользователя",
		},
		{
			name:         "service error",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			err:          errors.New("fail"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Внутренняя ошибка сервера",
		},
		{
			name:       "no withdrawals",
			ctx:        context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			list:       []*models.Withdrawal{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "success",
			ctx:        context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			list:       []*models.Withdrawal{{Order: "1", Sum: 5}},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockBalanceManager{
				listWithdrawalsFunc: func(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
					return tc.list, tc.err
				},
			}
			handler := NewBalanceHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
			req = req.WithContext(tc.ctx)
			w := httptest.NewRecorder()

			handler.GetUserWithdrawals(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("%s: got status %d, want %d", tc.name, res.StatusCode, tc.wantStatus)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			bodyStr := buf.String()
			if tc.wantContains != "" && !strings.Contains(bodyStr, tc.wantContains) {
				t.Errorf("%s: body %q does not contain %q", tc.name, bodyStr, tc.wantContains)
			}
			if tc.name == "success" {
				expected, _ := json.Marshal(tc.list)
				expectedStr := string(expected) + "\n"
				if bodyStr != expectedStr {
					t.Errorf("success: body %q, want %q", bodyStr, expectedStr)
				}
			}
		})
	}
}
