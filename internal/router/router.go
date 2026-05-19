package router

import (
	"hotel-backend/internal/handler"
	"hotel-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

var staffOnly = middleware.RequireRoles("employee", "admin")
var adminOnly = middleware.RequireRoles("admin")

func Setup(handlers *handler.Handlers) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Logger())

	api := r.Group("/api/v1")

	registerAuthRoutes(api, handlers.Auth)
	registerWSRoutes(api, handlers.WS)

	api.Use(middleware.Auth())

	registerGuestRoutes(api, handlers.Guest)
	registerEmployeeRoutes(api, handlers.Employee)
	registerAdminRoutes(api, handlers.Admin)
	registerWaiterRoutes(api, handlers.Waiter)
	registerRoomRoutes(api, handlers.Room)
	registerOrderRoutes(api, handlers.Order)
	registerCheckinRoutes(api, handlers.Checkin)
	registerNotificationRoutes(api, handlers.Notif)
	registerAuditLogRoutes(api, handlers.Audit)
	registerAppealRoutes(api, handlers.Appeal, handlers.Order)

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, authH *handler.AuthHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/staff-login", authH.StaffLogin)
		auth.POST("/admin-login", authH.AdminLogin)
		auth.POST("/register", authH.Register)
		auth.POST("/logout", middleware.Auth(), authH.Logout)
		auth.DELETE("/account", middleware.Auth(), authH.DeleteAccount)
	}
}

func registerGuestRoutes(api *gin.RouterGroup, guestH *handler.GuestHandler) {
	guests := api.Group("/guests")
	{
		guests.GET("", guestH.List)
		guests.GET("/:id", guestH.Get)
		guests.POST("", guestH.Create)
		guests.PUT("/:id", guestH.Update)
		guests.DELETE("/:id", guestH.Delete)
	}
}

func registerEmployeeRoutes(api *gin.RouterGroup, empH *handler.EmployeeHandler) {
	employees := api.Group("/employees", adminOnly)
	{
		employees.GET("", empH.List)
		employees.GET("/:id", empH.Get)
		employees.POST("", empH.Create)
		employees.PUT("/:id", empH.Update)
		employees.DELETE("/:id", empH.Delete)
	}
}

func registerAdminRoutes(api *gin.RouterGroup, adminH *handler.AdminHandler) {
	admins := api.Group("/admins", adminOnly)
	{
		admins.GET("", adminH.List)
		admins.GET("/:id", adminH.Get)
		admins.POST("", adminH.Create)
		admins.PUT("/:id", adminH.Update)
		admins.DELETE("/:id", adminH.Delete)
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
		orders.GET("/cancel-requests", staffOnly, orderH.ListCancelRequests)
		orders.GET("/:code", orderH.Get)
		orders.POST("", orderH.Create)
		orders.POST("/:code/cancel", orderH.Cancel)
		orders.POST("/:code/reject-cancel", staffOnly, orderH.RejectCancel)
		orders.POST("/:code/confirm", staffOnly, orderH.Confirm)
		orders.PUT("/:code", staffOnly, orderH.Update)
		orders.DELETE("/:code", staffOnly, orderH.Delete)
	}
}

func registerWaiterRoutes(api *gin.RouterGroup, waiterH *handler.WaiterHandler) {
	waiters := api.Group("/waiters", adminOnly)
	{
		waiters.GET("", waiterH.List)
		waiters.GET("/:id", waiterH.Get)
		waiters.POST("", waiterH.Create)
		waiters.PUT("/:id", waiterH.Update)
		waiters.DELETE("/:id", waiterH.Delete)
	}

	api.POST("/checkins/:id/service-request", waiterH.ServiceRequest)
	api.POST("/waiters/:id/complete-service", middleware.RequireRoles("waiter", "admin"), waiterH.CompleteService)
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
		notifs.PUT("/:code/read", notifH.MarkRead)
	}
}

func registerAppealRoutes(api *gin.RouterGroup, appealH *handler.AppealHandler, orderH *handler.OrderHandler) {
	orders := api.Group("/orders")
	orders.POST("/:code/appeal", appealH.Create)

	appeals := api.Group("/appeals", staffOnly)
	{
		appeals.GET("", appealH.List)
		appeals.POST("/:id/review", appealH.Review)
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
