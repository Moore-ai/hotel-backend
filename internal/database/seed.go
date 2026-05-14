package database

import (
	"log"

	"hotel-backend/config"
	"hotel-backend/internal/model"
	"hotel-backend/pkg/hash"
)

func SeedAdmin(cfg config.AdminConfig) error {
	var count int64
	DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		log.Println("Admin user already exists, skipping seed")
		return nil
	}

	pw, err := hash.HashPassword(cfg.Password)
	if err != nil {
		return err
	}
	admin := &model.User{
		Username:     cfg.Username,
		PasswordHash: pw,
		Role:         "admin",
		Name:         cfg.Name,
	}
	if err := DB.Create(admin).Error; err != nil {
		return err
	}
	log.Printf("Admin user seeded (%s/********)", cfg.Username)
	return nil
}
