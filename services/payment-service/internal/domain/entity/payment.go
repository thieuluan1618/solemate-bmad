package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSucceeded  PaymentStatus = "succeeded"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCanceled   PaymentStatus = "canceled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

type PaymentMethodType string

const (
	PaymentMethodCard   PaymentMethodType = "card"
	PaymentMethodBank   PaymentMethodType = "bank_transfer"
	PaymentMethodWallet PaymentMethodType = "digital_wallet"
)

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCanceled  RefundStatus = "canceled"
)

// Payment represents a payment transaction
type Payment struct {
	ID                uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID     `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderID           uuid.UUID     `json:"order_id" gorm:"type:uuid;not null;index"`
	PaymentMethodID   *uuid.UUID    `json:"payment_method_id" gorm:"type:uuid;index"`

	// Payment details
	Amount            float64       `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency          string        `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`
	Status            PaymentStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`

	// Stripe integration
	StripePaymentIntentID string    `json:"stripe_payment_intent_id" gorm:"type:varchar(255);unique;index"`
	StripeChargeID       string     `json:"stripe_charge_id" gorm:"type:varchar(255);index"`
	ClientSecret         string     `json:"client_secret" gorm:"type:varchar(255)"`

	// Payment metadata
	Description          string     `json:"description" gorm:"type:text"`
	FailureReason        string     `json:"failure_reason" gorm:"type:text"`
	FailureCode          string     `json:"failure_code" gorm:"type:varchar(100)"`

	// Timestamps
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ProcessedAt          *time.Time `json:"processed_at"`
	FailedAt             *time.Time `json:"failed_at"`

	// Relationships
	PaymentMethod        *PaymentMethod `json:"payment_method,omitempty" gorm:"foreignKey:PaymentMethodID"`
	Refunds              []Refund       `json:"refunds,omitempty" gorm:"foreignKey:PaymentID"`
}

// PaymentMethod represents a stored payment method
type PaymentMethod struct {
	ID            uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID         `json:"user_id" gorm:"type:uuid;not null;index"`
	Type          PaymentMethodType `json:"type" gorm:"type:varchar(20);not null"`

	// Stripe integration
	StripePaymentMethodID string `json:"stripe_payment_method_id" gorm:"type:varchar(255);unique;not null"`

	// Card details (masked)
	CardBrand     string `json:"card_brand" gorm:"type:varchar(20)"`
	CardLast4     string `json:"card_last4" gorm:"type:varchar(4)"`
	CardExpMonth  int    `json:"card_exp_month"`
	CardExpYear   int    `json:"card_exp_year"`

	// Billing address
	BillingAddress Address `json:"billing_address" gorm:"embedded;embeddedPrefix:billing_"`

	// Metadata
	IsDefault     bool      `json:"is_default" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Refund represents a payment refund
type Refund struct {
	ID                uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PaymentID         uuid.UUID    `json:"payment_id" gorm:"type:uuid;not null;index"`
	UserID            uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	OrderID           uuid.UUID    `json:"order_id" gorm:"type:uuid;not null;index"`

	// Refund details
	Amount            float64      `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency          string       `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`
	Status            RefundStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`

	// Stripe integration
	StripeRefundID    string       `json:"stripe_refund_id" gorm:"type:varchar(255);unique;index"`

	// Refund metadata
	Reason            string       `json:"reason" gorm:"type:varchar(100)"`
	Description       string       `json:"description" gorm:"type:text"`
	FailureReason     string       `json:"failure_reason" gorm:"type:text"`

	// Timestamps
	CreatedAt         time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	ProcessedAt       *time.Time   `json:"processed_at"`

	// Relationships
	Payment           *Payment     `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
}

// WebhookEvent represents a Stripe webhook event
type WebhookEvent struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StripeEventID   string    `json:"stripe_event_id" gorm:"type:varchar(255);unique;not null;index"`
	EventType       string    `json:"event_type" gorm:"type:varchar(100);not null"`
	PaymentID       *uuid.UUID `json:"payment_id" gorm:"type:uuid;index"`
	RefundID        *uuid.UUID `json:"refund_id" gorm:"type:uuid;index"`

	// Event data
	Data            string    `json:"data" gorm:"type:text"` // JSON data from Stripe
	Processed       bool      `json:"processed" gorm:"default:false"`
	ProcessingError string    `json:"processing_error" gorm:"type:text"`

	// Timestamps
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	ProcessedAt     *time.Time `json:"processed_at"`
}

// Address represents a billing address
type Address struct {
	FirstName    string `json:"first_name" gorm:"type:varchar(100)"`
	LastName     string `json:"last_name" gorm:"type:varchar(100)"`
	Company      string `json:"company" gorm:"type:varchar(100)"`
	AddressLine1 string `json:"address_line_1" gorm:"type:varchar(255)"`
	AddressLine2 string `json:"address_line_2" gorm:"type:varchar(255)"`
	City         string `json:"city" gorm:"type:varchar(100)"`
	State        string `json:"state" gorm:"type:varchar(100)"`
	PostalCode   string `json:"postal_code" gorm:"type:varchar(20)"`
	Country      string `json:"country" gorm:"type:varchar(100);default:'US'"`
}

// PaymentSummary for API responses
type PaymentSummary struct {
	ID          uuid.UUID     `json:"id"`
	OrderID     uuid.UUID     `json:"order_id"`
	Amount      float64       `json:"amount"`
	Currency    string        `json:"currency"`
	Status      PaymentStatus `json:"status"`
	CardLast4   string        `json:"card_last4,omitempty"`
	CardBrand   string        `json:"card_brand,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	ProcessedAt *time.Time    `json:"processed_at"`
}

// Payment business logic methods
func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusSucceeded
}

func (p *Payment) CanBeCanceled() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusProcessing
}

func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusSucceeded
}

func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

func (p *Payment) GetTotalRefunded() float64 {
	var total float64
	for _, refund := range p.Refunds {
		if refund.Status == RefundStatusSucceeded {
			total += refund.Amount
		}
	}
	return total
}

func (p *Payment) GetRefundableAmount() float64 {
	return p.Amount - p.GetTotalRefunded()
}

func (p *Payment) MarkAsSucceeded() {
	now := time.Now()
	p.Status = PaymentStatusSucceeded
	p.ProcessedAt = &now
	p.UpdatedAt = now
}

func (p *Payment) MarkAsFailed(reason, code string) {
	now := time.Now()
	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	p.FailureCode = code
	p.FailedAt = &now
	p.UpdatedAt = now
}

func (p *Payment) MarkAsProcessing() {
	p.Status = PaymentStatusProcessing
	p.UpdatedAt = time.Now()
}

// PaymentMethod business logic
func (pm *PaymentMethod) IsExpired() bool {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	if pm.CardExpYear < currentYear {
		return true
	}
	if pm.CardExpYear == currentYear && pm.CardExpMonth < currentMonth {
		return true
	}
	return false
}

func (pm *PaymentMethod) GetDisplayName() string {
	if pm.CardBrand != "" && pm.CardLast4 != "" {
		return fmt.Sprintf("%s ****%s", pm.CardBrand, pm.CardLast4)
	}
	return string(pm.Type)
}

// Refund business logic
func (r *Refund) MarkAsSucceeded() {
	now := time.Now()
	r.Status = RefundStatusSucceeded
	r.ProcessedAt = &now
	r.UpdatedAt = now
}

func (r *Refund) MarkAsFailed(reason string) {
	r.Status = RefundStatusFailed
	r.FailureReason = reason
	r.UpdatedAt = time.Now()
}

// Custom errors
type PaymentError struct {
	Message string
	Code    string
}

func (e PaymentError) Error() string {
	return e.Message
}

var (
	ErrPaymentNotFound        = PaymentError{Message: "payment not found", Code: "payment_not_found"}
	ErrPaymentMethodNotFound  = PaymentError{Message: "payment method not found", Code: "payment_method_not_found"}
	ErrInsufficientAmount     = PaymentError{Message: "insufficient amount for refund", Code: "insufficient_amount"}
	ErrPaymentNotRefundable   = PaymentError{Message: "payment is not refundable", Code: "payment_not_refundable"}
	ErrPaymentNotCancelable   = PaymentError{Message: "payment cannot be canceled", Code: "payment_not_cancelable"}
	ErrPaymentAlreadyProcessed = PaymentError{Message: "payment already processed", Code: "payment_already_processed"}
	ErrInvalidPaymentMethod   = PaymentError{Message: "invalid payment method", Code: "invalid_payment_method"}
	ErrPaymentMethodExpired   = PaymentError{Message: "payment method has expired", Code: "payment_method_expired"}
)