package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"solemate/pkg/auth"
	"solemate/pkg/database"
	"solemate/services/payment-service/internal/config"
	paymentHandlers "solemate/services/payment-service/internal/handler/http"
	paymentDatabase "solemate/services/payment-service/internal/infrastructure/database"
	paymentHttp "solemate/services/payment-service/internal/infrastructure/http"
	"solemate/services/payment-service/internal/infrastructure/stripe"
	"solemate/services/payment-service/internal/domain/service"
	"solemate/services/payment-service/internal/domain/entity"
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
		&entity.Payment{},
		&entity.PaymentMethod{},
		&entity.Refund{},
		&entity.WebhookEvent{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	paymentRepo := paymentDatabase.NewPaymentRepository(db)
	paymentMethodRepo := paymentDatabase.NewPaymentMethodRepository(db)
	refundRepo := paymentDatabase.NewRefundRepository(db)
	webhookRepo := paymentDatabase.NewWebhookRepository(db)
	stripeRepo := stripe.NewStripeRepository(cfg.Stripe.APIKey, cfg.Stripe.WebhookSecret)

	// Initialize order repository (HTTP client to order service)
	orderRepo := paymentHttp.NewOrderRepository("http://localhost:8082")

	// Initialize services
	paymentService := service.NewPaymentService(
		paymentRepo,
		paymentMethodRepo,
		refundRepo,
		webhookRepo,
		stripeRepo,
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
	paymentHandler := paymentHandlers.NewPaymentHandler(paymentService)

	// Setup router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "payment-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	paymentHandler.RegisterRoutes(v1, jwtMiddleware, adminMiddleware)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Payment service starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}