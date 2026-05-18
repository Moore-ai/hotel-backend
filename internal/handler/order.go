package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	role := c.GetString("role")
	userID := c.GetUint("user_id")

	var orders interface{}
	var total int64
	var err error

	if role == "guest" {
		orders, total, err = h.orderService.FindByUserID(userID, p.Page, p.PageSize)
	} else {
		orders, total, err = h.orderService.FindAll(p.Page, p.PageSize)
	}
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: orders, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := h.orderService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrOrderNotFound)
		return
	}
	role := c.GetString("role")
	userID := c.GetUint("user_id")
	if role == "guest" && order.UserID != userID {
		dto.Error(c, errcode.ErrForbidden)
		return
	}
	dto.Success(c, order)
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	userID := c.GetUint("user_id")

	var roomID *uint
	if req.RoomID != 0 {
		roomID = &req.RoomID
	}

	order, err := h.orderService.Create(userID, roomID, req.CheckInDate, req.CheckOutDate, req.TotalPrice, req.GuestCount, req.RoomTypePreference)
	if err != nil {
		if errors.Is(err, service.ErrNoRoomAvailable) {
			dto.Error(c, errcode.ErrNoRoomAvailable)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, order)
}

func (h *OrderHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	order, err := h.orderService.Update(uint(id), req.CheckInDate, req.CheckOutDate, req.Status, req.TotalPrice)
	if err != nil {
		dto.Error(c, errcode.ErrOrderNotFound)
		return
	}
	dto.Success(c, order)
}

func (h *OrderHandler) ListCancelRequests(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	orders, total, err := h.orderService.FindCancelRequests(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: orders, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")

	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}

	order, autoCancelled, err := h.orderService.Cancel(uint(id), userID, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFoundInCancel) {
			dto.Error(c, errcode.ErrOrderNotFound)
			return
		}
		if errors.Is(err, service.ErrOrderNotPending) {
			dto.Error(c, errcode.ErrOrderNotPending)
			return
		}
		if errors.Is(err, service.ErrOrderNotBelongToUser) {
			dto.Error(c, errcode.ErrForbidden)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}

	dto.Success(c, gin.H{
		"id":             order.ID,
		"status":         order.Status,
		"auto_cancelled": autoCancelled,
		"cancel_reason":  order.CancelReason,
	})
}

func (h *OrderHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.orderService.Delete(uint(id)); err != nil {
		dto.Error(c, errcode.ErrOrderNotFound)
		return
	}
	dto.Success(c, nil)
}
