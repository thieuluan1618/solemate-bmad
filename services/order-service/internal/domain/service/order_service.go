package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"solemate/services/order-service/internal/domain/entity"
	"solemate/services/order-service/internal/domain/repository"
)

type OrderService interface {
	// Order creation and management
	CreateOrderFromCart(ctx context.Context, userID uuid.UUID, shippingAddress, billingAddress entity.Address, shippingMethod string, notes string) (*entity.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entity.Order, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Order, int64, error)
	GetUserOrderSummaries(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.OrderSummary, int64, error)

	// Order status management
	ConfirmOrder(ctx context.Context, orderID uuid.UUID) error
	ProcessOrder(ctx context.Context, orderID uuid.UUID) error
	ShipOrder(ctx context.Context, orderID uuid.UUID, trackingNumber string, estimatedDelivery *time.Time) error
	DeliverOrder(ctx context.Context, orderID uuid.UUID) error
	CompleteOrder(ctx context.Context, orderID uuid.UUID) error
	CancelOrder(ctx context.Context, orderID uuid.UUID, reason string) error
	RefundOrder(ctx context.Context, orderID uuid.UUID, reason string) error

	// Payment management
	UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, status entity.PaymentStatus, transactionID string) error
	ProcessPayment(ctx context.Context, orderID uuid.UUID) error

	// Order modifications (for pending/confirmed orders)
	UpdateShippingAddress(ctx context.Context, orderID uuid.UUID, address entity.Address) error
	UpdateBillingAddress(ctx context.Context, orderID uuid.UUID, address entity.Address) error
	AddOrderNote(ctx context.Context, orderID uuid.UUID, note string) error

	// Administrative functions
	GetOrdersByStatus(ctx context.Context, status entity.OrderStatus, page, limit int) ([]*entity.Order, int64, error)
	GetOrdersByPaymentStatus(ctx context.Context, paymentStatus entity.PaymentStatus, page, limit int) ([]*entity.Order, int64, error)
	SearchOrders(ctx context.Context, filters *repository.OrderFilters) ([]*entity.Order, int64, error)

	// Analytics and reporting
	GetOrderStatistics(ctx context.Context, startDate, endDate time.Time) (*repository.OrderStatistics, error)
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]*repository.ProductSalesInfo, error)
	GetSalesMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.SalesMetrics, error)
}

type orderService struct {
	orderRepo        repository.OrderRepository
	cartRepo         repository.CartRepository
	productRepo      repository.ProductRepository
	notificationRepo repository.NotificationRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	notificationRepo repository.NotificationRepository,
) OrderService {
	return &orderService{
		orderRepo:        orderRepo,
		cartRepo:         cartRepo,
		productRepo:      productRepo,
		notificationRepo: notificationRepo,
	}
}

func (s *orderService) CreateOrderFromCart(ctx context.Context, userID uuid.UUID, shippingAddress, billingAddress entity.Address, shippingMethod string, notes string) (*entity.Order, error) {
	// Get user's cart
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, entity.ErrInvalidOrderData
	}

	// Validate product availability and reserve stock
	stockReservations := make([]repository.StockReservation, len(cart.Items))
	for i, cartItem := range cart.Items {
		// Validate product availability
		available, err := s.productRepo.ValidateProductAvailability(ctx, cartItem.ProductID, cartItem.VariantID, cartItem.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to validate product availability: %w", err)
		}
		if !available {
			return nil, fmt.Errorf("product %s is not available in requested quantity", cartItem.SKU)
		}

		stockReservations[i] = repository.StockReservation{
			ProductID: cartItem.ProductID,
			VariantID: cartItem.VariantID,
			Quantity:  cartItem.Quantity,
		}
	}

	// Reserve stock
	if err := s.productRepo.ReserveStock(ctx, stockReservations); err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Create order
	order := &entity.Order{
		ID:               uuid.New(),
		UserID:           userID,
		Status:           entity.OrderStatusPending,
		PaymentStatus:    entity.PaymentStatusPending,
		ShippingAddress:  shippingAddress,
		BillingAddress:   billingAddress,
		ShippingMethod:   shippingMethod,
		Notes:            notes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Convert cart items to order items
	order.Items = make([]entity.OrderItem, len(cart.Items))
	for i, cartItem := range cart.Items {
		orderItem := entity.OrderItem{
			ID:          uuid.New(),
			OrderID:     order.ID,
			ProductID:   cartItem.ProductID,
			VariantID:   cartItem.VariantID,
			ProductName: cartItem.ProductName,
			ProductSKU:  cartItem.SKU,
			Size:        cartItem.Size,
			Color:       cartItem.Color,
			UnitPrice:   cartItem.UnitPrice,
			Quantity:    cartItem.Quantity,
			ImageURL:    cartItem.ImageURL,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		orderItem.CalculateTotal()
		order.Items[i] = orderItem
	}

	// Calculate shipping cost (simplified - in reality this would use shipping service)
	order.ShippingCost = s.calculateShippingCost(shippingMethod, shippingAddress, order.Items)

	// Calculate tax (simplified - in reality this would use tax service)
	order.TaxAmount = s.calculateTax(order.Items, shippingAddress)

	// Save order
	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		// Release reserved stock on failure
		s.productRepo.ReleaseStock(ctx, stockReservations)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Clear user's cart
	if err := s.cartRepo.ClearCartByUserID(ctx, userID); err != nil {
		// Log warning but don't fail the order creation
		fmt.Printf("Warning: failed to clear cart for user %s: %v\n", userID, err)
	}

	// Send order confirmation notification
	if s.notificationRepo != nil {
		if err := s.notificationRepo.SendOrderConfirmation(ctx, order); err != nil {
			// Log warning but don't fail the order creation
			fmt.Printf("Warning: failed to send order confirmation: %v\n", err)
		}
	}

	return order, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*entity.Order, error) {
	return s.orderRepo.GetOrderByID(ctx, orderID)
}

func (s *orderService) GetOrderByNumber(ctx context.Context, orderNumber string) (*entity.Order, error) {
	return s.orderRepo.GetOrderByNumber(ctx, orderNumber)
}

func (s *orderService) GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Order, int64, error) {
	offset := (page - 1) * limit
	return s.orderRepo.GetOrdersByUserID(ctx, userID, limit, offset)
}

func (s *orderService) GetUserOrderSummaries(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.OrderSummary, int64, error) {
	offset := (page - 1) * limit
	return s.orderRepo.GetOrderSummariesByUserID(ctx, userID, limit, offset)
}

func (s *orderService) ConfirmOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(entity.OrderStatusConfirmed) {
		return entity.ErrInvalidStatusTransition
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusConfirmed); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Send status update notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) ProcessOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(entity.OrderStatusProcessing) {
		return entity.ErrInvalidStatusTransition
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusProcessing); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Send status update notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) ShipOrder(ctx context.Context, orderID uuid.UUID, trackingNumber string, estimatedDelivery *time.Time) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(entity.OrderStatusShipped) {
		return entity.ErrInvalidStatusTransition
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusShipped); err != nil {
		return err
	}

	// Update tracking information
	if trackingNumber != "" {
		order.TrackingNumber = trackingNumber
	}
	if estimatedDelivery != nil {
		order.EstimatedDelivery = estimatedDelivery
	}

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Send shipping notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendShippingNotification(ctx, order)
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) DeliverOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(entity.OrderStatusDelivered) {
		return entity.ErrInvalidStatusTransition
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusDelivered); err != nil {
		return err
	}

	order.ActualDelivery = &order.UpdatedAt

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Send delivery notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendDeliveryNotification(ctx, order)
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) CompleteOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanTransitionTo(entity.OrderStatusCompleted) {
		return entity.ErrInvalidStatusTransition
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusCompleted); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Send status update notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) CancelOrder(ctx context.Context, orderID uuid.UUID, reason string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.IsCancellable() {
		return entity.ErrOrderNotCancellable
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusCancelled); err != nil {
		return err
	}

	order.CancelReason = reason

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Release reserved stock
	stockReservations := make([]repository.StockReservation, len(order.Items))
	for i, item := range order.Items {
		stockReservations[i] = repository.StockReservation{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		}
	}

	if s.productRepo != nil {
		if err := s.productRepo.ReleaseStock(ctx, stockReservations); err != nil {
			fmt.Printf("Warning: failed to release stock for cancelled order %s: %v\n", order.ID, err)
		}
	}

	// Send status update notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) RefundOrder(ctx context.Context, orderID uuid.UUID, reason string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.IsRefundable() {
		return entity.ErrOrderNotRefundable
	}

	previousStatus := order.Status
	if err := order.TransitionTo(entity.OrderStatusRefunded); err != nil {
		return err
	}

	order.CancelReason = reason

	if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Send status update notification
	if s.notificationRepo != nil {
		s.notificationRepo.SendOrderStatusUpdate(ctx, order, previousStatus)
	}

	return nil
}

func (s *orderService) UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, status entity.PaymentStatus, transactionID string) error {
	return s.orderRepo.UpdatePaymentStatus(ctx, orderID, status, transactionID)
}

func (s *orderService) ProcessPayment(ctx context.Context, orderID uuid.UUID) error {
	// This would integrate with payment service
	// For now, we'll just update the payment status
	return s.orderRepo.UpdatePaymentStatus(ctx, orderID, entity.PaymentStatusCompleted, "")
}

func (s *orderService) UpdateShippingAddress(ctx context.Context, orderID uuid.UUID, address entity.Address) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.IsEditable() {
		return entity.ErrOrderNotEditable
	}

	order.ShippingAddress = address
	return s.orderRepo.UpdateOrder(ctx, order)
}

func (s *orderService) UpdateBillingAddress(ctx context.Context, orderID uuid.UUID, address entity.Address) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.IsEditable() {
		return entity.ErrOrderNotEditable
	}

	order.BillingAddress = address
	return s.orderRepo.UpdateOrder(ctx, order)
}

func (s *orderService) AddOrderNote(ctx context.Context, orderID uuid.UUID, note string) error {
	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.Notes != "" {
		order.Notes += "\n" + note
	} else {
		order.Notes = note
	}

	return s.orderRepo.UpdateOrder(ctx, order)
}

func (s *orderService) GetOrdersByStatus(ctx context.Context, status entity.OrderStatus, page, limit int) ([]*entity.Order, int64, error) {
	offset := (page - 1) * limit
	return s.orderRepo.GetOrdersByStatus(ctx, status, limit, offset)
}

func (s *orderService) GetOrdersByPaymentStatus(ctx context.Context, paymentStatus entity.PaymentStatus, page, limit int) ([]*entity.Order, int64, error) {
	offset := (page - 1) * limit
	return s.orderRepo.GetOrdersByPaymentStatus(ctx, paymentStatus, limit, offset)
}

func (s *orderService) SearchOrders(ctx context.Context, filters *repository.OrderFilters) ([]*entity.Order, int64, error) {
	return s.orderRepo.SearchOrders(ctx, filters)
}

func (s *orderService) GetOrderStatistics(ctx context.Context, startDate, endDate time.Time) (*repository.OrderStatistics, error) {
	return s.orderRepo.GetOrderStatistics(ctx, startDate, endDate)
}

func (s *orderService) GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]*repository.ProductSalesInfo, error) {
	return s.orderRepo.GetTopProducts(ctx, startDate, endDate, limit)
}

func (s *orderService) GetSalesMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.SalesMetrics, error) {
	return s.orderRepo.GetSalesMetrics(ctx, startDate, endDate)
}

// Helper functions for order calculations
func (s *orderService) calculateShippingCost(shippingMethod string, address entity.Address, items []entity.OrderItem) float64 {
	// Simplified shipping calculation
	// In reality, this would integrate with shipping providers
	switch shippingMethod {
	case "standard":
		return 5.99
	case "express":
		return 12.99
	case "overnight":
		return 24.99
	default:
		return 5.99
	}
}

func (s *orderService) calculateTax(items []entity.OrderItem, address entity.Address) float64 {
	// Simplified tax calculation
	// In reality, this would integrate with tax service based on address
	subtotal := 0.0
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	// Assume 8.5% tax rate for simplicity
	return subtotal * 0.085
}