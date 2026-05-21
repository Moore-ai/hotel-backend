package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	Checkout     CheckoutConfig
	Admin        AdminConfig
	Allocation   AllocationConfig
	Cancellation CancellationConfig
	Appeal       AppealConfig
	LLM          LLMConfig
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
	CutoffHours         int    `mapstructure:"cutoff_hours"`
	DefaultRejectReason string `mapstructure:"default_reject_reason"`
	NotifyStrategy      string `mapstructure:"notify_strategy"`
	NotifyStaffIDs      []uint `mapstructure:"notify_staff_ids"`
}

type AppealConfig struct {
	ReviewStrategy string `mapstructure:"review_strategy"`
	ReviewStaffIDs []uint `mapstructure:"review_staff_ids"`
}

type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	MaxRequestsPerMinute  int  `mapstructure:"max_requests_per_minute"`
	Algorithm             string `mapstructure:"algorithm"`
}

type LLMConfig struct {
	Provider     string          `mapstructure:"provider"`      // anthropic | ollama
	BaseURL      string          `mapstructure:"base_url"`
	APIKey       string          `mapstructure:"api_key"`
	Model        string          `mapstructure:"model"`
	MaxTokens    int             `mapstructure:"max_tokens"`
	Timeout      time.Duration   `mapstructure:"timeout"`
	SystemPrompt string          `mapstructure:"system_prompt"`
	MaxHistory   int             `mapstructure:"max_history"`
	RateLimit    RateLimitConfig `mapstructure:"rate_limit"`
}

// 读取 .env 文件并设置到环境变量（不覆盖已存在的系统环境变量）
func loadEnv(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	n := 0
	for line := range strings.SplitSeq(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
			n++
		}
	}
	log.Printf("Config: loaded %d variables from %s (existing env vars preserved)", n, path)
}

func Load(path string) (*Config, error) {
	loadEnv(".env")

	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	expanded := os.ExpandEnv(string(raw))
	if err := v.ReadConfig(bytes.NewReader([]byte(expanded))); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 记录 LLM API key 来源
	keySource := "config.yaml"
	if _, set := os.LookupEnv("LLM_API_KEY"); set {
		keySource = "system environment"
	} else if _, set := os.LookupEnv("LLM_API_KEY"); set {
		keySource = "system environment (LLM_API_KEY)"
	}
	if cfg.LLM.APIKey != "" {
		prefix := cfg.LLM.APIKey
		if len(prefix) > 8 {
			prefix = prefix[:8] + "..."
		}
		log.Printf("Config: LLM API key from %s (%s)", keySource, prefix)
	}

	return &cfg, nil
}
