package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func (h *RoomHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	rooms, total, err := h.roomService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: rooms, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *RoomHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	room, err := h.roomService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrRoomNotFound)
		return
	}
	dto.Success(c, room)
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	room, err := h.roomService.Create(req.RoomNumber, req.Type, req.Floor, req.PricePerNight, req.Status, req.Description)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, room)
}

func (h *RoomHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	room, err := h.roomService.Update(uint(id), req.RoomNumber, req.Type, req.Floor, req.PricePerNight, req.Status, req.Description)
	if err != nil {
		dto.Error(c, errcode.ErrRoomNotFound)
		return
	}
	dto.Success(c, room)
}

func (h *RoomHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.roomService.Delete(uint(id)); err != nil {
		dto.Error(c, errcode.ErrRoomNotFound)
		return
	}
	dto.Success(c, nil)
}
