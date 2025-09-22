package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SecurityTestSuite implements security testing
// per PDF requirement: "SQL injection, XSS, CSRF checks"
type SecurityTestSuite struct {
	suite.Suite
	server     *gin.Engine
	baseURL    string
	authToken  string
}

func (suite *SecurityTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.server = gin.New()
	suite.baseURL = "/api/v1"
	suite.authToken = "test_security_token_" + uuid.New().String()

	suite.setupSecurityTestRoutes()
}

func (suite *SecurityTestSuite) setupSecurityTestRoutes() {
	api := suite.server.Group("/api/v1")

	// Authentication routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", suite.mockSecureRegisterHandler)
		auth.POST("/login", suite.mockSecureLoginHandler)
	}

	// User routes
	users := api.Group("/users")
	{
		users.GET("/profile", suite.mockGetUserProfileHandler)
		users.PUT("/profile", suite.mockUpdateUserProfileHandler)
	}

	// Product routes
	products := api.Group("/products")
	{
		products.GET("", suite.mockSearchProductsHandler)
		products.GET("/:id", suite.mockGetProductHandler)
		products.POST("", suite.mockCreateProductHandler)
		products.PUT("/:id", suite.mockUpdateProductHandler)
	}

	// Order routes
	orders := api.Group("/orders")
	{
		orders.GET("", suite.mockGetOrdersHandler)
		orders.POST("", suite.mockCreateOrderHandler)
		orders.GET("/:id", suite.mockGetOrderHandler)
	}

	// Admin routes
	admin := api.Group("/admin")
	{
		admin.GET("/users", suite.mockAdminGetUsersHandler)
		admin.DELETE("/users/:id", suite.mockAdminDeleteUserHandler)
		admin.PUT("/orders/:id/status", suite.mockAdminUpdateOrderStatusHandler)
	}
}

// Test SQL Injection vulnerabilities (PDF requirement)
func (suite *SecurityTestSuite) TestSQLInjectionPrevention() {
	sqlInjectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users; --",
		"' UNION SELECT * FROM users --",
		"1' OR '1'='1' -- ",
		"admin'/*",
		"' OR 1=1#",
		"' OR 'x'='x",
		"'; EXEC xp_cmdshell('dir'); --",
		"1'; UPDATE users SET password='hacked' WHERE '1'='1",
		"' OR (SELECT COUNT(*) FROM users) > 0 --",
	}

	suite.Run("SQLInjection_Login_Protection", func() {
		for i, payload := range sqlInjectionPayloads {
			loginData := map[string]interface{}{
				"email":    payload,
				"password": payload,
			}

			resp := suite.makeRequest("POST", "/auth/login", loginData, false)

			// Should not return success for SQL injection attempts
			assert.NotEqual(suite.T(), http.StatusOK, resp.StatusCode,
				"SQL injection payload %d should not succeed: %s", i+1, payload)

			// Should return proper error response
			assert.Contains(suite.T(), []int{http.StatusBadRequest, http.StatusUnauthorized}, resp.StatusCode,
				"Should return 400 or 401 for SQL injection attempt %d", i+1)
		}
	})

	suite.Run("SQLInjection_Search_Protection", func() {
		for i, payload := range sqlInjectionPayloads {
			searchQuery := fmt.Sprintf("/products?query=%s", payload)
			resp := suite.makeRequest("GET", searchQuery, nil, false)

			// Should not expose database errors or return unexpected data
			assert.Equal(suite.T(), http.StatusOK, resp.StatusCode,
				"Search with SQL injection payload %d should return 200 but sanitized results", i+1)

			var response map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(suite.T(), err)

			// Should return empty or sanitized results, not database errors
			if products, exists := response["products"]; exists {
				assert.IsType(suite.T(), []interface{}{}, products,
					"Products should be array even with SQL injection attempt %d", i+1)
			}
		}
	})

	suite.Run("SQLInjection_UserProfile_Protection", func() {
		for i, payload := range sqlInjectionPayloads {
			profileData := map[string]interface{}{
				"first_name": payload,
				"last_name":  payload,
				"phone":      payload,
			}

			resp := suite.makeRequest("PUT", "/users/profile", profileData, true)

			// Should validate and sanitize input
			assert.Contains(suite.T(), []int{http.StatusOK, http.StatusBadRequest}, resp.StatusCode,
				"Profile update with SQL injection payload %d should be handled securely", i+1)
		}
	})
}

// Test XSS (Cross-Site Scripting) prevention
func (suite *SecurityTestSuite) TestXSSPrevention() {
	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
		"<iframe src=javascript:alert('XSS')>",
		"<body onload=alert('XSS')>",
		"<div onclick=alert('XSS')>Click me</div>",
		"<input onfocus=alert('XSS') autofocus>",
		"<select onfocus=alert('XSS') autofocus><option>test</option></select>",
		"<textarea onfocus=alert('XSS') autofocus>test</textarea>",
		"<details open ontoggle=alert('XSS')>",
		"<marquee onstart=alert('XSS')>test</marquee>",
	}

	suite.Run("XSS_UserInput_Sanitization", func() {
		for i, payload := range xssPayloads {
			userData := map[string]interface{}{
				"first_name": payload,
				"last_name":  "Test User",
				"phone":      "+1234567890",
			}

			resp := suite.makeRequest("PUT", "/users/profile", userData, true)

			var response map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&response)

			if resp.StatusCode == http.StatusOK && err == nil {
				// If update succeeded, verify data was sanitized
				if user, exists := response["user"]; exists {
					userMap := user.(map[string]interface{})
					firstName := userMap["first_name"].(string)

					// Should not contain script tags or event handlers
					assert.NotContains(suite.T(), firstName, "<script>",
						"XSS payload %d should be sanitized", i+1)
					assert.NotContains(suite.T(), firstName, "javascript:",
						"XSS payload %d should be sanitized", i+1)
					assert.NotContains(suite.T(), firstName, "onerror",
						"XSS payload %d should be sanitized", i+1)
					assert.NotContains(suite.T(), firstName, "onload",
						"XSS payload %d should be sanitized", i+1)
				}
			}
		}
	})

	suite.Run("XSS_ProductReview_Protection", func() {
		for i, payload := range xssPayloads {
			reviewData := map[string]interface{}{
				"product_id": uuid.New().String(),
				"rating":     5,
				"title":      payload,
				"comment":    "This product is " + payload,
			}

			resp := suite.makeRequest("POST", "/products/reviews", reviewData, true)

			if resp.StatusCode == http.StatusCreated {
				var response map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(suite.T(), err)

				// Verify review content is sanitized
				if review, exists := response["review"]; exists {
					reviewMap := review.(map[string]interface{})
					title := reviewMap["title"].(string)
					comment := reviewMap["comment"].(string)

					assert.NotContains(suite.T(), title, "<script>",
						"Review title XSS payload %d should be sanitized", i+1)
					assert.NotContains(suite.T(), comment, "<script>",
						"Review comment XSS payload %d should be sanitized", i+1)
				}
			}
		}
	})
}

// Test CSRF (Cross-Site Request Forgery) protection
func (suite *SecurityTestSuite) TestCSRFProtection() {
	suite.Run("CSRF_Token_Required", func() {
		// Test that state-changing operations require CSRF protection
		stateChangingEndpoints := []struct {
			method string
			path   string
			data   map[string]interface{}
		}{
			{
				method: "PUT",
				path:   "/users/profile",
				data: map[string]interface{}{
					"first_name": "Updated Name",
				},
			},
			{
				method: "POST",
				path:   "/orders",
				data: map[string]interface{}{
					"shipping_address": map[string]interface{}{
						"address_line_1": "123 Test St",
						"city":           "Test City",
						"state":          "TS",
						"postal_code":    "12345",
						"country":        "US",
					},
				},
			},
			{
				method: "DELETE",
				path:   "/admin/users/" + uuid.New().String(),
				data:   nil,
			},
		}

		for _, endpoint := range stateChangingEndpoints {
			// Request without CSRF token should fail
			resp := suite.makeRequestWithoutCSRF(endpoint.method, endpoint.path, endpoint.data, true)
			assert.Equal(suite.T(), http.StatusForbidden, resp.StatusCode,
				"Request to %s %s without CSRF token should fail", endpoint.method, endpoint.path)

			// Request with invalid CSRF token should fail
			resp = suite.makeRequestWithInvalidCSRF(endpoint.method, endpoint.path, endpoint.data, true)
			assert.Equal(suite.T(), http.StatusForbidden, resp.StatusCode,
				"Request to %s %s with invalid CSRF token should fail", endpoint.method, endpoint.path)

			// Request with valid CSRF token should succeed (or return proper business logic error)
			resp = suite.makeRequest(endpoint.method, endpoint.path, endpoint.data, true)
			assert.NotEqual(suite.T(), http.StatusForbidden, resp.StatusCode,
				"Request to %s %s with valid CSRF token should not fail due to CSRF", endpoint.method, endpoint.path)
		}
	})

	suite.Run("CSRF_Token_Generation", func() {
		// Test CSRF token endpoint
		resp := suite.makeRequest("GET", "/csrf-token", nil, false)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(suite.T(), err)

		token, exists := response["csrf_token"]
		assert.True(suite.T(), exists, "CSRF token should be provided")
		assert.NotEmpty(suite.T(), token, "CSRF token should not be empty")
		assert.IsType(suite.T(), "", token, "CSRF token should be string")

		// Token should be sufficiently long for security
		tokenStr := token.(string)
		assert.GreaterOrEqual(suite.T(), len(tokenStr), 32, "CSRF token should be at least 32 characters")
	})
}

// Test Authentication and Authorization security
func (suite *SecurityTestSuite) TestAuthenticationSecurity() {
	suite.Run("JWT_Token_Security", func() {
		// Test with invalid JWT tokens
		invalidTokens := []string{
			"invalid.token.here",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			"Bearer invalid_token",
			"expired.jwt.token.here",
			"malformed_token",
			"",
		}

		for i, invalidToken := range invalidTokens {
			resp := suite.makeRequestWithToken("GET", "/users/profile", nil, invalidToken)
			assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode,
				"Invalid token %d should return 401", i+1)
		}
	})

	suite.Run("Authorization_Levels", func() {
		// Test that admin endpoints require admin privileges
		adminEndpoints := []struct {
			method string
			path   string
			data   map[string]interface{}
		}{
			{"GET", "/admin/users", nil},
			{"DELETE", "/admin/users/" + uuid.New().String(), nil},
			{"PUT", "/admin/orders/" + uuid.New().String() + "/status", map[string]interface{}{"status": "shipped"}},
		}

		for _, endpoint := range adminEndpoints {
			// Regular user token should be denied
			userToken := "user_token_" + uuid.New().String()
			resp := suite.makeRequestWithToken(endpoint.method, endpoint.path, endpoint.data, userToken)
			assert.Equal(suite.T(), http.StatusForbidden, resp.StatusCode,
				"Regular user should not access admin endpoint: %s %s", endpoint.method, endpoint.path)

			// Admin token should be allowed
			adminToken := "admin_token_" + uuid.New().String()
			resp = suite.makeRequestWithToken(endpoint.method, endpoint.path, endpoint.data, adminToken)
			assert.NotEqual(suite.T(), http.StatusForbidden, resp.StatusCode,
				"Admin user should access admin endpoint: %s %s", endpoint.method, endpoint.path)
		}
	})

	suite.Run("Session_Security", func() {
		// Test session fixation prevention
		loginData := map[string]interface{}{
			"email":    "test@example.com",
			"password": "SecurePass123!",
		}

		resp1 := suite.makeRequest("POST", "/auth/login", loginData, false)
		resp2 := suite.makeRequest("POST", "/auth/login", loginData, false)

		if resp1.StatusCode == http.StatusOK && resp2.StatusCode == http.StatusOK {
			var response1, response2 map[string]interface{}
			json.NewDecoder(resp1.Body).Decode(&response1)
			json.NewDecoder(resp2.Body).Decode(&response2)

			token1 := response1["access_token"].(string)
			token2 := response2["access_token"].(string)

			// Each login should generate a new token (session fixation prevention)
			assert.NotEqual(suite.T(), token1, token2, "Each login should generate a unique token")
		}
	})
}

// Test Input Validation and Sanitization
func (suite *SecurityTestSuite) TestInputValidation() {
	suite.Run("Email_Validation", func() {
		invalidEmails := []string{
			"invalid-email",
			"@invalid.com",
			"invalid@",
			"<script>alert('xss')</script>@test.com",
			"test..test@example.com",
			"test@",
			"",
			"very-long-email-that-exceeds-normal-length-limits@very-long-domain-name-that-should-not-be-accepted.com",
		}

		for i, invalidEmail := range invalidEmails {
			registerData := map[string]interface{}{
				"email":      invalidEmail,
				"password":   "SecurePass123!",
				"first_name": "Test",
				"last_name":  "User",
			}

			resp := suite.makeRequest("POST", "/auth/register", registerData, false)
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"Invalid email %d should be rejected: %s", i+1, invalidEmail)
		}
	})

	suite.Run("Password_Strength_Validation", func() {
		weakPasswords := []string{
			"123456",
			"password",
			"12345678",
			"qwerty",
			"abc123",
			"",
			"pass",
			"PASSWORD123", // Missing special char
			"password123", // Missing uppercase
			"PASSWORD!",   // Missing number
		}

		for i, weakPassword := range weakPasswords {
			registerData := map[string]interface{}{
				"email":      fmt.Sprintf("test%d@example.com", i),
				"password":   weakPassword,
				"first_name": "Test",
				"last_name":  "User",
			}

			resp := suite.makeRequest("POST", "/auth/register", registerData, false)
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"Weak password %d should be rejected: %s", i+1, weakPassword)
		}
	})

	suite.Run("Numeric_Input_Validation", func() {
		// Test numeric field validation
		invalidPrices := []interface{}{
			-10.50,    // Negative price
			"not_a_number", // String instead of number
			999999.99, // Extremely high price
			0.001,     // Too many decimal places
		}

		for i, invalidPrice := range invalidPrices {
			productData := map[string]interface{}{
				"name":        "Test Product",
				"description": "Test description",
				"sku":         fmt.Sprintf("TEST-%d", i),
				"price":       invalidPrice,
				"category_id": uuid.New().String(),
				"brand_id":    uuid.New().String(),
				"stock":       10,
			}

			resp := suite.makeRequest("POST", "/products", productData, true)
			assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode,
				"Invalid price %d should be rejected: %v", i+1, invalidPrice)
		}
	})
}

// Test Rate Limiting and DDoS Protection
func (suite *SecurityTestSuite) TestRateLimiting() {
	suite.Run("Login_Rate_Limiting", func() {
		// Attempt multiple login requests rapidly
		loginData := map[string]interface{}{
			"email":    "attacker@example.com",
			"password": "wrong_password",
		}

		successfulRequests := 0
		rateLimitedRequests := 0

		// Make 20 rapid requests
		for i := 0; i < 20; i++ {
			resp := suite.makeRequest("POST", "/auth/login", loginData, false)

			if resp.StatusCode == http.StatusUnauthorized {
				successfulRequests++ // Request was processed (even if auth failed)
			} else if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitedRequests++
			}
		}

		// Should have rate limiting after several failed attempts
		assert.Greater(suite.T(), rateLimitedRequests, 0,
			"Should have rate limiting after multiple failed login attempts")
		assert.Less(suite.T(), successfulRequests, 20,
			"Not all requests should be processed when rate limited")
	})

	suite.Run("API_Rate_Limiting", func() {
		// Test general API rate limiting
		rapidRequestsBlocked := 0

		for i := 0; i < 100; i++ {
			resp := suite.makeRequest("GET", "/products", nil, false)

			if resp.StatusCode == http.StatusTooManyRequests {
				rapidRequestsBlocked++
			}

			// Small delay to simulate rapid requests
			time.Sleep(10 * time.Millisecond)
		}

		// Should have some rate limiting for rapid requests
		assert.Greater(suite.T(), rapidRequestsBlocked, 0,
			"Should block some rapid API requests")
	})
}

// Helper methods for security testing

func (suite *SecurityTestSuite) makeRequest(method, path string, data map[string]interface{}, requireAuth bool) *http.Response {
	return suite.makeRequestWithToken(method, path, data, suite.authToken)
}

func (suite *SecurityTestSuite) makeRequestWithToken(method, path string, data map[string]interface{}, token string) *http.Response {
	var bodyReader *bytes.Buffer
	if data != nil {
		jsonData, _ := json.Marshal(data)
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req := httptest.NewRequest(method, suite.baseURL+path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "valid_csrf_token_"+uuid.New().String())

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)

	return w.Result()
}

func (suite *SecurityTestSuite) makeRequestWithoutCSRF(method, path string, data map[string]interface{}, requireAuth bool) *http.Response {
	var bodyReader *bytes.Buffer
	if data != nil {
		jsonData, _ := json.Marshal(data)
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req := httptest.NewRequest(method, suite.baseURL+path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	if requireAuth {
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
	}

	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)

	return w.Result()
}

func (suite *SecurityTestSuite) makeRequestWithInvalidCSRF(method, path string, data map[string]interface{}, requireAuth bool) *http.Response {
	var bodyReader *bytes.Buffer
	if data != nil {
		jsonData, _ := json.Marshal(data)
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req := httptest.NewRequest(method, suite.baseURL+path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", "invalid_csrf_token")

	if requireAuth {
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
	}

	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)

	return w.Result()
}

// Mock handlers with security implementations

func (suite *SecurityTestSuite) mockSecureRegisterHandler(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate email format
	email, ok := request["email"].(string)
	if !ok || !suite.isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Validate password strength
	password, ok := request["password"].(string)
	if !ok || !suite.isStrongPassword(password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password does not meet security requirements"})
		return
	}

	// Sanitize string inputs
	firstName := suite.sanitizeInput(request["first_name"].(string))
	lastName := suite.sanitizeInput(request["last_name"].(string))

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         uuid.New().String(),
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
		},
		"access_token":  "secure_token_" + uuid.New().String(),
		"refresh_token": "secure_refresh_" + uuid.New().String(),
	})
}

func (suite *SecurityTestSuite) mockSecureLoginHandler(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	email := request["email"].(string)
	password := request["password"].(string)

	// Check for SQL injection patterns
	if suite.containsSQLInjection(email) || suite.containsSQLInjection(password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input detected"})
		return
	}

	// Simulate secure authentication
	if email == "test@example.com" && password == "SecurePass123!" {
		c.JSON(http.StatusOK, gin.H{
			"access_token":  "secure_token_" + uuid.New().String(),
			"refresh_token": "secure_refresh_" + uuid.New().String(),
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func (suite *SecurityTestSuite) mockGetUserProfileHandler(c *gin.Context) {
	// Verify CSRF token for state-changing operations
	if c.Request.Method != "GET" && !suite.hasValidCSRFToken(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         uuid.New().String(),
			"email":      "test@example.com",
			"first_name": "Test",
			"last_name":  "User",
		},
	})
}

func (suite *SecurityTestSuite) mockUpdateUserProfileHandler(c *gin.Context) {
	if !suite.hasValidCSRFToken(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
		return
	}

	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Sanitize inputs
	sanitizedData := make(map[string]interface{})
	for key, value := range request {
		if str, ok := value.(string); ok {
			sanitizedData[key] = suite.sanitizeInput(str)
		} else {
			sanitizedData[key] = value
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         uuid.New().String(),
			"first_name": sanitizedData["first_name"],
			"last_name":  sanitizedData["last_name"],
			"phone":      sanitizedData["phone"],
		},
	})
}

func (suite *SecurityTestSuite) mockSearchProductsHandler(c *gin.Context) {
	query := c.Query("query")

	// Sanitize search query to prevent SQL injection
	sanitizedQuery := suite.sanitizeInput(query)

	// Return empty results for suspicious queries
	if suite.containsSQLInjection(query) {
		c.JSON(http.StatusOK, gin.H{
			"products": []gin.H{},
			"total":    0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": []gin.H{
			{
				"id":    uuid.New().String(),
				"name":  "Secure Product",
				"query": sanitizedQuery,
			},
		},
		"total": 1,
	})
}

func (suite *SecurityTestSuite) mockGetProductHandler(c *gin.Context) {
	productID := c.Param("id")

	// Validate UUID format to prevent injection
	if _, err := uuid.Parse(productID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    productID,
		"name":  "Secure Product",
		"price": 99.99,
	})
}

func (suite *SecurityTestSuite) mockCreateProductHandler(c *gin.Context) {
	if !suite.hasValidCSRFToken(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
		return
	}

	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate price
	if price, ok := request["price"].(float64); ok {
		if price < 0 || price > 10000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price range"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be a number"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          uuid.New().String(),
		"name":        suite.sanitizeInput(request["name"].(string)),
		"description": suite.sanitizeInput(request["description"].(string)),
		"price":       request["price"],
	})
}

func (suite *SecurityTestSuite) mockUpdateProductHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Product updated securely"})
}

func (suite *SecurityTestSuite) mockGetOrdersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"orders": []gin.H{},
		"total":  0,
	})
}

func (suite *SecurityTestSuite) mockCreateOrderHandler(c *gin.Context) {
	if !suite.hasValidCSRFToken(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           uuid.New().String(),
		"order_number": "ORD-SEC-001",
		"status":       "pending",
	})
}

func (suite *SecurityTestSuite) mockGetOrderHandler(c *gin.Context) {
	orderID := c.Param("id")

	if _, err := uuid.Parse(orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     orderID,
		"status": "pending",
	})
}

func (suite *SecurityTestSuite) mockAdminGetUsersHandler(c *gin.Context) {
	// Check admin authorization
	authHeader := c.GetHeader("Authorization")
	if !strings.Contains(authHeader, "admin_token") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": []gin.H{},
		"total": 0,
	})
}

func (suite *SecurityTestSuite) mockAdminDeleteUserHandler(c *gin.Context) {
	if !suite.hasValidCSRFToken(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if !strings.Contains(authHeader, "admin_token") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted securely"})
}

func (suite *SecurityTestSuite) mockAdminUpdateOrderStatusHandler(c *gin.Context) {
	if !suite.hasValidCSRFToken(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token required"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if !strings.Contains(authHeader, "admin_token") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated securely"})
}

// Security validation helper methods

func (suite *SecurityTestSuite) isValidEmail(email string) bool {
	// Basic email validation
	return strings.Contains(email, "@") &&
		   strings.Contains(email, ".") &&
		   !strings.Contains(email, "<") &&
		   !strings.Contains(email, ">") &&
		   len(email) > 5 && len(email) < 100
}

func (suite *SecurityTestSuite) isStrongPassword(password string) bool {
	// Password strength requirements
	if len(password) < 8 {
		return false
	}

	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")
	hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func (suite *SecurityTestSuite) containsSQLInjection(input string) bool {
	sqlPatterns := []string{
		"'",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"union",
		"select",
		"insert",
		"update",
		"delete",
		"drop",
		"exec",
		"execute",
		";",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

func (suite *SecurityTestSuite) sanitizeInput(input string) string {
	// Remove dangerous characters and tags
	sanitized := input
	sanitized = strings.ReplaceAll(sanitized, "<script>", "")
	sanitized = strings.ReplaceAll(sanitized, "</script>", "")
	sanitized = strings.ReplaceAll(sanitized, "<img", "")
	sanitized = strings.ReplaceAll(sanitized, "javascript:", "")
	sanitized = strings.ReplaceAll(sanitized, "onerror=", "")
	sanitized = strings.ReplaceAll(sanitized, "onload=", "")
	sanitized = strings.ReplaceAll(sanitized, "onclick=", "")
	sanitized = strings.ReplaceAll(sanitized, "onfocus=", "")
	sanitized = strings.ReplaceAll(sanitized, "<iframe", "")
	sanitized = strings.ReplaceAll(sanitized, "<svg", "")
	sanitized = strings.ReplaceAll(sanitized, "<body", "")

	return sanitized
}

func (suite *SecurityTestSuite) hasValidCSRFToken(c *gin.Context) bool {
	token := c.GetHeader("X-CSRF-Token")
	// Simple validation - in real implementation, would verify against stored token
	return token != "" &&
		   strings.HasPrefix(token, "valid_csrf_token_") &&
		   len(token) > 20
}

func TestSecurityTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityTestSuite))
}