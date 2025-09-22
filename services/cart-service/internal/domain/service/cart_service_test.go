package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"solemate/services/cart-service/internal/domain/entity"
)

// MockCartRepository for testing
type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) CreateCart(ctx context.Context, cart *entity.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockCartRepository) GetCartByID(ctx context.Context, cartID uuid.UUID) (*entity.Cart, error) {
	args := m.Called(ctx, cartID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) GetCartBySessionID(ctx context.Context, sessionID string) (*entity.Cart, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartRepository) UpdateCart(ctx context.Context, cart *entity.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockCartRepository) DeleteCart(ctx context.Context, cartID uuid.UUID) error {
	args := m.Called(ctx, cartID)
	return args.Error(0)
}

func (m *MockCartRepository) DeleteExpiredCarts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockCartItemRepository for testing
type MockCartItemRepository struct {
	mock.Mock
}

func (m *MockCartItemRepository) CreateCartItem(ctx context.Context, item *entity.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartItemRepository) GetCartItemByID(ctx context.Context, itemID uuid.UUID) (*entity.CartItem, error) {
	args := m.Called(ctx, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

func (m *MockCartItemRepository) GetCartItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]*entity.CartItem, error) {
	args := m.Called(ctx, cartID)
	return args.Get(0).([]*entity.CartItem), args.Error(1)
}

func (m *MockCartItemRepository) GetCartItemByProductAndCart(ctx context.Context, cartID, productID uuid.UUID, variantID *uuid.UUID) (*entity.CartItem, error) {
	args := m.Called(ctx, cartID, productID, variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

func (m *MockCartItemRepository) UpdateCartItem(ctx context.Context, item *entity.CartItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockCartItemRepository) DeleteCartItem(ctx context.Context, itemID uuid.UUID) error {
	args := m.Called(ctx, itemID)
	return args.Error(0)
}

func (m *MockCartItemRepository) DeleteCartItemsByCartID(ctx context.Context, cartID uuid.UUID) error {
	args := m.Called(ctx, cartID)
	return args.Error(0)
}

// MockProductService for testing
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) ValidateProduct(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (*ProductInfo, error) {
	args := m.Called(ctx, productID, variantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductInfo), args.Error(1)
}

func (m *MockProductService) CheckStock(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, quantity int) (bool, error) {
	args := m.Called(ctx, productID, variantID, quantity)
	return args.Bool(0), args.Error(1)
}

func TestCartService_CreateCart(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockItemRepo := new(MockCartItemRepository)
	mockProductService := new(MockProductService)

	cartService := NewCartService(mockCartRepo, mockItemRepo, mockProductService)
	ctx := context.Background()

	t.Run("successful cart creation with user", func(t *testing.T) {
		userID := uuid.New()
		request := &CreateCartRequest{
			UserID: &userID,
		}

		mockCartRepo.On("CreateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		response, err := cartService.CreateCart(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, *response.UserID)
		assert.NotEmpty(t, response.SessionID)

		mockCartRepo.AssertExpectations(t)
	})

	t.Run("successful cart creation with session only", func(t *testing.T) {
		request := &CreateCartRequest{
			SessionID: "test-session-123",
		}

		mockCartRepo.On("CreateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		response, err := cartService.CreateCart(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Nil(t, response.UserID)
		assert.Equal(t, "test-session-123", response.SessionID)

		mockCartRepo.AssertExpectations(t)
	})
}

func TestCartService_GetCart(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockItemRepo := new(MockCartItemRepository)
	mockProductService := new(MockProductService)

	cartService := NewCartService(mockCartRepo, mockItemRepo, mockProductService)
	ctx := context.Background()

	t.Run("successful cart retrieval by ID", func(t *testing.T) {
		cartID := uuid.New()
		userID := uuid.New()

		cart := &entity.Cart{
			ID:          cartID,
			UserID:      &userID,
			SessionID:   "session-123",
			TotalAmount: 299.98,
			ItemCount:   2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		cartItems := []*entity.CartItem{
			{
				ID:        uuid.New(),
				CartID:    cartID,
				ProductID: uuid.New(),
				Quantity:  1,
				Price:     149.99,
			},
			{
				ID:        uuid.New(),
				CartID:    cartID,
				ProductID: uuid.New(),
				Quantity:  1,
				Price:     149.99,
			},
		}

		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockItemRepo.On("GetCartItemsByCartID", ctx, cartID).Return(cartItems, nil)

		response, err := cartService.GetCart(ctx, cartID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, cartID, response.ID)
		assert.Equal(t, userID, *response.UserID)
		assert.Len(t, response.Items, 2)
		assert.Equal(t, 299.98, response.TotalAmount)

		mockCartRepo.AssertExpectations(t)
		mockItemRepo.AssertExpectations(t)
	})

	t.Run("cart not found", func(t *testing.T) {
		cartID := uuid.New()

		mockCartRepo.On("GetCartByID", ctx, cartID).Return((*entity.Cart)(nil), entity.ErrCartNotFound)

		response, err := cartService.GetCart(ctx, cartID)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrCartNotFound, err)

		mockCartRepo.AssertExpectations(t)
	})
}

func TestCartService_AddItemToCart(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockItemRepo := new(MockCartItemRepository)
	mockProductService := new(MockProductService)

	cartService := NewCartService(mockCartRepo, mockItemRepo, mockProductService)
	ctx := context.Background()

	t.Run("successful item addition to cart", func(t *testing.T) {
		cartID := uuid.New()
		productID := uuid.New()

		cart := &entity.Cart{
			ID:          cartID,
			SessionID:   "session-123",
			TotalAmount: 0,
			ItemCount:   0,
		}

		request := &AddCartItemRequest{
			ProductID: productID,
			Quantity:  2,
		}

		productInfo := &ProductInfo{
			ID:       productID,
			Name:     "Nike Air Max",
			Price:    149.99,
			IsActive: true,
		}

		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockProductService.On("ValidateProduct", ctx, productID, (*uuid.UUID)(nil)).Return(productInfo, nil)
		mockProductService.On("CheckStock", ctx, productID, (*uuid.UUID)(nil), 2).Return(true, nil)
		mockItemRepo.On("GetCartItemByProductAndCart", ctx, cartID, productID, (*uuid.UUID)(nil)).Return((*entity.CartItem)(nil), entity.ErrCartItemNotFound)
		mockItemRepo.On("CreateCartItem", ctx, mock.AnythingOfType("*entity.CartItem")).Return(nil)
		mockCartRepo.On("UpdateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		response, err := cartService.AddItemToCart(ctx, cartID, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, productID, response.ProductID)
		assert.Equal(t, 2, response.Quantity)
		assert.Equal(t, 149.99, response.Price)

		mockCartRepo.AssertExpectations(t)
		mockItemRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
	})

	t.Run("update existing item quantity", func(t *testing.T) {
		cartID := uuid.New()
		productID := uuid.New()
		itemID := uuid.New()

		cart := &entity.Cart{
			ID:          cartID,
			SessionID:   "session-123",
			TotalAmount: 149.99,
			ItemCount:   1,
		}

		existingItem := &entity.CartItem{
			ID:        itemID,
			CartID:    cartID,
			ProductID: productID,
			Quantity:  1,
			Price:     149.99,
		}

		request := &AddCartItemRequest{
			ProductID: productID,
			Quantity:  2,
		}

		productInfo := &ProductInfo{
			ID:       productID,
			Name:     "Nike Air Max",
			Price:    149.99,
			IsActive: true,
		}

		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockProductService.On("ValidateProduct", ctx, productID, (*uuid.UUID)(nil)).Return(productInfo, nil)
		mockProductService.On("CheckStock", ctx, productID, (*uuid.UUID)(nil), 3).Return(true, nil) // 1 existing + 2 new
		mockItemRepo.On("GetCartItemByProductAndCart", ctx, cartID, productID, (*uuid.UUID)(nil)).Return(existingItem, nil)
		mockItemRepo.On("UpdateCartItem", ctx, mock.AnythingOfType("*entity.CartItem")).Return(nil)
		mockCartRepo.On("UpdateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		response, err := cartService.AddItemToCart(ctx, cartID, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, productID, response.ProductID)
		assert.Equal(t, 3, response.Quantity) // 1 + 2

		mockCartRepo.AssertExpectations(t)
		mockItemRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
	})

	t.Run("insufficient stock", func(t *testing.T) {
		cartID := uuid.New()
		productID := uuid.New()

		cart := &entity.Cart{
			ID:        cartID,
			SessionID: "session-123",
		}

		request := &AddCartItemRequest{
			ProductID: productID,
			Quantity:  10,
		}

		productInfo := &ProductInfo{
			ID:       productID,
			Name:     "Nike Air Max",
			Price:    149.99,
			IsActive: true,
		}

		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockProductService.On("ValidateProduct", ctx, productID, (*uuid.UUID)(nil)).Return(productInfo, nil)
		mockProductService.On("CheckStock", ctx, productID, (*uuid.UUID)(nil), 10).Return(false, nil)

		response, err := cartService.AddItemToCart(ctx, cartID, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "insufficient stock")

		mockCartRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
	})
}

func TestCartService_UpdateCartItem(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockItemRepo := new(MockCartItemRepository)
	mockProductService := new(MockProductService)

	cartService := NewCartService(mockCartRepo, mockItemRepo, mockProductService)
	ctx := context.Background()

	t.Run("successful item quantity update", func(t *testing.T) {
		itemID := uuid.New()
		cartID := uuid.New()
		productID := uuid.New()

		cartItem := &entity.CartItem{
			ID:        itemID,
			CartID:    cartID,
			ProductID: productID,
			Quantity:  2,
			Price:     149.99,
		}

		cart := &entity.Cart{
			ID:          cartID,
			TotalAmount: 299.98,
			ItemCount:   1,
		}

		request := &UpdateCartItemRequest{
			Quantity: 3,
		}

		mockItemRepo.On("GetCartItemByID", ctx, itemID).Return(cartItem, nil)
		mockProductService.On("CheckStock", ctx, productID, (*uuid.UUID)(nil), 3).Return(true, nil)
		mockItemRepo.On("UpdateCartItem", ctx, mock.AnythingOfType("*entity.CartItem")).Return(nil)
		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockCartRepo.On("UpdateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		response, err := cartService.UpdateCartItem(ctx, itemID, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 3, response.Quantity)

		mockItemRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("item not found", func(t *testing.T) {
		itemID := uuid.New()

		request := &UpdateCartItemRequest{
			Quantity: 3,
		}

		mockItemRepo.On("GetCartItemByID", ctx, itemID).Return((*entity.CartItem)(nil), entity.ErrCartItemNotFound)

		response, err := cartService.UpdateCartItem(ctx, itemID, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrCartItemNotFound, err)

		mockItemRepo.AssertExpectations(t)
	})
}

func TestCartService_RemoveItemFromCart(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockItemRepo := new(MockCartItemRepository)
	mockProductService := new(MockProductService)

	cartService := NewCartService(mockCartRepo, mockItemRepo, mockProductService)
	ctx := context.Background()

	t.Run("successful item removal", func(t *testing.T) {
		itemID := uuid.New()
		cartID := uuid.New()

		cartItem := &entity.CartItem{
			ID:       itemID,
			CartID:   cartID,
			Quantity: 2,
			Price:    149.99,
		}

		cart := &entity.Cart{
			ID:          cartID,
			TotalAmount: 299.98,
			ItemCount:   1,
		}

		mockItemRepo.On("GetCartItemByID", ctx, itemID).Return(cartItem, nil)
		mockItemRepo.On("DeleteCartItem", ctx, itemID).Return(nil)
		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockCartRepo.On("UpdateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		err := cartService.RemoveItemFromCart(ctx, itemID)

		assert.NoError(t, err)

		mockItemRepo.AssertExpectations(t)
		mockCartRepo.AssertExpectations(t)
	})

	t.Run("item not found", func(t *testing.T) {
		itemID := uuid.New()

		mockItemRepo.On("GetCartItemByID", ctx, itemID).Return((*entity.CartItem)(nil), entity.ErrCartItemNotFound)

		err := cartService.RemoveItemFromCart(ctx, itemID)

		assert.Error(t, err)
		assert.Equal(t, entity.ErrCartItemNotFound, err)

		mockItemRepo.AssertExpectations(t)
	})
}

func TestCartService_ClearCart(t *testing.T) {
	mockCartRepo := new(MockCartRepository)
	mockItemRepo := new(MockCartItemRepository)
	mockProductService := new(MockProductService)

	cartService := NewCartService(mockCartRepo, mockItemRepo, mockProductService)
	ctx := context.Background()

	t.Run("successful cart clear", func(t *testing.T) {
		cartID := uuid.New()

		cart := &entity.Cart{
			ID:          cartID,
			TotalAmount: 299.98,
			ItemCount:   2,
		}

		mockCartRepo.On("GetCartByID", ctx, cartID).Return(cart, nil)
		mockItemRepo.On("DeleteCartItemsByCartID", ctx, cartID).Return(nil)
		mockCartRepo.On("UpdateCart", ctx, mock.AnythingOfType("*entity.Cart")).Return(nil)

		err := cartService.ClearCart(ctx, cartID)

		assert.NoError(t, err)

		mockCartRepo.AssertExpectations(t)
		mockItemRepo.AssertExpectations(t)
	})

	t.Run("cart not found", func(t *testing.T) {
		cartID := uuid.New()

		mockCartRepo.On("GetCartByID", ctx, cartID).Return((*entity.Cart)(nil), entity.ErrCartNotFound)

		err := cartService.ClearCart(ctx, cartID)

		assert.Error(t, err)
		assert.Equal(t, entity.ErrCartNotFound, err)

		mockCartRepo.AssertExpectations(t)
	})
}