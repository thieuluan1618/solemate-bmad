# Makefile
.PHONY: help build run test clean docker-build docker-up docker-down migrate

# Variables
SERVICES = user-service product-service cart-service order-service payment-service
GO = go
GOFLAGS = -v
DOCKER_COMPOSE = docker-compose

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $1, $2}' $(MAKEFILE_LIST)

build: ## Build all services
	@for service in $(SERVICES); do \
		echo "Building $service..."; \
		cd services/$service && $(GO) build $(GOFLAGS) -o bin/$service ./cmd/main.go && cd ../..; \
	done

run-%: ## Run a specific service (e.g., make run-user-service)
	@echo "Running $*..."
	@cd services/$* && $(GO) run ./cmd/main.go

test: ## Run all tests
	@echo "Running tests..."
	@$(GO) test -v -cover -race ./...

test-service: ## Test a specific service (e.g., make test-service SERVICE=user-service)
	@echo "Testing $(SERVICE)..."
	@cd services/$(SERVICE) && $(GO) test -v -cover -race ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@goimports -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf services/*/bin
	@$(GO) clean -cache

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## View logs for all services
	@$(DOCKER_COMPOSE) logs -f

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path ./migrations -database "postgresql://solemate:password@localhost:5432/solemate_db?sslmode=disable" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path ./migrations -database "postgresql://solemate:password@localhost:5432/solemate_db?sslmode=disable" down

proto: ## Generate protobuf files
	@echo "Generating proto files..."
	@protoc --go_out=. --go-grpc_out=. proto/*.proto

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

.DEFAULT_GOAL := help
```

---

## 13. gRPC Service Definitions

### 13.1 Proto Files

```protobuf
// proto/user.proto
syntax = "proto3";

package user;
option go_package = "solemate/proto/user";

import "google/protobuf/timestamp.proto";

service UserService {
    rpc GetUser(GetUserRequest) returns (UserResponse);
    rpc CreateUser(CreateUserRequest) returns (UserResponse);
    rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
    rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

message GetUserRequest {
    string user_id = 1;
}

message CreateUserRequest {
    string email = 1;
    string password = 2;
    string first_name = 3;
    string last_name = 4;
}

message UpdateUserRequest {
    string user_id = 1;
    optional string first_name = 2;
    optional string last_name = 3;
    optional string phone_number = 4;
}

message DeleteUserRequest {
    string user_id = 1;
}

message DeleteUserResponse {
    bool success = 1;
}

message UserResponse {
    string id = 1;
    string email = 2;
    string first_name = 3;
    string last_name = 4;
    string phone_number = 5;
    string role = 6;
    bool is_active = 7;
    bool email_verified = 8;
    google.protobuf.Timestamp created_at = 9;
    google.protobuf.Timestamp updated_at = 10;
}

message ValidateTokenRequest {
    string token = 1;
}

message ValidateTokenResponse {
    bool valid = 1;
    string user_id = 2;
    string email = 3;
    string role = 4;
}
```

### 13.2 gRPC Server Implementation

```go
// internal/handler/grpc/user_grpc_handler.go
package grpc

import (
    "context"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "solemate/proto/user"
    "solemate/services/user-service/internal/domain/service"
)

type UserGRPCHandler struct {
    user.UnimplementedUserServiceServer
    userService *service.UserService
}

func NewUserGRPCHandler(userService *service.UserService) *UserGRPCHandler {
    return &UserGRPCHandler{
        userService: userService,
    }
}

func (h *UserGRPCHandler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
    userID, err := uuid.Parse(req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
    }
    
    u, err := h.userService.GetUserByID(ctx, userID)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
    }
    
    return &user.UserResponse{
        Id:            u.ID.String(),
        Email:         u.Email,
        FirstName:     u.FirstName,
        LastName:      u.LastName,
        PhoneNumber:   u.PhoneNumber,
        Role:          string(u.Role),
        IsActive:      u.IsActive,
        EmailVerified: u.EmailVerified,
        CreatedAt:     timestamppb.New(u.CreatedAt),
        UpdatedAt:     timestamppb.New(u.UpdatedAt),
    }, nil
}

func (h *UserGRPCHandler) ValidateToken(ctx context.Context, req *user.ValidateTokenRequest) (*user.ValidateTokenResponse, error) {
    claims, err := h.userService.ValidateToken(ctx, req.Token)
    if err != nil {
        return &user.ValidateTokenResponse{
            Valid: false,
        }, nil
    }
    
    return &user.ValidateTokenResponse{
        Valid:  true,
        UserId: claims.UserID.String(),
        Email:  claims.Email,
        Role:   claims.Role,
    }, nil
}
```

---

## 14. Testing Strategy

### 14.1 Unit Tests

```go
// internal/domain/service/user_service_test.go
package service_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "solemate/services/user-service/internal/domain/entity"
    "solemate/services/user-service/internal/domain/service"
    "solemate/services/user-service/mocks"
)

func TestUserService_RegisterUser(t *testing.T) {
    // Setup
    mockRepo := new(mocks.UserRepository)
    mockCache := new(mocks.Cache)
    mockJWT := new(mocks.JWTService)
    mockValidator := new(mocks.Validator)
    
    userService := service.NewUserService(mockRepo, mockCache, mockJWT, mockValidator)
    
    // Test data
    req := &service.RegisterRequest{
        Email:     "test@example.com",
        Password:  "SecurePass123!",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    // Mock expectations
    mockValidator.On("Validate", req).Return(nil)
    mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, errors.New("not found"))
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
    mockJWT.On("GenerateAccessToken", mock.AnythingOfType("*entity.User")).Return("access-token", nil)
    mockJWT.On("GenerateRefreshToken", mock.AnythingOfType("*entity.User")).Return("refresh-token", nil)
    mockRepo.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("*entity.RefreshToken")).Return(nil)
    mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
    
    // Execute
    result, err := userService.RegisterUser(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "test@example.com", result.User.Email)
    assert.Equal(t, "access-token", result.AccessToken)
    assert.Equal(t, "refresh-token", result.RefreshToken)
    
    // Verify mock expectations
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
    mockJWT.AssertExpectations(t)
    mockValidator.AssertExpectations(t)
}

func TestUserService_Login_InvalidCredentials(t *testing.T) {
    // Setup
    mockRepo := new(mocks.UserRepository)
    mockCache := new(mocks.Cache)
    mockJWT := new(mocks.JWTService)
    mockValidator := new(mocks.Validator)
    
    userService := service.NewUserService(mockRepo, mockCache, mockJWT, mockValidator)
    
    // Test data
    req := &service.LoginRequest{
        Email:    "test@example.com",
        Password: "WrongPassword",
    }
    
    user := &entity.User{
        ID:           uuid.New(),
        Email:        req.Email,
        PasswordHash: "$2a$10$validhash",
        IsActive:     true,
    }
    
    // Mock expectations
    mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(user, nil)
    
    // Execute
    result, err := userService.Login(context.Background(), req)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Equal(t, "invalid credentials", err.Error())
}
```

### 14.2 Integration Tests

```go
// internal/handler/http/user_handler_test.go
package http_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "solemate/services/user-service/internal/handler/http"
)

func TestUserHandler_Register(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    // Initialize handler with mocked service
    userHandler := setupUserHandler()
    userHandler.RegisterRoutes(router.Group("/api/v1"))
    
    // Test data
    payload := map[string]string{
        "email":      "test@example.com",
        "password":   "SecurePass123!",
        "first_name": "John",
        "last_name":  "Doe",
    }
    
    body, _ := json.Marshal(payload)
    
    // Create request
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Contains(t, response, "user")
    assert.Contains(t, response, "access_token")
    assert.Contains(t, response, "refresh_token")
}
```

### 14.3 Load Testing

```go
// tests/load/load_test.go
package load_test

import (
    "testing"
    "time"
    
    vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestLoadUserRegistration(t *testing.T) {
    rate := vegeta.Rate{Freq: 100, Per: time.Second} // 100 requests per second
    duration := 30 * time.Second
    
    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "POST",
        URL:    "http://localhost:8080/api/v1/auth/register",
        Body:   []byte(`{"email":"test@example.com","password":"Test123!"}`),
        Header: http.Header{
            "Content-Type": []string{"application/json"},
        },
    })
    
    attacker := vegeta.NewAttacker()
    
    var metrics vegeta.Metrics
    for res := range attacker.Attack(targeter, rate, duration, "Load Test") {
        metrics.Add(res)
    }
    metrics.Close()
    
    // Assert performance requirements
    assert.True(t, metrics.Latencies.P95 < 500*time.Millisecond, "95th percentile latency should be under 500ms")
    assert.True(t, metrics.Success > 0.95, "Success rate should be above 95%")
}
```

---

## 15. Monitoring and Logging

### 15.1 Structured Logging

```go
// pkg/logger/logger.go
package logger

import (
    "os"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(level string) {
    var err error
    
    config := zap.NewProductionConfig()
    
    // Set log level
    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
    default:
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    }
    
    // Configure encoder
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    config.EncoderConfig.StacktraceKey = "stacktrace"
    
    log, err = config.Build(zap.AddCallerSkip(1))
    if err != nil {
        panic(err)
    }
}

func Info(msg string, fields ...zap.Field) {
    log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
    log.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
    log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
    log.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
    log.Fatal(msg, fields...)
}
```

### 15.2 Prometheus Metrics

```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests in seconds",
        },
        []string{"method", "endpoint", "status"},
    )
    
    RequestTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    DatabaseQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_query_duration_seconds",
            Help: "Duration of database queries in seconds",
        },
        []string{"query_type", "table"},
    )
    
    CacheHitRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"cache_type"},
    )
    
    BusinessMetrics = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "business_events_total",
            Help: "Total number of business events",
        },
        []string{"event_type"},
    )
)

// Middleware for Gin
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
            RequestDuration.WithLabelValues(
                c.Request.Method,
                c.FullPath(),
                string(c.Writer.Status()),
            ).Observe(v)
        }))
        defer timer.ObserveDuration()
        
        c.Next()
        
        RequestTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            string(c.Writer.Status()),
        ).Inc()
    }
}
```

---

## 16. Configuration Management

### 16.1 Configuration Structure

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    
    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Services ServicesConfig `mapstructure:"services"`
}

type ServerConfig struct {
    Port         string `mapstructure:"port"`
    ReadTimeout  int    `mapstructure:"read_timeout"`
    WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     string `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
    SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     string `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
    AccessSecret  string `mapstructure:"access_secret"`
    RefreshSecret string `mapstructure:"refresh_secret"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")
    viper.AddConfigPath(".")
    
    // Environment variables override
    viper.AutomaticEnv()
    
    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // Override with environment variables
    if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
        config.Database.Host = dbHost
    }
    
    if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
        config.Redis.Host = redisHost
    }
    
    return &config, nil
}
```

---

**Document Status:** Complete  
**Implementation Language:** Go (Golang)  
**Architecture Pattern:** Microservices with Clean Architecture  
**Next Steps:** Implementation of individual services following this design
// pkg/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "solemate/pkg/auth"
)

type AuthMiddleware struct {
    jwtService *auth.JWTService
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
    return &AuthMiddleware{
        jwtService: jwtService,
    }
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }
        
        // Extract token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
            c.Abort()
            return
        }
        
        token := parts[1]
        
        // Validate token
        claims, err := m.jwtService.ValidateAccessToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("userID", claims.UserID.String())
        c.Set("userRole", claims.Role)
        c.Set("userEmail", claims.Email)
        
        c.Next()
    }
}

func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("userRole")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
            c.Abort()
            return
        }
        
        role := userRole.(string)
        for _, r := range roles {
            if r == role {
                c.Next()
                return
            }
        }
        
        c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
        c.Abort()
    }
}

// pkg/auth/jwt.go
package auth

import (
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type JWTService struct {
    accessSecret  []byte
    refreshSecret []byte
    accessTTL     time.Duration
    refreshTTL    time.Duration
}

type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    Role   string    `json:"role"`
    jwt.RegisteredClaims
}

func NewJWTService(accessSecret, refreshSecret string) *JWTService {
    return &JWTService{
        accessSecret:  []byte(accessSecret),
        refreshSecret: []byte(refreshSecret),
        accessTTL:     15 * time.Minute,
        refreshTTL:    7 * 24 * time.Hour,
    }
}

func (s *JWTService) GenerateAccessToken(user *entity.User) (string, error) {
    claims := Claims{
        UserID: user.ID,
        Email:  user.Email,
        Role:   string(user.Role),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   user.ID.String(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.accessSecret)
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return s.accessSecret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    return claims, nil
}
```

### 8.2 Rate Limiting and Circuit Breaker

```go
// pkg/middleware/rate_limiter.go
package middleware

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    redis   *redis.Client
    limiters map[string]*rate.Limiter
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
    return &RateLimiter{
        redis:    redis,
        limiters: make(map[string]*rate.Limiter),
    }
}

func (rl *RateLimiter) Limit(rps int, burst int) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := rl.getKey(c)
        
        // Try Redis-based distributed rate limiting first
        allowed, err := rl.checkRedisLimit(c.Request.Context(), key, rps)
        if err == nil && !allowed {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        // Fallback to in-memory rate limiting
        limiter := rl.getLimiter(key, rps, burst)
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func (rl *RateLimiter) getKey(c *gin.Context) string {
    // Use user ID if authenticated, otherwise use IP
    if userID, exists := c.Get("userID"); exists {
        return fmt.Sprintf("rate:%s", userID)
    }
    return fmt.Sprintf("rate:%s", c.ClientIP())
}

func (rl *RateLimiter) checkRedisLimit(ctx context.Context, key string, limit int) (bool, error) {
    pipe := rl.redis.Pipeline()
    incr := pipe.Incr(ctx, key)
    pipe.Expire(ctx, key, time.Second)
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        return true, err // Allow on error
    }
    
    return incr.Val() <= int64(limit), nil
}

// pkg/circuit/breaker.go
package circuit

import (
    "context"
    "errors"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    name          string
    maxFailures   int
    resetTimeout  time.Duration
    
    mu            sync.RWMutex
    state         State
    failures      int
    lastFailTime  time.Time
    successCount  int
}

func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        name:         name,
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        StateClosed,
    }
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
    if err := cb.canExecute(); err != nil {
        return err
    }
    
    err := fn()
    cb.recordResult(err)
    
    return err
}

func (cb *CircuitBreaker) canExecute() error {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    
    switch cb.state {
    case StateOpen:
        if time.Since(cb.lastFailTime) > cb.resetTimeout {
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = StateHalfOpen
            cb.successCount = 0
            cb.mu.Unlock()
            cb.mu.RLock()
            return nil
        }
        return errors.New("circuit breaker is open")
        
    case StateHalfOpen:
        return nil
        
    default: // StateClosed
        return nil
    }
}

func (cb *CircuitBreaker) recordResult(err error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.state == StateHalfOpen {
            cb.state = StateOpen
        } else if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
    } else {
        if cb.state == StateHalfOpen {
            cb.successCount++
            if cb.successCount >= 2 {
                cb.state = StateClosed
                cb.failures = 0
            }
        } else if cb.state == StateClosed {
            cb.failures = 0
        }
    }
}
```

### 8.3 Caching Layer

```go
// pkg/cache/cache.go
package cache

import (
    "context"
    "encoding/json"
    "errors"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    DeletePattern(ctx context.Context, pattern string) error
}

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
    return &RedisCache{
        client: client,
    }
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return errors.New("key not found")
    } else if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
    iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
    var keys []string
    
    for iter.Next(ctx) {
        keys = append(keys, iter.Val())
    }
    
    if err := iter.Err(); err != nil {
        return err
    }
    
    if len(keys) > 0 {
        return c.client.Del(ctx, keys...).Err()
    }
    
    return nil
}

// Cache-aside pattern helper
func GetOrSet[T any](ctx context.Context, cache Cache, key string, ttl time.Duration, fn func() (*T, error)) (*T, error) {
    var result T
    
    // Try to get from cache
    err := cache.Get(ctx, key, &result)
    if err == nil {
        return &result, nil
    }
    
    // Not in cache, fetch from source
    data, err := fn()
    if err != nil {
        return nil, err
    }
    
    // Store in cache (async)
    go cache.Set(context.Background(), key, data, ttl)
    
    return data, nil
}
```

---

## 9. Database Design

### 9.1 Database Connection and Migration

```go
// pkg/database/postgres.go
package database

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func NewPostgresDB(config *Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
    )
    
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logger.Info,
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // Connection pool settings
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db, nil
}

// migrations/001_create_users_table.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    role VARCHAR(50) DEFAULT 'customer',
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone_number);
```

---

## 10. API Gateway Design

### 10.1 Gateway Implementation

```go
// api-gateway/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "solemate/api-gateway/internal/config"
    "solemate/api-gateway/internal/proxy"
    "solemate/pkg/middleware"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize router
    router := gin.New()
    router.Use(gin.Logger())
    router.Use(gin.Recovery())
    
    // CORS middleware
    router.Use(middleware.CORS())
    
    // Rate limiting
    rateLimiter := middleware.NewRateLimiter(cfg.Redis)
    router.Use(rateLimiter.Limit(100, 200))
    
    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })
    
    // Service routes
    v1 := router.Group("/api/v1")
    {
        // User service
        v1.Any("/auth/*path", proxy.ReverseProxy(cfg.Services.UserService))
        v1.Any("/users/*path", proxy.ReverseProxy(cfg.Services.UserService))
        
        // Product service
        v1.Any("/products/*path", proxy.ReverseProxy(cfg.Services.ProductService))
        v1.Any("/categories/*path", proxy.ReverseProxy(cfg.Services.ProductService))
        
        // Cart service
        v1.Any("/cart/*path", proxy.ReverseProxy(cfg.Services.CartService))
        
        // Order service
        v1.Any("/orders/*path", proxy.ReverseProxy(cfg.Services.OrderService))
        
        // Payment service
        v1.Any("/payments/*path", proxy.ReverseProxy(cfg.Services.PaymentService))
    }
    
    // Server configuration
    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Graceful shutdown
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exiting")
}
```

---

## 11. Docker Configuration

### 11.1 Service Dockerfile

```dockerfile
# Dockerfile for Go services
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./services/${SERVICE_NAME}/cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/service .
COPY --from=builder /app/services/${SERVICE_NAME}/config ./config

EXPOSE 8080

CMD ["./service"]
```

### 11.2 Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: solemate
      POSTGRES_PASSWORD: password
      POSTGRES_DB: solemate_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  elasticsearch:
    image: elasticsearch:8.9.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  api-gateway:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - user-service
      - product-service
      - cart-service
      - order-service
      - payment-service
    environment:
      - PORT=8080

  user-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: user-service
    ports:
      - "8081:8080"
    depends_on:
      - postgres
      - redis
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis

  product-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: product-service
    ports:
      - "8082:8080"
    depends_on:
      - postgres
      - elasticsearch
    environment:
      - DB_HOST=postgres
      - ES_HOST=elasticsearch

  cart-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: cart-service
    ports:
      - "8083:8080"
    depends_on:
      - redis
    environment:
      - REDIS_HOST=redis

  order-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: order-service
    ports:
      - "8084:8080"
    depends_on:
      - postgres
    environment:
      - DB_HOST=postgres

  payment-service:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        SERVICE_NAME: payment-service
    ports:
      - "8085:8080"
    depends_on:
      - postgres
    environment:
      - DB_HOST=postgres

volumes:
  postgres_data:
  redis_data:
  elasticsearch_data:
```

---

## 12. Makefile for Development

```makefile
# Makefile
.PHONY: help

---

## 7. Payment Service Design

### 7.1 Payment Gateway Integration

```go
// internal/domain/entity/payment.go
package entity

import (
    "time"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "PENDING"
    PaymentStatusProcessing PaymentStatus = "PROCESSING"
    PaymentStatusCompleted PaymentStatus = "COMPLETED"
    PaymentStatusFailed    PaymentStatus = "FAILED"
    PaymentStatusRefunded  PaymentStatus = "REFUNDED"
    PaymentStatusPartialRefund PaymentStatus = "PARTIAL_REFUND"
)

type PaymentProvider string

const (
    ProviderStripe PaymentProvider = "stripe"
    ProviderPayPal PaymentProvider = "paypal"
    ProviderRazorpay PaymentProvider = "razorpay"
)

type Payment struct {
    ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key"`
    OrderID         uuid.UUID       `json:"order_id" gorm:"type:uuid;not null;index"`
    UserID          uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;index"`
    Amount          decimal.Decimal `json:"amount" gorm:"type:decimal(10,2);not null"`
    Currency        string          `json:"currency" gorm:"default:'USD'"`
    Status          PaymentStatus   `json:"status" gorm:"not null;index"`
    Provider        PaymentProvider `json:"provider" gorm:"not null"`
    TransactionID   string          `json:"transaction_id" gorm:"index"`
    PaymentMethod   string          `json:"payment_method"`
    FailureReason   string          `json:"failure_reason"`
    Metadata        json.RawMessage `json:"metadata" gorm:"type:jsonb"`
    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
    
    // Refund tracking
    RefundAmount    decimal.Decimal `json:"refund_amount" gorm:"type:decimal(10,2)"`
    RefundedAt      *time.Time      `json:"refunded_at"`
}

// internal/infrastructure/payment/stripe_gateway.go
package payment

import (
    "context"
    "fmt"
    
    "github.com/stripe/stripe-go/v74"
    "github.com/stripe/stripe-go/v74/paymentintent"
    "github.com/stripe/stripe-go/v74/refund"
    "github.com/shopspring/decimal"
)

type StripeGateway struct {
    apiKey string
}

func NewStripeGateway(apiKey string) *StripeGateway {
    stripe.Key = apiKey
    return &StripeGateway{apiKey: apiKey}
}

func (g *StripeGateway) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // Convert amount to cents
    amountCents := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()
    
    params := &stripe.PaymentIntentParams{
        Amount:   stripe.Int64(amountCents),
        Currency: stripe.String(string(req.Currency)),
        Metadata: map[string]string{
            "order_id": req.OrderID.String(),
            "user_id":  req.UserID.String(),
        },
        PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
    }
    
    // Add payment method if provided
    if req.PaymentMethodID != "" {
        params.PaymentMethod = stripe.String(req.PaymentMethodID)
        params.Confirm = stripe.Bool(true)
    }
    
    // Create payment intent
    pi, err := paymentintent.New(params)
    if err != nil {
        return nil, fmt.Errorf("failed to create payment intent: %w", err)
    }
    
    response := &PaymentResponse{
        Success:       false,
        TransactionID: pi.ID,
    }
    
    switch pi.Status {
    case stripe.PaymentIntentStatusSucceeded:
        response.Success = true
        response.Status = PaymentStatusCompleted
    case stripe.PaymentIntentStatusProcessing:
        response.Success = true
        response.Status = PaymentStatusProcessing
    case stripe.PaymentIntentStatusRequiresAction:
        response.Status = PaymentStatusPending
        response.ClientSecret = pi.ClientSecret
        response.RequiresAction = true
    default:
        response.Status = PaymentStatusFailed
        if pi.LastPaymentError != nil {
            response.Error = pi.LastPaymentError.Message
        }
    }
    
    return response, nil
}

func (g *StripeGateway) RefundPayment(ctx context.Context, transactionID string, amount decimal.Decimal) (*RefundResponse, error) {
    amountCents := amount.Mul(decimal.NewFromInt(100)).IntPart()
    
    params := &stripe.RefundParams{
        PaymentIntent: stripe.String(transactionID),
        Amount:        stripe.Int64(amountCents),
    }
    
    r, err := refund.New(params)
    if err != nil {
        return nil, fmt.Errorf("failed to create refund: %w", err)
    }
    
    return &RefundResponse{
        Success:      r.Status == stripe.RefundStatusSucceeded,
        RefundID:     r.ID,
        Amount:       amount,
        Status:       string(r.Status),
    }, nil
}

// internal/domain/service/payment_service.go
package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "sync"
    
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "solemate/services/payment-service/internal/domain/entity"
    "solemate/services/payment-service/internal/domain/repository"
    "solemate/services/payment-service/internal/infrastructure/payment"
)

type PaymentService struct {
    repo            repository.PaymentRepository
    stripeGateway   *payment.StripeGateway
    paypalGateway   *payment.PayPalGateway
    webhookHandlers map[string]WebhookHandler
    mu              sync.RWMutex
}

func NewPaymentService(
    repo repository.PaymentRepository,
    stripeGateway *payment.StripeGateway,
    paypalGateway *payment.PayPalGateway,
) *PaymentService {
    ps := &PaymentService{
        repo:            repo,
        stripeGateway:   stripeGateway,
        paypalGateway:   paypalGateway,
        webhookHandlers: make(map[string]WebhookHandler),
    }
    
    // Register webhook handlers
    ps.registerWebhookHandlers()
    
    return ps
}

func (s *PaymentService) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
    // Create payment record
    payment := &entity.Payment{
        ID:            uuid.New(),
        OrderID:       req.OrderID,
        UserID:        req.UserID,
        Amount:        req.Amount,
        Currency:      req.Currency,
        Status:        entity.PaymentStatusPending,
        Provider:      req.Provider,
        PaymentMethod: req.PaymentMethod,
        Metadata:      req.Metadata,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    
    // Save initial payment record
    if err := s.repo.Create(ctx, payment); err != nil {
        return nil, fmt.Errorf("failed to create payment record: %w", err)
    }
    
    // Process with appropriate gateway
    var result *payment.PaymentResponse
    var err error
    
    switch req.Provider {
    case entity.ProviderStripe:
        result, err = s.stripeGateway.ProcessPayment(ctx, &payment.PaymentRequest{
            OrderID:         req.OrderID,
            UserID:          req.UserID,
            Amount:          req.Amount,
            Currency:        req.Currency,
            PaymentMethodID: req.PaymentMethodID,
        })
        
    case entity.ProviderPayPal:
        result, err = s.paypalGateway.ProcessPayment(ctx, &payment.PaymentRequest{
            OrderID:  req.OrderID,
            UserID:   req.UserID,
            Amount:   req.Amount,
            Currency: req.Currency,
        })
        
    default:
        err = fmt.Errorf("unsupported payment provider: %s", req.Provider)
    }
    
    if err != nil {
        // Update payment as failed
        payment.Status = entity.PaymentStatusFailed
        payment.FailureReason = err.Error()
        s.repo.Update(ctx, payment)
        
        return nil, err
    }
    
    // Update payment record with result
    payment.TransactionID = result.TransactionID
    payment.Status = result.Status
    
    if !result.Success && result.Error != "" {
        payment.FailureReason = result.Error
    }
    
    if err := s.repo.Update(ctx, payment); err != nil {
        return nil, fmt.Errorf("failed to update payment record: %w", err)
    }
    
    return &ProcessPaymentResponse{
        Success:        result.Success,
        PaymentID:      payment.ID,
        TransactionID:  result.TransactionID,
        Status:         payment.Status,
        RequiresAction: result.RequiresAction,
        ClientSecret:   result.ClientSecret,
        Error:          result.Error,
    }, nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID uuid.UUID, amount decimal.Decimal) (*RefundResponse, error) {
    // Get payment record
    payment, err := s.repo.FindByID(ctx, paymentID)
    if err != nil {
        return nil, fmt.Errorf("payment not found: %w", err)
    }
    
    // Validate refund amount
    totalRefunded := payment.RefundAmount.Add(amount)
    if totalRefunded.GreaterThan(payment.Amount) {
        return nil, errors.New("refund amount exceeds payment amount")
    }
    
    // Process refund with appropriate gateway
    var result *payment.RefundResponse
    
    switch payment.Provider {
    case entity.ProviderStripe:
        result, err = s.stripeGateway.RefundPayment(ctx, payment.TransactionID, amount)
        
    case entity.ProviderPayPal:
        result, err = s.paypalGateway.RefundPayment(ctx, payment.TransactionID, amount)
        
    default:
        return nil, fmt.Errorf("unsupported payment provider: %s", payment.Provider)
    }
    
    if err != nil {
        return nil, fmt.Errorf("refund failed: %w", err)
    }
    
    // Update payment record
    payment.RefundAmount = totalRefunded
    now := time.Now()
    payment.RefundedAt = &now
    
    if totalRefunded.Equal(payment.Amount) {
        payment.Status = entity.PaymentStatusRefunded
    } else {
        payment.Status = entity.PaymentStatusPartialRefund
    }
    
    if err := s.repo.Update(ctx, payment); err != nil {
        return nil, fmt.Errorf("failed to update payment record: %w", err)
    }
    
    return &RefundResponse{
        Success:  result.Success,
        RefundID: result.RefundID,
        Amount:   amount,
        Status:   result.Status,
    }, nil
}

// Webhook handling
func (s *PaymentService) HandleWebhook(ctx context.Context, provider string, payload []byte, signature string) error {
    s.mu.RLock()
    handler, exists := s.webhookHandlers[provider]
    s.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("no webhook handler for provider: %s", provider)
    }
    
    return handler.Handle(ctx, payload, signature)
}

func (s *PaymentService) registerWebhookHandlers() {
    // Stripe webhook handler
    s.webhookHandlers["stripe"] = &StripeWebhookHandler{
        service: s,
        secret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
    }
    
    // PayPal webhook handler
    s.webhookHandlers["paypal"] = &PayPalWebhookHandler{
        service: s,
    }
}# Low-Level Design (LLD) Document - Golang Implementation
## SoleMate E-Commerce Platform

### Document Information
- **Project Name:** SoleMate E-Commerce Platform
- **Document Version:** 1.0
- **Date:** September 20, 2024
- **Prepared by:** Development Team
- **Language:** Go (Golang)
- **Framework:** Gin, GORM, gRPC
- **Status:** Final

---

## Table of Contents
1. [Introduction](#1-introduction)
2. [Project Structure](#2-project-structure)
3. [User Service Design](#3-user-service-design)
4. [Product Service Design](#4-product-service-design)
5. [Cart Service Design](#5-cart-service-design)
6. [Order Service Design](#6-order-service-design)
7. [Payment Service Design](#7-payment-service-design)
8. [Common Components](#8-common-components)
9. [Database Design](#9-database-design)
10. [API Gateway Design](#10-api-gateway-design)

---

## 1. Introduction

### 1.1 Technology Stack
- **Language:** Go 1.21+
- **Web Framework:** Gin for REST APIs
- **RPC Framework:** gRPC for inter-service communication
- **ORM:** GORM for database operations
- **Database:** PostgreSQL 15+
- **Cache:** Redis 7+
- **Message Queue:** RabbitMQ / NATS
- **Service Discovery:** Consul
- **Monitoring:** Prometheus + Grafana

### 1.2 Design Principles
- Clean Architecture (Domain-Driven Design)
- Dependency Injection
- Repository Pattern
- SOLID Principles
- Event-Driven Architecture
- Graceful error handling

---

## 2. Project Structure

### 2.1 Microservice Structure
```
solemate/
├── services/
│   ├── user-service/
│   ├── product-service/
│   ├── cart-service/
│   ├── order-service/
│   ├── payment-service/
│   ├── inventory-service/
│   └── notification-service/
├── api-gateway/
├── pkg/
│   ├── common/
│   ├── auth/
│   ├── database/
│   ├── cache/
│   └── utils/
├── proto/
├── deployments/
├── scripts/
└── docker-compose.yml
```

### 2.2 Service Structure Template
```
service-name/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   ├── handler/
│   │   ├── http/
│   │   └── grpc/
│   ├── infrastructure/
│   │   ├── database/
│   │   ├── cache/
│   │   └── messaging/
│   └── config/
├── pkg/
├── migrations/
├── Dockerfile
├── Makefile
└── go.mod
```

---

## 3. User Service Design

### 3.1 Domain Entities

```go
// internal/domain/entity/user.go
package entity

import (
    "time"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
    Email         string     `json:"email" gorm:"uniqueIndex;not null"`
    PasswordHash  string     `json:"-" gorm:"not null"`
    FirstName     string     `json:"first_name"`
    LastName      string     `json:"last_name"`
    PhoneNumber   string     `json:"phone_number" gorm:"index"`
    Role          UserRole   `json:"role" gorm:"default:'customer'"`
    IsActive      bool       `json:"is_active" gorm:"default:true"`
    EmailVerified bool       `json:"email_verified" gorm:"default:false"`
    LastLoginAt   *time.Time `json:"last_login_at"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
    
    // Associations
    Addresses     []Address  `json:"addresses,omitempty" gorm:"foreignKey:UserID"`
    RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID"`
}

type UserRole string

const (
    RoleCustomer UserRole = "customer"
    RoleAdmin    UserRole = "admin"
    RoleManager  UserRole = "manager"
)

// Password methods
func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.PasswordHash = string(hashedPassword)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
    return err == nil
}

// Address entity
type Address struct {
    ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
    UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
    Type        string    `json:"type" gorm:"default:'shipping'"`
    Name        string    `json:"name"`
    Street1     string    `json:"street_1"`
    Street2     string    `json:"street_2"`
    City        string    `json:"city"`
    State       string    `json:"state"`
    PostalCode  string    `json:"postal_code"`
    Country     string    `json:"country"`
    Phone       string    `json:"phone"`
    IsDefault   bool      `json:"is_default"`
    CreatedAt   time.Time `json:"created_at"`
}

// RefreshToken entity
type RefreshToken struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID    uuid.UUID `gorm:"type:uuid;not null"`
    Token     string    `gorm:"uniqueIndex;not null"`
    ExpiresAt time.Time `gorm:"not null"`
    CreatedAt time.Time
}
```

### 3.2 Repository Interface

```go
// internal/domain/repository/user_repository.go
package repository

import (
    "context"
    "github.com/google/uuid"
    "solemate/services/user-service/internal/domain/entity"
)

type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id uuid.UUID) error
    UpdateLastLogin(ctx context.Context, id uuid.UUID) error
    
    // Address methods
    CreateAddress(ctx context.Context, address *entity.Address) error
    GetUserAddresses(ctx context.Context, userID uuid.UUID) ([]entity.Address, error)
    UpdateAddress(ctx context.Context, address *entity.Address) error
    DeleteAddress(ctx context.Context, id uuid.UUID) error
    
    // Token methods
    SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error
    FindRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
    DeleteRefreshToken(ctx context.Context, token string) error
    DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}
```

### 3.3 Service Layer

```go
// internal/domain/service/user_service.go
package service

import (
    "context"
    "errors"
    "time"
    
    "github.com/google/uuid"
    "solemate/services/user-service/internal/domain/entity"
    "solemate/services/user-service/internal/domain/repository"
    "solemate/pkg/auth"
    "solemate/pkg/cache"
    "solemate/pkg/validator"
)

type UserService struct {
    repo      repository.UserRepository
    cache     cache.Cache
    jwtSvc    *auth.JWTService
    validator *validator.Validator
}

func NewUserService(
    repo repository.UserRepository,
    cache cache.Cache,
    jwtSvc *auth.JWTService,
    validator *validator.Validator,
) *UserService {
    return &UserService{
        repo:      repo,
        cache:     cache,
        jwtSvc:    jwtSvc,
        validator: validator,
    }
}

// RegisterUser creates a new user account
func (s *UserService) RegisterUser(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
    // Validate request
    if err := s.validator.Validate(req); err != nil {
        return nil, err
    }
    
    // Check if email already exists
    existingUser, _ := s.repo.FindByEmail(ctx, req.Email)
    if existingUser != nil {
        return nil, errors.New("email already registered")
    }
    
    // Create user entity
    user := &entity.User{
        ID:        uuid.New(),
        Email:     req.Email,
        FirstName: req.FirstName,
        LastName:  req.LastName,
        Role:      entity.RoleCustomer,
    }
    
    // Hash password
    if err := user.SetPassword(req.Password); err != nil {
        return nil, err
    }
    
    // Save to database
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Generate tokens
    tokens, err := s.generateTokens(ctx, user)
    if err != nil {
        return nil, err
    }
    
    // Cache user data
    s.cacheUser(ctx, user)
    
    return &AuthResponse{
        User:         s.sanitizeUser(user),
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
    }, nil
}

// Login authenticates a user
func (s *UserService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
    // Find user by email
    user, err := s.repo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.New("invalid credentials")
    }
    
    // Check password
    if !user.CheckPassword(req.Password) {
        return nil, errors.New("invalid credentials")
    }
    
    // Check if account is active
    if !user.IsActive {
        return nil, errors.New("account is deactivated")
    }
    
    // Update last login
    s.repo.UpdateLastLogin(ctx, user.ID)
    
    // Generate tokens
    tokens, err := s.generateTokens(ctx, user)
    if err != nil {
        return nil, err
    }
    
    return &AuthResponse{
        User:         s.sanitizeUser(user),
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
    }, nil
}

// RefreshToken generates new access token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
    // Verify refresh token
    claims, err := s.jwtSvc.ValidateRefreshToken(refreshToken)
    if err != nil {
        return nil, err
    }
    
    // Check if token exists in database
    storedToken, err := s.repo.FindRefreshToken(ctx, refreshToken)
    if err != nil || storedToken.ExpiresAt.Before(time.Now()) {
        return nil, errors.New("invalid refresh token")
    }
    
    // Get user
    user, err := s.repo.FindByID(ctx, claims.UserID)
    if err != nil {
        return nil, err
    }
    
    // Generate new access token
    accessToken, err := s.jwtSvc.GenerateAccessToken(user)
    if err != nil {
        return nil, err
    }
    
    return &TokenResponse{
        AccessToken: accessToken,
    }, nil
}

// Helper methods
func (s *UserService) generateTokens(ctx context.Context, user *entity.User) (*TokenPair, error) {
    // Generate access token
    accessToken, err := s.jwtSvc.GenerateAccessToken(user)
    if err != nil {
        return nil, err
    }
    
    // Generate refresh token
    refreshToken, err := s.jwtSvc.GenerateRefreshToken(user)
    if err != nil {
        return nil, err
    }
    
    // Store refresh token
    token := &entity.RefreshToken{
        ID:        uuid.New(),
        UserID:    user.ID,
        Token:     refreshToken,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    }
    
    if err := s.repo.SaveRefreshToken(ctx, token); err != nil {
        return nil, err
    }
    
    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

func (s *UserService) cacheUser(ctx context.Context, user *entity.User) {
    key := fmt.Sprintf("user:%s", user.ID.String())
    s.cache.Set(ctx, key, user, 1*time.Hour)
}

func (s *UserService) sanitizeUser(user *entity.User) *UserResponse {
    return &UserResponse{
        ID:            user.ID,
        Email:         user.Email,
        FirstName:     user.FirstName,
        LastName:      user.LastName,
        PhoneNumber:   user.PhoneNumber,
        Role:          string(user.Role),
        EmailVerified: user.EmailVerified,
        CreatedAt:     user.CreatedAt,
    }
}
```

### 3.4 HTTP Handler

```go
// internal/handler/http/user_handler.go
package http

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "solemate/services/user-service/internal/domain/service"
    "solemate/pkg/middleware"
)

type UserHandler struct {
    userService *service.UserService
    authMiddleware *middleware.AuthMiddleware
}

func NewUserHandler(userService *service.UserService, authMiddleware *middleware.AuthMiddleware) *UserHandler {
    return &UserHandler{
        userService: userService,
        authMiddleware: authMiddleware,
    }
}

func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
    auth := router.Group("/auth")
    {
        auth.POST("/register", h.Register)
        auth.POST("/login", h.Login)
        auth.POST("/refresh", h.RefreshToken)
        auth.POST("/logout", h.authMiddleware.Authenticate(), h.Logout)
    }
    
    users := router.Group("/users")
    users.Use(h.authMiddleware.Authenticate())
    {
        users.GET("/profile", h.GetProfile)
        users.PUT("/profile", h.UpdateProfile)
        users.POST("/change-password", h.ChangePassword)
        users.GET("/addresses", h.GetAddresses)
        users.POST("/addresses", h.CreateAddress)
        users.PUT("/addresses/:id", h.UpdateAddress)
        users.DELETE("/addresses/:id", h.DeleteAddress)
    }
}

func (h *UserHandler) Register(c *gin.Context) {
    var req service.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    response, err := h.userService.RegisterUser(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, response)
}

func (h *UserHandler) Login(c *gin.Context) {
    var req service.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    response, err := h.userService.Login(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := c.GetString("userID")
    
    user, err := h.userService.GetUserByID(c.Request.Context(), uuid.MustParse(userID))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    
    c.JSON(http.StatusOK, user)
}
```

---

## 4. Product Service Design

### 4.1 Domain Entities

```go
// internal/domain/entity/product.go
package entity

import (
    "time"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "database/sql/driver"
)

type Product struct {
    ID           uuid.UUID       `json:"id" gorm:"type:uuid;primary_key"`
    SKU          string          `json:"sku" gorm:"uniqueIndex;not null"`
    Name         string          `json:"name" gorm:"not null;index"`
    Slug         string          `json:"slug" gorm:"uniqueIndex;not null"`
    Description  string          `json:"description" gorm:"type:text"`
    CategoryID   uuid.UUID       `json:"category_id" gorm:"type:uuid;index"`
    BrandID      uuid.UUID       `json:"brand_id" gorm:"type:uuid;index"`
    Price        decimal.Decimal `json:"price" gorm:"type:decimal(10,2);not null"`
    ComparePrice decimal.Decimal `json:"compare_price" gorm:"type:decimal(10,2)"`
    Cost         decimal.Decimal `json:"cost" gorm:"type:decimal(10,2)"`
    Weight       float64         `json:"weight"`
    IsActive     bool            `json:"is_active" gorm:"default:true;index"`
    Tags         StringArray     `json:"tags" gorm:"type:text[]"`
    MetaTitle    string          `json:"meta_title"`
    MetaDesc     string          `json:"meta_description"`
    CreatedAt    time.Time       `json:"created_at"`
    UpdatedAt    time.Time       `json:"updated_at"`
    
    // Associations
    Category     *Category        `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
    Brand        *Brand           `json:"brand,omitempty" gorm:"foreignKey:BrandID"`
    Images       []ProductImage   `json:"images,omitempty" gorm:"foreignKey:ProductID"`
    Variants     []ProductVariant `json:"variants,omitempty" gorm:"foreignKey:ProductID"`
    Attributes   []ProductAttribute `json:"attributes,omitempty" gorm:"foreignKey:ProductID"`
    
    // Calculated fields (not stored in DB)
    Rating       float32         `json:"rating" gorm:"-"`
    ReviewCount  int             `json:"review_count" gorm:"-"`
    Stock        int             `json:"stock" gorm:"-"`
}

type ProductVariant struct {
    ID        uuid.UUID       `json:"id" gorm:"type:uuid;primary_key"`
    ProductID uuid.UUID       `json:"product_id" gorm:"type:uuid;not null;index"`
    SKU       string          `json:"sku" gorm:"uniqueIndex;not null"`
    Size      string          `json:"size"`
    Color     string          `json:"color"`
    Price     decimal.Decimal `json:"price" gorm:"type:decimal(10,2)"`
    Stock     int             `json:"stock" gorm:"default:0"`
    Weight    float64         `json:"weight"`
    Images    StringArray     `json:"images" gorm:"type:text[]"`
    IsActive  bool            `json:"is_active" gorm:"default:true"`
}

type Category struct {
    ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key"`
    ParentID    *uuid.UUID `json:"parent_id" gorm:"type:uuid;index"`
    Name        string     `json:"name" gorm:"not null"`
    Slug        string     `json:"slug" gorm:"uniqueIndex;not null"`
    Description string     `json:"description"`
    ImageURL    string     `json:"image_url"`
    SortOrder   int        `json:"sort_order" gorm:"default:0"`
    IsActive    bool       `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time  `json:"created_at"`
    
    // Self-referential relationship
    Parent      *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
    Children    []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// Custom type for PostgreSQL array support
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
    if len(a) == 0 {
        return nil, nil
    }
    return a, nil
}

func (a *StringArray) Scan(value interface{}) error {
    if value == nil {
        *a = []string{}
        return nil
    }
    // Implementation for scanning PostgreSQL array
    return nil
}
```

### 4.2 Search Service with Elasticsearch

```go
// internal/domain/service/search_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/elastic/go-elasticsearch/v8"
    "github.com/elastic/go-elasticsearch/v8/esapi"
    "solemate/services/product-service/internal/domain/entity"
)

type SearchService struct {
    client *elasticsearch.Client
    index  string
}

func NewSearchService(client *elasticsearch.Client) *SearchService {
    return &SearchService{
        client: client,
        index:  "products",
    }
}

type SearchRequest struct {
    Query      string    `json:"query"`
    Category   string    `json:"category"`
    Brands     []string  `json:"brands"`
    PriceMin   float64   `json:"price_min"`
    PriceMax   float64   `json:"price_max"`
    Sizes      []string  `json:"sizes"`
    Colors     []string  `json:"colors"`
    Rating     float32   `json:"rating"`
    SortBy     string    `json:"sort_by"`
    Page       int       `json:"page"`
    Limit      int       `json:"limit"`
}

func (s *SearchService) SearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // Set defaults
    if req.Page < 1 {
        req.Page = 1
    }
    if req.Limit < 1 || req.Limit > 100 {
        req.Limit = 20
    }
    
    // Build Elasticsearch query
    query := s.buildQuery(req)
    
    // Execute search
    res, err := s.client.Search(
        s.client.Search.WithContext(ctx),
        s.client.Search.WithIndex(s.index),
        s.client.Search.WithBody(query),
        s.client.Search.WithFrom((req.Page-1)*req.Limit),
        s.client.Search.WithSize(req.Limit),
        s.client.Search.WithSort(s.getSortOption(req.SortBy)),
    )
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()
    
    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    // Extract products from hits
    products := s.extractProducts(result)
    total := s.extractTotal(result)
    
    return &SearchResponse{
        Products:   products,
        Total:      total,
        Page:       req.Page,
        TotalPages: (total + req.Limit - 1) / req.Limit,
    }, nil
}

func (s *SearchService) buildQuery(req *SearchRequest) map[string]interface{} {
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "bool": map[string]interface{}{
                "must":   []interface{}{},
                "filter": []interface{}{},
            },
        },
        "aggs": s.buildAggregations(),
    }
    
    boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
    
    // Add text search if query provided
    if req.Query != "" {
        must := boolQuery["must"].([]interface{})
        must = append(must, map[string]interface{}{
            "multi_match": map[string]interface{}{
                "query": req.Query,
                "fields": []string{
                    "name^3",
                    "brand.name^2",
                    "description",
                    "tags",
                },
                "type":      "best_fields",
                "fuzziness": "AUTO",
            },
        })
        boolQuery["must"] = must
    }
    
    // Add filters
    filters := boolQuery["filter"].([]interface{})
    
    if req.Category != "" {
        filters = append(filters, map[string]interface{}{
            "term": map[string]interface{}{
                "category.slug": req.Category,
            },
        })
    }
    
    if len(req.Brands) > 0 {
        filters = append(filters, map[string]interface{}{
            "terms": map[string]interface{}{
                "brand.slug": req.Brands,
            },
        })
    }
    
    if req.PriceMin > 0 || req.PriceMax > 0 {
        rangeQuery := map[string]interface{}{
            "range": map[string]interface{}{
                "price": map[string]interface{}{},
            },
        }
        priceRange := rangeQuery["range"].(map[string]interface{})["price"].(map[string]interface{})
        
        if req.PriceMin > 0 {
            priceRange["gte"] = req.PriceMin
        }
        if req.PriceMax > 0 {
            priceRange["lte"] = req.PriceMax
        }
        
        filters = append(filters, rangeQuery)
    }
    
    boolQuery["filter"] = filters
    
    return query
}

func (s *SearchService) IndexProduct(ctx context.Context, product *entity.Product) error {
    // Prepare document for indexing
    doc := map[string]interface{}{
        "id":           product.ID,
        "sku":          product.SKU,
        "name":         product.Name,
        "slug":         product.Slug,
        "description":  product.Description,
        "price":        product.Price,
        "compare_price": product.ComparePrice,
        "brand": map[string]interface{}{
            "id":   product.BrandID,
            "name": product.Brand.Name,
            "slug": product.Brand.Slug,
        },
        "category": map[string]interface{}{
            "id":   product.CategoryID,
            "name": product.Category.Name,
            "slug": product.Category.Slug,
        },
        "tags":        product.Tags,
        "is_active":   product.IsActive,
        "created_at":  product.CreatedAt,
        "rating":      product.Rating,
        "review_count": product.ReviewCount,
    }
    
    // Add variants
    if len(product.Variants) > 0 {
        variants := make([]map[string]interface{}, len(product.Variants))
        for i, v := range product.Variants {
            variants[i] = map[string]interface{}{
                "sku":   v.SKU,
                "size":  v.Size,
                "color": v.Color,
                "price": v.Price,
                "stock": v.Stock,
            }
        }
        doc["variants"] = variants
    }
    
    // Index document
    data, err := json.Marshal(doc)
    if err != nil {
        return err
    }
    
    req := esapi.IndexRequest{
        Index:      s.index,
        DocumentID: product.ID.String(),
        Body:       bytes.NewReader(data),
        Refresh:    "true",
    }
    
    res, err := req.Do(ctx, s.client)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    
    if res.IsError() {
        return fmt.Errorf("error indexing document: %s", res.String())
    }
    
    return nil
}
```

---

## 5. Cart Service Design

### 5.1 Cart Entity and Service

```go
// internal/domain/entity/cart.go
package entity

import (
    "time"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

type Cart struct {
    ID         uuid.UUID       `json:"id"`
    UserID     uuid.UUID       `json:"user_id"`
    Items      []CartItem      `json:"items"`
    PromoCode  string          `json:"promo_code,omitempty"`
    Subtotal   decimal.Decimal `json:"subtotal"`
    Discount   decimal.Decimal `json:"discount"`
    Tax        decimal.Decimal `json:"tax"`
    Shipping   decimal.Decimal `json:"shipping"`
    Total      decimal.Decimal `json:"total"`
    ExpiresAt  time.Time       `json:"expires_at"`
    CreatedAt  time.Time       `json:"created_at"`
    UpdatedAt  time.Time       `json:"updated_at"`
}

type CartItem struct {
    ID         uuid.UUID       `json:"id"`
    ProductID  uuid.UUID       `json:"product_id"`
    VariantID  uuid.UUID       `json:"variant_id"`
    Quantity   int             `json:"quantity"`
    Price      decimal.Decimal `json:"price"`
    Discount   decimal.Decimal `json:"discount"`
    Total      decimal.Decimal `json:"total"`
    
    // Populated fields
    Product    *Product        `json:"product,omitempty"`
    Variant    *ProductVariant `json:"variant,omitempty"`
}

// internal/domain/service/cart_service.go
package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "solemate/services/cart-service/internal/domain/entity"
    "solemate/pkg/cache"
)

type CartService struct {
    cache           cache.Cache
    productClient   ProductServiceClient
    inventoryClient InventoryServiceClient
    promoClient     PromoServiceClient
}

func NewCartService(
    cache cache.Cache,
    productClient ProductServiceClient,
    inventoryClient InventoryServiceClient,
    promoClient PromoServiceClient,
) *CartService {
    return &CartService{
        cache:           cache,
        productClient:   productClient,
        inventoryClient: inventoryClient,
        promoClient:     promoClient,
    }
}

func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*entity.Cart, error) {
    // Try to get cart from cache
    cacheKey := fmt.Sprintf("cart:%s", userID.String())
    
    var cart entity.Cart
    err := s.cache.Get(ctx, cacheKey, &cart)
    if err == nil {
        // Refresh product data
        if err := s.refreshCartProducts(ctx, &cart); err != nil {
            return nil, err
        }
        
        // Recalculate totals
        s.calculateTotals(&cart)
        
        // Update cache
        s.saveCart(ctx, &cart)
        
        return &cart, nil
    }
    
    // Cart doesn't exist, create new one
    cart = entity.Cart{
        ID:        uuid.New(),
        UserID:    userID,
        Items:     []entity.CartItem{},
        Subtotal:  decimal.Zero,
        Discount:  decimal.Zero,
        Tax:       decimal.Zero,
        Shipping:  decimal.Zero,
        Total:     decimal.Zero,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    s.saveCart(ctx, &cart)
    return &cart, nil
}

func (s *CartService) AddItem(ctx context.Context, userID uuid.UUID, req *AddItemRequest) (*entity.Cart, error) {
    // Get current cart
    cart, err := s.GetCart(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Get product details
    product, err := s.productClient.GetProduct(ctx, req.ProductID)
    if err != nil {
        return nil, fmt.Errorf("product not found: %w", err)
    }
    
    if !product.IsActive {
        return nil, errors.New("product is not available")
    }
    
    // Find variant
    var variant *entity.ProductVariant
    for _, v := range product.Variants {
        if v.ID == req.VariantID {
            variant = &v
            break
        }
    }
    
    if variant == nil {
        return nil, errors.New("variant not found")
    }
    
    // Check inventory
    stock, err := s.inventoryClient.CheckStock(ctx, req.ProductID, req.VariantID)
    if err != nil {
        return nil, fmt.Errorf("failed to check stock: %w", err)
    }
    
    // Check if item already exists in cart
    existingIndex := -1
    currentQuantity := 0
    for i, item := range cart.Items {
        if item.ProductID == req.ProductID && item.VariantID == req.VariantID {
            existingIndex = i
            currentQuantity = item.Quantity
            break
        }
    }
    
    totalQuantity := currentQuantity + req.Quantity
    if totalQuantity > stock {
        return nil, fmt.Errorf("only %d items available in stock", stock)
    }
    
    // Determine price
    price := variant.Price
    if price.IsZero() {
        price = product.Price
    }
    
    if existingIndex >= 0 {
        // Update existing item
        cart.Items[existingIndex].Quantity = totalQuantity
        cart.Items[existingIndex].Total = price.Mul(decimal.NewFromInt(int64(totalQuantity)))
    } else {
        // Add new item
        newItem := entity.CartItem{
            ID:        uuid.New(),
            ProductID: req.ProductID,
            VariantID: req.VariantID,
            Quantity:  req.Quantity,
            Price:     price,
            Discount:  decimal.Zero,
            Total:     price.Mul(decimal.NewFromInt(int64(req.Quantity))),
            Product:   product,
            Variant:   variant,
        }
        cart.Items = append(cart.Items, newItem)
    }
    
    // Recalculate totals
    s.calculateTotals(cart)
    
    // Save to cache
    s.saveCart(ctx, cart)
    
    // Reserve inventory temporarily
    go s.inventoryClient.ReserveStock(context.Background(), req.ProductID, req.VariantID, req.Quantity, 15*time.Minute)
    
    return cart, nil
}

func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, quantity int) (*entity.Cart, error) {
    cart, err := s.GetCart(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    if quantity <= 0 {
        return s.RemoveItem(ctx, userID, itemID)
    }
    
    // Find item
    itemIndex := -1
    for i, item := range cart.Items {
        if item.ID == itemID {
            itemIndex = i
            break
        }
    }
    
    if itemIndex < 0 {
        return nil, errors.New("item not found in cart")
    }
    
    item := &cart.Items[itemIndex]
    
    // Check stock
    stock, err := s.inventoryClient.CheckStock(ctx, item.ProductID, item.VariantID)
    if err != nil {
        return nil, err
    }
    
    if quantity > stock {
        return nil, fmt.Errorf("only %d items available in stock", stock)
    }
    
    // Update quantity
    item.Quantity = quantity
    item.Total = item.Price.Mul(decimal.NewFromInt(int64(quantity)))
    
    // Recalculate totals
    s.calculateTotals(cart)
    
    // Save to cache
    s.saveCart(ctx, cart)
    
    return cart, nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*entity.Cart, error) {
    cart, err := s.GetCart(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Find and remove item
    newItems := []entity.CartItem{}
    for _, item := range cart.Items {
        if item.ID != itemID {
            newItems = append(newItems, item)
        }
    }
    
    cart.Items = newItems
    
    // Recalculate totals
    s.calculateTotals(cart)
    
    // Save to cache
    s.saveCart(ctx, cart)
    
    return cart, nil
}

func (s *CartService) ApplyPromoCode(ctx context.Context, userID uuid.UUID, code string) (*entity.Cart, error) {
    cart, err := s.GetCart(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Validate promo code
    promo, err := s.promoClient.ValidatePromoCode(ctx, code, cart.Subtotal)
    if err != nil {
        return nil, fmt.Errorf("invalid promo code: %w", err)
    }
    
    cart.PromoCode = code
    cart.Discount = promo.DiscountAmount
    
    // Recalculate totals
    s.calculateTotals(cart)
    
    // Save to cache
    s.saveCart(ctx, cart)
    
    return cart, nil
}

func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
    cacheKey := fmt.Sprintf("cart:%s", userID.String())
    return s.cache.Delete(ctx, cacheKey)
}

// Helper methods
func (s *CartService) calculateTotals(cart *entity.Cart) {
    // Calculate subtotal
    subtotal := decimal.Zero
    for _, item := range cart.Items {
        subtotal = subtotal.Add(item.Total)
    }
    cart.Subtotal = subtotal
    
    // Apply discount if exists
    if cart.Discount.IsZero() && cart.PromoCode != "" {
        // Re-validate promo code
        promo, err := s.promoClient.ValidatePromoCode(context.Background(), cart.PromoCode, subtotal)
        if err == nil {
            cart.Discount = promo.DiscountAmount
        } else {
            cart.PromoCode = ""
            cart.Discount = decimal.Zero
        }
    }
    
    // Calculate tax (8% for example)
    taxRate := decimal.NewFromFloat(0.08)
    taxableAmount := subtotal.Sub(cart.Discount)
    cart.Tax = taxableAmount.Mul(taxRate).Round(2)
    
    // Calculate shipping (free over $100)
    if subtotal.GreaterThanOrEqual(decimal.NewFromInt(100)) {
        cart.Shipping = decimal.Zero
    } else {
        cart.Shipping = decimal.NewFromInt(10)
    }
    
    // Calculate total
    cart.Total = subtotal.Sub(cart.Discount).Add(cart.Tax).Add(cart.Shipping).Round(2)
    cart.UpdatedAt = time.Now()
}

func (s *CartService) saveCart(ctx context.Context, cart *entity.Cart) {
    cacheKey := fmt.Sprintf("cart:%s", cart.UserID.String())
    s.cache.Set(ctx, cacheKey, cart, 7*24*time.Hour)
}

func (s *CartService) refreshCartProducts(ctx context.Context, cart *entity.Cart) error {
    for i := range cart.Items {
        item := &cart.Items[i]
        
        // Get latest product info
        product, err := s.productClient.GetProduct(ctx, item.ProductID)
        if err != nil {
            continue // Skip if product not found
        }
        
        item.Product = product
        
        // Find variant
        for _, v := range product.Variants {
            if v.ID == item.VariantID {
                item.Variant = &v
                
                // Update price if changed
                newPrice := v.Price
                if newPrice.IsZero() {
                    newPrice = product.Price
                }
                
                if !item.Price.Equal(newPrice) {
                    item.Price = newPrice
                    item.Total = newPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
                }
                break
            }
        }
    }
    
    return nil
}