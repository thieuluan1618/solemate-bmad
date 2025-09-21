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

type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

type CreateCategoryRequest struct {
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	ParentID    *string `json:"parent_id"`
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*entity.Category, error) {
	if req.Name == "" || req.Slug == "" {
		return nil, errors.New("name and slug are required")
	}

	// Check if slug already exists
	existing, _ := s.categoryRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, errors.New("category with this slug already exists")
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, errors.New("invalid parent category ID")
		}
		parent, err := s.categoryRepo.GetByID(ctx, id)
		if err != nil || parent == nil {
			return nil, errors.New("parent category not found")
		}
		parentID = &id
	}

	category := &entity.Category{
		ParentID:    parentID,
		Name:        utils.SanitizeString(req.Name),
		Slug:        utils.SanitizeString(req.Slug),
		Description: utils.SanitizeString(req.Description),
		ImageURL:    utils.SanitizeString(req.ImageURL),
		SortOrder:   req.SortOrder,
		IsActive:    true,
	}

	err := s.categoryRepo.Create(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return s.categoryRepo.GetByID(ctx, category.ID)
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *CategoryService) GetCategoryBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return s.categoryRepo.GetBySlug(ctx, slug)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, req *UpdateCategoryRequest) (*entity.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		category.Name = utils.SanitizeString(*req.Name)
	}

	if req.Slug != nil {
		existing, _ := s.categoryRepo.GetBySlug(ctx, *req.Slug)
		if existing != nil && existing.ID != id {
			return nil, errors.New("category with this slug already exists")
		}
		category.Slug = utils.SanitizeString(*req.Slug)
	}

	if req.Description != nil {
		category.Description = utils.SanitizeString(*req.Description)
	}

	if req.ImageURL != nil {
		category.ImageURL = utils.SanitizeString(*req.ImageURL)
	}

	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if req.ParentID != nil {
		if *req.ParentID == "" {
			category.ParentID = nil
		} else {
			parentUUID, err := uuid.Parse(*req.ParentID)
			if err != nil {
				return nil, errors.New("invalid parent category ID")
			}
			if parentUUID == id {
				return nil, errors.New("category cannot be its own parent")
			}
			parent, err := s.categoryRepo.GetByID(ctx, parentUUID)
			if err != nil || parent == nil {
				return nil, errors.New("parent category not found")
			}
			category.ParentID = &parentUUID
		}
	}

	err = s.categoryRepo.Update(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return s.categoryRepo.GetByID(ctx, category.ID)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.categoryRepo.Delete(ctx, id)
}

func (s *CategoryService) ListCategories(ctx context.Context, page, limit int) ([]*entity.Category, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	offset := (page - 1) * limit
	return s.categoryRepo.List(ctx, limit, offset)
}

func (s *CategoryService) GetCategoryTree(ctx context.Context) ([]*entity.Category, error) {
	return s.categoryRepo.GetTree(ctx)
}