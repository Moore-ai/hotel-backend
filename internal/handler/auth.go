// internal/handler/auth.go
package handler

import (
	"context"
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	access, refresh, expiresIn, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		dto.Error(c, errcode.ErrInvalidCredentials)
		return
	}
	dto.Success(c, dto.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	ctx := context.Background()
	ttl := 2 * time.Hour
	database.RDB.Set(ctx, "blacklist:"+token, "1", ttl)
	dto.Success(c, nil)
}
