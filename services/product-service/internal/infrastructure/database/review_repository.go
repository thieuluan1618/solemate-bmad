package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/product-service/internal/domain/entity"
	"solemate/services/product-service/internal/domain/repository"
)

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *entity.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *reviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	var review entity.Review
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("id = ?", id).
		First(&review).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("review not found")
		}
		return nil, err
	}

	return &review, nil
}

func (r *reviewRepository) GetByProductID(ctx context.Context, productID uuid.UUID, filters repository.ReviewFilters) ([]*entity.Review, int64, error) {
	var reviews []*entity.Review
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Review{}).Where("product_id = ?", productID)

	// Apply status filter if provided
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	} else {
		// By default, only show approved reviews
		query = query.Where("status = ?", "APPROVED")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filters.SortBy {
	case "highest_rated":
		query = query.Order("rating DESC, created_at DESC")
	case "lowest_rated":
		query = query.Order("rating ASC, created_at DESC")
	case "most_helpful":
		query = query.Order("helpful_count DESC, created_at DESC")
	default: // newest
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *entity.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

func (r *reviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Review{}, "id = ?", id).Error
}

func (r *reviewRepository) GetUserReviewForProduct(ctx context.Context, userID, productID uuid.UUID) (*entity.Review, error) {
	var review entity.Review
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&review).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not an error if review doesn't exist
		}
		return nil, err
	}

	return &review, nil
}
