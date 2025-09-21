package repository

import (
	"context"

	"github.com/google/uuid"
	"solemate/services/product-service/internal/domain/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters ProductFilters) ([]*entity.Product, int64, error)
	SearchByText(ctx context.Context, query string, filters ProductFilters) ([]*entity.Product, int64, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.Category, int64, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error)
	GetTree(ctx context.Context) ([]*entity.Category, error)
}

type BrandRepository interface {
	Create(ctx context.Context, brand *entity.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Brand, error)
	Update(ctx context.Context, brand *entity.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.Brand, int64, error)
}

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *entity.ProductVariant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ProductVariant, error)
	GetBySKU(ctx context.Context, sku string) (*entity.ProductVariant, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entity.ProductVariant, error)
	Update(ctx context.Context, variant *entity.ProductVariant) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error
}

type ProductImageRepository interface {
	Create(ctx context.Context, image *entity.ProductImage) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ProductImage, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entity.ProductImage, error)
	Update(ctx context.Context, image *entity.ProductImage) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetPrimary(ctx context.Context, productID, imageID uuid.UUID) error
}

// ProductFilters represents filters for product queries
type ProductFilters struct {
	CategoryID   *uuid.UUID `json:"category_id"`
	BrandID      *uuid.UUID `json:"brand_id"`
	MinPrice     *float64   `json:"min_price"`
	MaxPrice     *float64   `json:"max_price"`
	Tags         []string   `json:"tags"`
	IsActive     *bool      `json:"is_active"`
	InStock      *bool      `json:"in_stock"`
	SortBy       string     `json:"sort_by"`       // price, name, created_at, popularity
	SortOrder    string     `json:"sort_order"`    // asc, desc
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}