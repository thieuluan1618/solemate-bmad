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

type categoryRepositoryImpl struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) repository.CategoryRepository {
	return &categoryRepositoryImpl{db: db}
}

func (r *categoryRepositoryImpl) Create(ctx context.Context, category *entity.Category) error {
	category.ID = uuid.New()
	category.CreatedAt = time.Now()

	result := r.db.WithContext(ctx).Create(category)
	return result.Error
}

func (r *categoryRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var category entity.Category
	result := r.db.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		Where("id = ?", id).
		First(&category)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, result.Error
	}
	return &category, nil
}

func (r *categoryRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	var category entity.Category
	result := r.db.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		Where("slug = ?", slug).
		First(&category)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, result.Error
	}
	return &category, nil
}

func (r *categoryRepositoryImpl) Update(ctx context.Context, category *entity.Category) error {
	result := r.db.WithContext(ctx).Save(category)
	return result.Error
}

func (r *categoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Category{}, id)
	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}
	return result.Error
}

func (r *categoryRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entity.Category, int64, error) {
	var categories []*entity.Category
	var total int64

	countQuery := r.db.WithContext(ctx).Model(&entity.Category{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.WithContext(ctx).
		Preload("Parent").
		Order("sort_order ASC, name ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	result := query.Find(&categories)
	return categories, total, result.Error
}

func (r *categoryRepositoryImpl) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error) {
	var categories []*entity.Category
	result := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC, name ASC").
		Find(&categories)

	return categories, result.Error
}

func (r *categoryRepositoryImpl) GetTree(ctx context.Context) ([]*entity.Category, error) {
	var categories []*entity.Category

	// Get all categories with their children recursively
	result := r.db.WithContext(ctx).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, name ASC")
		}).
		Where("parent_id IS NULL").
		Order("sort_order ASC, name ASC").
		Find(&categories)

	return categories, result.Error
}