package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"solemate/services/payment-service/internal/domain/repository"
)

type orderRepositoryImpl struct {
	baseURL    string
	httpClient *http.Client
}

func NewOrderRepository(baseURL string) repository.OrderRepository {
	return &orderRepositoryImpl{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *orderRepositoryImpl) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*repository.OrderData, error) {
	url := fmt.Sprintf("%s/api/v1/orders/%s", r.baseURL, orderID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("order not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data *repository.OrderData `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

func (r *orderRepositoryImpl) UpdateOrderPaymentStatus(ctx context.Context, orderID uuid.UUID, status string, transactionID string) error {
	url := fmt.Sprintf("%s/api/v1/orders/%s/payment-status", r.baseURL, orderID.String())

	payload := map[string]interface{}{
		"payment_status":  status,
		"transaction_id":  transactionID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update order payment status: status code %d", resp.StatusCode)
	}

	return nil
}