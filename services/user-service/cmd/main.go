package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"solemate/pkg/auth"
	"solemate/pkg/database"
	"solemate/services/user-service/internal/config"
	"solemate/services/user-service/internal/domain/entity"
	"solemate/services/user-service/internal/domain/service"
	httpHandler "solemate/services/user-service/internal/handler/http"
	dbImpl "solemate/services/user-service/internal/infrastructure/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database schema
	if err := db.AutoMigrate(&entity.User{}, &entity.Address{}, &entity.WishlistItem{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	userRepo := dbImpl.NewUserRepository(db)
	addressRepo := dbImpl.NewAddressRepository(db)
	wishlistRepo := dbImpl.NewWishlistRepository(db)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager()

	// Initialize services
	userService := service.NewUserService(userRepo, addressRepo, jwtManager)
	wishlistService := service.NewWishlistService(wishlistRepo)

	// Initialize handlers
	userHandler := httpHandler.NewUserHandler(userService)
	wishlistHandler := httpHandler.NewWishlistHandler(wishlistService)

	// Setup routes
	router := httpHandler.SetupRoutes(userHandler, wishlistHandler, jwtManager)

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("User service starting on %s", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
