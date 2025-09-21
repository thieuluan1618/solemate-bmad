package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/product-service/internal/domain/entity"
	"solemate/services/product-service/internal/domain/repository"
)

type BrandService struct {
	brandRepo repository.BrandRepository
}

func NewBrandService(brandRepo repository.BrandRepository) *BrandService {
	return &BrandService{
		brandRepo: brandRepo,
	}
}

type CreateBrandRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	LogoURL     string `json:"logo_url"`
	Website     string `json:"website"`
	Description string `json:"description"`
}

type UpdateBrandRequest struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	LogoURL     *string `json:"logo_url"`
	Website     *string `json:"website"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

func (s *BrandService) CreateBrand(ctx context.Context, req *CreateBrandRequest) (*entity.Brand, error) {
	if req.Name == "" || req.Slug == "" {
		return nil, errors.New("name and slug are required")
	}

	// Check if name already exists
	existing, _ := s.brandRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.New("brand with this slug already exists")
	}

	brand := &entity.Brand{
		Name:        utils.SanitizeString(req.Name),
		Slug:        utils.SanitizeString(req.Slug),
		LogoURL:     utils.SanitizeString(req.LogoURL),
		Website:     utils.SanitizeString(req.Website),
		Description: utils.SanitizeString(req.Description),
		IsActive:    true,
	}

	err := s.brandRepo.Create(ctx, brand)
	if err != nil {
		return nil, fmt.Errorf("failed to create brand: %w", err)
	}

	return s.brandRepo.GetByID(ctx, brand.ID)
}

func (s *BrandService) GetBrandByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error) {
	return s.brandRepo.GetByID(ctx, id)
}

func (s *BrandService) GetBrandBySlug(ctx context.Context, slug string) (*entity.Brand, error) {
	return s.brandRepo.GetBySlug(ctx, slug)
}

func (s *BrandService) UpdateBrand(ctx context.Context, id uuid.UUID, req *UpdateBrandRequest) (*entity.Brand, error) {
	brand, err := s.brandRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		brand.Name = utils.SanitizeString(*req.Name)
	}

	if req.Slug != nil {
		existing, _ := s.brandRepo.GetBySlug(ctx, *req.Slug)
		if existing != nil && existing.ID != id {
			return nil, errors.New("brand with this slug already exists")
		}
		brand.Slug = utils.SanitizeString(*req.Slug)
	}

	if req.LogoURL != nil {
		brand.LogoURL = utils.SanitizeString(*req.LogoURL)
	}

	if req.Website != nil {
		brand.Website = utils.SanitizeString(*req.Website)
	}

	if req.Description != nil {
		brand.Description = utils.SanitizeString(*req.Description)
	}

	if req.IsActive != nil {
		brand.IsActive = *req.IsActive
	}

	err = s.brandRepo.Update(ctx, brand)
	if err != nil {
		return nil, fmt.Errorf("failed to update brand: %w", err)
	}

	return s.brandRepo.GetByID(ctx, brand.ID)
}

func (s *BrandService) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	return s.brandRepo.Delete(ctx, id)
}

func (s *BrandService) ListBrands(ctx context.Context, page, limit int) ([]*entity.Brand, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	offset := (page - 1) * limit
	return s.brandRepo.List(ctx, limit, offset)
}