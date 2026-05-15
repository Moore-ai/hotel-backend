package router

import (
	"github.com/gin-gonic/gin"
	"hotel-backend/internal/handler"
	"hotel-backend/internal/middleware"
)

var staffOnly = middleware.RequireRoles("employee", "admin")

func Setup(handlers *handler.Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Logger())

	api := r.Group("/api/v1")

	registerAuthRoutes(api, handlers.Auth)
	registerWSRoutes(api, handlers.WS)

	api.Use(middleware.Auth())

	registerUserRoutes(api, handlers.User)
	registerRoomRoutes(api, handlers.Room)
	registerOrderRoutes(api, handlers.Order)
	registerCheckinRoutes(api, handlers.Checkin)
	registerNotificationRoutes(api, handlers.Notif)
	registerAuditLogRoutes(api, handlers.Audit)

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, authH *handler.AuthHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/register", authH.Register)
		auth.POST("/logout", middleware.Auth(), authH.Logout)
		auth.DELETE("/account", middleware.Auth(), authH.DeleteAccount)
	}
}

func registerUserRoutes(api *gin.RouterGroup, userH *handler.UserHandler) {
	users := api.Group("/users", staffOnly)
	{
		users.GET("", userH.List)
		users.GET("/:id", userH.Get)
		users.POST("", userH.Create)
		users.PUT("/:id", userH.Update)
		users.DELETE("/:id", userH.Delete)
	}
}

func registerRoomRoutes(api *gin.RouterGroup, roomH *handler.RoomHandler) {
	rooms := api.Group("/rooms")
	{
		rooms.GET("", roomH.List)
		rooms.GET("/:id", roomH.Get)
		rooms.POST("", staffOnly, roomH.Create)
		rooms.PUT("/:id", staffOnly, roomH.Update)
		rooms.DELETE("/:id", staffOnly, roomH.Delete)
	}
}

func registerOrderRoutes(api *gin.RouterGroup, orderH *handler.OrderHandler) {
	orders := api.Group("/orders")
	{
		orders.GET("", orderH.List)
		orders.GET("/:id", orderH.Get)
		orders.POST("", orderH.Create)
		orders.PUT("/:id", staffOnly, orderH.Update)
		orders.DELETE("/:id", staffOnly, orderH.Delete)
	}
}

func registerCheckinRoutes(api *gin.RouterGroup, checkinH *handler.CheckinHandler) {
	checkins := api.Group("/checkins", staffOnly)
	{
		checkins.GET("", checkinH.List)
		checkins.GET("/:id", checkinH.Get)
		checkins.POST("", checkinH.Create)
		checkins.PUT("/:id/checkout", checkinH.Checkout)
		checkins.DELETE("/:id", checkinH.Delete)
	}
}

func registerNotificationRoutes(api *gin.RouterGroup, notifH *handler.NotificationHandler) {
	notifs := api.Group("/notifications")
	{
		notifs.GET("", notifH.List)
		notifs.GET("/unread", notifH.UnreadCount)
		notifs.PUT("/:id/read", notifH.MarkRead)
	}
}

func registerAuditLogRoutes(api *gin.RouterGroup, auditH *handler.AuditLogHandler) {
	audit := api.Group("/audit-logs", staffOnly)
	{
		audit.GET("", auditH.List)
	}
}

func registerWSRoutes(api *gin.RouterGroup, wsH *handler.WSHandler) {
	api.GET("/ws", wsH.Connect)
}
