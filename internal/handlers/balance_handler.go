package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aseptimu/internal/middlewares"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/services"
	"github.com/aseptimu/internal/utils"
	"net/http"
)

type BalanceHandler struct {
	orderService BalanceManager
}

type BalanceManager interface {
	GetUserBalance(ctx context.Context, userID int) (*models.Balance, error)
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
	ListWithdrawals(ctx context.Context, userID int) ([]*models.Withdrawal, error)
}

func NewBalanceHandler(orderService BalanceManager) *BalanceHandler {
	return &BalanceHandler{orderService: orderService}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDVal := r.Context().Value(middlewares.UserIDKey)
	if userIDVal == nil {
		http.Error(w, "Пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		http.Error(w, "Не удаётся определить пользователя", http.StatusInternalServerError)
		return
	}

	balance, err := h.orderService.GetUserBalance(r.Context(), userID)
	if err != nil {
		utils.LogWithError(r.Context(), "GetUserBalance: ошибка при получении баланса", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, "Ошибка сериализации JSON", http.StatusInternalServerError)
	}
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) WithdrawBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDVal := r.Context().Value(middlewares.UserIDKey)
	if userIDVal == nil {
		http.Error(w, "Пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		http.Error(w, "Не удаётся определить пользователя", http.StatusInternalServerError)
		return
	}

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogWithError(r.Context(), "WithdrawBalance: decode error", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	err := h.orderService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidOrderNumber):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, services.ErrInsufficientFunds):
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		default:
			utils.LogWithError(r.Context(), "WithdrawBalance: service error", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := r.Context().Value(middlewares.UserIDKey)
	if uid == nil {
		http.Error(w, "Пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}
	userID, ok := uid.(int)
	if !ok {
		http.Error(w, "Не удаётся определить пользователя", http.StatusInternalServerError)
		return
	}

	list, err := h.orderService.ListWithdrawals(r.Context(), userID)
	if err != nil {
		utils.LogWithError(r.Context(), "GetUserWithdrawals: ошибка получения данных", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "Ошибка сериализации JSON", http.StatusInternalServerError)
	}
}
