package config

import (
	"fmt"
	"os"

	sharedconfig "github.com/moneymate-2026/moneymate-backend/shared/config"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port        string `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type ServicesConfig struct {
	AuthAddr   string            `mapstructure:"auth_addr"`
	Downstream map[string]string `mapstructure:"downstream"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string
	DB       int `mapstructure:"db"`
}

type RateLimitConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	MaxRequests   int  `mapstructure:"max_requests"`
	WindowSeconds int  `mapstructure:"window_seconds"`
}

type CORSConfig struct {
	AllowOrigins string `mapstructure:"allow_origins"`
	AllowMethods string `mapstructure:"allow_methods"`
	AllowHeaders string `mapstructure:"allow_headers"`
}

type JWTConfig struct {
	Secret string
}

type TracingConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	CollectorURL string `mapstructure:"collector_url"`
}

type Config struct {
	Server       ServerConfig    `mapstructure:"server"`
	Services     ServicesConfig  `mapstructure:"services"`
	Redis        RedisConfig     `mapstructure:"redis"`
	JWT          JWTConfig
	RateLimiting RateLimitConfig `mapstructure:"rate_limiting"`
	CORS         CORSConfig      `mapstructure:"cors"`
	Tracing      TracingConfig   `mapstructure:"tracing"`
}

func LoadConfig() (*Config, error) {
	yamlPath := os.Getenv("CONFIG_PATH")
	if yamlPath == "" {
		yamlPath = "./config/config.yaml"
	}

	v := viper.New()
	v.SetConfigFile(yamlPath)
	v.AutomaticEnv()

	v.BindEnv("redis.address", "REDIS_ADDRESS")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("services.auth_addr", "SERVICES_AUTH_ADDR")
	v.BindEnv("server.port", "SERVER_PORT")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.JWT.Secret = sharedconfig.Get("JWT_ACCESS_SECRET", "")
	cfg.Redis.Password = sharedconfig.Get("REDIS_PASSWORD", "")

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	required := []struct {
		value string
		name  string
	}{
		{cfg.Services.AuthAddr, "services.auth_addr"},
		{cfg.Redis.Address, "redis.address"},
	}

	for _, r := range required {
		if r.value == "" {
			return fmt.Errorf("required config missing: %s", r.name)
		}
	}
	return nil
}
