package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"solemate/services/cart-service/internal/domain/entity"
	"solemate/services/cart-service/internal/domain/repository"
)

type cartRepositoryImpl struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewCartRepository(redisClient *redis.Client) repository.CartRepository {
	return &cartRepositoryImpl{
		redis: redisClient,
		ttl:   24 * time.Hour, // Default cart expiration: 24 hours
	}
}

func (r *cartRepositoryImpl) GetCart(ctx context.Context, userID uuid.UUID) (*entity.Cart, error) {
	key := r.getCartKey(userID)

	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Return empty cart if not found
			return r.createEmptyCart(userID), nil
		}
		return nil, fmt.Errorf("failed to get cart from Redis: %w", err)
	}

	var cart entity.Cart
	if err := json.Unmarshal([]byte(data), &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart data: %w", err)
	}

	// Check if cart is expired
	if cart.IsExpired() {
		r.DeleteCart(ctx, userID)
		return r.createEmptyCart(userID), nil
	}

	return &cart, nil
}

func (r *cartRepositoryImpl) SaveCart(ctx context.Context, cart *entity.Cart) error {
	key := r.getCartKey(cart.UserID)

	cart.UpdatedAt = time.Now()
	if cart.ExpiresAt.IsZero() {
		cart.ExpiresAt = time.Now().Add(r.ttl)
	}

	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart data: %w", err)
	}

	// Set with expiration
	expiration := time.Until(cart.ExpiresAt)
	if expiration <= 0 {
		expiration = r.ttl
		cart.ExpiresAt = time.Now().Add(expiration)
	}

	err = r.redis.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save cart to Redis: %w", err)
	}

	return nil
}

func (r *cartRepositoryImpl) DeleteCart(ctx context.Context, userID uuid.UUID) error {
	key := r.getCartKey(userID)
	return r.redis.Del(ctx, key).Err()
}

func (r *cartRepositoryImpl) ExtendCartExpiration(ctx context.Context, userID uuid.UUID, duration time.Duration) error {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.ExtendExpiration(duration)
	return r.SaveCart(ctx, cart)
}

func (r *cartRepositoryImpl) AddItem(ctx context.Context, userID uuid.UUID, item *entity.CartItem) error {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// Check if item already exists
	existingItem := cart.GetItemByProduct(item.ProductID, item.VariantID)
	if existingItem != nil {
		existingItem.Quantity += item.Quantity
		existingItem.TotalPrice = existingItem.Price * float64(existingItem.Quantity) - existingItem.Discount
		existingItem.UpdatedAt = time.Now()
	} else {
		item.ID = uuid.New()
		item.AddedAt = time.Now()
		item.UpdatedAt = time.Now()
		item.TotalPrice = item.Price * float64(item.Quantity) - item.Discount
		cart.Items = append(cart.Items, *item)
	}

	cart.CalculateTotals()
	return r.SaveCart(ctx, cart)
}

func (r *cartRepositoryImpl) UpdateItemQuantity(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) error {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	if quantity <= 0 {
		return r.RemoveItem(ctx, userID, itemID)
	}

	updated := cart.UpdateItemQuantity(itemID, quantity)
	if !updated {
		return fmt.Errorf("cart item not found")
	}

	return r.SaveCart(ctx, cart)
}

func (r *cartRepositoryImpl) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	removed := cart.RemoveItem(itemID)
	if !removed {
		return fmt.Errorf("cart item not found")
	}

	return r.SaveCart(ctx, cart)
}

func (r *cartRepositoryImpl) ClearCart(ctx context.Context, userID uuid.UUID) error {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.Clear()
	return r.SaveCart(ctx, cart)
}

func (r *cartRepositoryImpl) GetCartSummary(ctx context.Context, userID uuid.UUID) (*entity.CartSummary, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalDiscount := 0.0
	for _, item := range cart.Items {
		totalDiscount += item.Discount
	}

	return &entity.CartSummary{
		UserID:        cart.UserID,
		TotalItems:    cart.TotalItems,
		TotalPrice:    cart.TotalPrice,
		TotalDiscount: totalDiscount,
		ItemCount:     len(cart.Items),
		UpdatedAt:     cart.UpdatedAt,
	}, nil
}

func (r *cartRepositoryImpl) GetExpiredCarts(ctx context.Context, limit int) ([]uuid.UUID, error) {
	// Since Redis automatically expires keys, we don't need to implement this
	// This method would be useful for cleanup jobs if we were using persistent storage
	return []uuid.UUID{}, nil
}

func (r *cartRepositoryImpl) DeleteExpiredCarts(ctx context.Context, userIDs []uuid.UUID) error {
	// Redis automatically handles expiration, so this is a no-op
	return nil
}

// Helper methods
func (r *cartRepositoryImpl) getCartKey(userID uuid.UUID) string {
	return fmt.Sprintf("cart:%s", userID.String())
}

func (r *cartRepositoryImpl) createEmptyCart(userID uuid.UUID) *entity.Cart {
	now := time.Now()
	return &entity.Cart{
		UserID:     userID,
		Items:      []entity.CartItem{},
		TotalItems: 0,
		TotalPrice: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiresAt:  now.Add(r.ttl),
	}
}