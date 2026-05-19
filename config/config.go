package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Checkout   CheckoutConfig
	Admin      AdminConfig
	Allocation   AllocationConfig
	Cancellation CancellationConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret             string        `mapstructure:"secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

type CheckoutConfig struct {
	FirstThreshold    string `mapstructure:"first_threshold"`
	SecondThreshold   string `mapstructure:"second_threshold"`
	SchedulerInterval int    `mapstructure:"scheduler_interval"` // seconds
}

type AdminConfig struct {
	Username string
	Password string
	Name     string
}

type AllocationConfig struct {
	Strategy string `mapstructure:"strategy"`
}

type CancellationConfig struct {
	CutoffHours        int    `mapstructure:"cutoff_hours"`
	DefaultRejectReason string `mapstructure:"default_reject_reason"`
	NotifyStrategy     string `mapstructure:"notify_strategy"`
	NotifyStaffIDs     []uint `mapstructure:"notify_staff_ids"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
