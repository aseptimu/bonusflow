package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/aseptimu/internal/models"
)

func TestGetOrderByNumber_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %v", err)
	}
	defer db.Close()

	repo := NewOrderRepository(db)
	ctx := context.Background()
	number := "ord1"
	now := time.Now().Truncate(time.Second)
	expectedAccrual := 5.5

	rows := sqlmock.NewRows([]string{"id", "number", "user_id", "status", "uploaded_at", "accrual"}).
		AddRow(1, number, 10, "NEW", now, expectedAccrual)

	mock.ExpectQuery(regexp.QuoteMeta(getOrderByNumberQuery)).
		WithArgs(number).
		WillReturnRows(rows)

	order, err := repo.GetOrderByNumber(ctx, number)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != 1 || order.Number != number || order.UserID != 10 || order.Status != "NEW" {
		t.Errorf("got wrong order data: %+v", order)
	}
	if !order.UploadedAt.Equal(now) {
		t.Errorf("uploaded_at mismatch: got %v, want %v", order.UploadedAt, now)
	}
	if order.Accrual != expectedAccrual {
		t.Errorf("accrual mismatch: got %v, want %v", order.Accrual, expectedAccrual)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetOrderByNumber_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(getOrderByNumberQuery)).
		WithArgs("ordX").
		WillReturnError(errors.New("not found"))

	if _, err := repo.GetOrderByNumber(context.Background(), "ordX"); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCreateOrder_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	order := &models.Order{Number: "ord2", UserID: 20, Status: "NEW"}
	// Expect any time for UploadedAt
	mock.ExpectExec(regexp.QuoteMeta(createOrderQuery)).
		WithArgs(order.Number, order.UserID, order.Status, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.CreateOrder(context.Background(), order); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateOrder_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	number := "ord1"
	status := "PROCESSED"
	accrual := 7.7

	mock.ExpectExec(regexp.QuoteMeta(updateOrderQuery)).
		WithArgs(number, status, accrual).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateOrder(context.Background(), number, status, accrual); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetOrdersByUserID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	userID := 30
	now := time.Now().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "number", "user_id", "status", "uploaded_at", "accrual"}).
		AddRow(1, "o1", userID, "NEW", now, 1.1).
		AddRow(2, "o2", userID, "PROCESSED", now, 2.2)

	mock.ExpectQuery(regexp.QuoteMeta(getOrdersByUserIDQuery)).
		WithArgs(userID).
		WillReturnRows(rows)

	list, err := repo.GetOrdersByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(list))
	}
	if list[0].Number != "o1" || list[1].Number != "o2" {
		t.Errorf("order numbers mismatch: %+v", list)
	}
}

func TestGetOrdersByUserID_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(getOrdersByUserIDQuery)).
		WithArgs(99).
		WillReturnError(errors.New("query error"))

	if _, err := repo.GetOrdersByUserID(context.Background(), 99); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetOrdersByUserID_ScanError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	// uploaded_at wrong type
	rows := sqlmock.NewRows([]string{"id", "number", "user_id", "status", "uploaded_at", "accrual"}).
		AddRow(1, "o", 1, "NEW", "badtime", 0.0)
	mock.ExpectQuery(regexp.QuoteMeta(getOrdersByUserIDQuery)).
		WillReturnRows(rows)

	if _, err := repo.GetOrdersByUserID(context.Background(), 1); err == nil {
		t.Error("expected scan error, got nil")
	}
}

func TestGetOrdersByStatus_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	statuses := []string{"NEW", "PROCESSED"}
	placeholders := []string{"$1", "$2"}
	query := fmt.Sprintf(getOrdersByStatusQuery, strings.Join(placeholders, ","))

	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"number", "user_id", "status", "uploaded_at", "accrual"}).
		AddRow("x1", 11, "NEW", now, 3.3)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(statuses[0], statuses[1]).
		WillReturnRows(rows)

	list, err := repo.GetOrdersByStatus(context.Background(), statuses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].Number != "x1" {
		t.Errorf("unexpected result: %+v", list)
	}
}

func TestGetOrdersByStatus_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	statuses := []string{"A"}
	query := fmt.Sprintf(getOrdersByStatusQuery, "$1")

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(statuses[0]).
		WillReturnError(errors.New("query fail"))

	if _, err := repo.GetOrdersByStatus(context.Background(), statuses); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetOrdersByStatus_ScanError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrderRepository(db)

	statuses := []string{"X"}
	query := fmt.Sprintf(getOrdersByStatusQuery, "$1")

	// user_id wrong type
	rows := sqlmock.NewRows([]string{"number", "user_id", "status", "uploaded_at", "accrual"}).
		AddRow("y", "bad", "NEW", time.Now(), 0.0)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(statuses[0]).
		WillReturnRows(rows)

	if _, err := repo.GetOrdersByStatus(context.Background(), statuses); err == nil {
		t.Error("expected scan error, got nil")
	}
}
