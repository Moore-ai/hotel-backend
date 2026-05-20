package service

import (
	"hotel-backend/config"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/llm"
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
	Appeal   *AppealService
	Chat     *ChatService
}

func NewServices(repos *repository.Repositories, jwtCfg config.JWTConfig, hub *Hub, allocationCfg config.AllocationConfig, cancellationCfg config.CancellationConfig, appealCfg config.AppealConfig, llmCfg config.LLMConfig) *Services {
	audit := NewAuditLogService(repos.Audit)
	db := repos.DB()
	user := NewUserService(repos.User, repos.Guest, repos.Employee, repos.Admin, repos.Waiter, audit, db)
	strategy := NewStrategyFromConfig(allocationCfg.Strategy)
	allocator := NewRoomAllocator(repos.Room, repos.Order, strategy)
	obfKey := obfuscate.NewKey(jwtCfg.Secret)
	notifSvc := NewNotificationService(repos.Notif, hub, obfKey)
	waiterSvc := NewWaiterService(repos.Waiter, notifSvc, audit, db)
	orderSvc := NewOrderService(repos.Order, repos.Room, repos.User, allocator, audit, notifSvc, db, cancellationCfg.CutoffHours, obfKey, cancellationCfg.DefaultRejectReason, cancellationCfg.NotifyStrategy, cancellationCfg.NotifyStaffIDs)
	appealSvc := NewAppealService(repos.Appeal, repos.Order, repos.Room, repos.User, notifSvc, audit, db, appealCfg.ReviewStrategy, appealCfg.ReviewStaffIDs, obfKey)
	llmClientCfg := llm.Config{
		BaseURL:   llmCfg.BaseURL,
		APIKey:    llmCfg.APIKey,
		Model:     llmCfg.Model,
		MaxTokens: llmCfg.MaxTokens,
		Timeout:   llmCfg.Timeout,
	}
	llmClient := llm.NewAnthropicClient(llmClientCfg)
	return &Services{
		Auth:     NewAuthService(repos.User, repos.Guest, repos.Employee, repos.Admin, repos.Waiter, user, jwtCfg),
		User:     user,
		Guest:    NewGuestService(repos.Guest, audit),
		Employee: NewEmployeeService(repos.Employee, audit),
		Admin:    NewAdminService(repos.Admin, audit),
		Waiter:   waiterSvc,
		Room:     NewRoomService(repos.Room, audit),
		Order:    orderSvc,
		Checkin:  NewCheckinService(repos.Checkin, repos.Room, repos.Order, allocator, audit, db),
		Notif:    notifSvc,
		Audit:    audit,
		Appeal:   appealSvc,
		Chat:     NewChatService(llmClient, llmCfg.SystemPrompt, llmCfg.MaxHistory, waiterSvc, orderSvc, appealSvc),
	}
}
