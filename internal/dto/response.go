// internal/dto/response.go
package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hotel-backend/pkg/errcode"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.Success,
		Message: errcode.Message(errcode.Success),
		Data:    data,
	})
}

func Error(c *gin.Context, code int) {
	httpStatus := http.StatusOK
	switch code {
	case errcode.ErrBadRequest:
		httpStatus = http.StatusBadRequest
	case errcode.ErrUnauthorized:
		httpStatus = http.StatusUnauthorized
	case errcode.ErrForbidden:
		httpStatus = http.StatusForbidden
	case errcode.ErrConflict:
		httpStatus = http.StatusConflict
	case errcode.ErrNotFound:
		httpStatus = http.StatusNotFound
	case errcode.ErrInternal:
		httpStatus = http.StatusInternalServerError
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: errcode.Message(code),
	})
}

func ErrorWithMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
	})
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       uint   `json:"user_id"`
	Role         string `json:"role"`
	Username     string `json:"username"`
	Name         string `json:"name"`
}
