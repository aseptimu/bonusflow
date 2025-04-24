package services_test

import (
	"context"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/services"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type mockOrderRepo struct {
	order        *models.Order
	errGet       error
	errCreate    error
	errUpdate    error
	ordersByUser []*models.Order
	errGetByUser error
}

func (m *mockOrderRepo) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	return m.order, m.errGet
}
func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *models.Order) error {
	return m.errCreate
}
func (m *mockOrderRepo) UpdateOrder(ctx context.Context, number, status string, accrual float64) error {
	return m.errUpdate
}
func (m *mockOrderRepo) GetOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
	return m.ordersByUser, m.errGetByUser
}
func (m *mockOrderRepo) GetOrdersByStatus(ctx context.Context, statuses []string) ([]*models.Order, error) {
	return nil, nil
}

func TestUploadOrder_InvalidFormat(t *testing.T) {
	repo := &mockOrderRepo{}
	svc := services.NewOrderService(repo, "")

	status, err := svc.UploadOrder(context.Background(), 1, "abc123")
	assert.Equal(t, http.StatusUnprocessableEntity, status)
	assert.ErrorIs(t, err, services.ErrInvalidOrderFormat)
}

func TestUploadOrder_LuhnFail(t *testing.T) {
	repo := &mockOrderRepo{}
	svc := services.NewOrderService(repo, "")

	status, err := svc.UploadOrder(context.Background(), 1, "12345678901")
	assert.Equal(t, http.StatusUnprocessableEntity, status)
	assert.ErrorIs(t, err, services.ErrInvalidOrderFormat)
}

func TestUploadOrder_AlreadyExistsSameUser(t *testing.T) {
	repo := &mockOrderRepo{
		order: &models.Order{UserID: 1, Number: "79927398713"}, // valid Luhn
	}
	svc := services.NewOrderService(repo, "")

	status, err := svc.UploadOrder(context.Background(), 1, "79927398713")
	assert.Equal(t, http.StatusOK, status)
	assert.ErrorIs(t, err, services.ErrOrderAlreadyExistsSame)
}

func TestUploadOrder_AlreadyExistsOtherUser(t *testing.T) {
	repo := &mockOrderRepo{
		order: &models.Order{UserID: 2, Number: "79927398713"},
	}
	svc := services.NewOrderService(repo, "")

	status, err := svc.UploadOrder(context.Background(), 1, "79927398713")
	assert.Equal(t, http.StatusConflict, status)
	assert.ErrorIs(t, err, services.ErrOrderAlreadyExistsDifferent)
}

func TestUploadOrder_Success(t *testing.T) {
	repo := &mockOrderRepo{}
	svc := services.NewOrderService(repo, "")

	status, err := svc.UploadOrder(context.Background(), 1, "79927398713")
	assert.Equal(t, http.StatusAccepted, status)
	assert.NoError(t, err)
}

func TestGetUserOrders(t *testing.T) {
	expected := []*models.Order{
		{Number: "1"}, {Number: "2"},
	}
	repo := &mockOrderRepo{ordersByUser: expected}
	svc := services.NewOrderService(repo, "")

	orders, err := svc.GetUserOrders(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, orders)
}
