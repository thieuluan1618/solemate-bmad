package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"solemate/services/cart-service/internal/domain/entity"
	"solemate/services/cart-service/internal/domain/repository"
)

type CartService interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)
	AddItem(ctx context.Context, userID uuid.UUID, item *entity.CartItem) error
	UpdateItemQuantity(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) error
	RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
	GetCartSummary(ctx context.Context, userID uuid.UUID) (*entity.CartSummary, error)
	ExtendCartExpiration(ctx context.Context, userID uuid.UUID, duration time.Duration) error
	ValidateAndAddItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID, variantID *uuid.UUID, quantity int) error
	ApplyDiscount(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, discount float64) error
	GetItemCount(ctx context.Context, userID uuid.UUID) (int, error)
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) GetCart(ctx context.Context, userID uuid.UUID) (*entity.Cart, error) {
	return s.cartRepo.GetCart(ctx, userID)
}

func (s *cartService) AddItem(ctx context.Context, userID uuid.UUID, item *entity.CartItem) error {
	if item.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	if item.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}

	// Validate product exists and has stock
	if s.productRepo != nil {
		available, err := s.productRepo.CheckProductAvailability(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to check product availability: %w", err)
		}
		if !available {
			return fmt.Errorf("product is not available")
		}

		hasStock, err := s.productRepo.ValidateStock(ctx, item.ProductID, item.VariantID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to validate stock: %w", err)
		}
		if !hasStock {
			return fmt.Errorf("insufficient stock for requested quantity")
		}
	}

	return s.cartRepo.AddItem(ctx, userID, item)
}

func (s *cartService) ValidateAndAddItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID, variantID *uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// For now, skip product validation if productRepo is nil (cross-service communication not implemented)
	if s.productRepo == nil {
		// Create a basic cart item without product validation
		item := &entity.CartItem{
			ID:         uuid.New(),
			ProductID:  productID,
			VariantID:  variantID,
			SKU:        "UNKNOWN-SKU",
			Name:       "Unknown Product",
			Size:       "",
			Color:      "",
			Price:      100.0, // Default price for testing
			Quantity:   quantity,
			Discount:   0.0,
			TotalPrice: 100.0 * float64(quantity),
			ImageURL:   "",
			AddedAt:    time.Now(),
			UpdatedAt:  time.Now(),
		}
		return s.AddItem(ctx, userID, item)
	}

	// Get product information
	product, err := s.productRepo.GetProduct(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	var variant *entity.ProductVariantInfo
	if variantID != nil {
		variant, err = s.productRepo.GetProductVariant(ctx, *variantID)
		if err != nil {
			return fmt.Errorf("failed to get product variant: %w", err)
		}
	}

	// Validate stock
	hasStock, err := s.productRepo.ValidateStock(ctx, productID, variantID, quantity)
	if err != nil {
		return fmt.Errorf("failed to validate stock: %w", err)
	}
	if !hasStock {
		return fmt.Errorf("insufficient stock for requested quantity")
	}

	// Create cart item
	price := product.Price
	sku := product.SKU
	size := ""
	color := ""

	if variant != nil {
		if variant.Price > 0 {
			price = variant.Price
		}
		sku = variant.SKU
		size = variant.Size
		color = variant.Color
	}

	item := &entity.CartItem{
		ProductID:  productID,
		VariantID:  variantID,
		SKU:        sku,
		Name:       product.Name,
		Size:       size,
		Color:      color,
		Price:      price,
		Quantity:   quantity,
		Discount:   0,
		TotalPrice: price * float64(quantity),
		ImageURL:   product.ImageURL,
	}

	return s.cartRepo.AddItem(ctx, userID, item)
}

func (s *cartService) UpdateItemQuantity(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	// If quantity is 0, remove the item
	if quantity == 0 {
		return s.RemoveItem(ctx, userID, itemID)
	}

	// Validate stock for the new quantity if productRepo is available
	if s.productRepo != nil {
		cart, err := s.cartRepo.GetCart(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get cart: %w", err)
		}

		// Find the item to validate stock
		for _, item := range cart.Items {
			if item.ID == itemID {
				hasStock, err := s.productRepo.ValidateStock(ctx, item.ProductID, item.VariantID, quantity)
				if err != nil {
					return fmt.Errorf("failed to validate stock: %w", err)
				}
				if !hasStock {
					return fmt.Errorf("insufficient stock for requested quantity")
				}
				break
			}
		}
	}

	return s.cartRepo.UpdateItemQuantity(ctx, userID, itemID, quantity)
}

func (s *cartService) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	return s.cartRepo.RemoveItem(ctx, userID, itemID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.cartRepo.ClearCart(ctx, userID)
}

func (s *cartService) GetCartSummary(ctx context.Context, userID uuid.UUID) (*entity.CartSummary, error) {
	return s.cartRepo.GetCartSummary(ctx, userID)
}

func (s *cartService) ExtendCartExpiration(ctx context.Context, userID uuid.UUID, duration time.Duration) error {
	if duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	return s.cartRepo.ExtendCartExpiration(ctx, userID, duration)
}

func (s *cartService) ApplyDiscount(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, discount float64) error {
	if discount < 0 {
		return fmt.Errorf("discount cannot be negative")
	}

	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get cart: %w", err)
	}

	// Find and update the item
	for i := range cart.Items {
		if cart.Items[i].ID == itemID {
			cart.Items[i].Discount = discount
			cart.Items[i].TotalPrice = cart.Items[i].Price*float64(cart.Items[i].Quantity) - discount
			cart.Items[i].UpdatedAt = time.Now()
			break
		}
	}

	cart.CalculateTotals()
	return s.cartRepo.SaveCart(ctx, cart)
}

func (s *cartService) GetItemCount(ctx context.Context, userID uuid.UUID) (int, error) {
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return 0, err
	}
	return len(cart.Items), nil
}