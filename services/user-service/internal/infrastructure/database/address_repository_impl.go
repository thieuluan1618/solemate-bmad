package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/user-service/internal/domain/entity"
	"solemate/services/user-service/internal/domain/repository"
)

type addressRepositoryImpl struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) repository.AddressRepository {
	return &addressRepositoryImpl{db: db}
}

func (r *addressRepositoryImpl) Create(ctx context.Context, address *entity.Address) error {
	address.ID = uuid.New()
	address.CreatedAt = time.Now()

	result := r.db.WithContext(ctx).Create(address)
	return result.Error
}

func (r *addressRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Address, error) {
	var address entity.Address
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("address not found")
		}
		return nil, result.Error
	}
	return &address, nil
}

func (r *addressRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Address, error) {
	var addresses []*entity.Address
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&addresses)
	return addresses, result.Error
}

func (r *addressRepositoryImpl) Update(ctx context.Context, address *entity.Address) error {
	result := r.db.WithContext(ctx).Save(address)
	return result.Error
}

func (r *addressRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Address{}, id)
	if result.RowsAffected == 0 {
		return errors.New("address not found")
	}
	return result.Error
}

func (r *addressRepositoryImpl) SetDefault(ctx context.Context, userID, addressID uuid.UUID) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Set all addresses to non-default
	if err := tx.Model(&entity.Address{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set the specified address as default
	if err := tx.Model(&entity.Address{}).Where("id = ? AND user_id = ?", addressID, userID).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
