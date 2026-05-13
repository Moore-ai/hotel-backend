package database

import "hotel-backend/config"

func InitAll(cfg *config.Config) error {
	if err := InitPostgres(cfg.Database); err != nil {
		return err
	}
	InitRedis(cfg.Redis)
	if err := AutoMigrate(); err != nil {
		return err
	}
	if err := SeedAdmin(); err != nil {
		return err
	}
	return nil
}
