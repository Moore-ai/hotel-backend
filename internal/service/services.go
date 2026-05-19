package service

import (
	"hotel-backend/config"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/obfuscate"
)

type Services struct {
	Auth     *AuthService
	User     *UserService
	Guest    *GuestService
	Employee *EmployeeService
	Admin    *AdminService
	Waiter   *WaiterService
	Room     *RoomService
	Order    *OrderService
	Checkin  *CheckinService
	Notif    *NotificationService
	Audit    *AuditLogService
}

func NewServices(repos *repository.Repositories, jwtCfg config.JWTConfig, hub *Hub, allocationCfg config.AllocationConfig, cancellationCfg config.CancellationConfig) *Services {
	audit := NewAuditLogService(repos.Audit)
	db := repos.DB()
	user := NewUserService(repos.User, repos.Guest, repos.Employee, repos.Admin, repos.Waiter, audit, db)
	strategy := NewStrategyFromConfig(allocationCfg.Strategy)
	allocator := NewRoomAllocator(repos.Room, repos.Order, strategy)
	notifSvc := NewNotificationService(repos.Notif, hub)
	obfKey := obfuscate.NewKey(jwtCfg.Secret)
	return &Services{
		Auth:     NewAuthService(repos.User, repos.Guest, repos.Employee, repos.Admin, repos.Waiter, user, jwtCfg),
		User:     user,
		Guest:    NewGuestService(repos.Guest, audit),
		Employee: NewEmployeeService(repos.Employee, audit),
		Admin:    NewAdminService(repos.Admin, audit),
		Waiter:   NewWaiterService(repos.Waiter, notifSvc, audit, db),
		Room:     NewRoomService(repos.Room, audit),
		Order:    NewOrderService(repos.Order, repos.Room, repos.User, allocator, audit, notifSvc, db, cancellationCfg.CutoffHours, obfKey, cancellationCfg.DefaultRejectReason),
		Checkin:  NewCheckinService(repos.Checkin, repos.Room, repos.Order, allocator, audit, db),
		Notif:    notifSvc,
		Audit:    audit,
	}
}
