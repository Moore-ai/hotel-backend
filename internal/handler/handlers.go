package handler

import "hotel-backend/internal/service"

type Handlers struct {
	Auth    *AuthHandler
	User    *UserHandler
	Room    *RoomHandler
	Order   *OrderHandler
	Checkin *CheckinHandler
	Notif   *NotificationHandler
	Audit   *AuditLogHandler
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Auth:    NewAuthHandler(services.Auth),
		User:    NewUserHandler(services.User),
		Room:    NewRoomHandler(services.Room),
		Order:   NewOrderHandler(services.Order),
		Checkin: NewCheckinHandler(services.Checkin),
		Notif:   NewNotificationHandler(services.Notif),
		Audit:   NewAuditLogHandler(services.Audit),
	}
}
