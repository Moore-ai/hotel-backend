package service

import (
	"hotel-backend/config"
	"hotel-backend/internal/repository"
)

type Services struct {
	Auth    *AuthService
	User    *UserService
	Room    *RoomService
	Order   *OrderService
	Checkin *CheckinService
	Notif   *NotificationService
	Audit   *AuditLogService
}

func NewServices(repos *repository.Repositories, jwtCfg config.JWTConfig, hub *Hub) *Services {
	audit := NewAuditLogService(repos.Audit)
	return &Services{
		Auth:    NewAuthService(repos.User, jwtCfg),
		User:    NewUserService(repos.User, audit),
		Room:    NewRoomService(repos.Room, audit),
		Order:   NewOrderService(repos.Order, audit),
		Checkin: NewCheckinService(repos.Checkin, repos.Room, audit),
		Notif:   NewNotificationService(repos.Notif, hub),
		Audit:   audit,
	}
}
