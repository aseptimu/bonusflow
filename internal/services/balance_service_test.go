package services_test

import (
	"context"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/services"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockBalanceRepo struct {
	accrual     float64
	withdrawal  float64
	list        []*models.Withdrawal
	errAccrual  error
	errWithdraw error
	errCreate   error
	errList     error
}

func (m *mockBalanceRepo) GetAccrualSum(ctx context.Context, userID int) (float64, error) {
	return m.accrual, m.errAccrual
}
func (m *mockBalanceRepo) GetWithdrawalsSum(ctx context.Context, userID int) (float64, error) {
	return m.withdrawal, m.errWithdraw
}
func (m *mockBalanceRepo) CreateWithdrawal(ctx context.Context, userID int, orderNumber string, sum float64) error {
	return m.errCreate
}
func (m *mockBalanceRepo) ListWithdrawalsByUserID(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
	return m.list, m.errList
}

func TestGetUserBalance_Success(t *testing.T) {
	repo := &mockBalanceRepo{accrual: 120.555, withdrawal: 20.334}
	svc := services.NewBalanceService(repo)

	balance, err := svc.GetUserBalance(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 100.22, balance.Current)
	assert.Equal(t, 20.33, balance.Withdrawn)
}

func TestWithdraw_InvalidOrder(t *testing.T) {
	repo := &mockBalanceRepo{}
	svc := services.NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 1, "abc123", 10)
	assert.ErrorIs(t, err, services.ErrInvalidOrderNumber)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	repo := &mockBalanceRepo{accrual: 50, withdrawal: 20}
	svc := services.NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 1, "123456", 40)
	assert.ErrorIs(t, err, services.ErrInsufficientFunds)
}

func TestWithdraw_Success(t *testing.T) {
	repo := &mockBalanceRepo{accrual: 100, withdrawal: 30}
	svc := services.NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 1, "123456", 40)
	assert.NoError(t, err)
}

func TestListWithdrawals(t *testing.T) {
	expected := []*models.Withdrawal{
		{Order: "123", Sum: 10.0},
		{Order: "124", Sum: 20.5},
	}
	repo := &mockBalanceRepo{list: expected}
	svc := services.NewBalanceService(repo)

	list, err := svc.ListWithdrawals(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, list)
}
