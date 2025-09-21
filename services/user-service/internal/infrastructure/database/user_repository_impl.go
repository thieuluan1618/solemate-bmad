package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"solemate/services/user-service/internal/domain/entity"
	"solemate/services/user-service/internal/domain/repository"
)

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Create(user)
	return result.Error
}

func (r *userRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).Save(user)
	return result.Error
}

func (r *userRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entity.User{}, id)
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return result.Error
}

func (r *userRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users)
	return users, total, result.Error
}

func (r *userRepositoryImpl) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("last_login_at", now)
	return result.Error
}
