package http

import (
	"github.com/gin-gonic/gin"
	"solemate/pkg/auth"
)

func SetupRoutes(userHandler *UserHandler, wishlistHandler *WishlistHandler, jwtManager *auth.JWTManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "user-service"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(AuthMiddleware(jwtManager))
		{
			// User profile routes
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)

			// Wishlist routes
			wishlist := protected.Group("/wishlist")
			{
				wishlist.GET("", wishlistHandler.GetWishlist)
				wishlist.POST("/items", wishlistHandler.AddItem)
				wishlist.DELETE("/items/:product_id", wishlistHandler.RemoveItem)
				wishlist.DELETE("", wishlistHandler.ClearWishlist)
				wishlist.POST("/move-to-cart", wishlistHandler.MoveToCart) // Not implemented, frontend handles
			}

			// Admin only routes
			admin := protected.Group("/")
			admin.Use(AdminMiddleware())
			{
				admin.GET("/users", userHandler.ListUsers)
				admin.GET("/users/:id", userHandler.GetUser)
				admin.DELETE("/users/:id", userHandler.DeleteUser)
			}
		}
	}

	return r
}
