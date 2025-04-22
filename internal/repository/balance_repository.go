package repository

import (
	"context"
	"database/sql"
	"github.com/aseptimu/internal/models"
)

type BalanceRepository struct {
	DB *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{DB: db}
}

const getAccrualSum = `SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'`
const getWithdrawalsSum = `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`

func (r *BalanceRepository) GetAccrualSum(ctx context.Context, userID int) (float64, error) {
	var sum sql.NullFloat64
	err := r.DB.
		QueryRowContext(ctx, getAccrualSum, userID).
		Scan(&sum)
	return sum.Float64, err
}

func (r *BalanceRepository) GetWithdrawalsSum(ctx context.Context, userID int) (float64, error) {
	var sum sql.NullFloat64
	err := r.DB.
		QueryRowContext(ctx, getWithdrawalsSum, userID).
		Scan(&sum)
	return sum.Float64, err
}

const insertWithdrawalQuery = `
INSERT INTO withdrawals (user_id, order_number, sum)
VALUES ($1, $2, $3)
`

func (r *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int, orderNumber string, sum float64) error {
	_, err := r.DB.ExecContext(ctx, insertWithdrawalQuery, userID, orderNumber, sum)
	return err
}

const getWithdrawalsQuery = `
SELECT order_number, sum, processed_at
  FROM withdrawals
 WHERE user_id = $1
 ORDER BY processed_at DESC
`

func (r *BalanceRepository) ListWithdrawalsByUserID(ctx context.Context, userID int) ([]*models.Withdrawal, error) {
	rows, err := r.DB.QueryContext(ctx, getWithdrawalsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*models.Withdrawal
	for rows.Next() {
		w := new(models.Withdrawal)
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		res = append(res, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
