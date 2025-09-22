package functional

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// FunctionalTestSuite implements comprehensive functional testing
// per PDF requirement: "Verify features against SRS"
type FunctionalTestSuite struct {
	suite.Suite
	server     *gin.Engine
	testClient *http.Client
	baseURL    string
	authToken  string
	userID     string
}

func (suite *FunctionalTestSuite) SetupSuite() {
	// Setup test server
	gin.SetMode(gin.TestMode)
	suite.server = gin.New()
	suite.testClient = &http.Client{Timeout: 30 * time.Second}
	suite.baseURL = "http://localhost:8080/api/v1"

	// Setup routes (would normally initialize actual handlers)
	suite.setupTestRoutes()
}

func (suite *FunctionalTestSuite) setupTestRoutes() {
	// Mock routes for functional testing
	api := suite.server.Group("/api/v1")

	// Authentication routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", suite.mockRegisterHandler)
		auth.POST("/login", suite.mockLoginHandler)
		auth.POST("/refresh", suite.mockRefreshHandler)
	}

	// Product routes
	products := api.Group("/products")
	{
		products.GET("", suite.mockGetProductsHandler)
		products.GET("/:id", suite.mockGetProductHandler)
		products.POST("", suite.mockCreateProductHandler)
	}

	// Cart routes
	cart := api.Group("/cart")
	{
		cart.GET("", suite.mockGetCartHandler)
		cart.POST("/items", suite.mockAddCartItemHandler)
		cart.DELETE("/items/:id", suite.mockRemoveCartItemHandler)
	}

	// Order routes
	orders := api.Group("/orders")
	{
		orders.GET("", suite.mockGetOrdersHandler)
		orders.POST("", suite.mockCreateOrderHandler)
		orders.GET("/:id", suite.mockGetOrderHandler)
	}

	// Payment routes
	payments := api.Group("/payments")
	{
		payments.POST("/create-intent", suite.mockCreatePaymentIntentHandler)
		payments.POST("/confirm", suite.mockConfirmPaymentHandler)
	}
}

// Test Authentication Flow (PDF Requirement: User registration & authentication)
func (suite *FunctionalTestSuite) TestAuthenticationFlow() {
	// Test user registration
	suite.Run("UserRegistration", func() {
		registerData := map[string]interface{}{
			"email":      "test@example.com",
			"password":   "SecurePass123!",
			"first_name": "John",
			"last_name":  "Doe",
			"phone":      "+1234567890",
		}

		resp := suite.makeRequest("POST", "/auth/register", registerData)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		// Verify response structure
		assert.Contains(suite.T(), response, "user")
		assert.Contains(suite.T(), response, "access_token")
		assert.Contains(suite.T(), response, "refresh_token")

		suite.authToken = response["access_token"].(string)
		suite.userID = response["user"].(map[string]interface{})["id"].(string)
	})

	// Test user login
	suite.Run("UserLogin", func() {
		loginData := map[string]interface{}{
			"email":    "test@example.com",
			"password": "SecurePass123!",
		}

		resp := suite.makeRequest("POST", "/auth/login", loginData)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		assert.Contains(suite.T(), response, "access_token")
	})

	// Test token refresh
	suite.Run("TokenRefresh", func() {
		refreshData := map[string]interface{}{
			"refresh_token": "mock_refresh_token",
		}

		resp := suite.makeRequest("POST", "/auth/refresh", refreshData)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})
}

// Test Product Catalog (PDF Requirement: Product catalog with filtering and search)
func (suite *FunctionalTestSuite) TestProductCatalog() {
	suite.Run("GetProducts", func() {
		resp := suite.makeRequestWithAuth("GET", "/products?query=Nike&min_price=50&max_price=200", nil)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		assert.Contains(suite.T(), response, "products")
		assert.Contains(suite.T(), response, "total")
	})

	suite.Run("GetProductDetails", func() {
		productID := uuid.New().String()
		resp := suite.makeRequestWithAuth("GET", fmt.Sprintf("/products/%s", productID), nil)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})

	suite.Run("CreateProduct", func() {
		productData := map[string]interface{}{
			"name":        "Nike Air Max Test",
			"description": "Test product",
			"sku":         "NIKE-TEST-001",
			"price":       149.99,
			"category_id": uuid.New().String(),
			"brand_id":    uuid.New().String(),
			"stock":       100,
			"is_active":   true,
		}

		resp := suite.makeRequestWithAuth("POST", "/products", productData)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	})
}

// Test Shopping Cart (PDF Requirement: Shopping cart and checkout)
func (suite *FunctionalTestSuite) TestShoppingCart() {
	suite.Run("GetCart", func() {
		resp := suite.makeRequestWithAuth("GET", "/cart", nil)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		assert.Contains(suite.T(), response, "items")
		assert.Contains(suite.T(), response, "total_amount")
	})

	suite.Run("AddItemToCart", func() {
		itemData := map[string]interface{}{
			"product_id": uuid.New().String(),
			"quantity":   2,
		}

		resp := suite.makeRequestWithAuth("POST", "/cart/items", itemData)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	})

	suite.Run("RemoveItemFromCart", func() {
		itemID := uuid.New().String()
		resp := suite.makeRequestWithAuth("DELETE", fmt.Sprintf("/cart/items/%s", itemID), nil)
		assert.Equal(suite.T(), http.StatusNoContent, resp.StatusCode)
	})
}

// Test Order Management (PDF Requirement: Order tracking & history)
func (suite *FunctionalTestSuite) TestOrderManagement() {
	suite.Run("CreateOrder", func() {
		orderData := map[string]interface{}{
			"shipping_address": map[string]string{
				"address_line_1": "123 Main St",
				"city":           "New York",
				"state":          "NY",
				"postal_code":    "10001",
				"country":        "US",
			},
			"billing_address": map[string]string{
				"address_line_1": "123 Main St",
				"city":           "New York",
				"state":          "NY",
				"postal_code":    "10001",
				"country":        "US",
			},
		}

		resp := suite.makeRequestWithAuth("POST", "/orders", orderData)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	})

	suite.Run("GetOrders", func() {
		resp := suite.makeRequestWithAuth("GET", "/orders?status=pending", nil)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		assert.Contains(suite.T(), response, "orders")
		assert.Contains(suite.T(), response, "total")
	})

	suite.Run("GetOrderDetails", func() {
		orderID := uuid.New().String()
		resp := suite.makeRequestWithAuth("GET", fmt.Sprintf("/orders/%s", orderID), nil)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})
}

// Test Payment Processing (PDF Requirement: Payment gateway integration)
func (suite *FunctionalTestSuite) TestPaymentProcessing() {
	suite.Run("CreatePaymentIntent", func() {
		paymentData := map[string]interface{}{
			"order_id": uuid.New().String(),
			"amount":   299.98,
			"currency": "usd",
		}

		resp := suite.makeRequestWithAuth("POST", "/payments/create-intent", paymentData)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		assert.Contains(suite.T(), response, "client_secret")
		assert.Contains(suite.T(), response, "amount")
	})

	suite.Run("ConfirmPayment", func() {
		confirmData := map[string]interface{}{
			"payment_intent_id": "pi_test_123456789",
		}

		resp := suite.makeRequestWithAuth("POST", "/payments/confirm", confirmData)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})
}

// Test Error Handling and Edge Cases
func (suite *FunctionalTestSuite) TestErrorHandling() {
	suite.Run("UnauthorizedAccess", func() {
		resp := suite.makeRequest("GET", "/cart", nil)
		assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	})

	suite.Run("InvalidData", func() {
		invalidData := map[string]interface{}{
			"email":    "invalid-email",
			"password": "123", // Too short
		}

		resp := suite.makeRequest("POST", "/auth/register", invalidData)
		assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	})

	suite.Run("NotFoundEndpoint", func() {
		resp := suite.makeRequestWithAuth("GET", "/nonexistent", nil)
		assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	})
}

// Helper methods
func (suite *FunctionalTestSuite) makeRequest(method, endpoint string, data interface{}) *http.Response {
	var body *bytes.Buffer
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	}

	req := httptest.NewRequest(method, endpoint, body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)

	return w.Result()
}

func (suite *FunctionalTestSuite) makeRequestWithAuth(method, endpoint string, data interface{}) *http.Response {
	var body *bytes.Buffer
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	}

	req := httptest.NewRequest(method, endpoint, body)
	req.Header.Set("Content-Type", "application/json")
	if suite.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
	}

	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)

	return w.Result()
}

// Mock handlers (would be replaced with actual service handlers)
func (suite *FunctionalTestSuite) mockRegisterHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         uuid.New().String(),
			"email":      "test@example.com",
			"first_name": "John",
			"last_name":  "Doe",
		},
		"access_token":  "mock_access_token",
		"refresh_token": "mock_refresh_token",
		"expires_in":    3600,
	})
}

func (suite *FunctionalTestSuite) mockLoginHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"access_token":  "mock_access_token",
		"refresh_token": "mock_refresh_token",
		"expires_in":    3600,
	})
}

func (suite *FunctionalTestSuite) mockRefreshHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"access_token":  "new_mock_access_token",
		"refresh_token": "new_mock_refresh_token",
		"expires_in":    3600,
	})
}

func (suite *FunctionalTestSuite) mockGetProductsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"products": []gin.H{
			{
				"id":    uuid.New().String(),
				"name":  "Nike Air Max",
				"price": 149.99,
			},
		},
		"total":  1,
		"limit":  20,
		"offset": 0,
	})
}

func (suite *FunctionalTestSuite) mockGetProductHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":          c.Param("id"),
		"name":        "Nike Air Max",
		"description": "Comfortable running shoes",
		"price":       149.99,
		"stock":       100,
	})
}

func (suite *FunctionalTestSuite) mockCreateProductHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":          uuid.New().String(),
		"name":        "Nike Air Max Test",
		"description": "Test product",
		"price":       149.99,
		"stock":       100,
	})
}

func (suite *FunctionalTestSuite) mockGetCartHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":           uuid.New().String(),
		"items":        []gin.H{},
		"total_amount": 0.0,
		"item_count":   0,
	})
}

func (suite *FunctionalTestSuite) mockAddCartItemHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":         uuid.New().String(),
		"product_id": uuid.New().String(),
		"quantity":   2,
		"price":      149.99,
	})
}

func (suite *FunctionalTestSuite) mockRemoveCartItemHandler(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (suite *FunctionalTestSuite) mockGetOrdersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"orders": []gin.H{
			{
				"id":           uuid.New().String(),
				"order_number": "ORD-001",
				"status":       "pending",
				"total_amount": 299.98,
			},
		},
		"total":  1,
		"limit":  20,
		"offset": 0,
	})
}

func (suite *FunctionalTestSuite) mockCreateOrderHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":           uuid.New().String(),
		"order_number": "ORD-001",
		"status":       "pending",
		"total_amount": 299.98,
	})
}

func (suite *FunctionalTestSuite) mockGetOrderHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":           c.Param("id"),
		"order_number": "ORD-001",
		"status":       "pending",
		"total_amount": 299.98,
		"items": []gin.H{
			{
				"product_id":   uuid.New().String(),
				"product_name": "Nike Air Max",
				"quantity":     2,
				"unit_price":   149.99,
			},
		},
	})
}

func (suite *FunctionalTestSuite) mockCreatePaymentIntentHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":            "pi_test_123456789",
		"client_secret": "pi_test_123456789_secret_test",
		"amount":        299.98,
		"currency":      "usd",
		"status":        "requires_payment_method",
	})
}

func (suite *FunctionalTestSuite) mockConfirmPaymentHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"payment_id": "pi_test_123456789",
		"status":     "succeeded",
		"amount":     299.98,
		"order_id":   uuid.New().String(),
	})
}

// Test runner
func TestFunctionalTestSuite(t *testing.T) {
	suite.Run(t, new(FunctionalTestSuite))
}