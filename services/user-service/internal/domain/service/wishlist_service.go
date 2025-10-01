package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"solemate/services/user-service/internal/domain/entity"
	"solemate/services/user-service/internal/domain/repository"
)

type WishlistService struct {
	wishlistRepo repository.WishlistRepository
}

func NewWishlistService(wishlistRepo repository.WishlistRepository) *WishlistService {
	return &WishlistService{
		wishlistRepo: wishlistRepo,
	}
}

// WishlistResponse represents the response format expected by frontend
type WishlistResponse struct {
	Items      []*entity.WishlistItem `json:"items"`
	TotalItems int                    `json:"totalItems"`
}

// GetWishlist retrieves all wishlist items for a user
func (s *WishlistService) GetWishlist(ctx context.Context, userID uuid.UUID) (*WishlistResponse, error) {
	items, err := s.wishlistRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist: %w", err)
	}

	return &WishlistResponse{
		Items:      items,
		TotalItems: len(items),
	}, nil
}

// AddToWishlist adds a product to user's wishlist
func (s *WishlistService) AddToWishlist(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error) {
	// Check if product already exists in wishlist
	exists, err := s.wishlistRepo.ItemExists(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to check wishlist item: %w", err)
	}

	if exists {
		// Return existing item if already in wishlist (graceful handling)
		item, err := s.wishlistRepo.GetItemByProductID(ctx, userID, productID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing wishlist item: %w", err)
		}
		return item, nil
	}

	// Add new item to wishlist
	item, err := s.wishlistRepo.AddItem(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to add to wishlist: %w", err)
	}

	return item, nil
}

// RemoveFromWishlist removes a product from user's wishlist by product ID
func (s *WishlistService) RemoveFromWishlist(ctx context.Context, userID, productID uuid.UUID) error {
	err := s.wishlistRepo.RemoveItem(ctx, userID, productID)
	if err != nil {
		return fmt.Errorf("failed to remove from wishlist: %w", err)
	}

	return nil
}

// ClearWishlist removes all items from user's wishlist
func (s *WishlistService) ClearWishlist(ctx context.Context, userID uuid.UUID) error {
	err := s.wishlistRepo.ClearWishlist(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to clear wishlist: %w", err)
	}

	return nil
}

// IsInWishlist checks if a product is in user's wishlist
func (s *WishlistService) IsInWishlist(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	return s.wishlistRepo.ItemExists(ctx, userID, productID)
}
