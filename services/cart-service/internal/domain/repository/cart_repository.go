package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"solemate/services/cart-service/internal/domain/entity"
)

type CartRepository interface {
	// Cart operations
	GetCart(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)
	SaveCart(ctx context.Context, cart *entity.Cart) error
	DeleteCart(ctx context.Context, userID uuid.UUID) error
	ExtendCartExpiration(ctx context.Context, userID uuid.UUID, duration time.Duration) error

	// Cart item operations
	AddItem(ctx context.Context, userID uuid.UUID, item *entity.CartItem) error
	UpdateItemQuantity(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) error
	RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error

	// Cart utilities
	GetCartSummary(ctx context.Context, userID uuid.UUID) (*entity.CartSummary, error)
	GetExpiredCarts(ctx context.Context, limit int) ([]uuid.UUID, error)
	DeleteExpiredCarts(ctx context.Context, userIDs []uuid.UUID) error
}

type ProductRepository interface {
	// Product information for cart operations
	GetProduct(ctx context.Context, productID uuid.UUID) (*entity.ProductInfo, error)
	GetProductVariant(ctx context.Context, variantID uuid.UUID) (*entity.ProductVariantInfo, error)
	GetProductBySKU(ctx context.Context, sku string) (*entity.ProductInfo, error)
	GetVariantBySKU(ctx context.Context, sku string) (*entity.ProductVariantInfo, error)

	// Stock validation
	ValidateStock(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, quantity int) (bool, error)
	CheckProductAvailability(ctx context.Context, productID uuid.UUID) (bool, error)
}