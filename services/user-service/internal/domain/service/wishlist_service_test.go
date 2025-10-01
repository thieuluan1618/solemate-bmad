package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"solemate/services/user-service/internal/domain/entity"
)

// MockWishlistRepository is a mock implementation of repository.WishlistRepository
type MockWishlistRepository struct {
	mock.Mock
}

func (m *MockWishlistRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.WishlistItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WishlistItem), args.Error(1)
}

func (m *MockWishlistRepository) AddItem(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error) {
	args := m.Called(ctx, userID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WishlistItem), args.Error(1)
}

func (m *MockWishlistRepository) RemoveItem(ctx context.Context, userID, productID uuid.UUID) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockWishlistRepository) ClearWishlist(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockWishlistRepository) ItemExists(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, productID)
	return args.Bool(0), args.Error(1)
}

func (m *MockWishlistRepository) GetItemByProductID(ctx context.Context, userID, productID uuid.UUID) (*entity.WishlistItem, error) {
	args := m.Called(ctx, userID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WishlistItem), args.Error(1)
}

func TestWishlistService_GetWishlist(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	items := []*entity.WishlistItem{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			ProductID: uuid.New(),
		},
	}

	mockRepo.On("GetByUserID", ctx, userID).Return(items, nil)

	result, err := service.GetWishlist(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalItems)
	assert.Equal(t, items, result.Items)
	mockRepo.AssertExpectations(t)
}

func TestWishlistService_AddToWishlist_NewItem(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	expectedItem := &entity.WishlistItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
	}

	mockRepo.On("ItemExists", ctx, userID, productID).Return(false, nil)
	mockRepo.On("AddItem", ctx, userID, productID).Return(expectedItem, nil)

	result, err := service.AddToWishlist(ctx, userID, productID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedItem.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestWishlistService_AddToWishlist_ExistingItem(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	existingItem := &entity.WishlistItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
	}

	mockRepo.On("ItemExists", ctx, userID, productID).Return(true, nil)
	mockRepo.On("GetItemByProductID", ctx, userID, productID).Return(existingItem, nil)

	result, err := service.AddToWishlist(ctx, userID, productID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingItem.ID, result.ID)
	mockRepo.AssertExpectations(t)
}

func TestWishlistService_RemoveFromWishlist(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	mockRepo.On("RemoveItem", ctx, userID, productID).Return(nil)

	err := service.RemoveFromWishlist(ctx, userID, productID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWishlistService_RemoveFromWishlist_Error(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	mockRepo.On("RemoveItem", ctx, userID, productID).Return(errors.New("item not found"))

	err := service.RemoveFromWishlist(ctx, userID, productID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWishlistService_ClearWishlist(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("ClearWishlist", ctx, userID).Return(nil)

	err := service.ClearWishlist(ctx, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWishlistService_IsInWishlist(t *testing.T) {
	mockRepo := new(MockWishlistRepository)
	service := NewWishlistService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	mockRepo.On("ItemExists", ctx, userID, productID).Return(true, nil)

	result, err := service.IsInWishlist(ctx, userID, productID)

	assert.NoError(t, err)
	assert.True(t, result)
	mockRepo.AssertExpectations(t)
}
