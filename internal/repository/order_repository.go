package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aseptimu/internal/models"
	"strings"
	"time"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

const getOrderByNumberQuery = `
	SELECT id, number, user_id, status, uploaded_at, accrual
	FROM orders
	WHERE number = $1
`

func (r *OrderRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	row := r.DB.QueryRowContext(ctx, getOrderByNumberQuery, number)
	var order models.Order
	err := row.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.UploadedAt, &order.Accrual)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

const createOrderQuery = `
	INSERT INTO orders (number, user_id, status, uploaded_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id
`

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	order.UploadedAt = time.Now()
	_, err := r.DB.ExecContext(ctx, createOrderQuery, order.Number, order.UserID, order.Status, order.UploadedAt)
	return err
}

const updateOrderQuery = `
  UPDATE orders
     SET status = $2,
         accrual = $3
   WHERE number = $1
`

func (r *OrderRepository) UpdateOrder(ctx context.Context, number, status string, accrual float64) error {
	_, err := r.DB.ExecContext(ctx, updateOrderQuery, number, status, accrual)
	return err
}

const getOrdersByUserIDQuery = `
	SELECT id, number, user_id, status, uploaded_at, accrual
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded_at
`

func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
	rows, err := r.DB.QueryContext(ctx, getOrdersByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.UploadedAt, &order.Accrual); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

const getOrdersByStatusQuery = `
        SELECT number, user_id, status, uploaded_at, accrual
          FROM orders
         WHERE status IN (%s)
      ORDER BY uploaded_at
`

func (r *OrderRepository) GetOrdersByStatus(ctx context.Context, statuses []string) ([]*models.Order, error) {
	args := make([]interface{}, len(statuses))
	in := make([]string, len(statuses))
	for i, s := range statuses {
		args[i] = s
		in[i] = fmt.Sprintf("$%d", i+1)
	}
	query := fmt.Sprintf(getOrdersByStatusQuery, strings.Join(in, ","))
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.Number, &o.UserID, &o.Status, &o.UploadedAt, &o.Accrual); err != nil {
			return nil, err
		}
		out = append(out, &o)
	}
	return out, rows.Err()
}
