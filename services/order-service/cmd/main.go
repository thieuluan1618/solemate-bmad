package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"solemate/pkg/auth"
	"solemate/pkg/database"
	"solemate/services/order-service/internal/config"
	orderHttp "solemate/services/order-service/internal/handler/http"
	orderDatabase "solemate/services/order-service/internal/infrastructure/database"
	"solemate/services/order-service/internal/domain/service"
	"solemate/services/order-service/internal/domain/entity"
	"solemate/services/order-service/internal/domain/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
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

	// Auto-migrate database schema
	if err := db.AutoMigrate(&entity.Order{}, &entity.OrderItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	orderRepo := orderDatabase.NewOrderRepository(db)

	// For now, we'll pass nil for external service repositories
	// In production, these would be HTTP clients or gRPC clients
	var cartRepo repository.CartRepository = nil
	var productRepo repository.ProductRepository = nil
	var notificationRepo repository.NotificationRepository = nil

	// Initialize services
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, notificationRepo)

	// Initialize middleware
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

	// Initialize handlers
	orderHandler := orderHttp.NewOrderHandler(orderService)

	// Setup router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "order-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	orderHandler.RegisterRoutes(v1, jwtMiddleware, adminMiddleware)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Order service starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}