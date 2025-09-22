package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"solemate/pkg/auth"
	"solemate/pkg/database"
	"solemate/services/notification-service/internal/config"
	"solemate/services/notification-service/internal/domain/entity"
	"solemate/services/notification-service/internal/domain/service"
	notificationHttp "solemate/services/notification-service/internal/handler/http"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgresConnection(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&entity.Notification{},
		&entity.NotificationTemplate{},
		&entity.NotificationPreference{},
		&entity.NotificationLog{},
		&entity.NotificationProvider{},
		&entity.NotificationQueue{},
		&entity.NotificationEvent{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	var notificationRepo interface{} = nil
	var templateRepo interface{} = nil
	var preferenceRepo interface{} = nil
	var logRepo interface{} = nil
	var queueRepo interface{} = nil
	var eventRepo interface{} = nil
	var userRepo interface{} = nil

	notificationService := service.NewNotificationService(
		notificationRepo,
		templateRepo,
		preferenceRepo,
		logRepo,
		queueRepo,
		eventRepo,
		userRepo,
	)

	templateService := service.NewTemplateService(templateRepo)
	preferenceService := service.NewPreferenceService(preferenceRepo, userRepo)

	jwtMiddleware := auth.JWTMiddleware(cfg.JWT.AccessSecret)
	adminMiddleware := func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}

	notificationHandler := notificationHttp.NewNotificationHandler(
		notificationService,
		templateService,
		preferenceService,
	)

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "notification-service",
		})
	})

	v1 := router.Group("/api/v1")
	notificationHandler.RegisterRoutes(v1, jwtMiddleware, adminMiddleware)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Notification service starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}