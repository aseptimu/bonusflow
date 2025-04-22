package services

import (
	"context"
	"errors"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/repository"
	"math"
	"strconv"
)

var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)

type BalanceService struct {
	repo *repository.BalanceRepository
}

func NewBalanceService(repo *repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	accr, err := s.repo.GetAccrualSum(ctx, userID)
	if err != nil {
		return nil, err
	}

	with, err := s.repo.GetWithdrawalsSum(ctx, userID)
	if err != nil {
		return nil, err
	}

	rawCurrent := accr - with
	current := math.Round(rawCurrent*100) / 100
	withdrawn := math.Round(with*100) / 100

	return &models.Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}, nil
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	if _, err := strconv.ParseUint(orderNumber, 10, 64); err != nil {
		return ErrInvalidOrderNumber
	}

	withdrawals, err := s.repo.GetWithdrawalsSum(ctx, userID)
	if err != nil {
		return err
	}

	balance, err := s.repo.GetAccrualSum(ctx, userID)
	if err != nil {
		return err
	}

	rawCurrent := balance - withdrawals
	current := math.Round(rawCurrent*100) / 100

	if current < sum {
		return ErrInsufficientFunds
	}

	return s.repo.CreateWithdrawal(ctx, userID, orderNumber, sum)
}

func (s *BalanceService) ListWithdrawals(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
	return s.repo.ListWithdrawalsByUserID(ctx, userID)
}
