package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"solemate/pkg/auth"
	"solemate/pkg/database"
	"solemate/services/inventory-service/internal/config"
	inventoryHttp "solemate/services/inventory-service/internal/handler/http"
	"solemate/services/inventory-service/internal/domain/service"
	"solemate/services/inventory-service/internal/domain/entity"
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
	if err := db.AutoMigrate(
		&entity.Warehouse{},
		&entity.InventoryItem{},
		&entity.StockMovement{},
		&entity.StockReservation{},
		&entity.StockAlert{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// For now, we'll pass nil for repositories that need implementation
	// In production, these would be proper database implementations
	var inventoryRepo interface{} = nil
	var warehouseRepo interface{} = nil
	var movementRepo interface{} = nil
	var reservationRepo interface{} = nil
	var alertRepo interface{} = nil
	var productRepo interface{} = nil
	var orderRepo interface{} = nil

	// Initialize services
	inventoryService := service.NewInventoryService(
		inventoryRepo,
		warehouseRepo,
		movementRepo,
		reservationRepo,
		alertRepo,
		productRepo,
		orderRepo,
	)

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
	inventoryHandler := inventoryHttp.NewInventoryHandler(inventoryService)

	// Setup router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "inventory-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	inventoryHandler.RegisterRoutes(v1, jwtMiddleware, adminMiddleware)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Inventory service starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}