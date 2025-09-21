package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/product-service/internal/domain/entity"
	"solemate/services/product-service/internal/domain/repository"
)

type brandRepositoryImpl struct {
	db *gorm.DB
}

func NewBrandRepository(db *gorm.DB) repository.BrandRepository {
	return &brandRepositoryImpl{db: db}
}

func (r *brandRepositoryImpl) Create(ctx context.Context, brand *entity.Brand) error {
	brand.ID = uuid.New()
	brand.CreatedAt = time.Now()

	result := r.db.WithContext(ctx).Create(brand)
	return result.Error
}

func (r *brandRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error) {
	var brand entity.Brand
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&brand)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("brand not found")
		}
		return nil, result.Error
	}
	return &brand, nil
}

func (r *brandRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entity.Brand, error) {
	var brand entity.Brand
	result := r.db.WithContext(ctx).Where("slug = ?", slug).First(&brand)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("brand not found")
		}
		return nil, result.Error
	}
	return &brand, nil
}

func (r *brandRepositoryImpl) Update(ctx context.Context, brand *entity.Brand) error {
	result := r.db.WithContext(ctx).Save(brand)
	return result.Error
}

func (r *brandRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Brand{}, id)
	if result.RowsAffected == 0 {
		return errors.New("brand not found")
	}
	return result.Error
}

func (r *brandRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entity.Brand, int64, error) {
	var brands []*entity.Brand
	var total int64

	countQuery := r.db.WithContext(ctx).Model(&entity.Brand{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.WithContext(ctx).Order("name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	result := query.Find(&brands)
	return brands, total, result.Error
}