package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestGetAccrualSum_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock sql db: %v", err)
	}
	defer db.Close()

	repo := NewBalanceRepository(db)
	userID := 42
	expected := 123.45

	mock.ExpectQuery(regexp.QuoteMeta(getAccrualSum)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(expected))

	sum, err := repo.GetAccrualSum(context.Background(), userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sum != expected {
		t.Errorf("got sum %v, want %v", sum, expected)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetAccrualSum_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(getAccrualSum)).
		WillReturnError(errors.New("db error"))

	_, err := repo.GetAccrualSum(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetWithdrawalsSum_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)
	userID := 7
	expected := 200.0

	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsSum)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(expected))

	sum, err := repo.GetWithdrawalsSum(context.Background(), userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sum != expected {
		t.Errorf("got sum %v, want %v", sum, expected)
	}
}

func TestCreateWithdrawal_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)
	userID := 3
	order := "ord123"
	sum := 50.5

	mock.ExpectExec(regexp.QuoteMeta(insertWithdrawalQuery)).
		WithArgs(userID, order, sum).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.CreateWithdrawal(context.Background(), userID, order, sum); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateWithdrawal_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(insertWithdrawalQuery)).
		WillReturnError(errors.New("exec error"))

	err := repo.CreateWithdrawal(context.Background(), 1, "o", 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestListWithdrawalsByUserID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)
	userID := 5
	now := time.Now().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
		AddRow("o1", 10.0, now).
		AddRow("o2", 20.5, now)

	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsQuery)).
		WithArgs(userID).
		WillReturnRows(rows)

	list, err := repo.ListWithdrawalsByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 rows, got %d", len(list))
	}
	if list[0].Order != "o1" || list[0].Sum != 10.0 || !list[0].ProcessedAt.Equal(now) {
		t.Errorf("first row mismatch: %+v", list[0])
	}
}

func TestListWithdrawalsByUserID_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsQuery)).
		WillReturnError(errors.New("query error"))

	if _, err := repo.ListWithdrawalsByUserID(context.Background(), 1); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestListWithdrawalsByUserID_ScanError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewBalanceRepository(db)

	rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
		AddRow("o", "notfloat", time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(getWithdrawalsQuery)).
		WillReturnRows(rows)

	if _, err := repo.ListWithdrawalsByUserID(context.Background(), 1); err == nil {
		t.Error("expected scan error, got nil")
	}
}
