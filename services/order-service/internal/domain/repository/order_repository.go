package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"solemate/services/order-service/internal/domain/entity"
)

type OrderRepository interface {
	// Order CRUD operations
	CreateOrder(ctx context.Context, order *entity.Order) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entity.Order, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*entity.Order, error)
	UpdateOrder(ctx context.Context, order *entity.Order) error
	DeleteOrder(ctx context.Context, orderID uuid.UUID) error

	// Order querying
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Order, int64, error)
	GetOrdersByStatus(ctx context.Context, status entity.OrderStatus, limit, offset int) ([]*entity.Order, int64, error)
	GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Order, int64, error)
	GetOrdersByPaymentStatus(ctx context.Context, paymentStatus entity.PaymentStatus, limit, offset int) ([]*entity.Order, int64, error)

	// Order summaries
	GetOrderSummariesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.OrderSummary, int64, error)
	GetRecentOrders(ctx context.Context, limit int) ([]*entity.OrderSummary, error)

	// Order status management
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status entity.OrderStatus, notes string) error
	UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus entity.PaymentStatus, transactionID string) error
	UpdateTrackingInfo(ctx context.Context, orderID uuid.UUID, trackingNumber string, estimatedDelivery *time.Time) error

	// Order analytics
	GetOrderStatistics(ctx context.Context, startDate, endDate time.Time) (*OrderStatistics, error)
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]*ProductSalesInfo, error)
	GetSalesMetrics(ctx context.Context, startDate, endDate time.Time) (*SalesMetrics, error)

	// Order search and filtering
	SearchOrders(ctx context.Context, filters *OrderFilters) ([]*entity.Order, int64, error)
}

type CartRepository interface {
	// Integration with cart service
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*CartData, error)
	ClearCartByUserID(ctx context.Context, userID uuid.UUID) error
}

type ProductRepository interface {
	// Product validation for orders
	ValidateProductAvailability(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, quantity int) (bool, error)
	GetProductInfo(ctx context.Context, productID uuid.UUID) (*ProductInfo, error)
	GetProductVariantInfo(ctx context.Context, variantID uuid.UUID) (*ProductVariantInfo, error)
	ReserveStock(ctx context.Context, items []StockReservation) error
	ReleaseStock(ctx context.Context, items []StockReservation) error
}

type NotificationRepository interface {
	// Order notifications
	SendOrderConfirmation(ctx context.Context, order *entity.Order) error
	SendOrderStatusUpdate(ctx context.Context, order *entity.Order, previousStatus entity.OrderStatus) error
	SendShippingNotification(ctx context.Context, order *entity.Order) error
	SendDeliveryNotification(ctx context.Context, order *entity.Order) error
}

// Supporting types for repository operations
type OrderFilters struct {
	UserID           *uuid.UUID
	Status           *entity.OrderStatus
	PaymentStatus    *entity.PaymentStatus
	DateFrom         *time.Time
	DateTo           *time.Time
	MinAmount        *float64
	MaxAmount        *float64
	SearchTerm       string // Search in order number, customer name, etc.
	SortBy           string // "created_at", "total_price", "status"
	SortOrder        string // "asc", "desc"
	Limit            int
	Offset           int
}

type OrderStatistics struct {
	TotalOrders       int64               `json:"total_orders"`
	TotalRevenue      float64             `json:"total_revenue"`
	AverageOrderValue float64             `json:"average_order_value"`
	StatusBreakdown   map[entity.OrderStatus]int64 `json:"status_breakdown"`
	PaymentBreakdown  map[entity.PaymentStatus]int64 `json:"payment_breakdown"`
	OrdersByDay       []DailyOrderStats   `json:"orders_by_day"`
}

type DailyOrderStats struct {
	Date      time.Time `json:"date"`
	Orders    int64     `json:"orders"`
	Revenue   float64   `json:"revenue"`
}

type ProductSalesInfo struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	ProductSKU    string    `json:"product_sku"`
	TotalQuantity int       `json:"total_quantity"`
	TotalRevenue  float64   `json:"total_revenue"`
	OrderCount    int       `json:"order_count"`
}

type SalesMetrics struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalOrders       int64   `json:"total_orders"`
	AverageOrderValue float64 `json:"average_order_value"`
	TotalItems        int64   `json:"total_items"`
	AverageItemsPerOrder float64 `json:"average_items_per_order"`
	ConversionRate    float64 `json:"conversion_rate"`
	ReturnRate        float64 `json:"return_rate"`
}

// External service data types
type CartData struct {
	UserID     uuid.UUID  `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
}

type CartItem struct {
	ProductID   uuid.UUID  `json:"product_id"`
	VariantID   *uuid.UUID `json:"variant_id"`
	ProductName string     `json:"product_name"`
	SKU         string     `json:"sku"`
	Size        string     `json:"size"`
	Color       string     `json:"color"`
	UnitPrice   float64    `json:"unit_price"`
	Quantity    int        `json:"quantity"`
	ImageURL    string     `json:"image_url"`
}

type ProductInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	Price     float64   `json:"price"`
	ImageURL  string    `json:"image_url"`
	Available bool      `json:"available"`
	Stock     int       `json:"stock"`
}

type ProductVariantInfo struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	SKU       string    `json:"sku"`
	Size      string    `json:"size"`
	Color     string    `json:"color"`
	Price     float64   `json:"price"`
	Available bool      `json:"available"`
	Stock     int       `json:"stock"`
	ImageURL  string    `json:"image_url"`
}

type StockReservation struct {
	ProductID uuid.UUID  `json:"product_id"`
	VariantID *uuid.UUID `json:"variant_id"`
	Quantity  int        `json:"quantity"`
}