package handler

import (
	"github.com/gin-gonic/gin"
	"solemate/api-gateway/internal/middleware"
	"solemate/pkg/auth"
)

func SetupRoutes(proxyHandler *ProxyHandler, jwtManager *auth.JWTManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", proxyHandler.ProxyToUserService)
			auth.POST("/login", proxyHandler.ProxyToUserService)
			auth.POST("/refresh", proxyHandler.ProxyToUserService)
		}

		// Public product routes (no auth required)
		products := v1.Group("/products")
		{
			products.GET("", proxyHandler.ProxyToProductService)
			products.GET("/search", proxyHandler.ProxyToProductService)
			products.GET("/:id", proxyHandler.ProxyToProductService)
			products.GET("/:id/related", proxyHandler.ProxyToProductService)
		}

		// Public categories routes
		categories := v1.Group("/categories")
		{
			categories.GET("", proxyHandler.ProxyToProductService)
			categories.GET("/:id", proxyHandler.ProxyToProductService)
		}

		// Public brands routes
		brands := v1.Group("/brands")
		{
			brands.GET("", proxyHandler.ProxyToProductService)
			brands.GET("/:id", proxyHandler.ProxyToProductService)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// User profile routes
			protected.GET("/profile", proxyHandler.ProxyToUserService)
			protected.PUT("/profile", proxyHandler.ProxyToUserService)

			// Cart routes
			cart := protected.Group("/cart")
			{
				cart.GET("", proxyHandler.ProxyToCartService)
				cart.POST("/items", proxyHandler.ProxyToCartService)
				cart.PATCH("/items/:item_id/quantity", proxyHandler.ProxyToCartService)
				cart.DELETE("/items/:item_id", proxyHandler.ProxyToCartService)
				cart.DELETE("", proxyHandler.ProxyToCartService) // Clear cart
				cart.GET("/summary", proxyHandler.ProxyToCartService)
				cart.GET("/count", proxyHandler.ProxyToCartService)
				cart.POST("/extend", proxyHandler.ProxyToCartService)
				cart.POST("/items/:item_id/discount", proxyHandler.ProxyToCartService)
			}

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.POST("", proxyHandler.ProxyToOrderService)
				orders.GET("/me", proxyHandler.ProxyToOrderService)
				orders.GET("/me/summaries", proxyHandler.ProxyToOrderService)
				orders.GET("/:order_id", proxyHandler.ProxyToOrderService)
				orders.GET("/number/:order_number", proxyHandler.ProxyToOrderService)
				orders.PATCH("/:order_id/shipping-address", proxyHandler.ProxyToOrderService)
				orders.PATCH("/:order_id/billing-address", proxyHandler.ProxyToOrderService)

				// Admin order routes
				adminOrders := orders.Group("/admin")
				adminOrders.Use(middleware.AdminMiddleware())
				{
					adminOrders.POST("/search", proxyHandler.ProxyToOrderService)
					adminOrders.GET("/statistics", proxyHandler.ProxyToOrderService)
					adminOrders.GET("/top-products", proxyHandler.ProxyToOrderService)
					adminOrders.GET("/sales-metrics", proxyHandler.ProxyToOrderService)
					adminOrders.PATCH("/:order_id/status", proxyHandler.ProxyToOrderService)
					adminOrders.POST("/:order_id/ship", proxyHandler.ProxyToOrderService)
					adminOrders.PATCH("/:order_id/payment-status", proxyHandler.ProxyToOrderService)
				}
			}

			// Payment routes
			payments := protected.Group("/payments")
			{
				payments.POST("", proxyHandler.ProxyToPaymentService)
				payments.GET("/:id", proxyHandler.ProxyToPaymentService)
				payments.POST("/:id/refund", proxyHandler.ProxyToPaymentService)
			}

			// Wishlist routes
			wishlist := protected.Group("/wishlist")
			{
				wishlist.GET("", proxyHandler.ProxyToUserService)
				wishlist.POST("/items", proxyHandler.ProxyToUserService)
				wishlist.DELETE("/items/:id", proxyHandler.ProxyToUserService)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				// User management
				adminUsers := admin.Group("/users")
				{
					adminUsers.GET("", proxyHandler.ProxyToUserService)
					adminUsers.GET("/:id", proxyHandler.ProxyToUserService)
					adminUsers.DELETE("/:id", proxyHandler.ProxyToUserService)
				}

				// Product management
				adminProducts := admin.Group("/products")
				{
					adminProducts.POST("", proxyHandler.ProxyToProductService)
					adminProducts.PUT("/:id", proxyHandler.ProxyToProductService)
					adminProducts.DELETE("/:id", proxyHandler.ProxyToProductService)
				}

				// Order management
				adminOrders := admin.Group("/orders")
				{
					adminOrders.GET("", proxyHandler.ProxyToOrderService)
					adminOrders.PUT("/:id/status", proxyHandler.ProxyToOrderService)
				}

				// Analytics routes
				analytics := admin.Group("/analytics")
				{
					analytics.GET("/dashboard", proxyHandler.ProxyToOrderService)
					analytics.GET("/sales", proxyHandler.ProxyToOrderService)
					analytics.GET("/users", proxyHandler.ProxyToUserService)
					analytics.GET("/products", proxyHandler.ProxyToProductService)
				}
			}
		}
	}

	return r
}
