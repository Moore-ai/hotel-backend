package database

import (
	"log"
	"hotel-backend/internal/model"
	"hotel-backend/pkg/hash"
)

func SeedAdmin() error {
	var count int64
	DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		log.Println("Admin user already exists, skipping seed")
		return nil
	}

	pw, err := hash.HashPassword("admin123")
	if err != nil {
		return err
	}
	admin := &model.User{
		Username:     "admin",
		PasswordHash: pw,
		Role:         "admin",
		Name:         "超级管理员",
	}
	if err := DB.Create(admin).Error; err != nil {
		return err
	}
	log.Println("Admin user seeded (admin/admin123)")
	return nil
}
