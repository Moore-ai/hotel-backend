package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/repository"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type AuditLogHandler struct {
	auditLogService *service.AuditLogService
	userService     *service.UserService
}

func NewAuditLogHandler(auditLogService *service.AuditLogService, userService *service.UserService) *AuditLogHandler {
	return &AuditLogHandler{auditLogService: auditLogService, userService: userService}
}

func (h *AuditLogHandler) List(c *gin.Context) {
	entityID, _ := strconv.ParseUint(c.Query("entity_id"), 10, 64)
	userID := uint(0)
	if code := c.Query("user_code"); code != "" {
		id, err := h.userService.DecodeCode(code)
		if err == nil {
			userID = id
		}
	}
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))

	f := repository.AuditLogFilter{
		EntityType: c.Query("entity_type"),
		EntityID:   uint(entityID),
		UserID:     uint(userID),
		StartDate:  c.Query("start"),
		EndDate:    c.Query("end"),
		Page:       p.Page,
		PageSize:   p.PageSize,
	}
	logs, total, err := h.auditLogService.FindAll(f)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: logs, Total: total, Page: p.Page, PageSize: p.PageSize})
}
