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

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	adminUser := &model.User{
		Username:     cfg.Username,
		PasswordHash: pw,
		Role:         "admin",
	}
	if err := tx.Create(adminUser).Error; err != nil {
		tx.Rollback()
		return err
	}

	adminProfile := &model.Admin{
		UserID: adminUser.ID,
		Name:   cfg.Name,
	}
	if err := tx.Create(adminProfile).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Admin user seeded (%s/********)", cfg.Username)
	return nil
}
