package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Product struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SKU             string         `json:"sku" gorm:"uniqueIndex;not null"`
	Name            string         `json:"name" gorm:"not null"`
	Slug            string         `json:"slug" gorm:"uniqueIndex;not null"`
	Description     string         `json:"description" gorm:"type:text"`
	CategoryID      *uuid.UUID     `json:"category_id" gorm:"type:uuid"`
	BrandID         *uuid.UUID     `json:"brand_id" gorm:"type:uuid"`
	Price           float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	ComparePrice    *float64       `json:"compare_price" gorm:"type:decimal(10,2)"`
	Cost            *float64       `json:"cost" gorm:"type:decimal(10,2)"`
	Weight          *float64       `json:"weight" gorm:"type:decimal(10,3)"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	Tags            pq.StringArray `json:"tags" gorm:"type:text[]"`
	MetaTitle       string         `json:"meta_title"`
	MetaDescription string         `json:"meta_description" gorm:"type:text"`
	StockQuantity   int            `json:"stock_quantity" gorm:"-"` // Add this line, gorm:"-" to ignore this field in DB operations
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Category *Category        `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Brand    *Brand           `json:"brand,omitempty" gorm:"foreignKey:BrandID"`
	Variants []ProductVariant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`
	Images   []ProductImage   `json:"images,omitempty" gorm:"foreignKey:ProductID"`
}

type Category struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ParentID    *uuid.UUID `json:"parent_id" gorm:"type:uuid"`
	Name        string     `json:"name" gorm:"not null"`
	Slug        string     `json:"slug" gorm:"uniqueIndex;not null"`
	Description string     `json:"description" gorm:"type:text"`
	ImageURL    string     `json:"image_url"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Parent     *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children   []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Products   []Product  `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

type Brand struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null"`
	LogoURL     string    `json:"logo_url"`
	Website     string    `json:"website"`
	Description string    `json:"description" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Products []Product `json:"products,omitempty" gorm:"foreignKey:BrandID"`
}

type ProductVariant struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID uuid.UUID      `json:"product_id" gorm:"type:uuid;not null"`
	SKU       string         `json:"sku" gorm:"uniqueIndex;not null"`
	Size      string         `json:"size"`
	Color     string         `json:"color"`
	Price     *float64       `json:"price" gorm:"type:decimal(10,2)"`
	Stock     int            `json:"stock" gorm:"default:0"`
	Weight    *float64       `json:"weight" gorm:"type:decimal(10,3)"`
	Images    pq.StringArray `json:"images" gorm:"type:text[]"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

type ProductImage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	URL       string    `json:"url" gorm:"not null"`
	AltText   string    `json:"alt_text"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	IsPrimary bool      `json:"is_primary" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string {
	return "products"
}

func (Category) TableName() string {
	return "categories"
}

func (Brand) TableName() string {
	return "brands"
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

func (ProductImage) TableName() string {
	return "product_images"
}