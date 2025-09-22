package database

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/payment-service/internal/domain/entity"
	"solemate/services/payment-service/internal/domain/repository"
)

type paymentMethodRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) repository.PaymentMethodRepository {
	return &paymentMethodRepositoryImpl{db: db}
}

// Payment method CRUD operations
func (r *paymentMethodRepositoryImpl) CreatePaymentMethod(ctx context.Context, paymentMethod *entity.PaymentMethod) error {
	return r.db.WithContext(ctx).Create(paymentMethod).Error
}

func (r *paymentMethodRepositoryImpl) GetPaymentMethodByID(ctx context.Context, paymentMethodID uuid.UUID) (*entity.PaymentMethod, error) {
	var paymentMethod entity.PaymentMethod
	err := r.db.WithContext(ctx).First(&paymentMethod, "id = ?", paymentMethodID).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

func (r *paymentMethodRepositoryImpl) GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*entity.PaymentMethod, error) {
	var paymentMethod entity.PaymentMethod
	err := r.db.WithContext(ctx).First(&paymentMethod, "stripe_payment_method_id = ?", stripePaymentMethodID).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

func (r *paymentMethodRepositoryImpl) UpdatePaymentMethod(ctx context.Context, paymentMethod *entity.PaymentMethod) error {
	return r.db.WithContext(ctx).Save(paymentMethod).Error
}

func (r *paymentMethodRepositoryImpl) DeletePaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.PaymentMethod{}, "id = ?", paymentMethodID).Error
}

// Payment method querying
func (r *paymentMethodRepositoryImpl) GetPaymentMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.PaymentMethod, error) {
	var paymentMethods []*entity.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&paymentMethods).Error
	return paymentMethods, err
}

func (r *paymentMethodRepositoryImpl) GetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID) (*entity.PaymentMethod, error) {
	var paymentMethod entity.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", userID, true).
		First(&paymentMethod).Error
	if err != nil {
		return nil, err
	}
	return &paymentMethod, nil
}

func (r *paymentMethodRepositoryImpl) SetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, unset all default payment methods for the user
		if err := tx.Model(&entity.PaymentMethod{}).
			Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Then set the specified payment method as default
		return tx.Model(&entity.PaymentMethod{}).
			Where("id = ? AND user_id = ?", paymentMethodID, userID).
			Update("is_default", true).Error
	})
}

func (r *paymentMethodRepositoryImpl) UnsetDefaultPaymentMethods(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.PaymentMethod{}).
		Where("user_id = ?", userID).
		Update("is_default", false).Error
}