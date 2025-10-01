package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"solemate/pkg/auth"
	"solemate/pkg/database"
	"solemate/services/product-service/internal/config"
	"solemate/services/product-service/internal/domain/entity"
	"solemate/services/product-service/internal/domain/service"
	httpHandler "solemate/services/product-service/internal/handler/http"
	dbImpl "solemate/services/product-service/internal/infrastructure/database"
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
	if err := db.AutoMigrate(
		&entity.Product{},
		&entity.Category{},
		&entity.Brand{},
		&entity.ProductVariant{},
		&entity.ProductImage{},
		&entity.Review{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	productRepo := dbImpl.NewProductRepository(db)
	categoryRepo := dbImpl.NewCategoryRepository(db)
	brandRepo := dbImpl.NewBrandRepository(db)
	reviewRepo := dbImpl.NewReviewRepository(db)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager()

	// Initialize services
	productService := service.NewProductService(productRepo, categoryRepo, brandRepo, nil, nil)
	categoryService := service.NewCategoryService(categoryRepo)
	brandService := service.NewBrandService(brandRepo)
	reviewService := service.NewReviewService(reviewRepo, productRepo)

	// Initialize handlers
	productHandler := httpHandler.NewProductHandler(productService)
	categoryHandler := httpHandler.NewCategoryHandler(categoryService)
	brandHandler := httpHandler.NewBrandHandler(brandService)
	reviewHandler := httpHandler.NewReviewHandler(reviewService)

	// Setup routes
	router := httpHandler.SetupRoutes(productHandler, categoryHandler, brandHandler, reviewHandler, jwtManager)

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Product service starting on %s", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}