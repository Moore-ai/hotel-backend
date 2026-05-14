package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"hotel-backend/config"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
	"hotel-backend/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub    *service.Hub
	jwtCfg config.JWTConfig
}

func NewWSHandler(hub *service.Hub, jwtCfg config.JWTConfig) *WSHandler {
	return &WSHandler{hub: hub, jwtCfg: jwtCfg}
}

func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		errcode.Write(c, errcode.ErrUnauthorized)
		return
	}

	claims, err := jwt.ParseToken(token, h.jwtCfg.Secret)
	if err != nil {
		errcode.Write(c, errcode.ErrUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	client := &service.Client{
		Hub:    h.hub,
		Conn:   conn,
		UserID: claims.UserID,
		Send:   make(chan []byte, 64),
	}
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
