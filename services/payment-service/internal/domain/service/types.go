package service

import (
	"time"

	"github.com/google/uuid"
	"solemate/services/payment-service/internal/domain/entity"
	"solemate/services/payment-service/internal/domain/repository"
)

// Request types
type CreatePaymentRequest struct {
	UserID            uuid.UUID `json:"user_id" validate:"required"`
	OrderID           uuid.UUID `json:"order_id" validate:"required"`
	Amount            float64   `json:"amount" validate:"required,gt=0"`
	Currency          string    `json:"currency" validate:"required,len=3"`
	StripeCustomerID  string    `json:"stripe_customer_id,omitempty"`
	Description       string    `json:"description,omitempty"`
}

type CreatePaymentMethodRequest struct {
	UserID            uuid.UUID             `json:"user_id" validate:"required"`
	Type              string                `json:"type" validate:"required"`
	Card              *repository.CardDetails `json:"card,omitempty"`
	StripeCustomerID  string                `json:"stripe_customer_id,omitempty"`
	BillingAddress    entity.Address        `json:"billing_address"`
	IsDefault         bool                  `json:"is_default"`
}

type CreateRefundRequest struct {
	PaymentID   uuid.UUID `json:"payment_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Reason      string    `json:"reason" validate:"required"`
	Description string    `json:"description,omitempty"`
}

// Response types
type PaymentResponse struct {
	ID                    uuid.UUID                 `json:"id"`
	UserID                uuid.UUID                 `json:"user_id"`
	OrderID               uuid.UUID                 `json:"order_id"`
	PaymentMethodID       *uuid.UUID                `json:"payment_method_id,omitempty"`
	Amount                float64                   `json:"amount"`
	Currency              string                    `json:"currency"`
	Status                entity.PaymentStatus      `json:"status"`
	StripePaymentIntentID string                    `json:"stripe_payment_intent_id"`
	StripeChargeID        string                    `json:"stripe_charge_id,omitempty"`
	ClientSecret          string                    `json:"client_secret,omitempty"`
	Description           string                    `json:"description,omitempty"`
	FailureReason         string                    `json:"failure_reason,omitempty"`
	FailureCode           string                    `json:"failure_code,omitempty"`
	CreatedAt             time.Time                 `json:"created_at"`
	UpdatedAt             time.Time                 `json:"updated_at"`
	ProcessedAt           *time.Time                `json:"processed_at,omitempty"`
	FailedAt              *time.Time                `json:"failed_at,omitempty"`
	PaymentMethod         *PaymentMethodResponse    `json:"payment_method,omitempty"`
	Refunds               []*RefundResponse         `json:"refunds,omitempty"`
}

type PaymentSummaryResponse struct {
	ID          uuid.UUID            `json:"id"`
	OrderID     uuid.UUID            `json:"order_id"`
	Amount      float64              `json:"amount"`
	Currency    string               `json:"currency"`
	Status      entity.PaymentStatus `json:"status"`
	CardLast4   string               `json:"card_last4,omitempty"`
	CardBrand   string               `json:"card_brand,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	ProcessedAt *time.Time           `json:"processed_at,omitempty"`
}

type PaymentMethodResponse struct {
	ID                    uuid.UUID              `json:"id"`
	UserID                uuid.UUID              `json:"user_id"`
	Type                  entity.PaymentMethodType `json:"type"`
	StripePaymentMethodID string                 `json:"stripe_payment_method_id"`
	CardBrand             string                 `json:"card_brand,omitempty"`
	CardLast4             string                 `json:"card_last4,omitempty"`
	CardExpMonth          int                    `json:"card_exp_month,omitempty"`
	CardExpYear           int                    `json:"card_exp_year,omitempty"`
	BillingAddress        entity.Address         `json:"billing_address"`
	IsDefault             bool                   `json:"is_default"`
	IsExpired             bool                   `json:"is_expired"`
	DisplayName           string                 `json:"display_name"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

type RefundResponse struct {
	ID             uuid.UUID           `json:"id"`
	PaymentID      uuid.UUID           `json:"payment_id"`
	UserID         uuid.UUID           `json:"user_id"`
	OrderID        uuid.UUID           `json:"order_id"`
	Amount         float64             `json:"amount"`
	Currency       string              `json:"currency"`
	Status         entity.RefundStatus `json:"status"`
	StripeRefundID string              `json:"stripe_refund_id"`
	Reason         string              `json:"reason"`
	Description    string              `json:"description,omitempty"`
	FailureReason  string              `json:"failure_reason,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	ProcessedAt    *time.Time          `json:"processed_at,omitempty"`
}

// Pagination types
type PaginatedPaymentResponse struct {
	Payments []*PaymentSummaryResponse `json:"payments"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PerPage  int                       `json:"per_page"`
	Pages    int                       `json:"pages"`
}

type PaginatedRefundResponse struct {
	Refunds []*RefundResponse `json:"refunds"`
	Total   int64             `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
	Pages   int               `json:"pages"`
}

// Filter types for search and reporting
type PaymentFilterRequest struct {
	UserID    *uuid.UUID            `json:"user_id,omitempty"`
	OrderID   *uuid.UUID            `json:"order_id,omitempty"`
	Status    *entity.PaymentStatus `json:"status,omitempty"`
	StartDate *time.Time            `json:"start_date,omitempty"`
	EndDate   *time.Time            `json:"end_date,omitempty"`
	MinAmount *float64              `json:"min_amount,omitempty"`
	MaxAmount *float64              `json:"max_amount,omitempty"`
	Currency  string                `json:"currency,omitempty"`
	Page      int                   `json:"page,omitempty"`
	PerPage   int                   `json:"per_page,omitempty"`
}

type RefundFilterRequest struct {
	PaymentID *uuid.UUID          `json:"payment_id,omitempty"`
	UserID    *uuid.UUID          `json:"user_id,omitempty"`
	OrderID   *uuid.UUID          `json:"order_id,omitempty"`
	Status    *entity.RefundStatus `json:"status,omitempty"`
	StartDate *time.Time          `json:"start_date,omitempty"`
	EndDate   *time.Time          `json:"end_date,omitempty"`
	Page      int                 `json:"page,omitempty"`
	PerPage   int                 `json:"per_page,omitempty"`
}

// Analytics response types
type PaymentAnalyticsResponse struct {
	Statistics      *repository.PaymentStatistics   `json:"statistics"`
	RevenueMetrics  *repository.RevenueMetrics      `json:"revenue_metrics"`
	MethodStats     []*repository.PaymentMethodStats `json:"method_stats"`
}

type RevenueAnalyticsResponse struct {
	Period        string                      `json:"period"`
	StartDate     time.Time                   `json:"start_date"`
	EndDate       time.Time                   `json:"end_date"`
	Metrics       *repository.RevenueMetrics  `json:"metrics"`
	DailyStats    []repository.DailyPaymentStats `json:"daily_stats"`
}

// Webhook response types
type WebhookResponse struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Processed bool   `json:"processed"`
	Message   string `json:"message,omitempty"`
}