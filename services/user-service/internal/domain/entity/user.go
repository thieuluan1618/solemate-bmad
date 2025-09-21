package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string     `json:"-" gorm:"not null"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	PhoneNumber   string     `json:"phone_number"`
	Role          string     `json:"role" gorm:"default:customer"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	EmailVerified bool       `json:"email_verified" gorm:"default:false"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

type Address struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Type       string    `json:"type" gorm:"default:shipping"`
	Name       string    `json:"name" gorm:"not null"`
	Street1    string    `json:"street_1" gorm:"not null"`
	Street2    string    `json:"street_2"`
	City       string    `json:"city" gorm:"not null"`
	State      string    `json:"state" gorm:"not null"`
	PostalCode string    `json:"postal_code" gorm:"not null"`
	Country    string    `json:"country" gorm:"not null"`
	Phone      string    `json:"phone"`
	IsDefault  bool      `json:"is_default" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

func (Address) TableName() string {
	return "addresses"
}
