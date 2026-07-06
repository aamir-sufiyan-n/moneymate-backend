package config

import (
	"os"
)

type Config struct {
	AppPort    string
	AuthSvcURL string
}

func LoadConfig() *Config {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = ":3000"
	}

	authURL := os.Getenv("AUTH_SVC_URL")
	if authURL == "" {
		authURL = "localhost:50051"
	}

	return &Config{
		AppPort:    port,
		AuthSvcURL: authURL,
	}
}