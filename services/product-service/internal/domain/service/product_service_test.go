package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"solemate/services/product-service/internal/domain/entity"
)

// MockProductRepository for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) CreateProduct(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetProductByID(ctx context.Context, productID uuid.UUID) (*entity.Product, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) GetProductBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateProduct(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockProductRepository) SearchProducts(ctx context.Context, filters *ProductFilters) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, limit, offset int) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, categoryID, limit, offset)
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetProductsByBrand(ctx context.Context, brandID uuid.UUID, limit, offset int) ([]*entity.Product, int64, error) {
	args := m.Called(ctx, brandID, limit, offset)
	return args.Get(0).([]*entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetFeaturedProducts(ctx context.Context, limit int) ([]*entity.Product, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateProductStock(ctx context.Context, productID uuid.UUID, stockChange int) error {
	args := m.Called(ctx, productID, stockChange)
	return args.Error(0)
}

// MockCategoryRepository for testing
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) CreateCategory(ctx context.Context, category *entity.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*entity.Category, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetAllCategories(ctx context.Context) ([]*entity.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Category), args.Error(1)
}

func (m *MockCategoryRepository) UpdateCategory(ctx context.Context, category *entity.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepository) DeleteCategory(ctx context.Context, categoryID uuid.UUID) error {
	args := m.Called(ctx, categoryID)
	return args.Error(0)
}

// MockBrandRepository for testing
type MockBrandRepository struct {
	mock.Mock
}

func (m *MockBrandRepository) CreateBrand(ctx context.Context, brand *entity.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandRepository) GetBrandByID(ctx context.Context, brandID uuid.UUID) (*entity.Brand, error) {
	args := m.Called(ctx, brandID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Brand), args.Error(1)
}

func (m *MockBrandRepository) GetAllBrands(ctx context.Context) ([]*entity.Brand, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Brand), args.Error(1)
}

func (m *MockBrandRepository) UpdateBrand(ctx context.Context, brand *entity.Brand) error {
	args := m.Called(ctx, brand)
	return args.Error(0)
}

func (m *MockBrandRepository) DeleteBrand(ctx context.Context, brandID uuid.UUID) error {
	args := m.Called(ctx, brandID)
	return args.Error(0)
}

func TestProductService_CreateProduct(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockBrandRepo := new(MockBrandRepository)

	productService := NewProductService(mockProductRepo, mockCategoryRepo, mockBrandRepo)
	ctx := context.Background()

	t.Run("successful product creation", func(t *testing.T) {
		categoryID := uuid.New()
		brandID := uuid.New()

		request := &CreateProductRequest{
			Name:        "Nike Air Max",
			Description: "Comfortable running shoes",
			SKU:         "NIKE-AM-001",
			Price:       149.99,
			CategoryID:  categoryID,
			BrandID:     brandID,
			Stock:       100,
			IsActive:    true,
		}

		// Mock category and brand exist
		category := &entity.Category{ID: categoryID, Name: "Running Shoes"}
		brand := &entity.Brand{ID: brandID, Name: "Nike"}

		mockCategoryRepo.On("GetCategoryByID", ctx, categoryID).Return(category, nil)
		mockBrandRepo.On("GetBrandByID", ctx, brandID).Return(brand, nil)
		mockProductRepo.On("GetProductBySKU", ctx, request.SKU).Return((*entity.Product)(nil), entity.ErrProductNotFound)
		mockProductRepo.On("CreateProduct", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)

		response, err := productService.CreateProduct(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, request.Name, response.Name)
		assert.Equal(t, request.SKU, response.SKU)
		assert.Equal(t, request.Price, response.Price)

		mockProductRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
		mockBrandRepo.AssertExpectations(t)
	})

	t.Run("duplicate SKU error", func(t *testing.T) {
		categoryID := uuid.New()
		brandID := uuid.New()

		request := &CreateProductRequest{
			Name:       "Nike Air Max",
			SKU:        "EXISTING-SKU",
			CategoryID: categoryID,
			BrandID:    brandID,
		}

		existingProduct := &entity.Product{
			ID:  uuid.New(),
			SKU: "EXISTING-SKU",
		}

		mockProductRepo.On("GetProductBySKU", ctx, request.SKU).Return(existingProduct, nil)

		response, err := productService.CreateProduct(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "product with SKU already exists")

		mockProductRepo.AssertExpectations(t)
	})

	t.Run("invalid category", func(t *testing.T) {
		categoryID := uuid.New()
		brandID := uuid.New()

		request := &CreateProductRequest{
			Name:       "Nike Air Max",
			SKU:        "NIKE-AM-001",
			CategoryID: categoryID,
			BrandID:    brandID,
		}

		mockCategoryRepo.On("GetCategoryByID", ctx, categoryID).Return((*entity.Category)(nil), entity.ErrCategoryNotFound)

		response, err := productService.CreateProduct(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "category not found")

		mockCategoryRepo.AssertExpectations(t)
	})
}

func TestProductService_GetProduct(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockBrandRepo := new(MockBrandRepository)

	productService := NewProductService(mockProductRepo, mockCategoryRepo, mockBrandRepo)
	ctx := context.Background()

	t.Run("successful product retrieval", func(t *testing.T) {
		productID := uuid.New()
		categoryID := uuid.New()
		brandID := uuid.New()

		product := &entity.Product{
			ID:          productID,
			Name:        "Nike Air Max",
			Description: "Comfortable running shoes",
			SKU:         "NIKE-AM-001",
			Price:       149.99,
			CategoryID:  categoryID,
			BrandID:     brandID,
			Stock:       100,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockProductRepo.On("GetProductByID", ctx, productID).Return(product, nil)

		response, err := productService.GetProduct(ctx, productID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, product.Name, response.Name)
		assert.Equal(t, product.SKU, response.SKU)
		assert.Equal(t, product.Price, response.Price)

		mockProductRepo.AssertExpectations(t)
	})

	t.Run("product not found", func(t *testing.T) {
		productID := uuid.New()

		mockProductRepo.On("GetProductByID", ctx, productID).Return((*entity.Product)(nil), entity.ErrProductNotFound)

		response, err := productService.GetProduct(ctx, productID)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrProductNotFound, err)

		mockProductRepo.AssertExpectations(t)
	})
}

func TestProductService_SearchProducts(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockBrandRepo := new(MockBrandRepository)

	productService := NewProductService(mockProductRepo, mockCategoryRepo, mockBrandRepo)
	ctx := context.Background()

	t.Run("successful product search", func(t *testing.T) {
		request := &SearchProductsRequest{
			Query:    "Nike",
			MinPrice: floatPtr(50.0),
			MaxPrice: floatPtr(200.0),
			Limit:    10,
			Offset:   0,
		}

		products := []*entity.Product{
			{
				ID:    uuid.New(),
				Name:  "Nike Air Max",
				Price: 149.99,
			},
			{
				ID:    uuid.New(),
				Name:  "Nike Free Run",
				Price: 89.99,
			},
		}

		filters := &ProductFilters{
			Query:    request.Query,
			MinPrice: request.MinPrice,
			MaxPrice: request.MaxPrice,
			Limit:    request.Limit,
			Offset:   request.Offset,
		}

		mockProductRepo.On("SearchProducts", ctx, mock.MatchedBy(func(f *ProductFilters) bool {
			return f.Query == filters.Query &&
				   *f.MinPrice == *filters.MinPrice &&
				   *f.MaxPrice == *filters.MaxPrice
		})).Return(products, int64(2), nil)

		response, err := productService.SearchProducts(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Products, 2)
		assert.Equal(t, int64(2), response.Total)

		mockProductRepo.AssertExpectations(t)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockBrandRepo := new(MockBrandRepository)

	productService := NewProductService(mockProductRepo, mockCategoryRepo, mockBrandRepo)
	ctx := context.Background()

	t.Run("successful product update", func(t *testing.T) {
		productID := uuid.New()
		categoryID := uuid.New()
		brandID := uuid.New()

		existingProduct := &entity.Product{
			ID:         productID,
			Name:       "Nike Air Max",
			Price:      149.99,
			CategoryID: categoryID,
			BrandID:    brandID,
			Stock:      100,
		}

		request := &UpdateProductRequest{
			Name:  stringPtr("Nike Air Max 2023"),
			Price: floatPtr(159.99),
			Stock: intPtr(150),
		}

		mockProductRepo.On("GetProductByID", ctx, productID).Return(existingProduct, nil)
		mockProductRepo.On("UpdateProduct", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)

		response, err := productService.UpdateProduct(ctx, productID, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Nike Air Max 2023", response.Name)
		assert.Equal(t, 159.99, response.Price)
		assert.Equal(t, 150, response.Stock)

		mockProductRepo.AssertExpectations(t)
	})

	t.Run("product not found", func(t *testing.T) {
		productID := uuid.New()

		request := &UpdateProductRequest{
			Name: stringPtr("Updated Product"),
		}

		mockProductRepo.On("GetProductByID", ctx, productID).Return((*entity.Product)(nil), entity.ErrProductNotFound)

		response, err := productService.UpdateProduct(ctx, productID, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrProductNotFound, err)

		mockProductRepo.AssertExpectations(t)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	mockProductRepo := new(MockProductRepository)
	mockCategoryRepo := new(MockCategoryRepository)
	mockBrandRepo := new(MockBrandRepository)

	productService := NewProductService(mockProductRepo, mockCategoryRepo, mockBrandRepo)
	ctx := context.Background()

	t.Run("successful product deletion", func(t *testing.T) {
		productID := uuid.New()

		mockProductRepo.On("DeleteProduct", ctx, productID).Return(nil)

		err := productService.DeleteProduct(ctx, productID)

		assert.NoError(t, err)

		mockProductRepo.AssertExpectations(t)
	})

	t.Run("product not found", func(t *testing.T) {
		productID := uuid.New()

		mockProductRepo.On("DeleteProduct", ctx, productID).Return(entity.ErrProductNotFound)

		err := productService.DeleteProduct(ctx, productID)

		assert.Error(t, err)
		assert.Equal(t, entity.ErrProductNotFound, err)

		mockProductRepo.AssertExpectations(t)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}