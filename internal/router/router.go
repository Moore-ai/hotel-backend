package router

import (
	"github.com/gin-gonic/gin"
	"hotel-backend/internal/handler"
	"hotel-backend/internal/middleware"
)

func Setup(
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	roomH *handler.RoomHandler,
	orderH *handler.OrderHandler,
	checkinH *handler.CheckinHandler,
	notifH *handler.NotificationHandler,
	auditH *handler.AuditLogHandler,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Logger())

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/logout", middleware.Auth(), authH.Logout)
	}

	api.Use(middleware.Auth())

	users := api.Group("/users", middleware.RequireRoles("employee", "admin"))
	{
		users.GET("", userH.List)
		users.GET("/:id", userH.Get)
		users.POST("", userH.Create)
		users.PUT("/:id", userH.Update)
		users.DELETE("/:id", userH.Delete)
	}

	rooms := api.Group("/rooms")
	{
		rooms.GET("", roomH.List)
		rooms.GET("/:id", roomH.Get)
		rooms.POST("", middleware.RequireRoles("employee", "admin"), roomH.Create)
		rooms.PUT("/:id", middleware.RequireRoles("employee", "admin"), roomH.Update)
		rooms.DELETE("/:id", middleware.RequireRoles("employee", "admin"), roomH.Delete)
	}

	orders := api.Group("/orders")
	{
		orders.GET("", orderH.List)
		orders.GET("/:id", orderH.Get)
		orders.POST("", orderH.Create)
		orders.PUT("/:id", middleware.RequireRoles("employee", "admin"), orderH.Update)
		orders.DELETE("/:id", middleware.RequireRoles("employee", "admin"), orderH.Delete)
	}

	checkins := api.Group("/checkins", middleware.RequireRoles("employee", "admin"))
	{
		checkins.GET("", checkinH.List)
		checkins.GET("/:id", checkinH.Get)
		checkins.POST("", checkinH.Create)
		checkins.PUT("/:id/checkout", checkinH.Checkout)
		checkins.DELETE("/:id", checkinH.Delete)
	}

	notifs := api.Group("/notifications")
	{
		notifs.GET("", notifH.List)
		notifs.GET("/unread", notifH.UnreadCount)
		notifs.PUT("/:id/read", notifH.MarkRead)
	}

	audit := api.Group("/audit-logs", middleware.RequireRoles("employee", "admin"))
	{
		audit.GET("", auditH.List)
	}

	return r
}
