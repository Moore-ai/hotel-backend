package handler

import (
	"github.com/gin-gonic/gin"

	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}

	userID := c.GetUint("user_id")

	reply, convID, action, err := h.chatSvc.HandleMessage(userID, req.Message, req.ConversationID)
	if err != nil {
		dto.Error(c, errcode.ErrLLMUnavailable)
		return
	}

	dto.Success(c, gin.H{
		"reply":           reply,
		"conversation_id": convID,
		"action":          action,
	})
}
