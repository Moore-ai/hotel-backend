package handler

import (
	"hotel-backend/config"
	"hotel-backend/internal/repository"
	"hotel-backend/internal/service"
)

type Handlers struct {
	Auth     *AuthHandler
	Guest    *GuestHandler
	Employee *EmployeeHandler
	Admin    *AdminHandler
	Waiter   *WaiterHandler
	Room     *RoomHandler
	Order    *OrderHandler
	Checkin  *CheckinHandler
	Notif    *NotificationHandler
	Audit    *AuditLogHandler
	Appeal   *AppealHandler
	WS       *WSHandler
}

func NewHandlers(services *service.Services, repos *repository.Repositories, hub *service.Hub, jwtCfg config.JWTConfig) *Handlers {
	return &Handlers{
		Auth:     NewAuthHandler(services.Auth, jwtCfg),
		Guest:    NewGuestHandler(services.Guest, services.User),
		Employee: NewEmployeeHandler(services.Employee, services.User),
		Admin:    NewAdminHandler(services.Admin, services.User),
		Waiter:   NewWaiterHandler(services.Waiter, services.Checkin, services.User),
		Room:     NewRoomHandler(services.Room),
		Order:    NewOrderHandler(services.Order),
		Checkin:  NewCheckinHandler(services.Checkin),
		Notif:    NewNotificationHandler(services.Notif),
		Audit:    NewAuditLogHandler(services.Audit),
		Appeal:   NewAppealHandler(services.Appeal, services.Order),
		WS:       NewWSHandler(hub, jwtCfg),
	}
}
