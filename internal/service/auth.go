// internal/service/auth.go
package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"hotel-backend/config"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/hash"
	"hotel-backend/pkg/jwt"
)

type AuthService struct {
	userRepo *repository.UserRepo
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepo, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, jwtCfg: jwtCfg}
}

func (s *AuthService) Login(username, password string) (string, string, int64, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", 0, ErrInvalidCredentials
		}
		return "", "", 0, err
	}
	if !hash.CheckPassword(user.PasswordHash, password) {
		return "", "", 0, ErrInvalidCredentials
	}

	accessExpiry := s.jwtCfg.AccessTokenExpiry * time.Second
	refreshExpiry := s.jwtCfg.RefreshTokenExpiry * time.Second

	accessToken, err := jwt.GenerateToken(user.ID, user.Role, s.jwtCfg.Secret, accessExpiry)
	if err != nil {
		return "", "", 0, err
	}
	refreshToken, err := jwt.GenerateToken(user.ID, user.Role, s.jwtCfg.Secret, refreshExpiry)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, int64(accessExpiry.Seconds()), nil
}

func (s *AuthService) Logout(ctx context.Context, token string, expiry time.Duration) error {
	return nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")
