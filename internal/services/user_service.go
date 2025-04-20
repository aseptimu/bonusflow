package services

import (
	"context"
	"errors"
	"time"

	"github.com/aseptimu/internal/config"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/repository"
	"github.com/aseptimu/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const tokenExpirationTime = time.Hour * 72

var ErrorUserAlreadyExists = errors.New("user already exists")

type UserManager interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type userService struct {
	repo repository.UserStore
	conf *config.ConfigType
}

func NewUserService(repo repository.UserStore, conf *config.ConfigType) UserManager {
	return &userService{
		repo: repo,
		conf: conf,
	}
}

func (s *userService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", errors.New("login and password must be provided")
	}

	existingUser, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			existingUser = nil
		} else {
			utils.LogWithError(ctx, "Register: error checking user existence", err)
			return "", err
		}
	}
	if existingUser != nil {
		utils.LogWithError(ctx, "Register: user already exists", errors.New("user already exists"))
		return "", ErrorUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.LogWithError(ctx, "Register: error hashing password", err)
		return "", err
	}

	user := &models.User{
		Login:    login,
		Password: string(hashedPassword),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		utils.LogWithError(ctx, "Register: error creating user", err)
		return "", err
	}

	token, err := generateToken(user.ID, s.conf.SecretKey)
	if err != nil {
		utils.LogWithError(ctx, "Register: error generating token", err)
		return "", err
	}

	return token, nil
}

func (s *userService) Login(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", errors.New("login and password must be provided")
	}

	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		utils.LogWithError(ctx, "Login: error retrieving user", err)
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid login or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		utils.LogWithError(ctx, "Login: password mismatch", err)
		return "", errors.New("invalid login or password")
	}

	token, err := generateToken(user.ID, s.conf.SecretKey)
	if err != nil {
		utils.LogWithError(ctx, "Login: error generating token", err)
		return "", err
	}

	return token, nil
}

func generateToken(userID int, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(tokenExpirationTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
