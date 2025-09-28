package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/product-service/internal/domain/entity"
	"solemate/services/product-service/internal/domain/repository"
)

type productRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) repository.ProductRepository {
	return &productRepositoryImpl{db: db}
}

func (r *productRepositoryImpl) Create(ctx context.Context, product *entity.Product) error {
	product.ID = uuid.New()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Create(product)
	return result.Error
}

func (r *productRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var product entity.Product
	result := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Brand").
		Preload("Variants").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, created_at ASC")
		}).
		Where("id = ? ", id).
		First(&product)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, result.Error
	}
	r.calculateTotalStock(&product)
	return &product, nil
}

func (r *productRepositoryImpl) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var product entity.Product
	result := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Brand").
		Preload("Variants").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, created_at ASC")
		}).
		Where("sku = ? ", sku).
		First(&product)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, result.Error
	}
	r.calculateTotalStock(&product)
	return &product, nil
}

func (r *productRepositoryImpl) GetBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	var product entity.Product
	result := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Brand").
		Preload("Variants").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, created_at ASC")
		}).
		Where("slug = ? ", slug).
		First(&product)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, result.Error
	}
	r.calculateTotalStock(&product)
	return &product, nil
}

func (r *productRepositoryImpl) Update(ctx context.Context, product *entity.Product) error {
	product.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).Save(product)
	return result.Error
}

func (r *productRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.Product{}, id)
	if result.RowsAffected == 0 {
		return errors.New("product not found")
	}
	return result.Error
}

func (r *productRepositoryImpl) List(ctx context.Context, filters repository.ProductFilters) ([]*entity.Product, int64, error) {
	var products []*entity.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Product{})

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = r.applySorting(query, filters)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Preload relationships
	query = query.
		Preload("Category").
		Preload("Brand").
		Preload("Variants"). // Add this line
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true)
		})

	result := query.Find(&products)

	// Calculate total stock for each product
	for _, product := range products {
		r.calculateTotalStock(product)
	}

	return products, total, result.Error
}

func (r *productRepositoryImpl) SearchByText(ctx context.Context, searchQuery string, filters repository.ProductFilters) ([]*entity.Product, int64, error) {
	var products []*entity.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Product{})

	// Apply text search
	if searchQuery != "" {
		searchTerm := "%" + strings.ToLower(searchQuery) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR ? = ANY(LOWER(tags::text)::text[])",
			searchTerm, searchTerm, strings.ToLower(searchQuery),
		)
	}

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = r.applySorting(query, filters)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Preload relationships
	query = query.
		Preload("Category").
		Preload("Brand").
		Preload("Variants"). // Add this line
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true)
		})

	result := query.Find(&products)

	// Calculate total stock for each product
	for _, product := range products {
		r.calculateTotalStock(product)
	}

	return products, total, result.Error
}

func (r *productRepositoryImpl) GetRelatedProducts(ctx context.Context, productID uuid.UUID, limit int) ([]*entity.Product, error) {
	var currentProduct entity.Product
	if err := r.db.WithContext(ctx).Where("id = ?", productID).First(&currentProduct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	var relatedProducts []*entity.Product
	query := r.db.WithContext(ctx).
		Where("id != ?", productID). // Exclude the current product
		Limit(limit).
		Preload("Category").
		Preload("Brand").
		Preload("Variants"). // Add this line
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_primary = ?", true)
		})

	// Try to find products in the same category first
	if currentProduct.CategoryID != nil {
		query = query.Or("category_id = ?", *currentProduct.CategoryID)
	}

	// Then try to find products by the same brand
	if currentProduct.BrandID != nil {
		query = query.Or("brand_id = ?", *currentProduct.BrandID)
	}

	if err := query.Find(&relatedProducts).Error; err != nil {
		return nil, err
	}

	// Calculate total stock for each related product
	for _, product := range relatedProducts {
		r.calculateTotalStock(product)
	}

	return relatedProducts, nil
}

func (r *productRepositoryImpl) applyFilters(query *gorm.DB, filters repository.ProductFilters) *gorm.DB {
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}

	if filters.BrandID != nil {
		query = query.Where("brand_id = ?", *filters.BrandID)
	}

	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}

	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			query = query.Where("? = ANY(tags)", tag)
		}
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.InStock != nil && *filters.InStock {
		query = query.Joins("JOIN product_variants ON products.id = product_variants.product_id").
			Where("product_variants.stock > 0")
	}

	return query
}

func (r *productRepositoryImpl) applySorting(query *gorm.DB, filters repository.ProductFilters) *gorm.DB {
	sortBy := "created_at"
	sortOrder := "DESC"

	if filters.SortBy != "" {
		switch filters.SortBy {
		case "price", "name", "created_at":
			sortBy = filters.SortBy
		}
	}

	if filters.SortOrder != "" {
		if strings.ToUpper(filters.SortOrder) == "ASC" {
			sortOrder = "ASC"
		}
	}

	return query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
}

func (r *productRepositoryImpl) calculateTotalStock(product *entity.Product) {
	totalStock := 0
	for _, variant := range product.Variants {
		totalStock += variant.Stock
	}
	product.StockQuantity = totalStock
}