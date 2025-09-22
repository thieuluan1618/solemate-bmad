package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	External ExternalConfig
}

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxIdle  int
	MaxOpen  int
	MaxLife  time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
}

type ExternalConfig struct {
	CartServiceURL    string
	ProductServiceURL string
	PaymentServiceURL string
	NotificationServiceURL string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8084"),
			Host: getEnv("HOST", "0.0.0.0"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "solemate"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "solemate_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxIdle:  getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpen:  getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			MaxLife:  getEnvAsDuration("DB_CONN_MAX_LIFETIME", 1*time.Hour),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		},
		External: ExternalConfig{
			CartServiceURL:    getEnv("CART_SERVICE_URL", "http://localhost:8083"),
			ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
			PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
			NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8086"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}