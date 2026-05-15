// internal/handler/auth.go
package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/database"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) blacklistToken(c *gin.Context) {
	header := c.GetHeader("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	ctx := context.Background()
	database.RDB.Set(ctx, "blacklist:"+token, "1", 2*time.Hour)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	access, refresh, expiresIn, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		dto.Error(c, errcode.ErrInvalidCredentials)
		return
	}
	dto.Success(c, dto.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
		UserID:       user.ID,
		Role:         user.Role,
		Username:     user.Username,
		Name:         user.Name,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	user, err := h.authService.Register(req.Username, req.Password, req.Name, req.Phone, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			dto.Error(c, errcode.ErrUsernameDuplicate)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.blacklistToken(c)
	dto.Success(c, nil)
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	h.blacklistToken(c)
	userID := c.GetUint("user_id")

	if err := h.authService.DeleteAccount(userID); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}
