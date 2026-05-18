package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type WaiterHandler struct {
	waiterService  *service.WaiterService
	checkinService *service.CheckinService
	userService    *service.UserService
}

func NewWaiterHandler(waiterService *service.WaiterService, checkinService *service.CheckinService, userService *service.UserService) *WaiterHandler {
	return &WaiterHandler{waiterService: waiterService, checkinService: checkinService, userService: userService}
}

func (h *WaiterHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	waiters, total, err := h.waiterService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: waiters, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *WaiterHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	waiter, err := h.waiterService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrNotFound)
		return
	}
	dto.Success(c, waiter)
}

func (h *WaiterHandler) Create(c *gin.Context) {
	var req dto.CreateWaiterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	_, err := h.userService.Create(req.Username, req.Password, "waiter", req.Name, req.Phone, req.Email)
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

func (h *WaiterHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateWaiterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	waiter, err := h.waiterService.Update(uint(id), req.Name, req.Phone, req.Email)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, waiter)
}

func (h *WaiterHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	waiter, err := h.waiterService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	if err := h.userService.Delete(waiter.UserID); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}

func (h *WaiterHandler) ServiceRequest(c *gin.Context) {
	checkinID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	role := c.GetString("role")

	var req dto.ServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}

	checkin, err := h.checkinService.FindByID(uint(checkinID))
	if err != nil {
		dto.Error(c, errcode.ErrCheckinNotFound)
		return
	}
	if role == "guest" && checkin.UserID != userID {
		dto.Error(c, errcode.ErrForbidden)
		return
	}
	if checkin.Status != "active" {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}

	waiter, err := h.waiterService.Dispatch(checkin.RoomID, checkin.UserID, req.Content, req.Note)
	if err != nil {
		if errors.Is(err, service.ErrNoWaiterAvailable) {
			h.waiterService.NotifyWaiting(checkin.UserID)
			dto.Error(c, errcode.ErrNoWaiterAvailable)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, waiter)
}

func (h *WaiterHandler) CompleteService(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role := c.GetString("role")
	userID := c.GetUint("user_id")

	if role != "admin" {
		waiter, err := h.waiterService.FindByUserID(userID)
		if err != nil || waiter.ID != uint(id) {
			dto.Error(c, errcode.ErrForbidden)
			return
		}
	}

	if err := h.waiterService.CompleteService(uint(id)); err != nil {
		dto.Error(c, errcode.ErrNotFound)
		return
	}
	dto.Success(c, nil)
}
