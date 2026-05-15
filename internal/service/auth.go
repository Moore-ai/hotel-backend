// internal/service/auth.go
package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"hotel-backend/config"
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/hash"
	"hotel-backend/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameExists     = errors.New("username already exists")
)

type AuthService struct {
	userRepo    *repository.UserRepo
	userService *UserService
	jwtCfg      config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepo, userService *UserService, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, userService: userService, jwtCfg: jwtCfg}
}

func (s *AuthService) Login(username, password string) (string, string, int64, *model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", 0, nil, ErrInvalidCredentials
		}
		return "", "", 0, nil, err
	}
	if !hash.CheckPassword(user.PasswordHash, password) {
		return "", "", 0, nil, ErrInvalidCredentials
	}

	accessExpiry := s.jwtCfg.AccessTokenExpiry * time.Second
	refreshExpiry := s.jwtCfg.RefreshTokenExpiry * time.Second

	accessToken, err := jwt.GenerateToken(user.ID, user.Role, s.jwtCfg.Secret, accessExpiry)
	if err != nil {
		return "", "", 0, nil, err
	}
	refreshToken, err := jwt.GenerateToken(user.ID, user.Role, s.jwtCfg.Secret, refreshExpiry)
	if err != nil {
		return "", "", 0, nil, err
	}

	return accessToken, refreshToken, int64(accessExpiry.Seconds()), user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string, expiry time.Duration) error {
	return nil
}

func (s *AuthService) Register(username, password, name, phone, email string) (*model.User, error) {
	return s.userService.Create(username, password, "guest", name, phone, email)
}

func (s *AuthService) DeleteAccount(userID uint) error {
	return s.userService.Delete(userID)
}
