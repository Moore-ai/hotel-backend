package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type GuestHandler struct {
	guestService *service.GuestService
	userService  *service.UserService
}

func NewGuestHandler(guestService *service.GuestService, userService *service.UserService) *GuestHandler {
	return &GuestHandler{guestService: guestService, userService: userService}
}

func (h *GuestHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	guests, total, err := h.guestService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: guests, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *GuestHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	guest, err := h.guestService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrNotFound)
		return
	}
	dto.Success(c, guest)
}

func (h *GuestHandler) Create(c *gin.Context) {
	var req dto.CreateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	_, err := h.userService.Create(req.Username, req.Password, "guest", req.Name, req.Phone, req.Email)
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

func (h *GuestHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	guest, err := h.guestService.Update(uint(id), req.Name, req.Phone, req.Email)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, guest)
}

func (h *GuestHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	guest, err := h.guestService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	if err := h.userService.Delete(guest.UserID); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}
