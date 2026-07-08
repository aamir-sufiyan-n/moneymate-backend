package config

import (
	"fmt"
	"os"
	"time"

	sharedconfig "github.com/moneymate-2026/moneymate-backend/shared/config"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	HTTPAddr     string        `mapstructure:"http_addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type ServicesConfig struct {
    AuthGRPCAddr string `mapstructure:"auth_grpc_addr"` // auth:9091
    AuthHTTPAddr string `mapstructure:"auth_http_addr"` // http://auth:8081
    // add CoreGRPCAddr, MerchantGRPCAddr later
}

type RedisConfig struct {
    Addr              string        `mapstructure:"addr"`
    Password          string        // from env
    RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`
}

type RateLimitConfig struct {
    GlobalPerMinute  int `mapstructure:"global_requests_per_minute"`
    AuthPerMinute    int `mapstructure:"auth_requests_per_minute"`
    PublicPerMinute  int `mapstructure:"public_requests_per_minute"`
}

type CORSConfig struct {
    AllowedOrigins []string `mapstructure:"allowed_origins"`
    AllowedMethods []string `mapstructure:"allowed_methods"`
    AllowedHeaders []string `mapstructure:"allowed_headers"`
}
type JWTConfig struct {
    AccessSecret        string // from env — same value as auth-svc
    AccessExpiryMinutes int    `mapstructure:"access_expiry_minutes"`
}

type LogConfig struct {
    Level string `mapstructure:"level"`
}

type Config struct {
    Env       string
    Server    ServerConfig    `mapstructure:"server"`
    Services  ServicesConfig  `mapstructure:"services"`
    Redis     RedisConfig     `mapstructure:"redis"`
    RateLimit RateLimitConfig `mapstructure:"rate_limit"`
    CORS      CORSConfig      `mapstructure:"cors"`
    JWT       JWTConfig       `mapstructure:"jwt"`
    Log       LogConfig       `mapstructure:"log"`
}



func LoadConfig() (*Config, error) {
    yamlPath := os.Getenv("CONFIG_PATH")
    if yamlPath == "" {
        yamlPath = "./config/config.yaml"
    }

    v := viper.New()
    v.SetConfigFile(yamlPath)
    v.AutomaticEnv()

    // gateway only binds redis from env
    // no database bindings — gateway has no DB
    v.BindEnv("redis.addr",     "REDIS_ADDR")
    v.BindEnv("redis.password", "REDIS_PASSWORD")

    // allow env vars to override service addresses
    // useful in different deployment environments
    v.BindEnv("services.auth_grpc_addr", "AUTH_GRPC_ADDR")
    v.BindEnv("services.auth_http_addr", "AUTH_HTTP_ADDR")

    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    // secrets from env only — never in yaml
    cfg.JWT.AccessSecret = sharedconfig.MustGet("JWT_ACCESS_SECRET")
    cfg.Redis.Password   = sharedconfig.Get("REDIS_PASSWORD", "")
    cfg.Env              = sharedconfig.Get("ENVIRONMENT", "dev")

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
        {cfg.Services.AuthGRPCAddr, "services.auth_grpc_addr"},
        {cfg.Services.AuthHTTPAddr, "services.auth_http_addr"},
        {cfg.Redis.Addr,            "REDIS_ADDR"},
        {cfg.JWT.AccessSecret,      "JWT_ACCESS_SECRET"},
    }

    for _, r := range required {
        if r.value == "" {
            return fmt.Errorf("required config missing: %s", r.name)
        }
    }
    return nil
}