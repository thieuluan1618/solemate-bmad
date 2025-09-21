package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"solemate/pkg/auth"
	"solemate/pkg/cache"
	"solemate/services/cart-service/internal/config"
	cartHttp "solemate/services/cart-service/internal/handler/http"
	cartCache "solemate/services/cart-service/internal/infrastructure/cache"
	"solemate/services/cart-service/internal/domain/service"
	"solemate/services/cart-service/internal/domain/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	// Test Redis connection
	if err := cache.TestRedisConnection(redisClient); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repositories
	cartRepo := cartCache.NewCartRepository(redisClient)

	// For now, we'll pass nil for productRepo since we don't have cross-service communication yet
	// In production, this would be a gRPC client to product service
	var productRepo repository.ProductRepository = nil

	// Initialize services
	cartService := service.NewCartService(cartRepo, productRepo)

	// Initialize JWT middleware
	jwtMiddleware := auth.JWTMiddleware(cfg.JWT.AccessSecret)

	// Initialize handlers
	cartHandler := cartHttp.NewCartHandler(cartService)

	// Setup router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "cart-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	cartHandler.RegisterRoutes(v1, jwtMiddleware)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Cart service starting on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}