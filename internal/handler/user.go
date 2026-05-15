package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	users, total, err := h.userService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: users, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := h.userService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrUserNotFound)
		return
	}
	dto.Success(c, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	user, err := h.userService.Create(req.Username, req.Password, req.Role, req.Name, req.Phone, req.Email)
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

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	user, err := h.userService.Update(uint(id), req.Username, req.Password, req.Role, req.Name, req.Phone, req.Email)
	if err != nil {
		dto.Error(c, errcode.ErrUserNotFound)
		return
	}
	dto.Success(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.userService.Delete(uint(id)); err != nil {
		dto.Error(c, errcode.ErrUserNotFound)
		return
	}
	dto.Success(c, nil)
}
