package handler

import (
	"errors"
	"time"

	"hotel-backend/config"
	"hotel-backend/internal/database"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/model"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtCfg      config.JWTConfig
}

func NewAuthHandler(authService *service.AuthService, jwtCfg config.JWTConfig) *AuthHandler {
	return &AuthHandler{authService: authService, jwtCfg: jwtCfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	accessToken, refreshToken, expiresIn, user, name, err := h.authService.Login(req.Username, req.Password, "guest")
	if err != nil {
		dto.Error(c, errcode.ErrInvalidCredentials)
		return
	}
	dto.Success(c, loginResponse(accessToken, refreshToken, expiresIn, user, name))
}

func (h *AuthHandler) StaffLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	accessToken, refreshToken, expiresIn, user, name, err := h.authService.Login(req.Username, req.Password, "waiter", "employee", "admin")
	if err != nil {
		dto.Error(c, errcode.ErrInvalidCredentials)
		return
	}
	dto.Success(c, loginResponse(accessToken, refreshToken, expiresIn, user, name))
}

func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	accessToken, refreshToken, expiresIn, user, name, err := h.authService.Login(req.Username, req.Password, "admin")
	if err != nil {
		dto.Error(c, errcode.ErrInvalidCredentials)
		return
	}
	dto.Success(c, loginResponse(accessToken, refreshToken, expiresIn, user, name))
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
		dto.Error(c, errcode.ErrConflict)
		return
	}
	dto.Success(c, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetString("raw_token")
	if token == "" {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	expiry := int64(h.jwtCfg.AccessTokenExpiry.Seconds())
	key := "blacklist:" + token
	if err := database.RDB.Set(c.Request.Context(), key, "1", time.Duration(expiry)*time.Second).Err(); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}

func loginResponse(accessToken, refreshToken string, expiresIn int64, user *model.User, name string) dto.LoginResponse {
	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserID:       user.ID,
		Role:         user.Role,
		Username:     user.Username,
		Name:         name,
	}
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetUint("user_id")
	token := c.GetString("raw_token")

	if err := h.authService.DeleteAccount(userID); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}

	expiry := int64(h.jwtCfg.AccessTokenExpiry.Seconds())
	key := "blacklist:" + token
	if err := database.RDB.Set(c.Request.Context(), key, "1", time.Duration(expiry)*time.Second).Err(); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}

	dto.Success(c, nil)
}
