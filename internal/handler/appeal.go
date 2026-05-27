package handler

import (
	"errors"

	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type AppealHandler struct {
	appealService *service.AppealService
	orderService  *service.OrderService
}

func NewAppealHandler(appealService *service.AppealService, orderService *service.OrderService) *AppealHandler {
	return &AppealHandler{appealService: appealService, orderService: orderService}
}

func (h *AppealHandler) Create(c *gin.Context) {
	orderID, err := h.orderService.DecodeCode(c.Param("code"))
	if err != nil {
		dto.Error(c, errcode.ErrNotFound)
		return
	}
	userID := c.GetUint("user_id")

	var req dto.CreateAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}

	appeal, err := h.appealService.Create(orderID, userID, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFoundInCancel) {
			dto.Error(c, errcode.ErrOrderNotFound)
			return
		}
		if errors.Is(err, service.ErrOrderNotBelongToUser) {
			dto.Error(c, errcode.ErrForbidden)
			return
		}
		if errors.Is(err, service.ErrOrderNotPending) {
			dto.Error(c, errcode.ErrOrderNotPending)
			return
		}
		if errors.Is(err, service.ErrAppealExists) {
			dto.Error(c, errcode.ErrAppealExists)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, appeal)
}

func (h *AppealHandler) List(c *gin.Context) {
	status := c.Query("status")
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	list, total, err := h.appealService.FindAll(status, p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: list, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *AppealHandler) ListMine(c *gin.Context) {
	userID := c.GetUint("user_id")
	status := c.Query("status")
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	list, total, err := h.appealService.FindByUserID(userID, status, p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: list, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *AppealHandler) Review(c *gin.Context) {
	id, err := h.appealService.DecodeCode(c.Param("code"))
	if err != nil {
		dto.Error(c, errcode.ErrAppealNotFound)
		return
	}
	reviewerID := c.GetUint("user_id")

	var req dto.ReviewAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}

	appeal, err := h.appealService.Review(uint(id), reviewerID, req.Action, req.ReviewNote)
	if err != nil {
		if errors.Is(err, service.ErrAppealNotFound) {
			dto.Error(c, errcode.ErrAppealNotFound)
			return
		}
		if errors.Is(err, service.ErrAppealNotPending) {
			dto.Error(c, errcode.ErrAppealNotPending)
			return
		}
		if errors.Is(err, service.ErrInvalidAppealAction) {
			dto.Error(c, errcode.ErrBadRequest)
			return
		}
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, appeal)
}
