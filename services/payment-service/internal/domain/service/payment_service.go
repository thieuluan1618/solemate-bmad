package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"solemate/services/payment-service/internal/domain/entity"
	"solemate/services/payment-service/internal/domain/repository"
)

type PaymentService interface {
	// Payment operations
	CreatePayment(ctx context.Context, request *CreatePaymentRequest) (*PaymentResponse, error)
	GetPayment(ctx context.Context, paymentID uuid.UUID) (*PaymentResponse, error)
	GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*PaymentResponse, error)
	GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*PaymentSummaryResponse, int64, error)
	ProcessPayment(ctx context.Context, paymentID uuid.UUID, paymentMethodID string) (*PaymentResponse, error)
	CancelPayment(ctx context.Context, paymentID uuid.UUID) (*PaymentResponse, error)

	// Payment method operations
	CreatePaymentMethod(ctx context.Context, request *CreatePaymentMethodRequest) (*PaymentMethodResponse, error)
	GetPaymentMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*PaymentMethodResponse, error)
	SetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID uuid.UUID) error
	DeletePaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error

	// Refund operations
	CreateRefund(ctx context.Context, request *CreateRefundRequest) (*RefundResponse, error)
	GetRefund(ctx context.Context, refundID uuid.UUID) (*RefundResponse, error)
	GetRefundsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*RefundResponse, error)

	// Webhook handling
	ProcessWebhook(ctx context.Context, payload []byte, signature string) error

	// Analytics and reporting
	GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time) (*repository.PaymentStatistics, error)
	GetRevenueMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.RevenueMetrics, error)
}

type paymentService struct {
	paymentRepo       repository.PaymentRepository
	paymentMethodRepo repository.PaymentMethodRepository
	refundRepo        repository.RefundRepository
	webhookRepo       repository.WebhookRepository
	stripeRepo        repository.StripeRepository
	orderRepo         repository.OrderRepository
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	paymentMethodRepo repository.PaymentMethodRepository,
	refundRepo repository.RefundRepository,
	webhookRepo repository.WebhookRepository,
	stripeRepo repository.StripeRepository,
	orderRepo repository.OrderRepository,
) PaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		paymentMethodRepo: paymentMethodRepo,
		refundRepo:        refundRepo,
		webhookRepo:       webhookRepo,
		stripeRepo:        stripeRepo,
		orderRepo:         orderRepo,
	}
}

// Payment operations
func (s *paymentService) CreatePayment(ctx context.Context, request *CreatePaymentRequest) (*PaymentResponse, error) {
	// Validate order exists
	_, err := s.orderRepo.GetOrderByID(ctx, request.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if payment already exists for this order
	existingPayment, err := s.paymentRepo.GetPaymentByOrderID(ctx, request.OrderID)
	if err == nil && existingPayment != nil {
		return nil, entity.ErrPaymentAlreadyProcessed
	}

	// Create Stripe payment intent
	stripeRequest := &repository.CreatePaymentIntentRequest{
		Amount:                  int64(request.Amount * 100), // Convert to cents
		Currency:                request.Currency,
		CustomerID:              request.StripeCustomerID,
		Description:             fmt.Sprintf("Payment for order %s", request.OrderID.String()),
		OrderID:                 request.OrderID.String(),
		AutomaticPaymentMethods: true,
	}

	stripeResponse, err := s.stripeRepo.CreatePaymentIntent(ctx, stripeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe payment intent: %w", err)
	}

	// Create payment entity
	payment := &entity.Payment{
		ID:                    uuid.New(),
		UserID:                request.UserID,
		OrderID:               request.OrderID,
		Amount:                request.Amount,
		Currency:              request.Currency,
		Status:                entity.PaymentStatusPending,
		StripePaymentIntentID: stripeResponse.ID,
		ClientSecret:          stripeResponse.ClientSecret,
		Description:           request.Description,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return s.mapPaymentToResponse(payment), nil
}

func (s *paymentService) GetPayment(ctx context.Context, paymentID uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, entity.ErrPaymentNotFound
	}

	return s.mapPaymentToResponse(payment), nil
}

func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, entity.ErrPaymentNotFound
	}

	return s.mapPaymentToResponse(payment), nil
}

func (s *paymentService) GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*PaymentSummaryResponse, int64, error) {
	summaries, total, err := s.paymentRepo.GetPaymentSummariesByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payment summaries: %w", err)
	}

	responses := make([]*PaymentSummaryResponse, len(summaries))
	for i, summary := range summaries {
		responses[i] = &PaymentSummaryResponse{
			ID:          summary.ID,
			OrderID:     summary.OrderID,
			Amount:      summary.Amount,
			Currency:    summary.Currency,
			Status:      summary.Status,
			CardLast4:   summary.CardLast4,
			CardBrand:   summary.CardBrand,
			CreatedAt:   summary.CreatedAt,
			ProcessedAt: summary.ProcessedAt,
		}
	}

	return responses, total, nil
}

func (s *paymentService) ProcessPayment(ctx context.Context, paymentID uuid.UUID, paymentMethodID string) (*PaymentResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, entity.ErrPaymentNotFound
	}

	// Check if payment can be processed
	if !payment.CanBeCanceled() {
		return nil, entity.ErrPaymentAlreadyProcessed
	}

	// Mark payment as processing
	payment.MarkAsProcessing()
	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Confirm payment intent with Stripe
	stripeResponse, err := s.stripeRepo.ConfirmPaymentIntent(ctx, payment.StripePaymentIntentID, paymentMethodID)
	if err != nil {
		// Mark payment as failed
		payment.MarkAsFailed("Payment confirmation failed", "stripe_error")
		s.paymentRepo.UpdatePayment(ctx, payment)
		return nil, fmt.Errorf("failed to confirm payment: %w", err)
	}

	// Update payment based on Stripe response
	if stripeResponse.Status == "succeeded" {
		payment.MarkAsSucceeded()
		payment.StripeChargeID = stripeResponse.ChargeID

		// Update order status
		if err := s.orderRepo.UpdateOrderPaymentStatus(ctx, payment.OrderID, "confirmed", payment.ID.String()); err != nil {
			// Log error but don't fail the payment
			fmt.Printf("Failed to update order status: %v", err)
		}
	} else if stripeResponse.Status == "requires_action" {
		// Payment requires additional action (3D Secure, etc.)
		payment.Status = entity.PaymentStatusPending
		payment.UpdatedAt = time.Now()
	} else {
		payment.MarkAsFailed("Payment not successful", "payment_failed")
	}

	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return s.mapPaymentToResponse(payment), nil
}

func (s *paymentService) CancelPayment(ctx context.Context, paymentID uuid.UUID) (*PaymentResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, entity.ErrPaymentNotFound
	}

	// Check if payment can be canceled
	if !payment.CanBeCanceled() {
		return nil, entity.ErrPaymentNotCancelable
	}

	// Cancel payment intent with Stripe
	_, err = s.stripeRepo.CancelPaymentIntent(ctx, payment.StripePaymentIntentID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	// Update payment status
	payment.Status = entity.PaymentStatusCanceled
	payment.UpdatedAt = time.Now()

	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return s.mapPaymentToResponse(payment), nil
}

// Payment method operations
func (s *paymentService) CreatePaymentMethod(ctx context.Context, request *CreatePaymentMethodRequest) (*PaymentMethodResponse, error) {
	// Create payment method with Stripe
	stripeRequest := &repository.CreatePaymentMethodRequest{
		Type: request.Type,
		Card: request.Card,
	}

	stripeResponse, err := s.stripeRepo.CreatePaymentMethod(ctx, stripeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe payment method: %w", err)
	}

	// Attach to customer if customer ID provided
	if request.StripeCustomerID != "" {
		_, err = s.stripeRepo.AttachPaymentMethod(ctx, stripeResponse.ID, request.StripeCustomerID)
		if err != nil {
			return nil, fmt.Errorf("failed to attach payment method to customer: %w", err)
		}
	}

	// Create payment method entity
	paymentMethod := &entity.PaymentMethod{
		ID:                    uuid.New(),
		UserID:                request.UserID,
		Type:                  entity.PaymentMethodType(stripeResponse.Type),
		StripePaymentMethodID: stripeResponse.ID,
		BillingAddress:        request.BillingAddress,
		IsDefault:             request.IsDefault,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if stripeResponse.Card != nil {
		paymentMethod.CardBrand = stripeResponse.Card.Brand
		paymentMethod.CardLast4 = stripeResponse.Card.Last4
		paymentMethod.CardExpMonth = stripeResponse.Card.ExpMonth
		paymentMethod.CardExpYear = stripeResponse.Card.ExpYear
	}

	// If this is set as default, unset other defaults first
	if request.IsDefault {
		if err := s.paymentMethodRepo.UnsetDefaultPaymentMethods(ctx, request.UserID); err != nil {
			return nil, fmt.Errorf("failed to unset default payment methods: %w", err)
		}
	}

	if err := s.paymentMethodRepo.CreatePaymentMethod(ctx, paymentMethod); err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	return s.mapPaymentMethodToResponse(paymentMethod), nil
}

func (s *paymentService) GetPaymentMethodsByUserID(ctx context.Context, userID uuid.UUID) ([]*PaymentMethodResponse, error) {
	paymentMethods, err := s.paymentMethodRepo.GetPaymentMethodsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment methods: %w", err)
	}

	responses := make([]*PaymentMethodResponse, len(paymentMethods))
	for i, pm := range paymentMethods {
		responses[i] = s.mapPaymentMethodToResponse(pm)
	}

	return responses, nil
}

func (s *paymentService) SetDefaultPaymentMethod(ctx context.Context, userID uuid.UUID, paymentMethodID uuid.UUID) error {
	return s.paymentMethodRepo.SetDefaultPaymentMethod(ctx, userID, paymentMethodID)
}

func (s *paymentService) DeletePaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error {
	// Get payment method
	paymentMethod, err := s.paymentMethodRepo.GetPaymentMethodByID(ctx, paymentMethodID)
	if err != nil {
		return entity.ErrPaymentMethodNotFound
	}

	// Detach from Stripe
	if _, err := s.stripeRepo.DetachPaymentMethod(ctx, paymentMethod.StripePaymentMethodID); err != nil {
		return fmt.Errorf("failed to detach payment method from Stripe: %w", err)
	}

	// Delete from database
	return s.paymentMethodRepo.DeletePaymentMethod(ctx, paymentMethodID)
}

// Refund operations
func (s *paymentService) CreateRefund(ctx context.Context, request *CreateRefundRequest) (*RefundResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.GetPaymentByID(ctx, request.PaymentID)
	if err != nil {
		return nil, entity.ErrPaymentNotFound
	}

	// Check if payment can be refunded
	if !payment.CanBeRefunded() {
		return nil, entity.ErrPaymentNotRefundable
	}

	// Check refund amount
	refundableAmount := payment.GetRefundableAmount()
	if request.Amount > refundableAmount {
		return nil, entity.ErrInsufficientAmount
	}

	// Create refund with Stripe
	stripeRequest := &repository.CreateRefundRequest{
		PaymentIntentID: payment.StripePaymentIntentID,
		Reason:          request.Reason,
		Metadata:        map[string]string{"order_id": payment.OrderID.String()},
	}

	if request.Amount < payment.Amount {
		amountCents := int64(request.Amount * 100)
		stripeRequest.Amount = &amountCents
	}

	stripeResponse, err := s.stripeRepo.CreateRefund(ctx, stripeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe refund: %w", err)
	}

	// Create refund entity
	refund := &entity.Refund{
		ID:             uuid.New(),
		PaymentID:      request.PaymentID,
		UserID:         payment.UserID,
		OrderID:        payment.OrderID,
		Amount:         request.Amount,
		Currency:       payment.Currency,
		Status:         entity.RefundStatusPending,
		StripeRefundID: stripeResponse.ID,
		Reason:         request.Reason,
		Description:    request.Description,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if stripeResponse.Status == "succeeded" {
		refund.MarkAsSucceeded()
	}

	if err := s.refundRepo.CreateRefund(ctx, refund); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return s.mapRefundToResponse(refund), nil
}

func (s *paymentService) GetRefund(ctx context.Context, refundID uuid.UUID) (*RefundResponse, error) {
	refund, err := s.refundRepo.GetRefundByID(ctx, refundID)
	if err != nil {
		return nil, fmt.Errorf("refund not found: %w", err)
	}

	return s.mapRefundToResponse(refund), nil
}

func (s *paymentService) GetRefundsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]*RefundResponse, error) {
	refunds, err := s.refundRepo.GetRefundsByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refunds: %w", err)
	}

	responses := make([]*RefundResponse, len(refunds))
	for i, refund := range refunds {
		responses[i] = s.mapRefundToResponse(refund)
	}

	return responses, nil
}

// Webhook handling
func (s *paymentService) ProcessWebhook(ctx context.Context, payload []byte, signature string) error {
	// Construct webhook event
	eventData, err := s.stripeRepo.ConstructWebhookEvent(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to construct webhook event: %w", err)
	}

	// Check if event already processed
	existingEvent, err := s.webhookRepo.GetWebhookEventByStripeID(ctx, eventData.ID)
	if err == nil && existingEvent != nil && existingEvent.Processed {
		return nil // Already processed
	}

	// Create webhook event record
	webhookEvent := &entity.WebhookEvent{
		ID:            uuid.New(),
		StripeEventID: eventData.ID,
		EventType:     eventData.Type,
		Data:          string(payload),
		Processed:     false,
		CreatedAt:     time.Now(),
	}

	if existingEvent == nil {
		if err := s.webhookRepo.CreateWebhookEvent(ctx, webhookEvent); err != nil {
			return fmt.Errorf("failed to create webhook event: %w", err)
		}
	} else {
		webhookEvent = existingEvent
	}

	// Process the event
	if err := s.processWebhookEvent(ctx, webhookEvent, eventData); err != nil {
		s.webhookRepo.MarkWebhookEventAsFailed(ctx, webhookEvent.ID, err.Error())
		return fmt.Errorf("failed to process webhook event: %w", err)
	}

	// Mark as processed
	return s.webhookRepo.MarkWebhookEventAsProcessed(ctx, webhookEvent.ID)
}

func (s *paymentService) processWebhookEvent(ctx context.Context, webhookEvent *entity.WebhookEvent, eventData *repository.WebhookEventData) error {
	switch eventData.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(ctx, eventData)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(ctx, eventData)
	case "payment_intent.canceled":
		return s.handlePaymentIntentCanceled(ctx, eventData)
	case "charge.dispute.created":
		return s.handleChargeDisputeCreated(ctx, eventData)
	default:
		// Log unhandled event type but don't fail
		fmt.Printf("Unhandled webhook event type: %s", eventData.Type)
		return nil
	}
}

func (s *paymentService) handlePaymentIntentSucceeded(ctx context.Context, eventData *repository.WebhookEventData) error {
	// Parse payment intent data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(eventData.Data.(string)), &data); err != nil {
		if paymentIntent, ok := eventData.Data.(map[string]interface{}); ok {
			data = paymentIntent
		} else {
			return fmt.Errorf("failed to parse payment intent data")
		}
	}

	paymentIntentID, ok := data["id"].(string)
	if !ok {
		return fmt.Errorf("payment intent ID not found")
	}

	// Get payment by Stripe payment intent ID
	payment, err := s.paymentRepo.GetPaymentByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil {
		return fmt.Errorf("payment not found for payment intent %s", paymentIntentID)
	}

	// Update payment status
	payment.MarkAsSucceeded()
	if charges, ok := data["charges"].(map[string]interface{}); ok {
		if chargeData, ok := charges["data"].([]interface{}); ok && len(chargeData) > 0 {
			if charge, ok := chargeData[0].(map[string]interface{}); ok {
				if chargeID, ok := charge["id"].(string); ok {
					payment.StripeChargeID = chargeID
				}
			}
		}
	}

	if err := s.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Update order status
	return s.orderRepo.UpdateOrderPaymentStatus(ctx, payment.OrderID, "confirmed", payment.ID.String())
}

func (s *paymentService) handlePaymentIntentFailed(ctx context.Context, eventData *repository.WebhookEventData) error {
	// Similar implementation for failed payments
	return nil
}

func (s *paymentService) handlePaymentIntentCanceled(ctx context.Context, eventData *repository.WebhookEventData) error {
	// Similar implementation for canceled payments
	return nil
}

func (s *paymentService) handleChargeDisputeCreated(ctx context.Context, eventData *repository.WebhookEventData) error {
	// Handle dispute creation - update payment status, notify admin, etc.
	return nil
}

// Analytics and reporting
func (s *paymentService) GetPaymentStatistics(ctx context.Context, startDate, endDate time.Time) (*repository.PaymentStatistics, error) {
	return s.paymentRepo.GetPaymentStatistics(ctx, startDate, endDate)
}

func (s *paymentService) GetRevenueMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.RevenueMetrics, error) {
	return s.paymentRepo.GetRevenueMetrics(ctx, startDate, endDate)
}

// Helper methods
func (s *paymentService) mapPaymentToResponse(payment *entity.Payment) *PaymentResponse {
	response := &PaymentResponse{
		ID:                    payment.ID,
		UserID:                payment.UserID,
		OrderID:               payment.OrderID,
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		Status:                payment.Status,
		StripePaymentIntentID: payment.StripePaymentIntentID,
		StripeChargeID:        payment.StripeChargeID,
		ClientSecret:          payment.ClientSecret,
		Description:           payment.Description,
		FailureReason:         payment.FailureReason,
		FailureCode:           payment.FailureCode,
		CreatedAt:             payment.CreatedAt,
		UpdatedAt:             payment.UpdatedAt,
		ProcessedAt:           payment.ProcessedAt,
		FailedAt:              payment.FailedAt,
	}

	if payment.PaymentMethod != nil {
		response.PaymentMethod = s.mapPaymentMethodToResponse(payment.PaymentMethod)
	}

	for _, refund := range payment.Refunds {
		response.Refunds = append(response.Refunds, s.mapRefundToResponse(&refund))
	}

	return response
}

func (s *paymentService) mapPaymentMethodToResponse(pm *entity.PaymentMethod) *PaymentMethodResponse {
	return &PaymentMethodResponse{
		ID:                    pm.ID,
		UserID:                pm.UserID,
		Type:                  pm.Type,
		StripePaymentMethodID: pm.StripePaymentMethodID,
		CardBrand:             pm.CardBrand,
		CardLast4:             pm.CardLast4,
		CardExpMonth:          pm.CardExpMonth,
		CardExpYear:           pm.CardExpYear,
		BillingAddress:        pm.BillingAddress,
		IsDefault:             pm.IsDefault,
		IsExpired:             pm.IsExpired(),
		DisplayName:           pm.GetDisplayName(),
		CreatedAt:             pm.CreatedAt,
		UpdatedAt:             pm.UpdatedAt,
	}
}

func (s *paymentService) mapRefundToResponse(refund *entity.Refund) *RefundResponse {
	return &RefundResponse{
		ID:             refund.ID,
		PaymentID:      refund.PaymentID,
		UserID:         refund.UserID,
		OrderID:        refund.OrderID,
		Amount:         refund.Amount,
		Currency:       refund.Currency,
		Status:         refund.Status,
		StripeRefundID: refund.StripeRefundID,
		Reason:         refund.Reason,
		Description:    refund.Description,
		FailureReason:  refund.FailureReason,
		CreatedAt:      refund.CreatedAt,
		UpdatedAt:      refund.UpdatedAt,
		ProcessedAt:    refund.ProcessedAt,
	}
}