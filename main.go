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
	cfg := mustLoadConfig()

	if err := database.InitAll(cfg); err != nil {
		log.Fatalf("Failed to init infrastructure: %v", err)
	}

	middleware.JWTSecret = cfg.JWT.Secret

	hub := service.NewHub()
	go hub.Run()

	repos := repository.NewRepositories(database.DB)
	services := service.NewServices(repos, cfg.JWT, hub, cfg.Allocation, cfg.Cancellation, cfg.Appeal)
	handlers := handler.NewHandlers(services, repos, hub, cfg.JWT)

	go service.NewScheduler(repos.Checkin, repos.Schedule, services.Notif, services.Checkin, repos.User, repos.Guest, cfg.Checkout).Run()

	r := router.Setup(handlers)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func mustLoadConfig() *config.Config {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}
