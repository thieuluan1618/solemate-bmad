package http

import (
	"github.com/gin-gonic/gin"
	"solemate/pkg/auth"
)

func SetupRoutes(productHandler *ProductHandler, categoryHandler *CategoryHandler, brandHandler *BrandHandler, reviewHandler *ReviewHandler, jwtManager *auth.JWTManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "product-service"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public product routes (no authentication required)
		products := v1.Group("/products")
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/slug/:slug", productHandler.GetProductBySlug)
			products.GET("/:id/related", productHandler.GetRelatedProducts)

			// Public review routes (no authentication for GET)
			products.GET("/:id/reviews", reviewHandler.GetReviewsByProductID)
		}

		// Public category routes
		categories := v1.Group("/categories")
		{
			categories.GET("", categoryHandler.ListCategories)
			categories.GET("/tree", categoryHandler.GetCategoryTree)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.GET("/slug/:slug", categoryHandler.GetCategoryBySlug)
		}

		// Public brand routes
		brands := v1.Group("/brands")
		{
			brands.GET("", brandHandler.ListBrands)
			brands.GET("/:id", brandHandler.GetBrand)
			brands.GET("/slug/:slug", brandHandler.GetBrandBySlug)
		}

		// Public review routes (no authentication for GET)
		reviews := v1.Group("/reviews")
		{
			reviews.GET("/:id", reviewHandler.GetReview)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(AuthMiddleware(jwtManager))
		{
			// Review routes (authentication required for POST, PUT, DELETE)
			protectedReviews := protected.Group("/products/:id/reviews")
			{
				protectedReviews.POST("", reviewHandler.CreateReview)
			}

			reviewsAuth := protected.Group("/reviews")
			{
				reviewsAuth.PUT("/:id", reviewHandler.UpdateReview)
				reviewsAuth.DELETE("/:id", reviewHandler.DeleteReview)
			}

			// Admin only routes
			admin := protected.Group("/admin")
			admin.Use(AdminMiddleware())
			{
				// Product management
				adminProducts := admin.Group("/products")
				{
					adminProducts.POST("", productHandler.CreateProduct)
					adminProducts.PUT("/:id", productHandler.UpdateProduct)
					adminProducts.DELETE("/:id", productHandler.DeleteProduct)
				}

				// Category management
				adminCategories := admin.Group("/categories")
				{
					adminCategories.POST("", categoryHandler.CreateCategory)
					adminCategories.PUT("/:id", categoryHandler.UpdateCategory)
					adminCategories.DELETE("/:id", categoryHandler.DeleteCategory)
				}

				// Brand management
				adminBrands := admin.Group("/brands")
				{
					adminBrands.POST("", brandHandler.CreateBrand)
					adminBrands.PUT("/:id", brandHandler.UpdateBrand)
					adminBrands.DELETE("/:id", brandHandler.DeleteBrand)
				}
			}
		}
	}

	return r
}

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from headers set by API Gateway
		userID := c.GetHeader("X-User-ID")
		email := c.GetHeader("X-User-Email")
		role := c.GetHeader("X-User-Role")

		if userID == "" {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("role", role)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" && role != "manager" {
			c.JSON(403, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}