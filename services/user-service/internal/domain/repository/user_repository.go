package repository

import (
	"context"

	"github.com/google/uuid"
	"solemate/services/user-service/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type AddressRepository interface {
	Create(ctx context.Context, address *entity.Address) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Address, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Address, error)
	Update(ctx context.Context, address *entity.Address) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetDefault(ctx context.Context, userID, addressID uuid.UUID) error
}
