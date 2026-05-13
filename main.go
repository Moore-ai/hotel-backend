package main

import (
	"log"

	"hotel-backend/config"
	"hotel-backend/internal/database"
	"hotel-backend/internal/handler"
	"hotel-backend/internal/middleware"
	"hotel-backend/internal/repository"
	"hotel-backend/internal/router"
	"hotel-backend/internal/service"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitPostgres(cfg.Database); err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	database.InitRedis(cfg.Redis)
	database.AutoMigrate()
	database.SeedAdmin()

	middleware.JWTSecret = cfg.JWT.Secret

	userRepo := repository.NewUserRepo(database.DB)
	roomRepo := repository.NewRoomRepo(database.DB)
	orderRepo := repository.NewOrderRepo(database.DB)
	checkinRepo := repository.NewCheckinRepo(database.DB)
	notifRepo := repository.NewNotificationRepo(database.DB)
	auditRepo := repository.NewAuditLogRepo(database.DB)
	scheduleRepo := repository.NewCheckoutScheduleRepo(database.DB)

	auditSvc := service.NewAuditLogService(auditRepo)
	authSvc := service.NewAuthService(userRepo, cfg.JWT)
	userSvc := service.NewUserService(userRepo, auditSvc)
	roomSvc := service.NewRoomService(roomRepo, auditSvc)
	orderSvc := service.NewOrderService(orderRepo, auditSvc)
	checkinSvc := service.NewCheckinService(checkinRepo, roomRepo, auditSvc)
	notifSvc := service.NewNotificationService(notifRepo)

	scheduler := service.NewScheduler(checkinRepo, scheduleRepo, notifSvc, checkinSvc, userRepo, cfg.Checkout)
	go scheduler.Run()

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	roomH := handler.NewRoomHandler(roomSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	checkinH := handler.NewCheckinHandler(checkinSvc)
	notifH := handler.NewNotificationHandler(notifSvc)
	auditH := handler.NewAuditLogHandler(auditSvc)

	r := router.Setup(authH, userH, roomH, orderH, checkinH, notifH, auditH)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
