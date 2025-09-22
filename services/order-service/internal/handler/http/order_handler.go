package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/order-service/internal/domain/entity"
	"solemate/services/order-service/internal/domain/repository"
	"solemate/services/order-service/internal/domain/service"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// Request DTOs
type CreateOrderRequest struct {
	ShippingAddress entity.Address `json:"shipping_address" binding:"required"`
	BillingAddress  entity.Address `json:"billing_address" binding:"required"`
	ShippingMethod  string         `json:"shipping_method" binding:"required"`
	Notes           string         `json:"notes"`
}

type UpdateOrderStatusRequest struct {
	Status entity.OrderStatus `json:"status" binding:"required"`
	Notes  string             `json:"notes"`
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus entity.PaymentStatus `json:"payment_status" binding:"required"`
	TransactionID string               `json:"transaction_id"`
}

type ShipOrderRequest struct {
	TrackingNumber    string     `json:"tracking_number" binding:"required"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
}

type UpdateAddressRequest struct {
	Address entity.Address `json:"address" binding:"required"`
}

type OrderSearchRequest struct {
	UserID        *uuid.UUID              `json:"user_id"`
	Status        *entity.OrderStatus     `json:"status"`
	PaymentStatus *entity.PaymentStatus   `json:"payment_status"`
	DateFrom      *time.Time              `json:"date_from"`
	DateTo        *time.Time              `json:"date_to"`
	MinAmount     *float64                `json:"min_amount"`
	MaxAmount     *float64                `json:"max_amount"`
	SearchTerm    string                  `json:"search_term"`
	SortBy        string                  `json:"sort_by"`
	SortOrder     string                  `json:"sort_order"`
}

// Order creation and retrieval
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	order, err := h.orderService.CreateOrderFromCart(
		c.Request.Context(),
		userUUID,
		req.ShippingAddress,
		req.BillingAddress,
		req.ShippingMethod,
		req.Notes,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create order", err.Error())
		return
	}

	utils.CreatedResponse(c, "Order created successfully", order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format", "invalid_order_id")
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Order not found", err.Error())
		return
	}

	// Check if user can access this order
	userID, exists := c.Get("user_id")
	if exists {
		userUUID, ok := userID.(uuid.UUID)
		if ok && order.UserID != userUUID {
			// Check if user has admin role
			userRole, roleExists := c.Get("user_role")
			if !roleExists || userRole != "admin" {
				utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "insufficient_permissions")
				return
			}
		}
	}

	utils.SuccessResponse(c, "Order retrieved successfully", order)
}

func (h *OrderHandler) GetOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")

	order, err := h.orderService.GetOrderByNumber(c.Request.Context(), orderNumber)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Order not found", err.Error())
		return
	}

	// Check if user can access this order
	userID, exists := c.Get("user_id")
	if exists {
		userUUID, ok := userID.(uuid.UUID)
		if ok && order.UserID != userUUID {
			// Check if user has admin role
			userRole, roleExists := c.Get("user_role")
			if !roleExists || userRole != "admin" {
				utils.ErrorResponse(c, http.StatusForbidden, "Access denied", "insufficient_permissions")
				return
			}
		}
	}

	utils.SuccessResponse(c, "Order retrieved successfully", order)
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orders, total, err := h.orderService.GetUserOrders(c.Request.Context(), userUUID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get orders", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.PaginatedSuccessResponse(c, "Orders retrieved successfully", orders, pagination)
}

func (h *OrderHandler) GetUserOrderSummaries(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	summaries, total, err := h.orderService.GetUserOrderSummaries(c.Request.Context(), userUUID, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get order summaries", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.PaginatedSuccessResponse(c, "Order summaries retrieved successfully", summaries, pagination)
}

// Order status management
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format", "invalid_order_id")
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Route to appropriate status update method
	switch req.Status {
	case entity.OrderStatusConfirmed:
		err = h.orderService.ConfirmOrder(c.Request.Context(), orderID)
	case entity.OrderStatusProcessing:
		err = h.orderService.ProcessOrder(c.Request.Context(), orderID)
	case entity.OrderStatusDelivered:
		err = h.orderService.DeliverOrder(c.Request.Context(), orderID)
	case entity.OrderStatusCompleted:
		err = h.orderService.CompleteOrder(c.Request.Context(), orderID)
	case entity.OrderStatusCancelled:
		err = h.orderService.CancelOrder(c.Request.Context(), orderID, req.Notes)
	case entity.OrderStatusRefunded:
		err = h.orderService.RefundOrder(c.Request.Context(), orderID, req.Notes)
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid status transition", "invalid_status")
		return
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update order status", err.Error())
		return
	}

	utils.SuccessResponse(c, "Order status updated successfully", gin.H{
		"order_id": orderID,
		"status":   req.Status,
	})
}

func (h *OrderHandler) ShipOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format", "invalid_order_id")
		return
	}

	var req ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err = h.orderService.ShipOrder(c.Request.Context(), orderID, req.TrackingNumber, req.EstimatedDelivery)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to ship order", err.Error())
		return
	}

	utils.SuccessResponse(c, "Order shipped successfully", gin.H{
		"order_id":         orderID,
		"tracking_number":  req.TrackingNumber,
		"estimated_delivery": req.EstimatedDelivery,
	})
}

func (h *OrderHandler) UpdatePaymentStatus(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format", "invalid_order_id")
		return
	}

	var req UpdatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err = h.orderService.UpdatePaymentStatus(c.Request.Context(), orderID, req.PaymentStatus, req.TransactionID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update payment status", err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment status updated successfully", gin.H{
		"order_id":       orderID,
		"payment_status": req.PaymentStatus,
		"transaction_id": req.TransactionID,
	})
}

// Order modifications
func (h *OrderHandler) UpdateShippingAddress(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format", "invalid_order_id")
		return
	}

	var req UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err = h.orderService.UpdateShippingAddress(c.Request.Context(), orderID, req.Address)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update shipping address", err.Error())
		return
	}

	utils.SuccessResponse(c, "Shipping address updated successfully", gin.H{
		"order_id": orderID,
	})
}

func (h *OrderHandler) UpdateBillingAddress(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format", "invalid_order_id")
		return
	}

	var req UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err = h.orderService.UpdateBillingAddress(c.Request.Context(), orderID, req.Address)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update billing address", err.Error())
		return
	}

	utils.SuccessResponse(c, "Billing address updated successfully", gin.H{
		"order_id": orderID,
	})
}

// Administrative functions
func (h *OrderHandler) SearchOrders(c *gin.Context) {
	// Check admin permissions
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", "insufficient_permissions")
		return
	}

	var req OrderSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	filters := &repository.OrderFilters{
		UserID:        req.UserID,
		Status:        req.Status,
		PaymentStatus: req.PaymentStatus,
		DateFrom:      req.DateFrom,
		DateTo:        req.DateTo,
		MinAmount:     req.MinAmount,
		MaxAmount:     req.MaxAmount,
		SearchTerm:    req.SearchTerm,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
		Limit:         limit,
		Offset:        (page - 1) * limit,
	}

	orders, total, err := h.orderService.SearchOrders(c.Request.Context(), filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search orders", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.PaginatedSuccessResponse(c, "Orders retrieved successfully", orders, pagination)
}

func (h *OrderHandler) GetOrderStatistics(c *gin.Context) {
	// Check admin permissions
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", "insufficient_permissions")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "invalid_date")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "invalid_date")
			return
		}
	} else {
		endDate = time.Now()
	}

	statistics, err := h.orderService.GetOrderStatistics(c.Request.Context(), startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get order statistics", err.Error())
		return
	}

	utils.SuccessResponse(c, "Order statistics retrieved successfully", statistics)
}

func (h *OrderHandler) GetTopProducts(c *gin.Context) {
	// Check admin permissions
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", "insufficient_permissions")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "10")

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "invalid_date")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "invalid_date")
			return
		}
	} else {
		endDate = time.Now()
	}

	products, err := h.orderService.GetTopProducts(c.Request.Context(), startDate, endDate, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get top products", err.Error())
		return
	}

	utils.SuccessResponse(c, "Top products retrieved successfully", products)
}

func (h *OrderHandler) GetSalesMetrics(c *gin.Context) {
	// Check admin permissions
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", "insufficient_permissions")
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start date format", "invalid_date")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end date format", "invalid_date")
			return
		}
	} else {
		endDate = time.Now()
	}

	metrics, err := h.orderService.GetSalesMetrics(c.Request.Context(), startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get sales metrics", err.Error())
		return
	}

	utils.SuccessResponse(c, "Sales metrics retrieved successfully", metrics)
}

// Route registration
func (h *OrderHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware, adminMiddleware gin.HandlerFunc) {
	orders := router.Group("/orders")
	orders.Use(authMiddleware)

	// User order routes
	orders.POST("", h.CreateOrder)
	orders.GET("/me", h.GetUserOrders)
	orders.GET("/me/summaries", h.GetUserOrderSummaries)
	orders.GET("/:order_id", h.GetOrder)
	orders.GET("/number/:order_number", h.GetOrderByNumber)

	// Order modification routes (user can modify pending/confirmed orders)
	orders.PATCH("/:order_id/shipping-address", h.UpdateShippingAddress)
	orders.PATCH("/:order_id/billing-address", h.UpdateBillingAddress)

	// Admin routes
	admin := orders.Group("/admin")
	admin.Use(adminMiddleware)
	{
		admin.POST("/search", h.SearchOrders)
		admin.GET("/statistics", h.GetOrderStatistics)
		admin.GET("/top-products", h.GetTopProducts)
		admin.GET("/sales-metrics", h.GetSalesMetrics)
		admin.PATCH("/:order_id/status", h.UpdateOrderStatus)
		admin.POST("/:order_id/ship", h.ShipOrder)
		admin.PATCH("/:order_id/payment-status", h.UpdatePaymentStatus)
	}
}