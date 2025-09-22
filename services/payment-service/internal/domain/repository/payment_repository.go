package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"solemate/services/payment-service/internal/domain/entity"
)

type PaymentRepository interface {
	// Payment CRUD operations
	CreatePayment(ctx context.Context, payment *entity.Payment) error
	GetPaymentByID(ctx context.Context, paymentID uuid.UUID) (*entity.Payment, error)
	GetPaymentByStripePaymentIntentID(ctx context.Context, stripePaymentIntentID string) (*entity.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Payment, error)
	UpdatePayment(ctx context.Context, payment *entity.Payment) error
	DeletePayment(ctx context.Context, paymentID uuid.UUID) error

	// Payment querying
	GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Payment, int64, error)
	GetPaymentsByStatus(ctx context.Context, status entity.PaymentStatus, limit, offset int) ([]*entity.Payment, int64, error)
	GetPaymentsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Payment, int64, error)
	GetPaymentSummariesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.PaymentSummary, int64, error)

	// Payment analytics
	GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time) (*PaymentStatistics, error)
	GetPaymentMethodStats(ctx context.Context, startDate, endDate time.Time) ([]*PaymentMethodStats, error)
	GetRevenueMetrics(ctx context.Context, startDate, endDate time.Time) (*RevenueMetrics, error)
}

type PaymentMethodRepository interface {
	// Payment method CRUD operations
	CreatePaymentMethod(ctx context.Context, paymentMethod *entity.PaymentMethod) error
	GetPaymentMethodByID(ctx context.Context, paymentMethodID uuid.UUID) (*entity.PaymentMethod, error)
	GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*entity.PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, paymentMethod *entity.PaymentMethod) error
	DeletePaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error

	// Payment method querying
	GetPaymentMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PaymentMethod, error)
	GetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (*entity.PaymentMethod, error)
	SetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID uuid.UUID) error
	UnsetDefaultPaymentMethods(ctx context.Context, userID uuid.UUID) error
}

type RefundRepository interface {
	// Refund CRUD operations
	CreateRefund(ctx context.Context, refund *entity.Refund) error
	GetRefundByID(ctx context.Context, refundID uuid.UUID) (*entity.Refund, error)
	GetRefundByStripeID(ctx context.Context, stripeRefundID string) (*entity.Refund, error)
	UpdateRefund(ctx context.Context, refund *entity.Refund) error
	DeleteRefund(ctx context.Context, refundID uuid.UUID) error

	// Refund querying
	GetRefundsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entity.Refund, error)
	GetRefundsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Refund, int64, error)
	GetRefundsByStatus(ctx context.Context, status entity.RefundStatus, limit, offset int) ([]*entity.Refund, int64, error)
	GetRefundsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Refund, int64, error)
}

type WebhookRepository interface {
	// Webhook event operations
	CreateWebhookEvent(ctx context.Context, event *entity.WebhookEvent) error
	GetWebhookEventByStripeID(ctx context.Context, stripeEventID string) (*entity.WebhookEvent, error)
	UpdateWebhookEvent(ctx context.Context, event *entity.WebhookEvent) error
	GetUnprocessedWebhookEvents(ctx context.Context, limit int) ([]*entity.WebhookEvent, error)
	MarkWebhookEventAsProcessed(ctx context.Context, eventID uuid.UUID) error
	MarkWebhookEventAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error
}

type StripeRepository interface {
	// Stripe payment operations
	CreatePaymentIntent(ctx context.Context, request *CreatePaymentIntentRequest) (*PaymentIntentResponse, error)
	ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (*PaymentIntentResponse, error)
	CancelPaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentIntentResponse, error)
	RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (*PaymentIntentResponse, error)

	// Stripe payment method operations
	CreatePaymentMethod(ctx context.Context, request *CreatePaymentMethodRequest) (*PaymentMethodResponse, error)
	AttachPaymentMethod(ctx context.Context, paymentMethodID string, customerID string) (*PaymentMethodResponse, error)
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethodResponse, error)
	RetrievePaymentMethod(ctx context.Context, paymentMethodID string) (*PaymentMethodResponse, error)

	// Stripe customer operations
	CreateCustomer(ctx context.Context, request *CreateCustomerRequest) (*CustomerResponse, error)
	RetrieveCustomer(ctx context.Context, customerID string) (*CustomerResponse, error)
	UpdateCustomer(ctx context.Context, customerID string, request *UpdateCustomerRequest) (*CustomerResponse, error)

	// Stripe refund operations
	CreateRefund(ctx context.Context, request *CreateRefundRequest) (*RefundResponse, error)
	RetrieveRefund(ctx context.Context, refundID string) (*RefundResponse, error)

	// Stripe webhook operations
	ConstructWebhookEvent(ctx context.Context, payload []byte, signature string) (*WebhookEventData, error)
}

type OrderRepository interface {
	// Integration with order service
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderData, error)
	UpdateOrderPaymentStatus(ctx context.Context, orderID uuid.UUID, status string, transactionID string) error
}

// Supporting types for repository operations
type PaymentStatistics struct {
	TotalPayments     int64                            `json:"total_payments"`
	TotalRevenue      float64                          `json:"total_revenue"`
	SuccessfulPayments int64                           `json:"successful_payments"`
	FailedPayments    int64                            `json:"failed_payments"`
	RefundedAmount    float64                          `json:"refunded_amount"`
	StatusBreakdown   map[entity.PaymentStatus]int64   `json:"status_breakdown"`
	PaymentsByDay     []DailyPaymentStats              `json:"payments_by_day"`
}

type DailyPaymentStats struct {
	Date     time.Time `json:"date"`
	Payments int64     `json:"payments"`
	Revenue  float64   `json:"revenue"`
}

type PaymentMethodStats struct {
	Type         entity.PaymentMethodType `json:"type"`
	Brand        string                   `json:"brand"`
	Count        int64                    `json:"count"`
	TotalAmount  float64                  `json:"total_amount"`
}

type RevenueMetrics struct {
	TotalRevenue       float64 `json:"total_revenue"`
	NetRevenue         float64 `json:"net_revenue"` // After refunds
	RefundedRevenue    float64 `json:"refunded_revenue"`
	AverageTransaction float64 `json:"average_transaction"`
	TransactionCount   int64   `json:"transaction_count"`
	RefundRate         float64 `json:"refund_rate"`
}

// Stripe API request/response types
type CreatePaymentIntentRequest struct {
	Amount        int64  `json:"amount"` // Amount in cents
	Currency      string `json:"currency"`
	CustomerID    string `json:"customer_id,omitempty"`
	Description   string `json:"description,omitempty"`
	OrderID       string `json:"order_id,omitempty"`
	AutomaticPaymentMethods bool `json:"automatic_payment_methods"`
}

type PaymentIntentResponse struct {
	ID           string `json:"id"`
	ClientSecret string `json:"client_secret"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	CustomerID   string `json:"customer_id,omitempty"`
	ChargeID     string `json:"charge_id,omitempty"`
}

type CreatePaymentMethodRequest struct {
	Type string          `json:"type"`
	Card *CardDetails    `json:"card,omitempty"`
}

type CardDetails struct {
	Number   string `json:"number"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
	CVC      string `json:"cvc"`
}

type PaymentMethodResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Card     *Card  `json:"card,omitempty"`
	CustomerID string `json:"customer_id,omitempty"`
}

type Card struct {
	Brand    string `json:"brand"`
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

type CreateCustomerRequest struct {
	Email       string  `json:"email"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Address     *StripeAddress `json:"address,omitempty"`
}

type UpdateCustomerRequest struct {
	Email       string  `json:"email,omitempty"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Address     *StripeAddress `json:"address,omitempty"`
}

type CustomerResponse struct {
	ID          string         `json:"id"`
	Email       string         `json:"email"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Address     *StripeAddress `json:"address,omitempty"`
}

type StripeAddress struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

type CreateRefundRequest struct {
	PaymentIntentID string  `json:"payment_intent_id"`
	Amount          *int64  `json:"amount,omitempty"` // Amount in cents, nil for full refund
	Reason          string  `json:"reason,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type RefundResponse struct {
	ID              string `json:"id"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	PaymentIntentID string `json:"payment_intent_id"`
	Status          string `json:"status"`
	Reason          string `json:"reason,omitempty"`
}

type WebhookEventData struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Created int64       `json:"created"`
}

// External service data types
type OrderData struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
}

// Search and filtering types
type PaymentFilters struct {
	UserID       *uuid.UUID
	Status       *entity.PaymentStatus
	OrderID      *uuid.UUID
	DateFrom     *time.Time
	DateTo       *time.Time
	MinAmount    *float64
	MaxAmount    *float64
	Currency     string
	SearchTerm   string
	SortBy       string
	SortOrder    string
	Limit        int
	Offset       int
}