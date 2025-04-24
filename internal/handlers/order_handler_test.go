package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aseptimu/internal/middlewares"
	"github.com/aseptimu/internal/models"
)

type mockOrderManager struct {
	uploadFunc func(ctx context.Context, userID int, orderNumber string) (int, error)
	getFunc    func(ctx context.Context, userID int) ([]*models.Order, error)
}

func (m *mockOrderManager) UploadOrder(ctx context.Context, userID int, orderNumber string) (int, error) {
	return m.uploadFunc(ctx, userID, orderNumber)
}

func (m *mockOrderManager) GetUserOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	return m.getFunc(ctx, userID)
}

func TestUploadOrder(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		body         io.Reader
		uploadCode   int
		uploadErr    error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "unauthenticated",
			ctx:          context.Background(),
			body:         strings.NewReader("123"),
			wantStatus:   http.StatusUnauthorized,
			wantContains: "Пользователь не аутентифицирован",
		},
		{
			name:         "invalid userID type",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, "x"),
			body:         strings.NewReader("123"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Не удается определить пользователя",
		},
		{
			name:         "read body error",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         io.NopCloser(strings.NewReader("fail")),
			uploadCode:   0,
			uploadErr:    nil,
			wantStatus:   http.StatusUnprocessableEntity,
			wantContains: "Номер заказа должен состоять из цифр\n",
		},
		{
			name:         "empty order number",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("   "),
			wantStatus:   http.StatusBadRequest,
			wantContains: "Номер заказа не передан",
		},
		{
			name:         "non-numeric order number",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("abc"),
			wantStatus:   http.StatusUnprocessableEntity,
			wantContains: "Номер заказа должен состоять из цифр",
		},
		{
			name:         "already uploaded",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("123"),
			uploadCode:   http.StatusOK,
			wantStatus:   http.StatusOK,
			wantContains: "Заказ уже загружен",
		},
		{
			name:         "accepted new order",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("456"),
			uploadCode:   http.StatusAccepted,
			wantStatus:   http.StatusAccepted,
			wantContains: "Новый номер заказа принят в обработку",
		},
		{
			name:         "conflict other user",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("789"),
			uploadCode:   http.StatusConflict,
			wantStatus:   http.StatusConflict,
			wantContains: "Номер заказа уже загружен другим пользователем",
		},
		{
			name:         "unprocessable format",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("000"),
			uploadCode:   http.StatusUnprocessableEntity,
			wantStatus:   http.StatusUnprocessableEntity,
			wantContains: "Неверный формат номера заказа",
		},
		{
			name:         "internal error",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			body:         strings.NewReader("321"),
			uploadCode:   999,
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Внутренняя ошибка сервера",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockOrderManager{
				uploadFunc: func(ctx context.Context, userID int, orderNumber string) (int, error) {
					return tc.uploadCode, tc.uploadErr
				},
			}
			handler := NewOrderHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/orders", nil)
			req = req.WithContext(tc.ctx)
			// Override body if provided
			if tc.body != nil {
				rc := io.NopCloser(tc.body)
				req.Body = rc
			}
			w := httptest.NewRecorder()

			handler.UploadOrder(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("%s: got status %d, want %d", tc.name, res.StatusCode, tc.wantStatus)
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			bodyStr := buf.String()
			if !strings.Contains(bodyStr, tc.wantContains) {
				t.Errorf("%s: response body %q does not contain %q", tc.name, bodyStr, tc.wantContains)
			}
		})
	}
}

func TestGetUserOrders(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		orders       []*models.Order
		getErr       error
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
			wantContains: "Не удается определить пользователя",
		},
		{
			name:         "service error",
			ctx:          context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			getErr:       errors.New("fail"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Ошибка сервера",
		},
		{
			name:       "no orders",
			ctx:        context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			orders:     []*models.Order{},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "success",
			ctx:        context.WithValue(context.Background(), middlewares.UserIDKey, 1),
			orders:     []*models.Order{{ID: 1, Number: "111", Status: "NEW"}},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockOrderManager{
				getFunc: func(ctx context.Context, userID int) ([]*models.Order, error) {
					return tc.orders, tc.getErr
				},
			}
			handler := NewOrderHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/orders", nil)
			req = req.WithContext(tc.ctx)
			w := httptest.NewRecorder()

			handler.GetUserOrders(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.wantStatus {
				t.Errorf("%s: got status %d, want %d", tc.name, res.StatusCode, tc.wantStatus)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)
			bodyStr := buf.String()
			if tc.wantContains != "" && !strings.Contains(bodyStr, tc.wantContains) {
				t.Errorf("%s: response body %q does not contain %q", tc.name, bodyStr, tc.wantContains)
			}
			if tc.name == "success" {
				expected, _ := json.Marshal(tc.orders)
				expectedStr := string(expected) + "\n"
				if bodyStr != expectedStr {
					t.Errorf("success: body %q, want %q", bodyStr, expectedStr)
				}
			}
		})
	}
}
