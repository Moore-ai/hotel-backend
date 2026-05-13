package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type CheckinHandler struct {
	checkinService *service.CheckinService
}

func NewCheckinHandler(checkinService *service.CheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

func (h *CheckinHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	checkins, total, err := h.checkinService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: checkins, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *CheckinHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	checkin, err := h.checkinService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrCheckinNotFound)
		return
	}
	dto.Success(c, checkin)
}

func (h *CheckinHandler) Create(c *gin.Context) {
	var req dto.CreateCheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	expectedTime, err := time.Parse(time.RFC3339, req.ExpectedCheckoutTime)
	if err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	checkin, err := h.checkinService.Create(req.OrderID, req.UserID, req.RoomID, expectedTime)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, checkin)
}

func (h *CheckinHandler) Checkout(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	checkin, err := h.checkinService.Checkout(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrCheckinNotFound)
		return
	}
	if checkin == nil {
		dto.Error(c, errcode.ErrAlreadyCheckedOut)
		return
	}
	dto.Success(c, checkin)
}

func (h *CheckinHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.checkinService.Delete(uint(id)); err != nil {
		dto.Error(c, errcode.ErrCheckinNotFound)
		return
	}
	dto.Success(c, nil)
}
