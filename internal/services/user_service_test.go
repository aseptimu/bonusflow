package services

import (
	"context"
	"errors"
	"github.com/aseptimu/internal/config"
	"github.com/aseptimu/internal/models"
	"github.com/aseptimu/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

// mockUserRepo implements UserRepository interface for testing
type mockUserRepo struct {
	getByLogin func(ctx context.Context, login string) (*models.User, error)
	createUser func(ctx context.Context, user *models.User) error
}

func (m *mockUserRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return m.getByLogin(ctx, login)
}
func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	return m.createUser(ctx, user)
}

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()
	conf := &config.ConfigType{SecretKey: "testsecret"}

	tests := []struct {
		name            string
		login           string
		password        string
		getUserResult   *models.User
		getUserErr      error
		createUserErr   error
		wantErr         error
		wantTokenUserID int
	}{
		{name: "empty login", login: "", password: "p", wantErr: errors.New("login and password must be provided")},
		{name: "empty password", login: "u", password: "", wantErr: errors.New("login and password must be provided")},
		{name: "existing user", login: "u", password: "p", getUserResult: &models.User{ID: 1}, getUserErr: nil, wantErr: ErrorUserAlreadyExists},
		{name: "repo error checking", login: "u", password: "p", getUserResult: nil, getUserErr: errors.New("db fail"), wantErr: errors.New("db fail")},
		{name: "create user error", login: "u", password: "p", getUserResult: nil, getUserErr: repository.ErrUserNotFound, createUserErr: errors.New("insert fail"), wantErr: errors.New("insert fail")},
		{name: "success", login: "user1", password: "pass1", getUserResult: nil, getUserErr: repository.ErrUserNotFound, createUserErr: nil, wantErr: nil, wantTokenUserID: 42},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*models.User, error) {
					return tc.getUserResult, tc.getUserErr
				},
				createUser: func(ctx context.Context, user *models.User) error {
					user.ID = tc.wantTokenUserID
					return tc.createUserErr
				},
			}
			svc := NewUserService(repo, conf)
			token, err := svc.Register(ctx, tc.login, tc.password)
			if tc.wantErr != nil {
				if err == nil || err.Error() != tc.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// parse token and verify user_id claim
			parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
				return []byte(conf.SecretKey), nil
			})
			if err != nil || !parsed.Valid {
				t.Fatalf("invalid token: %v", err)
			}
			claims := parsed.Claims.(jwt.MapClaims)
			uid := int(claims["user_id"].(float64))
			if uid != tc.wantTokenUserID {
				t.Errorf("token user_id = %d; want %d", uid, tc.wantTokenUserID)
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	conf := &config.ConfigType{SecretKey: "testsecret"}
	// generate valid hash for password
	pass := "mypassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		login         string
		password      string
		getUserResult *models.User
		getUserErr    error
		wantErr       string
		wantUserID    int
	}{
		{name: "empty login", login: "", password: "p", wantErr: "login and password must be provided"},
		{name: "empty password", login: "u", password: "", wantErr: "login and password must be provided"},
		{name: "db error", login: "u", password: "p", getUserErr: errors.New("dbfail"), wantErr: "dbfail"},
		{name: "user not found", login: "u", password: "p", getUserResult: nil, getUserErr: nil, wantErr: "invalid login or password"},
		{name: "wrong password", login: "u", password: "wrong", getUserResult: &models.User{ID: 5, Password: string(hash)}, getUserErr: nil, wantErr: "invalid login or password"},
		{name: "success", login: "u", password: pass, getUserResult: &models.User{ID: 7, Password: string(hash)}, getUserErr: nil, wantErr: "", wantUserID: 7},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUserRepo{
				getByLogin: func(ctx context.Context, login string) (*models.User, error) {
					return tc.getUserResult, tc.getUserErr
				},
			}
			svc := NewUserService(repo, conf)
			token, err := svc.Login(ctx, tc.login, tc.password)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("expected error %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
				return []byte(conf.SecretKey), nil
			})
			if err != nil || !parsed.Valid {
				t.Fatalf("invalid token: %v", err)
			}
			claims := parsed.Claims.(jwt.MapClaims)
			uid := int(claims["user_id"].(float64))
			if uid != tc.wantUserID {
				t.Errorf("token user_id = %d; want %d", uid, tc.wantUserID)
			}
		})
	}
}
