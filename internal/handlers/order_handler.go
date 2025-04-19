package handlers

import (
	"encoding/json"
	"github.com/aseptimu/internal/middlewares"
	"github.com/aseptimu/internal/services"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/aseptimu/internal/utils"
)

type OrderHandler struct {
	orderService services.OrderService
}

func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	if userIDVal == nil {
		http.Error(w, "Пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		http.Error(w, "Не удается определить пользователя", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.LogWithError(r.Context(), "UploadOrder: ошибка чтения тела запроса", err)
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Номер заказа не передан", http.StatusBadRequest)
		return
	}
	if _, err := strconv.ParseUint(orderNumber, 10, 64); err != nil {
		http.Error(w, "Номер заказа должен состоять из цифр", http.StatusUnprocessableEntity)
		return
	}

	code, err := h.orderService.UploadOrder(r.Context(), userID, orderNumber)

	switch code {
	case http.StatusOK:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Заказ уже загружен"))
	case http.StatusAccepted:
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Новый номер заказа принят в обработку"))
	case http.StatusConflict:
		http.Error(w, "Номер заказа уже загружен другим пользователем", http.StatusConflict)
	case http.StatusUnprocessableEntity:
		http.Error(w, "Неверный формат номера заказа", http.StatusUnprocessableEntity)
	default:
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value(middlewares.UserIDKey)
	if userIDVal == nil {
		http.Error(w, "Пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		http.Error(w, "Не удается определить пользователя", http.StatusInternalServerError)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		utils.LogWithError(r.Context(), "GetUserOrders: ошибка при получении заказов пользователя", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Ошибка сериализации JSON", http.StatusInternalServerError)
	}
}
