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

func NewServices(repos *repository.Repositories, jwtCfg config.JWTConfig, hub *Hub, allocationCfg config.AllocationConfig) *Services {
	audit := NewAuditLogService(repos.Audit)
	user := NewUserService(repos.User, audit)
	strategy := NewStrategyFromConfig(allocationCfg.Strategy)
	allocator := NewRoomAllocator(repos.Room, repos.Order, strategy)
	db := repos.DB()
	return &Services{
		Auth:    NewAuthService(repos.User, user, jwtCfg),
		User:    user,
		Room:    NewRoomService(repos.Room, audit),
		Order:   NewOrderService(repos.Order, repos.Room, allocator, audit, db),
		Checkin: NewCheckinService(repos.Checkin, repos.Room, repos.Order, allocator, audit, db),
		Notif:   NewNotificationService(repos.Notif, hub),
		Audit:   audit,
	}
}
