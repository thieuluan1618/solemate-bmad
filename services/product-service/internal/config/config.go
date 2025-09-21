package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Elasticsearch ElasticsearchConfig
}

type ServerConfig struct {
	Port string
	Host string
	ENV  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
	Index    string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8081"),
			Host: getEnv("HOST", "0.0.0.0"),
			ENV:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "solemate"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "solemate_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Elasticsearch: ElasticsearchConfig{
			URL:      getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
			Username: getEnv("ELASTICSEARCH_USERNAME", ""),
			Password: getEnv("ELASTICSEARCH_PASSWORD", ""),
			Index:    getEnv("ELASTICSEARCH_INDEX", "products"),
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