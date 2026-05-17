package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type AdminHandler struct {
	adminService *service.AdminService
	userService  *service.UserService
}

func NewAdminHandler(adminService *service.AdminService, userService *service.UserService) *AdminHandler {
	return &AdminHandler{adminService: adminService, userService: userService}
}

func (h *AdminHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	admins, total, err := h.adminService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: admins, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *AdminHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	admin, err := h.adminService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrNotFound)
		return
	}
	dto.Success(c, admin)
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req dto.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	_, err := h.userService.Create(req.Username, req.Password, "admin", req.Name, req.Phone, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrUsernameExists) {
			dto.Error(c, errcode.ErrUsernameDuplicate)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}

func (h *AdminHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	admin, err := h.adminService.Update(uint(id), req.Name, req.Phone, req.Email)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, admin)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	admin, err := h.adminService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	if err := h.userService.Delete(admin.UserID); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}
