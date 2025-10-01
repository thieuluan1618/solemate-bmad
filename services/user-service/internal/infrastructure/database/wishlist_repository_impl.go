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

type wishlistRepositoryImpl struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) repository.WishlistRepository {
	return &wishlistRepositoryImpl{db: db}
}

func (r *wishlistRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WishlistItem, error) {
	var items []*entity.WishlistItem

	// Preload product details and order by most recently added
	result := r.db.WithContext(ctx).
		Preload("Product").
		Where("user_id = ?", userID).
		Order("added_at DESC").
		Find(&items)

	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

func (r *wishlistRepositoryImpl) AddItem(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error) {
	// Check if item already exists
	exists, err := r.ItemExists(ctx, userID, productID)
	if err != nil {
		return nil, err
	}

	if exists {
		// If already exists, return the existing item
		return r.GetItemByProductID(ctx, userID, productID)
	}

	// Create new wishlist item
	item := &entity.WishlistItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		AddedAt:   time.Now(),
	}

	result := r.db.WithContext(ctx).Create(item)
	if result.Error != nil {
		return nil, result.Error
	}

	// Load product details before returning
	r.db.WithContext(ctx).Preload("Product").First(item, item.ID)

	return item, nil
}

func (r *wishlistRepositoryImpl) RemoveItem(ctx context.Context, userID, productID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&entity.WishlistItem{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("wishlist item not found")
	}

	return nil
}

func (r *wishlistRepositoryImpl) ClearWishlist(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.WishlistItem{})

	return result.Error
}

func (r *wishlistRepositoryImpl) ItemExists(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&entity.WishlistItem{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

func (r *wishlistRepositoryImpl) GetItemByProductID(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error) {
	var item entity.WishlistItem
	result := r.db.WithContext(ctx).
		Preload("Product").
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&item)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("wishlist item not found")
		}
		return nil, result.Error
	}

	return &item, nil
}
