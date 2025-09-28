package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Host string
	ENV  string
}

type ServicesConfig struct {
	UserServiceURL    string
	ProductServiceURL string
	CartServiceURL    string
	OrderServiceURL   string
	PaymentServiceURL string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8000"),
			Host: getEnv("HOST", "0.0.0.0"),
			ENV:  getEnv("ENV", "development"),
		},
		Services: ServicesConfig{
			UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8080"),
			ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
			CartServiceURL:    getEnv("CART_SERVICE_URL", "http://localhost:8083"),
			OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8084"),
			PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "default-access-secret"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "default-refresh-secret"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
