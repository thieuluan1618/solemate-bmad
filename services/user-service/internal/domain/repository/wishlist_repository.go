package repository

import (
	"context"

	"github.com/google/uuid"
	"solemate/services/user-service/internal/domain/entity"
)

type WishlistRepository interface {
	// GetByUserID retrieves all wishlist items for a user with product details
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WishlistItem, error)

	// AddItem adds a product to user's wishlist
	AddItem(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error)

	// RemoveItem removes a product from user's wishlist by product ID
	RemoveItem(ctx context.Context, userID, productID uuid.UUID) error

	// ClearWishlist removes all items from user's wishlist
	ClearWishlist(ctx context.Context, userID uuid.UUID) error

	// ItemExists checks if a product is already in user's wishlist
	ItemExists(ctx context.Context, userID, productID uuid.UUID) (bool, error)

	// GetItemByProductID retrieves a specific wishlist item by product ID
	GetItemByProductID(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error)
}
