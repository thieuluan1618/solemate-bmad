package repository

import (
	"context"

	"github.com/google/uuid"
	"solemate/services/product-service/internal/domain/entity"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *entity.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error)
	GetByProductID(ctx context.Context, productID uuid.UUID, filters ReviewFilters) ([]*entity.Review, int64, error)
	Update(ctx context.Context, review *entity.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUserReviewForProduct(ctx context.Context, userID, productID uuid.UUID) (*entity.Review, error)
}

// ReviewFilters represents filters for review queries
type ReviewFilters struct {
	Status    string `json:"status"`    // PENDING, APPROVED, REJECTED
	SortBy    string `json:"sort_by"`   // newest, highest_rated, lowest_rated, most_helpful
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}
