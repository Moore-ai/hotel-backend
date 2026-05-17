package service

import (
	"hotel-backend/config"
	"hotel-backend/internal/repository"
)

type Services struct {
	Auth     *AuthService
	User     *UserService
	Guest    *GuestService
	Employee *EmployeeService
	Admin    *AdminService
	Room     *RoomService
	Order    *OrderService
	Checkin  *CheckinService
	Notif    *NotificationService
	Audit    *AuditLogService
}

func NewServices(repos *repository.Repositories, jwtCfg config.JWTConfig, hub *Hub, allocationCfg config.AllocationConfig) *Services {
	audit := NewAuditLogService(repos.Audit)
	db := repos.DB()
	user := NewUserService(repos.User, repos.Guest, repos.Employee, repos.Admin, audit, db)
	strategy := NewStrategyFromConfig(allocationCfg.Strategy)
	allocator := NewRoomAllocator(repos.Room, repos.Order, strategy)
	return &Services{
		Auth:     NewAuthService(repos.User, repos.Guest, repos.Employee, repos.Admin, user, jwtCfg),
		User:     user,
		Guest:    NewGuestService(repos.Guest, audit),
		Employee: NewEmployeeService(repos.Employee, audit),
		Admin:    NewAdminService(repos.Admin, audit),
		Room:     NewRoomService(repos.Room, audit),
		Order:    NewOrderService(repos.Order, repos.Room, allocator, audit, db),
		Checkin:  NewCheckinService(repos.Checkin, repos.Room, repos.Order, allocator, audit, db),
		Notif:    NewNotificationService(repos.Notif, hub),
		Audit:    audit,
	}
}
