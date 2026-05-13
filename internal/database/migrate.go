package database

import (
	"log"
	"hotel-backend/internal/model"
)

func AutoMigrate() error {
	err := DB.AutoMigrate(
		&model.User{},
		&model.Room{},
		&model.Order{},
		&model.Checkin{},
		&model.Notification{},
		&model.CheckoutSchedule{},
		&model.AuditLog{},
	)
	if err != nil {
		return err
	}
	log.Println("Database migration completed")
	return nil
}
