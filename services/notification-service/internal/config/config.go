package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	JWT           JWTConfig
	Redis         RedisConfig
	Email         EmailConfig
	SMS           SMSConfig
	Push          PushConfig
	Queue         QueueConfig
	Notification  NotificationConfig
}

type ServerConfig struct {
	Host string
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type EmailConfig struct {
	Provider    string
	SMTPHost    string
	SMTPPort    string
	Username    string
	Password    string
	FromEmail   string
	FromName    string
	APIKey      string
	APISecret   string
	Enabled     bool
}

type SMSConfig struct {
	Provider    string
	AccountSID  string
	AuthToken   string
	FromNumber  string
	APIKey      string
	APISecret   string
	Enabled     bool
}

type PushConfig struct {
	Provider      string
	ServerKey     string
	ProjectID     string
	PrivateKey    string
	ClientEmail   string
	Enabled       bool
}

type QueueConfig struct {
	BatchSize       int
	RetryInterval   int
	MaxRetries      int
	ProcessInterval int
}

type NotificationConfig struct {
	MaxBulkSize           int
	DefaultChannel        string
	DefaultPriority       string
	TemplateRenderTimeout int
	RetentionDays         int
	EnableScheduled       bool
	EnableBatching        bool
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8087"),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "solemate_notification_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Email: EmailConfig{
			Provider:  getEnv("EMAIL_PROVIDER", "smtp"),
			SMTPHost:  getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:  getEnv("SMTP_PORT", "587"),
			Username:  getEnv("SMTP_USERNAME", ""),
			Password:  getEnv("SMTP_PASSWORD", ""),
			FromEmail: getEnv("EMAIL_FROM", "noreply@solemate.com"),
			FromName:  getEnv("EMAIL_FROM_NAME", "SoleMate"),
			APIKey:    getEnv("EMAIL_API_KEY", ""),
			APISecret: getEnv("EMAIL_API_SECRET", ""),
			Enabled:   getEnvAsBool("EMAIL_ENABLED", true),
		},
		SMS: SMSConfig{
			Provider:   getEnv("SMS_PROVIDER", "twilio"),
			AccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
			AuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
			FromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
			APIKey:     getEnv("SMS_API_KEY", ""),
			APISecret:  getEnv("SMS_API_SECRET", ""),
			Enabled:    getEnvAsBool("SMS_ENABLED", false),
		},
		Push: PushConfig{
			Provider:    getEnv("PUSH_PROVIDER", "fcm"),
			ServerKey:   getEnv("FCM_SERVER_KEY", ""),
			ProjectID:   getEnv("FCM_PROJECT_ID", ""),
			PrivateKey:  getEnv("FCM_PRIVATE_KEY", ""),
			ClientEmail: getEnv("FCM_CLIENT_EMAIL", ""),
			Enabled:     getEnvAsBool("PUSH_ENABLED", false),
		},
		Queue: QueueConfig{
			BatchSize:       getEnvAsInt("QUEUE_BATCH_SIZE", 10),
			RetryInterval:   getEnvAsInt("QUEUE_RETRY_INTERVAL", 300),
			MaxRetries:      getEnvAsInt("QUEUE_MAX_RETRIES", 3),
			ProcessInterval: getEnvAsInt("QUEUE_PROCESS_INTERVAL", 60),
		},
		Notification: NotificationConfig{
			MaxBulkSize:           getEnvAsInt("NOTIFICATION_MAX_BULK_SIZE", 1000),
			DefaultChannel:        getEnv("NOTIFICATION_DEFAULT_CHANNEL", "email"),
			DefaultPriority:       getEnv("NOTIFICATION_DEFAULT_PRIORITY", "medium"),
			TemplateRenderTimeout: getEnvAsInt("NOTIFICATION_TEMPLATE_TIMEOUT", 30),
			RetentionDays:         getEnvAsInt("NOTIFICATION_RETENTION_DAYS", 90),
			EnableScheduled:       getEnvAsBool("NOTIFICATION_ENABLE_SCHEDULED", true),
			EnableBatching:        getEnvAsBool("NOTIFICATION_ENABLE_BATCHING", true),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}