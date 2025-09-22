package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"solemate/services/notification-service/internal/domain/entity"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetPendingNotifications(ctx context.Context, limit int) ([]*entity.Notification, error)
	GetScheduledNotifications(ctx context.Context, before time.Time, limit int) ([]*entity.Notification, error)
	GetByStatus(ctx context.Context, status entity.NotificationStatus, limit, offset int) ([]*entity.Notification, error)
	GetByType(ctx context.Context, notificationType entity.NotificationType, limit, offset int) ([]*entity.Notification, error)
	GetByChannel(ctx context.Context, channel entity.NotificationChannel, limit, offset int) ([]*entity.Notification, error)
	GetByRelatedEntity(ctx context.Context, entityID uuid.UUID, entityType string) ([]*entity.Notification, error)
	Update(ctx context.Context, notification *entity.Notification) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.NotificationStatus) error
	MarkAsSent(ctx context.Context, id uuid.UUID, sentAt time.Time, externalID *string) error
	MarkAsDelivered(ctx context.Context, id uuid.UUID, deliveredAt time.Time) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, failedAt time.Time, errorMessage string) error
	IncrementRetryCount(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOldNotifications(ctx context.Context, olderThan time.Time) error
	GetStatistics(ctx context.Context, from, to time.Time) (*NotificationStatistics, error)
	GetDeliveryReport(ctx context.Context, from, to time.Time, groupBy string) ([]*DeliveryReport, error)
}

type TemplateRepository interface {
	Create(ctx context.Context, template *entity.NotificationTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationTemplate, error)
	GetByName(ctx context.Context, name string) (*entity.NotificationTemplate, error)
	GetByTypeAndChannel(ctx context.Context, notificationType entity.NotificationType, channel entity.NotificationChannel) (*entity.NotificationTemplate, error)
	GetActive(ctx context.Context) ([]*entity.NotificationTemplate, error)
	GetByType(ctx context.Context, notificationType entity.NotificationType) ([]*entity.NotificationTemplate, error)
	Update(ctx context.Context, template *entity.NotificationTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.NotificationTemplate, error)
}

type PreferenceRepository interface {
	Create(ctx context.Context, preference *entity.NotificationPreference) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreference, error)
	Update(ctx context.Context, preference *entity.NotificationPreference) error
	Delete(ctx context.Context, userID uuid.UUID) error
	GetUsersWithPreference(ctx context.Context, channel entity.NotificationChannel, notificationType entity.NotificationType) ([]uuid.UUID, error)
	BulkGetPreferences(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*entity.NotificationPreference, error)
}

type LogRepository interface {
	Create(ctx context.Context, log *entity.NotificationLog) error
	GetByNotificationID(ctx context.Context, notificationID uuid.UUID) ([]*entity.NotificationLog, error)
	GetByStatus(ctx context.Context, status entity.NotificationStatus, limit, offset int) ([]*entity.NotificationLog, error)
	GetDeliveryStats(ctx context.Context, from, to time.Time) (*DeliveryStats, error)
	DeleteOldLogs(ctx context.Context, olderThan time.Time) error
}

type ProviderRepository interface {
	Create(ctx context.Context, provider *entity.NotificationProvider) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationProvider, error)
	GetByChannel(ctx context.Context, channel entity.NotificationChannel) ([]*entity.NotificationProvider, error)
	GetActive(ctx context.Context) ([]*entity.NotificationProvider, error)
	Update(ctx context.Context, provider *entity.NotificationProvider) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.NotificationProvider, error)
}

type QueueRepository interface {
	Enqueue(ctx context.Context, queueItem *entity.NotificationQueue) error
	Dequeue(ctx context.Context, priority entity.NotificationPriority, limit int) ([]*entity.NotificationQueue, error)
	GetPending(ctx context.Context, before time.Time, limit int) ([]*entity.NotificationQueue, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID, processedAt time.Time) error
	Reschedule(ctx context.Context, id uuid.UUID, retryAfter time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetQueueStats(ctx context.Context) (*QueueStats, error)
	CleanupProcessed(ctx context.Context, olderThan time.Time) error
}

type EventRepository interface {
	Create(ctx context.Context, event *entity.NotificationEvent) error
	GetUnprocessed(ctx context.Context, limit int) ([]*entity.NotificationEvent, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
	GetByEntityID(ctx context.Context, entityID uuid.UUID, entityType string) ([]*entity.NotificationEvent, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOldEvents(ctx context.Context, olderThan time.Time) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]*User, error)
	ValidateUserExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

type NotificationStatistics struct {
	TotalSent      int64   `json:"total_sent"`
	TotalDelivered int64   `json:"total_delivered"`
	TotalFailed    int64   `json:"total_failed"`
	TotalPending   int64   `json:"total_pending"`
	DeliveryRate   float64 `json:"delivery_rate"`
	FailureRate    float64 `json:"failure_rate"`
	ByChannel      map[entity.NotificationChannel]ChannelStats `json:"by_channel"`
	ByType         map[entity.NotificationType]TypeStats `json:"by_type"`
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

type DeliveryReport struct {
	Period         string  `json:"period"`
	TotalSent      int64   `json:"total_sent"`
	TotalDelivered int64   `json:"total_delivered"`
	TotalFailed    int64   `json:"total_failed"`
	DeliveryRate   float64 `json:"delivery_rate"`
}

type DeliveryStats struct {
	AverageDeliveryTime time.Duration                           `json:"average_delivery_time"`
	SuccessRate         float64                                 `json:"success_rate"`
	ByChannel           map[entity.NotificationChannel]float64 `json:"by_channel"`
	ByPriority          map[entity.NotificationPriority]float64 `json:"by_priority"`
}

type QueueStats struct {
	PendingCount    int64                                         `json:"pending_count"`
	ProcessedCount  int64                                         `json:"processed_count"`
	ByPriority      map[entity.NotificationPriority]int64         `json:"by_priority"`
	OldestPending   *time.Time                                    `json:"oldest_pending"`
	AverageWaitTime time.Duration                                 `json:"average_wait_time"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Phone    *string   `json:"phone"`
	FullName string    `json:"full_name"`
	Language string    `json:"language"`
	TimeZone string    `json:"time_zone"`
	IsActive bool      `json:"is_active"`
}