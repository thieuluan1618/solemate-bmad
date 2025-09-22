package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/payment-service/internal/domain/entity"
	"solemate/services/payment-service/internal/domain/repository"
)

type webhookRepositoryImpl struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) repository.WebhookRepository {
	return &webhookRepositoryImpl{db: db}
}

// Webhook event operations
func (r *webhookRepositoryImpl) CreateWebhookEvent(ctx context.Context, event *entity.WebhookEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *webhookRepositoryImpl) GetWebhookEventByStripeID(ctx context.Context, stripeEventID string) (*entity.WebhookEvent, error) {
	var event entity.WebhookEvent
	err := r.db.WithContext(ctx).First(&event, "stripe_event_id = ?", stripeEventID).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *webhookRepositoryImpl) UpdateWebhookEvent(ctx context.Context, event *entity.WebhookEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *webhookRepositoryImpl) GetUnprocessedWebhookEvents(ctx context.Context, limit int) ([]*entity.WebhookEvent, error) {
	var events []*entity.WebhookEvent
	err := r.db.WithContext(ctx).
		Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *webhookRepositoryImpl) MarkWebhookEventAsProcessed(ctx context.Context, eventID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.WebhookEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"processed":     true,
			"processed_at":  &now,
			"processing_error": "",
		}).Error
}

func (r *webhookRepositoryImpl) MarkWebhookEventAsFailed(ctx context.Context, eventID uuid.UUID, errorMessage string) error {
	return r.db.WithContext(ctx).Model(&entity.WebhookEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"processed":        false,
			"processing_error": errorMessage,
		}).Error
}