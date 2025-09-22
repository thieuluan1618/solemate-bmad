package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type Order struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderNumber       string         `json:"order_number" gorm:"unique;not null;index"`
	Status            OrderStatus    `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	PaymentStatus     PaymentStatus  `json:"payment_status" gorm:"type:varchar(20);not null;default:'pending'"`

	// Items and pricing
	Items             []OrderItem    `json:"items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	ItemCount         int            `json:"item_count" gorm:"not null;default:0"`
	SubtotalPrice     float64        `json:"subtotal_price" gorm:"type:decimal(10,2);not null;default:0"`
	TaxAmount         float64        `json:"tax_amount" gorm:"type:decimal(10,2);not null;default:0"`
	ShippingCost      float64        `json:"shipping_cost" gorm:"type:decimal(10,2);not null;default:0"`
	DiscountAmount    float64        `json:"discount_amount" gorm:"type:decimal(10,2);not null;default:0"`
	TotalPrice        float64        `json:"total_price" gorm:"type:decimal(10,2);not null;default:0"`

	// Addresses
	ShippingAddress   Address        `json:"shipping_address" gorm:"embedded;embeddedPrefix:shipping_"`
	BillingAddress    Address        `json:"billing_address" gorm:"embedded;embeddedPrefix:billing_"`

	// Shipping and tracking
	ShippingMethod    string         `json:"shipping_method" gorm:"type:varchar(100)"`
	TrackingNumber    string         `json:"tracking_number" gorm:"type:varchar(100);index"`
	EstimatedDelivery *time.Time     `json:"estimated_delivery"`
	ActualDelivery    *time.Time     `json:"actual_delivery"`

	// Payment information
	PaymentMethodID   *uuid.UUID     `json:"payment_method_id" gorm:"type:uuid"`
	TransactionID     string         `json:"transaction_id" gorm:"type:varchar(100);index"`

	// Order metadata
	Notes             string         `json:"notes" gorm:"type:text"`
	CancelReason      string         `json:"cancel_reason" gorm:"type:text"`

	// Timestamps
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	ConfirmedAt       *time.Time     `json:"confirmed_at"`
	ShippedAt         *time.Time     `json:"shipped_at"`
	DeliveredAt       *time.Time     `json:"delivered_at"`
	CompletedAt       *time.Time     `json:"completed_at"`
	CancelledAt       *time.Time     `json:"cancelled_at"`
}

type OrderItem struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID       uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	ProductID     uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	VariantID     *uuid.UUID `json:"variant_id" gorm:"type:uuid"`

	// Product details (snapshot at time of order)
	ProductName   string    `json:"product_name" gorm:"type:varchar(255);not null"`
	ProductSKU    string    `json:"product_sku" gorm:"type:varchar(100);not null"`
	VariantName   string    `json:"variant_name" gorm:"type:varchar(255)"`
	Size          string    `json:"size" gorm:"type:varchar(20)"`
	Color         string    `json:"color" gorm:"type:varchar(50)"`
	ImageURL      string    `json:"image_url" gorm:"type:text"`

	// Pricing and quantity
	UnitPrice     float64   `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Quantity      int       `json:"quantity" gorm:"not null;check:quantity > 0"`
	Discount      float64   `json:"discount" gorm:"type:decimal(10,2);not null;default:0"`
	TotalPrice    float64   `json:"total_price" gorm:"type:decimal(10,2);not null"`

	// Timestamps
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Address struct {
	FirstName   string `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName    string `json:"last_name" gorm:"type:varchar(100);not null"`
	Company     string `json:"company" gorm:"type:varchar(100)"`
	AddressLine1 string `json:"address_line_1" gorm:"type:varchar(255);not null"`
	AddressLine2 string `json:"address_line_2" gorm:"type:varchar(255)"`
	City        string `json:"city" gorm:"type:varchar(100);not null"`
	StateProvince string `json:"state_province" gorm:"type:varchar(100);not null"`
	PostalCode  string `json:"postal_code" gorm:"type:varchar(20);not null"`
	Country     string `json:"country" gorm:"type:varchar(100);not null;default:'US'"`
	Phone       string `json:"phone" gorm:"type:varchar(20)"`
}

type OrderSummary struct {
	ID              uuid.UUID     `json:"id"`
	OrderNumber     string        `json:"order_number"`
	Status          OrderStatus   `json:"status"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	ItemCount       int           `json:"item_count"`
	TotalPrice      float64       `json:"total_price"`
	CreatedAt       time.Time     `json:"created_at"`
	EstimatedDelivery *time.Time  `json:"estimated_delivery"`
}

// Order state machine methods
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:    {OrderStatusDelivered, OrderStatusCancelled},
		OrderStatusDelivered:  {OrderStatusCompleted, OrderStatusRefunded},
		OrderStatusCompleted:  {OrderStatusRefunded},
		OrderStatusCancelled:  {}, // Terminal state
		OrderStatusRefunded:   {}, // Terminal state
	}

	allowedStatuses, exists := validTransitions[o.Status]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}
	return false
}

func (o *Order) TransitionTo(newStatus OrderStatus) error {
	if !o.CanTransitionTo(newStatus) {
		return ErrInvalidStatusTransition
	}

	now := time.Now()
	previousStatus := o.Status
	o.Status = newStatus
	o.UpdatedAt = now

	// Set timestamp fields based on status
	switch newStatus {
	case OrderStatusConfirmed:
		o.ConfirmedAt = &now
	case OrderStatusShipped:
		o.ShippedAt = &now
	case OrderStatusDelivered:
		o.DeliveredAt = &now
	case OrderStatusCompleted:
		o.CompletedAt = &now
	case OrderStatusCancelled:
		o.CancelledAt = &now
	}

	return nil
}

func (o *Order) CalculateTotals() {
	o.SubtotalPrice = 0
	o.ItemCount = 0

	for _, item := range o.Items {
		o.SubtotalPrice += item.TotalPrice
		o.ItemCount += item.Quantity
	}

	o.TotalPrice = o.SubtotalPrice + o.TaxAmount + o.ShippingCost - o.DiscountAmount
}

func (o *Order) IsEditable() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

func (o *Order) IsCancellable() bool {
	return o.Status == OrderStatusPending ||
		   o.Status == OrderStatusConfirmed ||
		   o.Status == OrderStatusProcessing
}

func (o *Order) IsRefundable() bool {
	return o.Status == OrderStatusDelivered ||
		   o.Status == OrderStatusCompleted
}

// Order item methods
func (oi *OrderItem) CalculateTotal() {
	oi.TotalPrice = (oi.UnitPrice * float64(oi.Quantity)) - oi.Discount
}

// Custom errors
type OrderError struct {
	Message string
}

func (e OrderError) Error() string {
	return e.Message
}

var (
	ErrInvalidStatusTransition = OrderError{Message: "invalid status transition"}
	ErrOrderNotEditable       = OrderError{Message: "order is not editable"}
	ErrOrderNotCancellable    = OrderError{Message: "order cannot be cancelled"}
	ErrOrderNotRefundable     = OrderError{Message: "order is not refundable"}
	ErrInvalidOrderData       = OrderError{Message: "invalid order data"}
)