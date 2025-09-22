package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"solemate/services/user-service/internal/domain/entity"
)

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetUsersByRole(ctx context.Context, role entity.Role) ([]*entity.User, error) {
	args := m.Called(ctx, role)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MockJWTService for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokens(userID uuid.UUID, email string, role entity.Role) (*TokenPair, error) {
	args := m.Called(userID, email, role)
	return args.Get(0).(*TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *MockJWTService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	args := m.Called(refreshToken)
	return args.Get(0).(*TokenPair), args.Error(1)
}

func TestUserService_RegisterUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := NewUserService(mockRepo, mockJWT)

	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		request := &RegisterUserRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
		}

		// Mock email doesn't exist
		mockRepo.On("EmailExists", ctx, request.Email).Return(false, nil)

		// Mock successful user creation
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

		// Mock token generation
		tokens := &TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}
		mockRepo.On("GetUserByEmail", ctx, request.Email).Return(&entity.User{
			ID:        uuid.New(),
			Email:     request.Email,
			FirstName: request.FirstName,
			LastName:  request.LastName,
			Role:      entity.RoleCustomer,
		}, nil)
		mockJWT.On("GenerateTokens", mock.AnythingOfType("uuid.UUID"), request.Email, entity.RoleCustomer).Return(tokens, nil)

		response, err := userService.RegisterUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, request.Email, response.User.Email)
		assert.Equal(t, "access_token", response.AccessToken)
		assert.Equal(t, "refresh_token", response.RefreshToken)

		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		request := &RegisterUserRequest{
			Email:     "existing@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		}

		mockRepo.On("EmailExists", ctx, request.Email).Return(true, nil)

		response, err := userService.RegisterUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "email already exists")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_LoginUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := NewUserService(mockRepo, mockJWT)

	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := &entity.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
			FirstName:    "John",
			LastName:     "Doe",
			Role:         entity.RoleCustomer,
			IsActive:     true,
		}

		request := &LoginUserRequest{
			Email:    email,
			Password: password,
		}

		mockRepo.On("GetUserByEmail", ctx, email).Return(user, nil)
		mockRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)

		tokens := &TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}
		mockJWT.On("GenerateTokens", user.ID, user.Email, user.Role).Return(tokens, nil)

		response, err := userService.LoginUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.Email, response.User.Email)
		assert.Equal(t, "access_token", response.AccessToken)

		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})

	t.Run("invalid email", func(t *testing.T) {
		request := &LoginUserRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		mockRepo.On("GetUserByEmail", ctx, request.Email).Return((*entity.User)(nil), entity.ErrUserNotFound)

		response, err := userService.LoginUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrInvalidCredentials, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

		user := &entity.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
			IsActive:     true,
		}

		request := &LoginUserRequest{
			Email:    email,
			Password: "wrong_password",
		}

		mockRepo.On("GetUserByEmail", ctx, email).Return(user, nil)

		response, err := userService.LoginUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrInvalidCredentials, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("inactive user", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := &entity.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
			IsActive:     false,
		}

		request := &LoginUserRequest{
			Email:    email,
			Password: password,
		}

		mockRepo.On("GetUserByEmail", ctx, email).Return(user, nil)

		response, err := userService.LoginUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "account is deactivated")

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := NewUserService(mockRepo, mockJWT)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful profile retrieval", func(t *testing.T) {
		user := &entity.User{
			ID:        userID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      entity.RoleCustomer,
			IsActive:  true,
			CreatedAt: time.Now(),
		}

		mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)

		response, err := userService.GetUserProfile(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.Email, response.Email)
		assert.Equal(t, user.FirstName, response.FirstName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetUserByID", ctx, userID).Return((*entity.User)(nil), entity.ErrUserNotFound)

		response, err := userService.GetUserProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrUserNotFound, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUserProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := NewUserService(mockRepo, mockJWT)

	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful profile update", func(t *testing.T) {
		existingUser := &entity.User{
			ID:        userID,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Phone:     "+1234567890",
		}

		request := &UpdateUserProfileRequest{
			FirstName: stringPtr("Jane"),
			LastName:  stringPtr("Smith"),
			Phone:     stringPtr("+0987654321"),
		}

		mockRepo.On("GetUserByID", ctx, userID).Return(existingUser, nil)
		mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*entity.User")).Return(nil)

		response, err := userService.UpdateUserProfile(ctx, userID, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Jane", response.FirstName)
		assert.Equal(t, "Smith", response.LastName)
		assert.Equal(t, "+0987654321", response.Phone)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_RefreshTokens(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	userService := NewUserService(mockRepo, mockJWT)

	ctx := context.Background()

	t.Run("successful token refresh", func(t *testing.T) {
		refreshToken := "valid_refresh_token"
		newTokens := &TokenPair{
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
		}

		mockJWT.On("RefreshTokens", refreshToken).Return(newTokens, nil)

		response, err := userService.RefreshTokens(ctx, refreshToken)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "new_access_token", response.AccessToken)
		assert.Equal(t, "new_refresh_token", response.RefreshToken)

		mockJWT.AssertExpectations(t)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshToken := "invalid_refresh_token"

		mockJWT.On("RefreshTokens", refreshToken).Return((*TokenPair)(nil), entity.ErrInvalidToken)

		response, err := userService.RefreshTokens(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, entity.ErrInvalidToken, err)

		mockJWT.AssertExpectations(t)
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}