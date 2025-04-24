package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/aseptimu/internal/models"
)

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	user := &models.User{Login: "u1", Password: "p1"}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(7)
	mock.ExpectQuery(regexp.QuoteMeta(createUserQuery)).
		WithArgs(user.Login, user.Password).
		WillReturnRows(rows)

	if err := repo.CreateUser(context.Background(), user); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user.ID != 7 {
		t.Errorf("expected ID=7, got %d", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestCreateUser_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewUserRepository(db)
	user := &models.User{Login: "u2", Password: "p2"}

	mock.ExpectQuery(regexp.QuoteMeta(createUserQuery)).
		WithArgs(user.Login, user.Password).
		WillReturnError(errors.New("insert failed"))

	err := repo.CreateUser(context.Background(), user)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetUserByLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)
	login := "test"
	expectedID := 9
	expectedPwd := "hash"

	rows := sqlmock.NewRows([]string{"id", "login", "password"}).
		AddRow(expectedID, login, expectedPwd)
	mock.ExpectQuery(regexp.QuoteMeta(getUserByLoginQuery)).
		WithArgs(login).
		WillReturnRows(rows)

	user, err := repo.GetUserByLogin(context.Background(), login)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != expectedID || user.Login != login || user.Password != expectedPwd {
		t.Errorf("got user %+v, want ID=%d, Login=%s, Password=%s", user, expectedID, login, expectedPwd)
	}
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewUserRepository(db)
	login := "missing"

	mock.ExpectQuery(regexp.QuoteMeta(getUserByLoginQuery)).
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetUserByLogin(context.Background(), login)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetUserByLogin_OtherError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewUserRepository(db)
	login := "err"

	mock.ExpectQuery(regexp.QuoteMeta(getUserByLoginQuery)).
		WithArgs(login).
		WillReturnError(errors.New("query fail"))

	_, err := repo.GetUserByLogin(context.Background(), login)
	if err == nil || errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected generic error, got %v", err)
	}
}
