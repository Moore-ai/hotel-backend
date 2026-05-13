// internal/database/database.go
package database

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"hotel-backend/config"
)

var DB *gorm.DB
var RDB *redis.Client

func InitPostgres(cfg config.DatabaseConfig) error {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	log.Println("PostgreSQL connected")
	return nil
}

func InitRedis(cfg config.RedisConfig) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	log.Println("Redis connected")
}
