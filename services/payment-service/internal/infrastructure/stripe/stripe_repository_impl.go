package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
	"solemate/services/payment-service/internal/domain/repository"
)

type stripeRepositoryImpl struct {
	secretKey    string
	webhookSecret string
}

func NewStripeRepository(secretKey, webhookSecret string) repository.StripeRepository {
	stripe.Key = secretKey
	return &stripeRepositoryImpl{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
	}
}

// Payment Intent operations
func (s *stripeRepositoryImpl) CreatePaymentIntent(ctx context.Context, request *repository.CreatePaymentIntentRequest) (*repository.PaymentIntentResponse, error) {
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(request.Amount),
		Currency:    stripe.String(request.Currency),
		Description: stripe.String(request.Description),
	}

	if request.CustomerID != "" {
		params.Customer = stripe.String(request.CustomerID)
	}

	if request.AutomaticPaymentMethods {
		params.AutomaticPaymentMethods = &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		}
	}

	// Add order ID to metadata
	if request.OrderID != "" {
		params.Metadata = map[string]string{
			"order_id": request.OrderID,
		}
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return &repository.PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		CustomerID:   getStringValue(pi.Customer),
	}, nil
}

func (s *stripeRepositoryImpl) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, paymentMethodID string) (*repository.PaymentIntentResponse, error) {
	params := &stripe.PaymentIntentConfirmParams{
		PaymentMethod: stripe.String(paymentMethodID),
	}

	pi, err := paymentintent.Confirm(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm payment intent: %w", err)
	}

	response := &repository.PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		CustomerID:   getStringValue(pi.Customer),
	}

	// Get charge ID if payment succeeded
	if pi.Status == stripe.PaymentIntentStatusSucceeded && pi.LatestCharge != nil {
		response.ChargeID = pi.LatestCharge.ID
	}

	return response, nil
}

func (s *stripeRepositoryImpl) CancelPaymentIntent(ctx context.Context, paymentIntentID string) (*repository.PaymentIntentResponse, error) {
	params := &stripe.PaymentIntentCancelParams{}

	pi, err := paymentintent.Cancel(paymentIntentID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	return &repository.PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		CustomerID:   getStringValue(pi.Customer),
	}, nil
}

func (s *stripeRepositoryImpl) RetrievePaymentIntent(ctx context.Context, paymentIntentID string) (*repository.PaymentIntentResponse, error) {
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %w", err)
	}

	response := &repository.PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		CustomerID:   getStringValue(pi.Customer),
	}

	// Get charge ID if payment succeeded
	if pi.Status == stripe.PaymentIntentStatusSucceeded && pi.LatestCharge != nil {
		response.ChargeID = pi.LatestCharge.ID
	}

	return response, nil
}

// Payment Method operations
func (s *stripeRepositoryImpl) CreatePaymentMethod(ctx context.Context, request *repository.CreatePaymentMethodRequest) (*repository.PaymentMethodResponse, error) {
	params := &stripe.PaymentMethodParams{
		Type: stripe.String(request.Type),
	}

	if request.Card != nil {
		params.Card = &stripe.PaymentMethodCardParams{
			Number:   stripe.String(request.Card.Number),
			ExpMonth: stripe.Int64(int64(request.Card.ExpMonth)),
			ExpYear:  stripe.Int64(int64(request.Card.ExpYear)),
			CVC:      stripe.String(request.Card.CVC),
		}
	}

	pm, err := paymentmethod.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	response := &repository.PaymentMethodResponse{
		ID:   pm.ID,
		Type: string(pm.Type),
	}

	if pm.Card != nil {
		response.Card = &repository.Card{
			Brand:    string(pm.Card.Brand),
			Last4:    pm.Card.Last4,
			ExpMonth: int(pm.Card.ExpMonth),
			ExpYear:  int(pm.Card.ExpYear),
		}
	}

	return response, nil
}

func (s *stripeRepositoryImpl) AttachPaymentMethod(ctx context.Context, paymentMethodID string, customerID string) (*repository.PaymentMethodResponse, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	pm, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to attach payment method: %w", err)
	}

	response := &repository.PaymentMethodResponse{
		ID:         pm.ID,
		Type:       string(pm.Type),
		CustomerID: getStringValue(pm.Customer),
	}

	if pm.Card != nil {
		response.Card = &repository.Card{
			Brand:    string(pm.Card.Brand),
			Last4:    pm.Card.Last4,
			ExpMonth: int(pm.Card.ExpMonth),
			ExpYear:  int(pm.Card.ExpYear),
		}
	}

	return response, nil
}

func (s *stripeRepositoryImpl) DetachPaymentMethod(ctx context.Context, paymentMethodID string) (*repository.PaymentMethodResponse, error) {
	pm, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to detach payment method: %w", err)
	}

	response := &repository.PaymentMethodResponse{
		ID:   pm.ID,
		Type: string(pm.Type),
	}

	if pm.Card != nil {
		response.Card = &repository.Card{
			Brand:    string(pm.Card.Brand),
			Last4:    pm.Card.Last4,
			ExpMonth: int(pm.Card.ExpMonth),
			ExpYear:  int(pm.Card.ExpYear),
		}
	}

	return response, nil
}

func (s *stripeRepositoryImpl) RetrievePaymentMethod(ctx context.Context, paymentMethodID string) (*repository.PaymentMethodResponse, error) {
	pm, err := paymentmethod.Get(paymentMethodID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment method: %w", err)
	}

	response := &repository.PaymentMethodResponse{
		ID:         pm.ID,
		Type:       string(pm.Type),
		CustomerID: getStringValue(pm.Customer),
	}

	if pm.Card != nil {
		response.Card = &repository.Card{
			Brand:    string(pm.Card.Brand),
			Last4:    pm.Card.Last4,
			ExpMonth: int(pm.Card.ExpMonth),
			ExpYear:  int(pm.Card.ExpYear),
		}
	}

	return response, nil
}

// Customer operations
func (s *stripeRepositoryImpl) CreateCustomer(ctx context.Context, request *repository.CreateCustomerRequest) (*repository.CustomerResponse, error) {
	params := &stripe.CustomerParams{
		Email:       stripe.String(request.Email),
		Name:        stripe.String(request.Name),
		Description: stripe.String(request.Description),
	}

	if request.Address != nil {
		params.Address = &stripe.AddressParams{
			Line1:      stripe.String(request.Address.Line1),
			Line2:      stripe.String(request.Address.Line2),
			City:       stripe.String(request.Address.City),
			State:      stripe.String(request.Address.State),
			PostalCode: stripe.String(request.Address.PostalCode),
			Country:    stripe.String(request.Address.Country),
		}
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	response := &repository.CustomerResponse{
		ID:          cust.ID,
		Email:       cust.Email,
		Name:        cust.Name,
		Description: cust.Description,
	}

	if cust.Address != nil {
		response.Address = &repository.StripeAddress{
			Line1:      cust.Address.Line1,
			Line2:      cust.Address.Line2,
			City:       cust.Address.City,
			State:      cust.Address.State,
			PostalCode: cust.Address.PostalCode,
			Country:    cust.Address.Country,
		}
	}

	return response, nil
}

func (s *stripeRepositoryImpl) RetrieveCustomer(ctx context.Context, customerID string) (*repository.CustomerResponse, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}

	response := &repository.CustomerResponse{
		ID:          cust.ID,
		Email:       cust.Email,
		Name:        cust.Name,
		Description: cust.Description,
	}

	if cust.Address != nil {
		response.Address = &repository.StripeAddress{
			Line1:      cust.Address.Line1,
			Line2:      cust.Address.Line2,
			City:       cust.Address.City,
			State:      cust.Address.State,
			PostalCode: cust.Address.PostalCode,
			Country:    cust.Address.Country,
		}
	}

	return response, nil
}

func (s *stripeRepositoryImpl) UpdateCustomer(ctx context.Context, customerID string, request *repository.UpdateCustomerRequest) (*repository.CustomerResponse, error) {
	params := &stripe.CustomerParams{}

	if request.Email != "" {
		params.Email = stripe.String(request.Email)
	}
	if request.Name != "" {
		params.Name = stripe.String(request.Name)
	}
	if request.Description != "" {
		params.Description = stripe.String(request.Description)
	}

	if request.Address != nil {
		params.Address = &stripe.AddressParams{
			Line1:      stripe.String(request.Address.Line1),
			Line2:      stripe.String(request.Address.Line2),
			City:       stripe.String(request.Address.City),
			State:      stripe.String(request.Address.State),
			PostalCode: stripe.String(request.Address.PostalCode),
			Country:    stripe.String(request.Address.Country),
		}
	}

	cust, err := customer.Update(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	response := &repository.CustomerResponse{
		ID:          cust.ID,
		Email:       cust.Email,
		Name:        cust.Name,
		Description: cust.Description,
	}

	if cust.Address != nil {
		response.Address = &repository.StripeAddress{
			Line1:      cust.Address.Line1,
			Line2:      cust.Address.Line2,
			City:       cust.Address.City,
			State:      cust.Address.State,
			PostalCode: cust.Address.PostalCode,
			Country:    cust.Address.Country,
		}
	}

	return response, nil
}

// Refund operations
func (s *stripeRepositoryImpl) CreateRefund(ctx context.Context, request *repository.CreateRefundRequest) (*repository.RefundResponse, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(request.PaymentIntentID),
		Reason:        stripe.String(request.Reason),
	}

	if request.Amount != nil {
		params.Amount = stripe.Int64(*request.Amount)
	}

	if request.Metadata != nil {
		params.Metadata = request.Metadata
	}

	ref, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return &repository.RefundResponse{
		ID:              ref.ID,
		Amount:          ref.Amount,
		Currency:        string(ref.Currency),
		PaymentIntentID: ref.PaymentIntent.ID,
		Status:          string(ref.Status),
		Reason:          string(ref.Reason),
	}, nil
}

func (s *stripeRepositoryImpl) RetrieveRefund(ctx context.Context, refundID string) (*repository.RefundResponse, error) {
	ref, err := refund.Get(refundID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve refund: %w", err)
	}

	return &repository.RefundResponse{
		ID:              ref.ID,
		Amount:          ref.Amount,
		Currency:        string(ref.Currency),
		PaymentIntentID: ref.PaymentIntent.ID,
		Status:          string(ref.Status),
		Reason:          string(ref.Reason),
	}, nil
}

// Webhook operations
func (s *stripeRepositoryImpl) ConstructWebhookEvent(ctx context.Context, payload []byte, signature string) (*repository.WebhookEventData, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to construct webhook event: %w", err)
	}

	// Convert the event data to our format
	eventData := &repository.WebhookEventData{
		ID:      event.ID,
		Type:    string(event.Type),
		Created: event.Created,
	}

	// Parse the data based on event type
	switch event.Type {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err == nil {
			eventData.Data = pi
		}
	case "charge.dispute.created":
		var dispute stripe.Dispute
		if err := json.Unmarshal(event.Data.Raw, &dispute); err == nil {
			eventData.Data = dispute
		}
	default:
		// For other events, store the raw data
		eventData.Data = event.Data.Object
	}

	return eventData, nil
}

// Helper functions
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	if obj, ok := value.(*stripe.Customer); ok && obj != nil {
		return obj.ID
	}
	return ""
}