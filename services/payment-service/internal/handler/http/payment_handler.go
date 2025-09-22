package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/services/payment-service/internal/domain/service"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup, jwtMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	payments := router.Group("/payments")
	payments.Use(jwtMiddleware)
	{
		// Payment operations
		payments.POST("", h.CreatePayment)
		payments.GET("/:id", h.GetPayment)
		payments.GET("", h.GetPayments)
		payments.GET("/order/:orderId", h.GetPaymentByOrderID)
		payments.POST("/:id/process", h.ProcessPayment)
		payments.POST("/:id/cancel", h.CancelPayment)

		// Payment method operations
		payments.POST("/methods", h.CreatePaymentMethod)
		payments.GET("/methods", h.GetPaymentMethods)
		payments.POST("/methods/:id/default", h.SetDefaultPaymentMethod)
		payments.DELETE("/methods/:id", h.DeletePaymentMethod)

		// Refund operations
		payments.POST("/:id/refunds", h.CreateRefund)
		payments.GET("/:id/refunds", h.GetRefundsByPaymentID)
		payments.GET("/refunds/:id", h.GetRefund)

		// Analytics (admin only)
		admin := payments.Group("/analytics")
		admin.Use(adminMiddleware)
		{
			admin.GET("/statistics", h.GetPaymentStatistics)
			admin.GET("/revenue", h.GetRevenueMetrics)
		}
	}

	// Webhook endpoint (no auth required)
	router.POST("/webhooks/stripe", h.HandleStripeWebhook)
}

// Payment operations
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var request service.CreatePaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	request.UserID = userID.(uuid.UUID)

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := h.paymentService.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Check if user owns this payment or is admin
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	if userRole != "admin" && payment.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) GetPayments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse pagination parameters
	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	offset := (page - 1) * perPage
	payments, total, err := h.paymentService.GetPaymentsByUserID(c.Request.Context(), userID.(uuid.UUID), perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := service.PaginatedPaymentResponse{
		Payments: payments,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
		Pages:    int((total + int64(perPage) - 1) / int64(perPage)),
	}

	c.JSON(http.StatusOK, response)
}

func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	payment, err := h.paymentService.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Check if user owns this payment or is admin
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	if userRole != "admin" && payment.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var request struct {
		PaymentMethodID string `json:"payment_method_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.ProcessPayment(c.Request.Context(), paymentID, request.PaymentMethodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := h.paymentService.CancelPayment(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// Payment method operations
func (h *PaymentHandler) CreatePaymentMethod(c *gin.Context) {
	var request service.CreatePaymentMethodRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	request.UserID = userID.(uuid.UUID)

	paymentMethod, err := h.paymentService.CreatePaymentMethod(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, paymentMethod)
}

func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	paymentMethods, err := h.paymentService.GetPaymentMethodsByUserID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment_methods": paymentMethods})
}

func (h *PaymentHandler) SetDefaultPaymentMethod(c *gin.Context) {
	paymentMethodID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.paymentService.SetDefaultPaymentMethod(c.Request.Context(), userID.(uuid.UUID), paymentMethodID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default payment method updated"})
}

func (h *PaymentHandler) DeletePaymentMethod(c *gin.Context) {
	paymentMethodID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method ID"})
		return
	}

	if err := h.paymentService.DeletePaymentMethod(c.Request.Context(), paymentMethodID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment method deleted"})
}

// Refund operations
func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	var request service.CreateRefundRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.PaymentID = paymentID

	refund, err := h.paymentService.CreateRefund(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, refund)
}

func (h *PaymentHandler) GetRefundsByPaymentID(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	refunds, err := h.paymentService.GetRefundsByPaymentID(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"refunds": refunds})
}

func (h *PaymentHandler) GetRefund(c *gin.Context) {
	refundID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refund ID"})
		return
	}

	refund, err := h.paymentService.GetRefund(c.Request.Context(), refundID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Refund not found"})
		return
	}

	c.JSON(http.StatusOK, refund)
}

// Analytics operations (admin only)
func (h *PaymentHandler) GetPaymentStatistics(c *gin.Context) {
	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0) // Default: last month
	endDate := time.Now()

	if sd := c.Query("start_date"); sd != "" {
		if parsed, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = parsed
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if parsed, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = parsed
		}
	}

	statistics, err := h.paymentService.GetPaymentStatistics(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statistics)
}

func (h *PaymentHandler) GetRevenueMetrics(c *gin.Context) {
	// Parse date range
	startDate := time.Now().AddDate(0, -1, 0) // Default: last month
	endDate := time.Now()

	if sd := c.Query("start_date"); sd != "" {
		if parsed, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = parsed
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if parsed, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = parsed
		}
	}

	metrics, err := h.paymentService.GetRevenueMetrics(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := service.RevenueAnalyticsResponse{
		Period:    "custom",
		StartDate: startDate,
		EndDate:   endDate,
		Metrics:   metrics,
	}

	c.JSON(http.StatusOK, response)
}

// Webhook handling
func (h *PaymentHandler) HandleStripeWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe signature"})
		return
	}

	if err := h.paymentService.ProcessWebhook(c.Request.Context(), payload, signature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}