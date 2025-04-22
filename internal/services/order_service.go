package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aseptimu/internal/repository"
	"net/http"
	"regexp"
	"strings"

	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/utils"
)

var (
	ErrInvalidOrderFormat          = errors.New("неверный формат номера заказа")
	ErrOrderAlreadyExistsDifferent = errors.New("номер заказа уже загружен другим пользователем")
	ErrOrderAlreadyExistsSame      = errors.New("номер заказа уже загружен этим пользователем")
)

type OrderService struct {
	repo             *repository.OrderRepository
	accrualSystemURL string
}

func NewOrderService(repo *repository.OrderRepository, accrualSystemURL string) *OrderService {
	return &OrderService{repo, accrualSystemURL}
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int) ([]*models.Order, error) {
	return s.repo.GetOrdersByUserID(ctx, userID)
}

func (s *OrderService) UploadOrder(ctx context.Context, userID int, orderNumber string) (int, error) {
	matched, err := regexp.MatchString(`^\d+$`, orderNumber)
	if err != nil || !matched {
		utils.LogWithError(ctx, "UploadOrder: номер заказа должен содержать только цифры", err)
		return http.StatusUnprocessableEntity, ErrInvalidOrderFormat
	}

	if !validateLuhn(orderNumber) {
		utils.LogWithError(ctx, "UploadOrder: неверный формат номера заказа", nil)
		return http.StatusUnprocessableEntity, ErrInvalidOrderFormat
	}

	existingOrder, err := s.repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		utils.LogWithError(ctx, "UploadOrder: ошибка при проверке существования заказа", err)
		return http.StatusInternalServerError, err
	}
	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return http.StatusOK, ErrOrderAlreadyExistsSame
		} else {
			return http.StatusConflict, ErrOrderAlreadyExistsDifferent
		}
	}

	newOrder := &models.Order{
		Number: orderNumber,
		UserID: userID,
		Status: "NEW",
	}

	if err := s.repo.CreateOrder(ctx, newOrder); err != nil {
		utils.LogWithError(ctx, "UploadOrder: ошибка создания заказа", err)
		return http.StatusInternalServerError, err
	}

	return http.StatusAccepted, nil
}

func validateLuhn(number string) bool {
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}
	return sum%10 == 0
}

type AccrualSystemResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

var ErrTooManyRequests = errors.New("too many requests to accrual server")

func (s *OrderService) fetchAndStoreAccrual(ctx context.Context, number string) error {
	url := fmt.Sprintf("%s/api/orders/%s", strings.TrimRight(s.accrualSystemURL, "/"), number)
	resp, err := http.Get(url)
	if err != nil {
		utils.LogWithError(ctx, "fetchAccrual: не удалось выполнить запрос", err)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		ext := AccrualSystemResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&ext); err != nil {
			utils.LogWithError(ctx, "fetchAccrual: ошибка парсинга JSON", err)
			return err
		}

		var internalStatus string
		switch ext.Status {
		case "REGISTERED", "PROCESSING":
			internalStatus = "PROCESSING"
		case "INVALID":
			internalStatus = "INVALID"
		case "PROCESSED":
			internalStatus = "PROCESSED"
		}

		if err := s.repo.UpdateOrder(ctx, number, internalStatus, ext.Accrual); err != nil {
			utils.LogWithError(ctx, "fetchAccrual: ошибка обновления заказа в БД", err)
			return err
		}

	case http.StatusTooManyRequests:
		utils.LogWithError(ctx, "fetchAccrual: слишком частые запросы", err)
		return ErrTooManyRequests
	}
	return nil
}
