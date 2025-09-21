package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"solemate/api-gateway/internal/config"
	"solemate/api-gateway/internal/handler"
	"solemate/pkg/auth"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager()

	// Initialize proxy handler
	proxyHandler := handler.NewProxyHandler(
		cfg.Services.UserServiceURL,
		cfg.Services.ProductServiceURL,
		cfg.Services.CartServiceURL,
		cfg.Services.OrderServiceURL,
		cfg.Services.PaymentServiceURL,
	)

	// Setup routes
	router := handler.SetupRoutes(proxyHandler, jwtManager)

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("API Gateway starting on %s", serverAddr)
	log.Printf("Proxying to services:")
	log.Printf("  User Service: %s", cfg.Services.UserServiceURL)
	log.Printf("  Product Service: %s", cfg.Services.ProductServiceURL)
	log.Printf("  Cart Service: %s", cfg.Services.CartServiceURL)
	log.Printf("  Order Service: %s", cfg.Services.OrderServiceURL)
	log.Printf("  Payment Service: %s", cfg.Services.PaymentServiceURL)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
