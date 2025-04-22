package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aseptimu/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

const createUserQuery = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.DB.QueryRowContext(ctx, createUserQuery, user.Login, user.Password).Scan(&user.ID)
}

var ErrUserNotFound = errors.New("пользователь не найден")

const getUserByLoginQuery = `SELECT id, login, password FROM users WHERE login = $1`

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
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
