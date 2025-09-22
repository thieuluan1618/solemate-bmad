package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string
type NotificationChannel string
type NotificationStatus string
type NotificationPriority string

const (
	NotificationTypeOrderCreated      NotificationType = "order_created"
	NotificationTypeOrderConfirmed    NotificationType = "order_confirmed"
	NotificationTypeOrderShipped      NotificationType = "order_shipped"
	NotificationTypeOrderDelivered    NotificationType = "order_delivered"
	NotificationTypeOrderCancelled    NotificationType = "order_cancelled"
	NotificationTypePaymentSuccessful NotificationType = "payment_successful"
	NotificationTypePaymentFailed     NotificationType = "payment_failed"
	NotificationTypeStockAlert        NotificationType = "stock_alert"
	NotificationTypeWelcome           NotificationType = "welcome"
	NotificationTypePasswordReset     NotificationType = "password_reset"
	NotificationTypePromotion         NotificationType = "promotion"
	NotificationTypeNewsletter        NotificationType = "newsletter"
)

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
	ChannelInApp NotificationChannel = "in_app"
)

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRetrying  NotificationStatus = "retrying"
	StatusCancelled NotificationStatus = "cancelled"
)

const (
	PriorityLow      NotificationPriority = "low"
	PriorityMedium   NotificationPriority = "medium"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

type Notification struct {
	ID                uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID            `json:"user_id" gorm:"type:uuid;not null;index"`
	Type              NotificationType     `json:"type" gorm:"type:varchar(50);not null;index"`
	Channel           NotificationChannel  `json:"channel" gorm:"type:varchar(20);not null"`
	Status            NotificationStatus   `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Priority          NotificationPriority `json:"priority" gorm:"type:varchar(20);not null;default:'medium'"`
	Subject           string               `json:"subject" gorm:"type:text;not null"`
	Content           string               `json:"content" gorm:"type:text;not null"`
	HTMLContent       *string              `json:"html_content" gorm:"type:text"`
	RecipientEmail    *string              `json:"recipient_email" gorm:"type:varchar(255);index"`
	RecipientPhone    *string              `json:"recipient_phone" gorm:"type:varchar(20);index"`
	SenderEmail       *string              `json:"sender_email" gorm:"type:varchar(255)"`
	SenderName        *string              `json:"sender_name" gorm:"type:varchar(255)"`
	TemplateID        *string              `json:"template_id" gorm:"type:varchar(100);index"`
	TemplateData      map[string]interface{} `json:"template_data" gorm:"type:jsonb"`
	Metadata          map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	RelatedEntityID   *uuid.UUID           `json:"related_entity_id" gorm:"type:uuid;index"`
	RelatedEntityType *string              `json:"related_entity_type" gorm:"type:varchar(50);index"`
	ScheduledAt       *time.Time           `json:"scheduled_at" gorm:"index"`
	SentAt            *time.Time           `json:"sent_at" gorm:"index"`
	DeliveredAt       *time.Time           `json:"delivered_at"`
	FailedAt          *time.Time           `json:"failed_at"`
	RetryCount        int                  `json:"retry_count" gorm:"default:0"`
	MaxRetries        int                  `json:"max_retries" gorm:"default:3"`
	ErrorMessage      *string              `json:"error_message" gorm:"type:text"`
	ExternalID        *string              `json:"external_id" gorm:"type:varchar(255);index"`
	CreatedAt         time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
}

type NotificationTemplate struct {
	ID           uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string               `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Type         NotificationType     `json:"type" gorm:"type:varchar(50);not null;index"`
	Channel      NotificationChannel  `json:"channel" gorm:"type:varchar(20);not null"`
	Subject      string               `json:"subject" gorm:"type:text;not null"`
	Content      string               `json:"content" gorm:"type:text;not null"`
	HTMLContent  *string              `json:"html_content" gorm:"type:text"`
	Variables    []string             `json:"variables" gorm:"type:jsonb"`
	IsActive     bool                 `json:"is_active" gorm:"default:true"`
	Version      int                  `json:"version" gorm:"default:1"`
	Description  *string              `json:"description" gorm:"type:text"`
	SenderEmail  *string              `json:"sender_email" gorm:"type:varchar(255)"`
	SenderName   *string              `json:"sender_name" gorm:"type:varchar(255)"`
	CreatedAt    time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
}

type NotificationPreference struct {
	ID                    uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID                uuid.UUID            `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	EmailNotifications    bool                 `json:"email_notifications" gorm:"default:true"`
	SMSNotifications      bool                 `json:"sms_notifications" gorm:"default:false"`
	PushNotifications     bool                 `json:"push_notifications" gorm:"default:true"`
	InAppNotifications    bool                 `json:"in_app_notifications" gorm:"default:true"`
	OrderUpdates          bool                 `json:"order_updates" gorm:"default:true"`
	PaymentUpdates        bool                 `json:"payment_updates" gorm:"default:true"`
	PromotionalEmails     bool                 `json:"promotional_emails" gorm:"default:false"`
	Newsletter            bool                 `json:"newsletter" gorm:"default:false"`
	SecurityAlerts        bool                 `json:"security_alerts" gorm:"default:true"`
	StockAlerts           bool                 `json:"stock_alerts" gorm:"default:false"`
	PreferredChannel      NotificationChannel  `json:"preferred_channel" gorm:"type:varchar(20);default:'email'"`
	TimeZone              string               `json:"time_zone" gorm:"type:varchar(50);default:'UTC'"`
	QuietHoursStart       *string              `json:"quiet_hours_start" gorm:"type:varchar(5)"`
	QuietHoursEnd         *string              `json:"quiet_hours_end" gorm:"type:varchar(5)"`
	CreatedAt             time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
}

type NotificationLog struct {
	ID               uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NotificationID   uuid.UUID            `json:"notification_id" gorm:"type:uuid;not null;index"`
	Status           NotificationStatus   `json:"status" gorm:"type:varchar(20);not null"`
	AttemptNumber    int                  `json:"attempt_number" gorm:"not null"`
	ProviderResponse *string              `json:"provider_response" gorm:"type:text"`
	ErrorMessage     *string              `json:"error_message" gorm:"type:text"`
	DeliveryTime     *time.Duration       `json:"delivery_time"`
	CreatedAt        time.Time            `json:"created_at" gorm:"autoCreateTime"`

	Notification     Notification         `json:"notification" gorm:"foreignKey:NotificationID"`
}

type NotificationProvider struct {
	ID          uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string               `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	Channel     NotificationChannel  `json:"channel" gorm:"type:varchar(20);not null"`
	IsActive    bool                 `json:"is_active" gorm:"default:true"`
	Priority    int                  `json:"priority" gorm:"default:1"`
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	RateLimit   *int                 `json:"rate_limit"`
	CreatedAt   time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
}

type NotificationQueue struct {
	ID             uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NotificationID uuid.UUID            `json:"notification_id" gorm:"type:uuid;not null;index"`
	Priority       NotificationPriority `json:"priority" gorm:"type:varchar(20);not null;index"`
	ScheduledAt    time.Time            `json:"scheduled_at" gorm:"not null;index"`
	ProcessedAt    *time.Time           `json:"processed_at"`
	IsProcessed    bool                 `json:"is_processed" gorm:"default:false;index"`
	RetryAfter     *time.Time           `json:"retry_after"`
	CreatedAt      time.Time            `json:"created_at" gorm:"autoCreateTime"`

	Notification   Notification         `json:"notification" gorm:"foreignKey:NotificationID"`
}

type NotificationEvent struct {
	ID         uuid.UUID                `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EventType  string                   `json:"event_type" gorm:"type:varchar(100);not null;index"`
	EntityID   uuid.UUID                `json:"entity_id" gorm:"type:uuid;not null;index"`
	EntityType string                   `json:"entity_type" gorm:"type:varchar(50);not null;index"`
	UserID     *uuid.UUID               `json:"user_id" gorm:"type:uuid;index"`
	Payload    map[string]interface{}   `json:"payload" gorm:"type:jsonb"`
	Processed  bool                     `json:"processed" gorm:"default:false;index"`
	CreatedAt  time.Time                `json:"created_at" gorm:"autoCreateTime"`
}