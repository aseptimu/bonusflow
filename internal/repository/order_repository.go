package repository

import (
	"context"
	"database/sql"
	"github.com/aseptimu/internal/models"
	"time"
)

type OrderRepository interface {
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) error
	UpdateOrder(ctx context.Context, number, status string, accrual float64) error
	GetOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error)
}

type PostgresOrderRepository struct {
	DB *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{DB: db}
}

const getOrderByNumberQuery = `
	SELECT id, number, user_id, status, uploaded_at, accrual
	FROM orders
	WHERE number = $1
`

func (r *PostgresOrderRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
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

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
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

func (r *PostgresOrderRepository) UpdateOrder(ctx context.Context, number, status string, accrual float64) error {
	_, err := r.DB.ExecContext(ctx, updateOrderQuery, number, status, accrual)
	return err
}

const getOrdersByUserIDQuery = `
	SELECT id, number, user_id, status, uploaded_at, accrual
	FROM orders
	WHERE user_id = $1
	ORDER BY uploaded_at
`

func (r *PostgresOrderRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]*models.Order, error) {
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
	return orders, nil
}
