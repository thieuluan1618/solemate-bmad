package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/product-service/internal/domain/entity"
	"solemate/services/product-service/internal/domain/repository"
)

type ProductService struct {
	productRepo   repository.ProductRepository
	categoryRepo  repository.CategoryRepository
	brandRepo     repository.BrandRepository
	variantRepo   repository.ProductVariantRepository
	imageRepo     repository.ProductImageRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	brandRepo repository.BrandRepository,
	variantRepo repository.ProductVariantRepository,
	imageRepo repository.ProductImageRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
		variantRepo:  variantRepo,
		imageRepo:    imageRepo,
	}
}

// Product CRUD operations
type CreateProductRequest struct {
	SKU             string    `json:"sku" binding:"required"`
	Name            string    `json:"name" binding:"required"`
	Slug            string    `json:"slug" binding:"required"`
	Description     string    `json:"description"`
	CategoryID      *string   `json:"category_id"`
	BrandID         *string   `json:"brand_id"`
	Price           float64   `json:"price" binding:"required,min=0"`
	ComparePrice    *float64  `json:"compare_price"`
	Cost            *float64  `json:"cost"`
	Weight          *float64  `json:"weight"`
	Tags            []string  `json:"tags"`
	MetaTitle       string    `json:"meta_title"`
	MetaDescription string    `json:"meta_description"`
}

type UpdateProductRequest struct {
	Name            *string   `json:"name"`
	Slug            *string   `json:"slug"`
	Description     *string   `json:"description"`
	CategoryID      *string   `json:"category_id"`
	BrandID         *string   `json:"brand_id"`
	Price           *float64  `json:"price"`
	ComparePrice    *float64  `json:"compare_price"`
	Cost            *float64  `json:"cost"`
	Weight          *float64  `json:"weight"`
	Tags            *[]string `json:"tags"`
	MetaTitle       *string   `json:"meta_title"`
	MetaDescription *string   `json:"meta_description"`
	IsActive        *bool     `json:"is_active"`
}

type ProductSearchRequest struct {
	Query        string    `json:"query"`
	CategoryID   *string   `json:"category_id"`
	BrandID      *string   `json:"brand_id"`
	MinPrice     *float64  `json:"min_price"`
	MaxPrice     *float64  `json:"max_price"`
	Tags         []string  `json:"tags"`
	InStock      *bool     `json:"in_stock"`
	SortBy       string    `json:"sort_by"`
	SortOrder    string    `json:"sort_order"`
	Page         int       `json:"page"`
	Limit        int       `json:"limit"`
}

func (s *ProductService) CreateProduct(ctx context.Context, req *CreateProductRequest) (*entity.Product, error) {
	// Validate required fields
	if req.SKU == "" || req.Name == "" || req.Slug == "" {
		return nil, errors.New("SKU, name, and slug are required")
	}

	if req.Price < 0 {
		return nil, errors.New("price must be non-negative")
	}

	// Check if SKU already exists
	existing, _ := s.productRepo.GetBySKU(ctx, req.SKU)
	if existing != nil {
		return nil, errors.New("product with this SKU already exists")
	}

	// Check if slug already exists
	existing, _ = s.productRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.New("product with this slug already exists")
	}

	// Validate category if provided
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, errors.New("invalid category ID")
		}
		category, err := s.categoryRepo.GetByID(ctx, id)
		if err != nil || category == nil {
			return nil, errors.New("category not found")
		}
		categoryID = &id
	}

	// Validate brand if provided
	var brandID *uuid.UUID
	if req.BrandID != nil && *req.BrandID != "" {
		id, err := uuid.Parse(*req.BrandID)
		if err != nil {
			return nil, errors.New("invalid brand ID")
		}
		brand, err := s.brandRepo.GetByID(ctx, id)
		if err != nil || brand == nil {
			return nil, errors.New("brand not found")
		}
		brandID = &id
	}

	// Create product
	product := &entity.Product{
		SKU:             utils.SanitizeString(req.SKU),
		Name:            utils.SanitizeString(req.Name),
		Slug:            utils.SanitizeString(req.Slug),
		Description:     utils.SanitizeString(req.Description),
		CategoryID:      categoryID,
		BrandID:         brandID,
		Price:           req.Price,
		ComparePrice:    req.ComparePrice,
		Cost:            req.Cost,
		Weight:          req.Weight,
		Tags:            req.Tags,
		MetaTitle:       utils.SanitizeString(req.MetaTitle),
		MetaDescription: utils.SanitizeString(req.MetaDescription),
		IsActive:        true,
	}

	err := s.productRepo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return s.productRepo.GetByID(ctx, product.ID)
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	return s.productRepo.GetBySlug(ctx, slug)
}

func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req *UpdateProductRequest) (*entity.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		product.Name = utils.SanitizeString(*req.Name)
	}

	if req.Slug != nil {
		// Check if new slug already exists
		existing, _ := s.productRepo.GetBySlug(ctx, *req.Slug)
		if existing != nil && existing.ID != id {
			return nil, errors.New("product with this slug already exists")
		}
		product.Slug = utils.SanitizeString(*req.Slug)
	}

	if req.Description != nil {
		product.Description = utils.SanitizeString(*req.Description)
	}

	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			product.CategoryID = nil
		} else {
			categoryUUID, err := uuid.Parse(*req.CategoryID)
			if err != nil {
				return nil, errors.New("invalid category ID")
			}
			category, err := s.categoryRepo.GetByID(ctx, categoryUUID)
			if err != nil || category == nil {
				return nil, errors.New("category not found")
			}
			product.CategoryID = &categoryUUID
		}
	}

	if req.BrandID != nil {
		if *req.BrandID == "" {
			product.BrandID = nil
		} else {
			brandUUID, err := uuid.Parse(*req.BrandID)
			if err != nil {
				return nil, errors.New("invalid brand ID")
			}
			brand, err := s.brandRepo.GetByID(ctx, brandUUID)
			if err != nil || brand == nil {
				return nil, errors.New("brand not found")
			}
			product.BrandID = &brandUUID
		}
	}

	if req.Price != nil {
		if *req.Price < 0 {
			return nil, errors.New("price must be non-negative")
		}
		product.Price = *req.Price
	}

	if req.ComparePrice != nil {
		product.ComparePrice = req.ComparePrice
	}

	if req.Cost != nil {
		product.Cost = req.Cost
	}

	if req.Weight != nil {
		product.Weight = req.Weight
	}

	if req.Tags != nil {
		product.Tags = *req.Tags
	}

	if req.MetaTitle != nil {
		product.MetaTitle = utils.SanitizeString(*req.MetaTitle)
	}

	if req.MetaDescription != nil {
		product.MetaDescription = utils.SanitizeString(*req.MetaDescription)
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	err = s.productRepo.Update(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return s.productRepo.GetByID(ctx, product.ID)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.productRepo.Delete(ctx, id)
}

func (s *ProductService) SearchProducts(ctx context.Context, req *ProductSearchRequest) ([]*entity.Product, int64, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// Build filters
	filters := repository.ProductFilters{
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Limit:     req.Limit,
		Offset:    (req.Page - 1) * req.Limit,
		IsActive:  boolPtr(true), // Only show active products by default
		InStock:   req.InStock,
		MinPrice:  req.MinPrice,
		MaxPrice:  req.MaxPrice,
		Tags:      req.Tags,
	}

	// Parse UUIDs if provided
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, 0, errors.New("invalid category ID")
		}
		filters.CategoryID = &id
	}

	if req.BrandID != nil && *req.BrandID != "" {
		id, err := uuid.Parse(*req.BrandID)
		if err != nil {
			return nil, 0, errors.New("invalid brand ID")
		}
		filters.BrandID = &id
	}

	// Use text search if query is provided
	if req.Query != "" && strings.TrimSpace(req.Query) != "" {
		return s.productRepo.SearchByText(ctx, strings.TrimSpace(req.Query), filters)
	}

	return s.productRepo.List(ctx, filters)
}

func (s *ProductService) ListProducts(ctx context.Context, page, limit int) ([]*entity.Product, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	filters := repository.ProductFilters{
		Limit:    limit,
		Offset:   (page - 1) * limit,
		IsActive: boolPtr(true),
		SortBy:   "created_at",
		SortOrder: "desc",
	}

	return s.productRepo.List(ctx, filters)
}

func (s *ProductService) GetRelatedProducts(ctx context.Context, productID uuid.UUID, limit int) ([]*entity.Product, error) {
	if limit <= 0 {
		limit = 4 // Default limit
	}
	return s.productRepo.GetRelatedProducts(ctx, productID, limit)
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}