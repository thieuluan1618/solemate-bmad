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

type ReviewService struct {
	reviewRepo  repository.ReviewRepository
	productRepo repository.ProductRepository
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	productRepo repository.ProductRepository,
) *ReviewService {
	return &ReviewService{
		reviewRepo:  reviewRepo,
		productRepo: productRepo,
	}
}

// CreateReviewRequest represents the request to create a review
type CreateReviewRequest struct {
	ProductID string   `json:"product_id" binding:"required"`
	Rating    int      `json:"rating" binding:"required,min=1,max=5"`
	Title     string   `json:"title"`
	Comment   string   `json:"comment"`
	Images    []string `json:"images"`
	OrderID   string   `json:"order_id"`
}

// UpdateReviewRequest represents the request to update a review
type UpdateReviewRequest struct {
	Rating  *int     `json:"rating" binding:"omitempty,min=1,max=5"`
	Title   *string  `json:"title"`
	Comment *string  `json:"comment"`
	Images  []string `json:"images"`
}

// GetReviewsRequest represents the request to get reviews for a product
type GetReviewsRequest struct {
	Status   string `json:"status"`
	SortBy   string `json:"sort_by"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

func (s *ReviewService) CreateReview(ctx context.Context, userID uuid.UUID, req *CreateReviewRequest) (*entity.Review, error) {
	// Validate rating
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	// Parse and validate product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, errors.New("invalid product ID")
	}

	// Check if product exists
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil || product == nil {
		return nil, errors.New("product not found")
	}

	// Check if user has already reviewed this product
	existingReview, err := s.reviewRepo.GetUserReviewForProduct(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}
	if existingReview != nil {
		return nil, errors.New("you have already reviewed this product")
	}

	// Parse order ID if provided
	var orderID *uuid.UUID
	if req.OrderID != "" {
		id, err := uuid.Parse(req.OrderID)
		if err != nil {
			return nil, errors.New("invalid order ID")
		}
		orderID = &id
	}

	// Create review
	review := &entity.Review{
		ProductID:    productID,
		UserID:       userID,
		OrderID:      orderID,
		Rating:       req.Rating,
		Title:        utils.SanitizeString(req.Title),
		Comment:      utils.SanitizeString(req.Comment),
		Images:       req.Images,
		IsVerified:   orderID != nil, // Mark as verified if order ID is provided
		HelpfulCount: 0,
		Status:       "APPROVED", // Auto-approve reviews (can be changed to PENDING for moderation)
	}

	err = s.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return s.reviewRepo.GetByID(ctx, review.ID)
}

func (s *ReviewService) GetReviewByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	return s.reviewRepo.GetByID(ctx, id)
}

func (s *ReviewService) GetReviewsByProductID(ctx context.Context, productID uuid.UUID, req *GetReviewsRequest) ([]*entity.Review, int64, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Build filters
	filters := repository.ReviewFilters{
		Status: req.Status,
		SortBy: req.SortBy,
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
	}

	return s.reviewRepo.GetByProductID(ctx, productID, filters)
}

func (s *ReviewService) UpdateReview(ctx context.Context, userID, reviewID uuid.UUID, req *UpdateReviewRequest) (*entity.Review, error) {
	// Get existing review
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	// Check if user owns this review
	if review.UserID != userID {
		return nil, errors.New("you can only update your own reviews")
	}

	// Update fields if provided
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, errors.New("rating must be between 1 and 5")
		}
		review.Rating = *req.Rating
	}

	if req.Title != nil {
		review.Title = utils.SanitizeString(*req.Title)
	}

	if req.Comment != nil {
		review.Comment = utils.SanitizeString(*req.Comment)
	}

	if req.Images != nil {
		review.Images = req.Images
	}

	err = s.reviewRepo.Update(ctx, review)
	if err != nil {
		return nil, fmt.Errorf("failed to update review: %w", err)
	}

	return s.reviewRepo.GetByID(ctx, review.ID)
}

func (s *ReviewService) DeleteReview(ctx context.Context, userID, reviewID uuid.UUID) error {
	// Get existing review
	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Check if user owns this review
	if review.UserID != userID {
		return errors.New("you can only delete your own reviews")
	}

	return s.reviewRepo.Delete(ctx, reviewID)
}
