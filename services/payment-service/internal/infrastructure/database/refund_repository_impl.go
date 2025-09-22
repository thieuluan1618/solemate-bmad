package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/payment-service/internal/domain/entity"
	"solemate/services/payment-service/internal/domain/repository"
)

type refundRepositoryImpl struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) repository.RefundRepository {
	return &refundRepositoryImpl{db: db}
}

// Refund CRUD operations
func (r *refundRepositoryImpl) CreateRefund(ctx context.Context, refund *entity.Refund) error {
	return r.db.WithContext(ctx).Create(refund).Error
}

func (r *refundRepositoryImpl) GetRefundByID(ctx context.Context, refundID uuid.UUID) (*entity.Refund, error) {
	var refund entity.Refund
	err := r.db.WithContext(ctx).
		Preload("Payment").
		First(&refund, "id = ?", refundID).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepositoryImpl) GetRefundByStripeID(ctx context.Context, stripeRefundID string) (*entity.Refund, error) {
	var refund entity.Refund
	err := r.db.WithContext(ctx).
		Preload("Payment").
		First(&refund, "stripe_refund_id = ?", stripeRefundID).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepositoryImpl) UpdateRefund(ctx context.Context, refund *entity.Refund) error {
	return r.db.WithContext(ctx).Save(refund).Error
}

func (r *refundRepositoryImpl) DeleteRefund(ctx context.Context, refundID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Refund{}, "id = ?", refundID).Error
}

// Refund querying
func (r *refundRepositoryImpl) GetRefundsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*entity.Refund, error) {
	var refunds []*entity.Refund
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&refunds).Error
	return refunds, err
}

func (r *refundRepositoryImpl) GetRefundsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Refund, int64, error) {
	var refunds []*entity.Refund
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Refund{}).Where("user_id = ?", userID)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Payment").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error

	return refunds, count, err
}

func (r *refundRepositoryImpl) GetRefundsByStatus(ctx context.Context, status entity.RefundStatus, limit, offset int) ([]*entity.Refund, int64, error) {
	var refunds []*entity.Refund
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Refund{}).Where("status = ?", status)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Payment").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error

	return refunds, count, err
}

func (r *refundRepositoryImpl) GetRefundsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.Refund, int64, error) {
	var refunds []*entity.Refund
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Refund{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Payment").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error

	return refunds, count, err
}