package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"solemate/pkg/auth"
	"solemate/pkg/utils"
	"solemate/services/user-service/internal/domain/entity"
	"solemate/services/user-service/internal/domain/repository"
)

type UserService struct {
	userRepo    repository.UserRepository
	addressRepo repository.AddressRepository
	jwtManager  *auth.JWTManager
}

func NewUserService(userRepo repository.UserRepository, addressRepo repository.AddressRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		jwtManager:  jwtManager,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone_number"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         *entity.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type UpdateUserRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	PhoneNumber *string `json:"phone_number"`
}

func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*entity.User, error) {
	// Validate input
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	if !utils.IsValidPassword(req.Password) {
		return nil, errors.New("password must be at least 8 characters")
	}

	if req.Phone != "" && !utils.IsValidPhoneNumber(req.Phone) {
		return nil, errors.New("invalid phone number format")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entity.User{
		Email:        utils.SanitizeString(req.Email),
		PasswordHash: string(hashedPassword),
		FirstName:    utils.SanitizeString(req.FirstName),
		LastName:     utils.SanitizeString(req.LastName),
		PhoneNumber:  utils.SanitizeString(req.Phone),
		Role:         "customer",
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		user.FirstName = utils.SanitizeString(*req.FirstName)
	}

	if req.LastName != nil {
		user.LastName = utils.SanitizeString(*req.LastName)
	}

	if req.PhoneNumber != nil {
		if *req.PhoneNumber != "" && !utils.IsValidPhoneNumber(*req.PhoneNumber) {
			return nil, errors.New("invalid phone number format")
		}
		user.PhoneNumber = utils.SanitizeString(*req.PhoneNumber)
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit
	return s.userRepo.List(ctx, limit, offset)
}

func (s *UserService) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	return s.jwtManager.ValidateAccessToken(token)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	return s.jwtManager.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}
