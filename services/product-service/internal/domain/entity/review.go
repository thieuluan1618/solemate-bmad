package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Review struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID    uuid.UUID      `json:"product_id" gorm:"type:uuid;not null"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	OrderID      *uuid.UUID     `json:"order_id" gorm:"type:uuid"`
	Rating       int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Title        string         `json:"title" gorm:"size:255"`
	Comment      string         `json:"comment" gorm:"type:text"`
	Images       pq.StringArray `json:"images" gorm:"type:text[]"`
	IsVerified   bool           `json:"is_verified" gorm:"default:false"`
	HelpfulCount int            `json:"helpful_count" gorm:"default:0"`
	Status       string         `json:"status" gorm:"size:20;default:PENDING;check:status IN ('PENDING', 'APPROVED', 'REJECTED')"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

func (Review) TableName() string {
	return "reviews"
}
