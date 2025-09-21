package entity

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	UserID     uuid.UUID  `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	TotalPrice float64    `json:"total_price"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
}

type CartItem struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	VariantID   *uuid.UUID `json:"variant_id,omitempty"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Size        string    `json:"size,omitempty"`
	Color       string    `json:"color,omitempty"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Discount    float64   `json:"discount"`
	TotalPrice  float64   `json:"total_price"`
	ImageURL    string    `json:"image_url,omitempty"`
	AddedAt     time.Time `json:"added_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CartSummary provides aggregated cart information
type CartSummary struct {
	UserID       uuid.UUID `json:"user_id"`
	TotalItems   int       `json:"total_items"`
	TotalPrice   float64   `json:"total_price"`
	TotalDiscount float64  `json:"total_discount"`
	ItemCount    int       `json:"item_count"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProductInfo represents product information for cart operations
type ProductInfo struct {
	ID       uuid.UUID `json:"id"`
	SKU      string    `json:"sku"`
	Name     string    `json:"name"`
	Price    float64   `json:"price"`
	ImageURL string    `json:"image_url"`
	InStock  bool      `json:"in_stock"`
	Stock    int       `json:"stock"`
}

// ProductVariantInfo represents variant information for cart operations
type ProductVariantInfo struct {
	ID       uuid.UUID `json:"id"`
	SKU      string    `json:"sku"`
	Size     string    `json:"size"`
	Color    string    `json:"color"`
	Price    float64   `json:"price"`
	InStock  bool      `json:"in_stock"`
	Stock    int       `json:"stock"`
	ImageURL string    `json:"image_url"`
}

func (c *Cart) CalculateTotals() {
	c.TotalItems = 0
	c.TotalPrice = 0

	for _, item := range c.Items {
		c.TotalItems += item.Quantity
		c.TotalPrice += item.TotalPrice
	}
}

func (c *Cart) GetItemByProduct(productID uuid.UUID, variantID *uuid.UUID) *CartItem {
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			// Check variant match
			if variantID == nil && c.Items[i].VariantID == nil {
				return &c.Items[i]
			}
			if variantID != nil && c.Items[i].VariantID != nil && *variantID == *c.Items[i].VariantID {
				return &c.Items[i]
			}
		}
	}
	return nil
}

func (c *Cart) RemoveItem(itemID uuid.UUID) bool {
	for i, item := range c.Items {
		if item.ID == itemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.CalculateTotals()
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

func (c *Cart) UpdateItemQuantity(itemID uuid.UUID, quantity int) bool {
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			if quantity <= 0 {
				return c.RemoveItem(itemID)
			}
			c.Items[i].Quantity = quantity
			c.Items[i].TotalPrice = c.Items[i].Price * float64(quantity) - c.Items[i].Discount
			c.Items[i].UpdatedAt = time.Now()
			c.CalculateTotals()
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

func (c *Cart) AddItem(product ProductInfo, variant *ProductVariantInfo, quantity int) {
	now := time.Now()

	// Determine which price and details to use
	price := product.Price
	sku := product.SKU
	size := ""
	color := ""
	var variantID *uuid.UUID

	if variant != nil {
		if variant.Price > 0 {
			price = variant.Price
		}
		sku = variant.SKU
		size = variant.Size
		color = variant.Color
		variantID = &variant.ID
	}

	// Check if item already exists
	existingItem := c.GetItemByProduct(product.ID, variantID)
	if existingItem != nil {
		existingItem.Quantity += quantity
		existingItem.TotalPrice = existingItem.Price * float64(existingItem.Quantity) - existingItem.Discount
		existingItem.UpdatedAt = now
	} else {
		// Add new item
		newItem := CartItem{
			ID:         uuid.New(),
			ProductID:  product.ID,
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
			AddedAt:    now,
			UpdatedAt:  now,
		}
		c.Items = append(c.Items, newItem)
	}

	c.CalculateTotals()
	c.UpdatedAt = now
}

func (c *Cart) Clear() {
	c.Items = []CartItem{}
	c.TotalItems = 0
	c.TotalPrice = 0
	c.UpdatedAt = time.Now()
}

func (c *Cart) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *Cart) ExtendExpiration(duration time.Duration) {
	c.ExpiresAt = time.Now().Add(duration)
	c.UpdatedAt = time.Now()
}