package service

import (
	"time"

	"github.com/google/uuid"
	"solemate/services/notification-service/internal/domain/entity"
)

type SendNotificationRequest struct {
	UserID            uuid.UUID                      `json:"user_id" validate:"required"`
	Type              entity.NotificationType        `json:"type" validate:"required"`
	Channel           entity.NotificationChannel     `json:"channel" validate:"required"`
	Priority          entity.NotificationPriority    `json:"priority"`
	Subject           string                         `json:"subject" validate:"required"`
	Content           string                         `json:"content" validate:"required"`
	HTMLContent       *string                        `json:"html_content"`
	RecipientEmail    *string                        `json:"recipient_email"`
	RecipientPhone    *string                        `json:"recipient_phone"`
	TemplateID        *string                        `json:"template_id"`
	TemplateData      map[string]interface{}         `json:"template_data"`
	Metadata          map[string]interface{}         `json:"metadata"`
	RelatedEntityID   *uuid.UUID                     `json:"related_entity_id"`
	RelatedEntityType *string                        `json:"related_entity_type"`
	ScheduledAt       *time.Time                     `json:"scheduled_at"`
}

type SendBulkNotificationRequest struct {
	UserIDs           []uuid.UUID                    `json:"user_ids" validate:"required,min=1"`
	Type              entity.NotificationType        `json:"type" validate:"required"`
	Channel           entity.NotificationChannel     `json:"channel" validate:"required"`
	Priority          entity.NotificationPriority    `json:"priority"`
	Subject           string                         `json:"subject" validate:"required"`
	Content           string                         `json:"content" validate:"required"`
	HTMLContent       *string                        `json:"html_content"`
	TemplateID        *string                        `json:"template_id"`
	TemplateData      map[string]interface{}         `json:"template_data"`
	Metadata          map[string]interface{}         `json:"metadata"`
	ScheduledAt       *time.Time                     `json:"scheduled_at"`
}

type SendTemplateNotificationRequest struct {
	UserID            uuid.UUID                      `json:"user_id" validate:"required"`
	TemplateID        string                         `json:"template_id" validate:"required"`
	Channel           entity.NotificationChannel     `json:"channel" validate:"required"`
	TemplateData      map[string]interface{}         `json:"template_data"`
	RecipientEmail    *string                        `json:"recipient_email"`
	RecipientPhone    *string                        `json:"recipient_phone"`
	Priority          entity.NotificationPriority    `json:"priority"`
	Metadata          map[string]interface{}         `json:"metadata"`
	RelatedEntityID   *uuid.UUID                     `json:"related_entity_id"`
	RelatedEntityType *string                        `json:"related_entity_type"`
	ScheduledAt       *time.Time                     `json:"scheduled_at"`
}

type NotificationResponse struct {
	ID                uuid.UUID                      `json:"id"`
	UserID            uuid.UUID                      `json:"user_id"`
	Type              entity.NotificationType        `json:"type"`
	Channel           entity.NotificationChannel     `json:"channel"`
	Status            entity.NotificationStatus      `json:"status"`
	Priority          entity.NotificationPriority    `json:"priority"`
	Subject           string                         `json:"subject"`
	Content           string                         `json:"content"`
	RecipientEmail    *string                        `json:"recipient_email"`
	RecipientPhone    *string                        `json:"recipient_phone"`
	RelatedEntityID   *uuid.UUID                     `json:"related_entity_id"`
	RelatedEntityType *string                        `json:"related_entity_type"`
	ScheduledAt       *time.Time                     `json:"scheduled_at"`
	SentAt            *time.Time                     `json:"sent_at"`
	DeliveredAt       *time.Time                     `json:"delivered_at"`
	FailedAt          *time.Time                     `json:"failed_at"`
	RetryCount        int                            `json:"retry_count"`
	ErrorMessage      *string                        `json:"error_message"`
	CreatedAt         time.Time                      `json:"created_at"`
	UpdatedAt         time.Time                      `json:"updated_at"`
}

type CreateTemplateRequest struct {
	Name         string                         `json:"name" validate:"required"`
	Type         entity.NotificationType        `json:"type" validate:"required"`
	Channel      entity.NotificationChannel     `json:"channel" validate:"required"`
	Subject      string                         `json:"subject" validate:"required"`
	Content      string                         `json:"content" validate:"required"`
	HTMLContent  *string                        `json:"html_content"`
	Variables    []string                       `json:"variables"`
	Description  *string                        `json:"description"`
	SenderEmail  *string                        `json:"sender_email"`
	SenderName   *string                        `json:"sender_name"`
}

type UpdateTemplateRequest struct {
	Name         *string                        `json:"name"`
	Subject      *string                        `json:"subject"`
	Content      *string                        `json:"content"`
	HTMLContent  *string                        `json:"html_content"`
	Variables    []string                       `json:"variables"`
	IsActive     *bool                          `json:"is_active"`
	Description  *string                        `json:"description"`
	SenderEmail  *string                        `json:"sender_email"`
	SenderName   *string                        `json:"sender_name"`
}

type TemplateResponse struct {
	ID          uuid.UUID                      `json:"id"`
	Name        string                         `json:"name"`
	Type        entity.NotificationType        `json:"type"`
	Channel     entity.NotificationChannel     `json:"channel"`
	Subject     string                         `json:"subject"`
	Content     string                         `json:"content"`
	HTMLContent *string                        `json:"html_content"`
	Variables   []string                       `json:"variables"`
	IsActive    bool                           `json:"is_active"`
	Version     int                            `json:"version"`
	Description *string                        `json:"description"`
	SenderEmail *string                        `json:"sender_email"`
	SenderName  *string                        `json:"sender_name"`
	CreatedAt   time.Time                      `json:"created_at"`
	UpdatedAt   time.Time                      `json:"updated_at"`
}

type UpdatePreferenceRequest struct {
	EmailNotifications    *bool                          `json:"email_notifications"`
	SMSNotifications      *bool                          `json:"sms_notifications"`
	PushNotifications     *bool                          `json:"push_notifications"`
	InAppNotifications    *bool                          `json:"in_app_notifications"`
	OrderUpdates          *bool                          `json:"order_updates"`
	PaymentUpdates        *bool                          `json:"payment_updates"`
	PromotionalEmails     *bool                          `json:"promotional_emails"`
	Newsletter            *bool                          `json:"newsletter"`
	SecurityAlerts        *bool                          `json:"security_alerts"`
	StockAlerts           *bool                          `json:"stock_alerts"`
	PreferredChannel      *entity.NotificationChannel    `json:"preferred_channel"`
	TimeZone              *string                        `json:"time_zone"`
	QuietHoursStart       *string                        `json:"quiet_hours_start"`
	QuietHoursEnd         *string                        `json:"quiet_hours_end"`
}

type PreferenceResponse struct {
	ID                    uuid.UUID                      `json:"id"`
	UserID                uuid.UUID                      `json:"user_id"`
	EmailNotifications    bool                           `json:"email_notifications"`
	SMSNotifications      bool                           `json:"sms_notifications"`
	PushNotifications     bool                           `json:"push_notifications"`
	InAppNotifications    bool                           `json:"in_app_notifications"`
	OrderUpdates          bool                           `json:"order_updates"`
	PaymentUpdates        bool                           `json:"payment_updates"`
	PromotionalEmails     bool                           `json:"promotional_emails"`
	Newsletter            bool                           `json:"newsletter"`
	SecurityAlerts        bool                           `json:"security_alerts"`
	StockAlerts           bool                           `json:"stock_alerts"`
	PreferredChannel      entity.NotificationChannel     `json:"preferred_channel"`
	TimeZone              string                         `json:"time_zone"`
	QuietHoursStart       *string                        `json:"quiet_hours_start"`
	QuietHoursEnd         *string                        `json:"quiet_hours_end"`
	CreatedAt             time.Time                      `json:"created_at"`
	UpdatedAt             time.Time                      `json:"updated_at"`
}

type NotificationListResponse struct {
	Notifications []*NotificationResponse            `json:"notifications"`
	Total         int64                              `json:"total"`
	Limit         int                                `json:"limit"`
	Offset        int                                `json:"offset"`
}

type TemplateListResponse struct {
	Templates []*TemplateResponse                    `json:"templates"`
	Total     int64                                  `json:"total"`
	Limit     int                                    `json:"limit"`
	Offset    int                                    `json:"offset"`
}

type StatisticsResponse struct {
	TotalSent      int64                                          `json:"total_sent"`
	TotalDelivered int64                                          `json:"total_delivered"`
	TotalFailed    int64                                          `json:"total_failed"`
	TotalPending   int64                                          `json:"total_pending"`
	DeliveryRate   float64                                        `json:"delivery_rate"`
	FailureRate    float64                                        `json:"failure_rate"`
	ByChannel      map[entity.NotificationChannel]ChannelStats   `json:"by_channel"`
	ByType         map[entity.NotificationType]TypeStats         `json:"by_type"`
	Period         string                                         `json:"period"`
	From           time.Time                                      `json:"from"`
	To             time.Time                                      `json:"to"`
}

type ChannelStats struct {
	Sent      int64   `json:"sent"`
	Delivered int64   `json:"delivered"`
	Failed    int64   `json:"failed"`
	Rate      float64 `json:"rate"`
}

type TypeStats struct {
	Sent      int64   `json:"sent"`
	Delivered int64   `json:"delivered"`
	Failed    int64   `json:"failed"`
	Rate      float64 `json:"rate"`
}

type DeliveryReportResponse struct {
	Reports []*DeliveryReportItem                  `json:"reports"`
	Summary *DeliveryReportSummary                 `json:"summary"`
	Period  string                                 `json:"period"`
	From    time.Time                              `json:"from"`
	To      time.Time                              `json:"to"`
}

type DeliveryReportItem struct {
	Period         string  `json:"period"`
	TotalSent      int64   `json:"total_sent"`
	TotalDelivered int64   `json:"total_delivered"`
	TotalFailed    int64   `json:"total_failed"`
	DeliveryRate   float64 `json:"delivery_rate"`
}

type DeliveryReportSummary struct {
	TotalSent        int64   `json:"total_sent"`
	TotalDelivered   int64   `json:"total_delivered"`
	TotalFailed      int64   `json:"total_failed"`
	OverallRate      float64 `json:"overall_rate"`
	BestPeriod       string  `json:"best_period"`
	WorstPeriod      string  `json:"worst_period"`
	AverageRate      float64 `json:"average_rate"`
}

type EventProcessingRequest struct {
	EventType  string                         `json:"event_type" validate:"required"`
	EntityID   uuid.UUID                      `json:"entity_id" validate:"required"`
	EntityType string                         `json:"entity_type" validate:"required"`
	UserID     *uuid.UUID                     `json:"user_id"`
	Payload    map[string]interface{}         `json:"payload"`
}

type EventProcessingResponse struct {
	EventID              uuid.UUID `json:"event_id"`
	NotificationsCreated int       `json:"notifications_created"`
	ProcessedAt          time.Time `json:"processed_at"`
}

type RetryFailedRequest struct {
	NotificationIDs []uuid.UUID `json:"notification_ids"`
	MaxRetries      *int        `json:"max_retries"`
}

type RetryFailedResponse struct {
	Processed int       `json:"processed"`
	Failed    int       `json:"failed"`
	StartedAt time.Time `json:"started_at"`
}

type HealthResponse struct {
	Status           string                            `json:"status"`
	Service          string                            `json:"service"`
	Database         string                            `json:"database"`
	Redis            string                            `json:"redis"`
	EmailProvider    string                            `json:"email_provider"`
	SMSProvider      string                            `json:"sms_provider"`
	QueueStatus      map[string]interface{}            `json:"queue_status"`
	LastProcessed    *time.Time                        `json:"last_processed"`
	PendingCount     int64                             `json:"pending_count"`
	Timestamp        time.Time                         `json:"timestamp"`
}