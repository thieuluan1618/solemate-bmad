package entity

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a minimal product entity for wishlist item relationship
// Full product details are fetched from Product Service
type Product struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url" gorm:"column:image_url"`
	CategoryID  uuid.UUID `json:"category_id" gorm:"type:uuid"`
	BrandID     uuid.UUID `json:"brand_id" gorm:"type:uuid"`
	Stock       int       `json:"stock"`
	IsActive    bool      `json:"is_active"`
}

// WishlistItem represents a product saved to user's wishlist
type WishlistItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	AddedAt   time.Time `json:"added_at" gorm:"autoCreateTime"`

	// Product relationship - eagerly loaded when fetching wishlist
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID"`
}

func (Product) TableName() string {
	return "products"
}

func (WishlistItem) TableName() string {
	return "wishlist_items"
}
