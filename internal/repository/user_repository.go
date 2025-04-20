package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aseptimu/internal/models"
)

type UserStore interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type PostgresUserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserStore {
	return &PostgresUserRepository{DB: db}
}

const createUserQuery = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.DB.QueryRowContext(ctx, createUserQuery, user.Login, user.Password).Scan(&user.ID)
}

var ErrUserNotFound = errors.New("пользователь не найден")

const getUserByLoginQuery = `SELECT id, login, password FROM users WHERE login = $1`

func (r *PostgresUserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	row := r.DB.QueryRowContext(ctx, getUserByLoginQuery, login)

	var user models.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
