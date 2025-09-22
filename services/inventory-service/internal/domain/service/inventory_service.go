package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"solemate/services/inventory-service/internal/domain/entity"
	"solemate/services/inventory-service/internal/domain/repository"
)

type InventoryService interface {
	// Inventory management
	CreateInventoryItem(ctx context.Context, request *CreateInventoryItemRequest) (*InventoryItemResponse, error)
	GetInventoryItem(ctx context.Context, itemID uuid.UUID) (*InventoryItemResponse, error)
	UpdateInventoryItem(ctx context.Context, itemID uuid.UUID, request *UpdateInventoryItemRequest) (*InventoryItemResponse, error)
	DeleteInventoryItem(ctx context.Context, itemID uuid.UUID) error
	SearchInventoryItems(ctx context.Context, request *SearchInventoryRequest) (*PaginatedInventoryResponse, error)

	// Stock operations
	CheckStockAvailability(ctx context.Context, request *StockAvailabilityRequest) (*StockAvailabilityResponse, error)
	ReserveStock(ctx context.Context, request *ReserveStockRequest) (*StockReservationResponse, error)
	ReleaseStockReservation(ctx context.Context, reservationID uuid.UUID) error
	FulfillStockReservation(ctx context.Context, reservationID uuid.UUID) error
	AdjustStock(ctx context.Context, request *AdjustStockRequest) (*StockMovementResponse, error)
	TransferStock(ctx context.Context, request *TransferStockRequest) ([]*StockMovementResponse, error)

	// Bulk operations
	BulkStockUpdate(ctx context.Context, request *BulkStockUpdateRequest) (*BulkOperationResponse, error)
	BulkReserveStock(ctx context.Context, request *BulkReserveStockRequest) (*BulkReservationResponse, error)

	// Warehouse management
	CreateWarehouse(ctx context.Context, request *CreateWarehouseRequest) (*WarehouseResponse, error)
	GetWarehouse(ctx context.Context, warehouseID uuid.UUID) (*WarehouseResponse, error)
	UpdateWarehouse(ctx context.Context, warehouseID uuid.UUID, request *UpdateWarehouseRequest) (*WarehouseResponse, error)
	GetAllWarehouses(ctx context.Context, activeOnly bool) ([]*WarehouseResponse, error)
	GetWarehouseInventorySummary(ctx context.Context, warehouseID uuid.UUID) (*WarehouseInventorySummaryResponse, error)

	// Stock movements and history
	GetStockMovements(ctx context.Context, request *StockMovementHistoryRequest) (*PaginatedStockMovementResponse, error)
	GetMovementSummary(ctx context.Context, request *MovementSummaryRequest) (*MovementSummaryResponse, error)

	// Alerts and notifications
	GetStockAlerts(ctx context.Context, request *StockAlertsRequest) (*PaginatedStockAlertResponse, error)
	MarkAlertAsRead(ctx context.Context, alertID uuid.UUID) error
	MarkAlertAsResolved(ctx context.Context, alertID uuid.UUID) error
	GenerateStockAlerts(ctx context.Context) (*AlertGenerationResponse, error)

	// Analytics and reporting
	GetInventoryAnalytics(ctx context.Context, request *InventoryAnalyticsRequest) (*InventoryAnalyticsResponse, error)
	GetTurnoverReport(ctx context.Context, request *TurnoverReportRequest) (*TurnoverReportResponse, error)
	GetStockValuationReport(ctx context.Context, warehouseID *uuid.UUID) (*StockValuationResponse, error)

	// Order integration
	AllocateStockForOrder(ctx context.Context, orderID uuid.UUID) (*OrderStockAllocationResponse, error)
	ReleaseOrderStock(ctx context.Context, orderID uuid.UUID) error
}

type inventoryService struct {
	inventoryRepo     repository.InventoryRepository
	warehouseRepo     repository.WarehouseRepository
	movementRepo      repository.StockMovementRepository
	reservationRepo   repository.StockReservationRepository
	alertRepo         repository.StockAlertRepository
	productRepo       repository.ProductRepository
	orderRepo         repository.OrderRepository
}

func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	warehouseRepo repository.WarehouseRepository,
	movementRepo repository.StockMovementRepository,
	reservationRepo repository.StockReservationRepository,
	alertRepo repository.StockAlertRepository,
	productRepo repository.ProductRepository,
	orderRepo repository.OrderRepository,
) InventoryService {
	return &inventoryService{
		inventoryRepo:   inventoryRepo,
		warehouseRepo:   warehouseRepo,
		movementRepo:    movementRepo,
		reservationRepo: reservationRepo,
		alertRepo:       alertRepo,
		productRepo:     productRepo,
		orderRepo:       orderRepo,
	}
}

// Inventory management
func (s *inventoryService) CreateInventoryItem(ctx context.Context, request *CreateInventoryItemRequest) (*InventoryItemResponse, error) {
	// Validate product exists
	if s.productRepo != nil {
		exists, err := s.productRepo.ValidateProductExists(ctx, request.ProductID, request.VariantID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate product: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("product or variant does not exist")
		}
	}

	// Validate warehouse exists
	warehouse, err := s.warehouseRepo.GetWarehouseByID(ctx, request.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	// Check if inventory item already exists
	existing, err := s.inventoryRepo.GetInventoryItemByProductAndWarehouse(ctx, request.ProductID, request.WarehouseID, request.VariantID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("inventory item already exists for this product and warehouse")
	}

	// Create new inventory item
	item := &entity.InventoryItem{
		ID:                uuid.New(),
		ProductID:         request.ProductID,
		VariantID:         request.VariantID,
		WarehouseID:       request.WarehouseID,
		QuantityAvailable: request.InitialQuantity,
		QuantityTotal:     request.InitialQuantity,
		MinStockLevel:     request.MinStockLevel,
		MaxStockLevel:     request.MaxStockLevel,
		ReorderPoint:      request.ReorderPoint,
		Location:          request.Location,
		SKU:               request.SKU,
		Barcode:           request.Barcode,
		CostPrice:         request.CostPrice,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	item.UpdateStatus()

	if err := s.inventoryRepo.CreateInventoryItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}

	// Create initial stock movement if quantity > 0
	if request.InitialQuantity > 0 {
		movement := &entity.StockMovement{
			ID:               uuid.New(),
			InventoryItemID:  item.ID,
			Type:             entity.MovementTypeInbound,
			Quantity:         request.InitialQuantity,
			PreviousQuantity: 0,
			NewQuantity:      request.InitialQuantity,
			ReferenceType:    "initial_stock",
			Reason:           "Initial inventory setup",
			UnitCost:         request.CostPrice,
			TotalCost:        float64(request.InitialQuantity) * request.CostPrice,
			UserName:         request.UserName,
			CreatedAt:        time.Now(),
			MovementDate:     time.Now(),
		}

		if err := s.movementRepo.CreateStockMovement(ctx, movement); err != nil {
			// Log error but don't fail the creation
			fmt.Printf("Failed to create initial stock movement: %v", err)
		}
	}

	return s.mapInventoryItemToResponse(item, warehouse), nil
}

func (s *inventoryService) GetInventoryItem(ctx context.Context, itemID uuid.UUID) (*InventoryItemResponse, error) {
	item, err := s.inventoryRepo.GetInventoryItemByID(ctx, itemID)
	if err != nil {
		return nil, entity.ErrInventoryNotFound
	}

	var warehouse *entity.Warehouse
	if item.WarehouseID != uuid.Nil {
		warehouse, _ = s.warehouseRepo.GetWarehouseByID(ctx, item.WarehouseID)
	}

	return s.mapInventoryItemToResponse(item, warehouse), nil
}

func (s *inventoryService) UpdateInventoryItem(ctx context.Context, itemID uuid.UUID, request *UpdateInventoryItemRequest) (*InventoryItemResponse, error) {
	item, err := s.inventoryRepo.GetInventoryItemByID(ctx, itemID)
	if err != nil {
		return nil, entity.ErrInventoryNotFound
	}

	// Update fields if provided
	if request.MinStockLevel != nil {
		item.MinStockLevel = *request.MinStockLevel
	}
	if request.MaxStockLevel != nil {
		item.MaxStockLevel = *request.MaxStockLevel
	}
	if request.ReorderPoint != nil {
		item.ReorderPoint = *request.ReorderPoint
	}
	if request.Location != nil {
		item.Location = *request.Location
	}
	if request.CostPrice != nil {
		item.LastCostPrice = item.CostPrice
		item.CostPrice = *request.CostPrice
	}

	item.UpdatedAt = time.Now()
	item.UpdateStatus()

	if err := s.inventoryRepo.UpdateInventoryItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}

	var warehouse *entity.Warehouse
	if item.WarehouseID != uuid.Nil {
		warehouse, _ = s.warehouseRepo.GetWarehouseByID(ctx, item.WarehouseID)
	}

	return s.mapInventoryItemToResponse(item, warehouse), nil
}

func (s *inventoryService) DeleteInventoryItem(ctx context.Context, itemID uuid.UUID) error {
	// Check if item has active reservations
	reservations, err := s.reservationRepo.GetStockReservationsByItem(ctx, itemID, true)
	if err != nil {
		return fmt.Errorf("failed to check reservations: %w", err)
	}

	if len(reservations) > 0 {
		return fmt.Errorf("cannot delete inventory item with active reservations")
	}

	return s.inventoryRepo.DeleteInventoryItem(ctx, itemID)
}

func (s *inventoryService) SearchInventoryItems(ctx context.Context, request *SearchInventoryRequest) (*PaginatedInventoryResponse, error) {
	filters := &repository.InventoryFilters{
		ProductID:   request.ProductID,
		VariantID:   request.VariantID,
		WarehouseID: request.WarehouseID,
		Status:      request.Status,
		SKU:         request.SKU,
		Barcode:     request.Barcode,
		MinQuantity: request.MinQuantity,
		MaxQuantity: request.MaxQuantity,
		LowStock:    request.LowStock,
		OutOfStock:  request.OutOfStock,
		Location:    request.Location,
		SearchTerm:  request.SearchTerm,
		SortBy:      request.SortBy,
		SortOrder:   request.SortOrder,
		Limit:       request.Limit,
		Offset:      request.Offset,
	}

	items, total, err := s.inventoryRepo.SearchInventoryItems(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search inventory items: %w", err)
	}

	responses := make([]*InventoryItemResponse, len(items))
	for i, item := range items {
		var warehouse *entity.Warehouse
		if item.WarehouseID != uuid.Nil {
			warehouse, _ = s.warehouseRepo.GetWarehouseByID(ctx, item.WarehouseID)
		}
		responses[i] = s.mapInventoryItemToResponse(item, warehouse)
	}

	return &PaginatedInventoryResponse{
		Items:   responses,
		Total:   total,
		Page:    (request.Offset / request.Limit) + 1,
		PerPage: request.Limit,
		Pages:   (int(total) + request.Limit - 1) / request.Limit,
	}, nil
}

// Stock operations
func (s *inventoryService) CheckStockAvailability(ctx context.Context, request *StockAvailabilityRequest) (*StockAvailabilityResponse, error) {
	availability, err := s.inventoryRepo.CheckAvailability(ctx, request.ProductID, request.VariantID, request.Quantity, request.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	return &StockAvailabilityResponse{
		ProductID:      request.ProductID,
		VariantID:      request.VariantID,
		RequestedQuantity: request.Quantity,
		IsAvailable:    availability.IsAvailable,
		TotalAvailable: availability.TotalAvailable,
		WarehouseStock: availability.WarehouseStock,
		AllocationSuggestion: availability.AllocationSuggestion,
	}, nil
}

func (s *inventoryService) ReserveStock(ctx context.Context, request *ReserveStockRequest) (*StockReservationResponse, error) {
	reservationReq := &repository.StockReservationRequest{
		ProductID:          request.ProductID,
		VariantID:          request.VariantID,
		OrderID:            request.OrderID,
		Quantity:           request.Quantity,
		PreferredWarehouse: request.PreferredWarehouse,
		ExpirationHours:    request.ExpirationHours,
		ReservedPrice:      request.ReservedPrice,
	}

	reservation, err := s.inventoryRepo.ReserveStock(ctx, reservationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	return s.mapReservationToResponse(reservation), nil
}

func (s *inventoryService) ReleaseStockReservation(ctx context.Context, reservationID uuid.UUID) error {
	return s.inventoryRepo.ReleaseStock(ctx, reservationID)
}

func (s *inventoryService) FulfillStockReservation(ctx context.Context, reservationID uuid.UUID) error {
	return s.inventoryRepo.FulfillStock(ctx, reservationID)
}

func (s *inventoryService) AdjustStock(ctx context.Context, request *AdjustStockRequest) (*StockMovementResponse, error) {
	adjustmentReq := &repository.StockAdjustmentRequest{
		InventoryItemID: request.InventoryItemID,
		Quantity:        request.Quantity,
		Type:            request.Type,
		Reason:          request.Reason,
		Notes:           request.Notes,
		UnitCost:        request.UnitCost,
		UserID:          request.UserID,
		UserName:        request.UserName,
	}

	if err := s.inventoryRepo.AdjustStock(ctx, adjustmentReq); err != nil {
		return nil, fmt.Errorf("failed to adjust stock: %w", err)
	}

	// Get the latest movement
	movements, _, err := s.movementRepo.GetStockMovementsByItem(ctx, request.InventoryItemID, 1, 0)
	if err != nil || len(movements) == 0 {
		return nil, fmt.Errorf("failed to retrieve stock movement")
	}

	return s.mapMovementToResponse(movements[0]), nil
}

func (s *inventoryService) TransferStock(ctx context.Context, request *TransferStockRequest) ([]*StockMovementResponse, error) {
	// Get source inventory item
	sourceItem, err := s.inventoryRepo.GetInventoryItemByProductAndWarehouse(ctx, request.ProductID, request.FromWarehouseID, request.VariantID)
	if err != nil {
		return nil, fmt.Errorf("source inventory item not found: %w", err)
	}

	// Check availability
	if !sourceItem.IsAvailable(request.Quantity) {
		return nil, entity.ErrInsufficientStock
	}

	// Get or create destination inventory item
	destItem, err := s.inventoryRepo.GetInventoryItemByProductAndWarehouse(ctx, request.ProductID, request.ToWarehouseID, request.VariantID)
	if err != nil {
		// Create destination item if it doesn't exist
		destItem = &entity.InventoryItem{
			ID:                uuid.New(),
			ProductID:         request.ProductID,
			VariantID:         request.VariantID,
			WarehouseID:       request.ToWarehouseID,
			QuantityAvailable: 0,
			QuantityTotal:     0,
			MinStockLevel:     sourceItem.MinStockLevel,
			MaxStockLevel:     sourceItem.MaxStockLevel,
			ReorderPoint:      sourceItem.ReorderPoint,
			SKU:               sourceItem.SKU,
			CostPrice:         sourceItem.CostPrice,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := s.inventoryRepo.CreateInventoryItem(ctx, destItem); err != nil {
			return nil, fmt.Errorf("failed to create destination inventory item: %w", err)
		}
	}

	// Create outbound movement from source
	outboundMovement := &entity.StockMovement{
		ID:               uuid.New(),
		InventoryItemID:  sourceItem.ID,
		Type:             entity.MovementTypeTransfer,
		Quantity:         -request.Quantity,
		PreviousQuantity: sourceItem.QuantityTotal,
		NewQuantity:      sourceItem.QuantityTotal - request.Quantity,
		ReferenceType:    "transfer",
		ReferenceID:      &destItem.ID,
		Reason:           request.Reason,
		Notes:            fmt.Sprintf("Transfer to warehouse %s", request.ToWarehouseID.String()),
		UnitCost:         sourceItem.CostPrice,
		TotalCost:        float64(request.Quantity) * sourceItem.CostPrice,
		UserID:           request.UserID,
		UserName:         request.UserName,
		CreatedAt:        time.Now(),
		MovementDate:     time.Now(),
	}

	// Create inbound movement to destination
	inboundMovement := &entity.StockMovement{
		ID:               uuid.New(),
		InventoryItemID:  destItem.ID,
		Type:             entity.MovementTypeTransfer,
		Quantity:         request.Quantity,
		PreviousQuantity: destItem.QuantityTotal,
		NewQuantity:      destItem.QuantityTotal + request.Quantity,
		ReferenceType:    "transfer",
		ReferenceID:      &sourceItem.ID,
		Reason:           request.Reason,
		Notes:            fmt.Sprintf("Transfer from warehouse %s", request.FromWarehouseID.String()),
		UnitCost:         sourceItem.CostPrice,
		TotalCost:        float64(request.Quantity) * sourceItem.CostPrice,
		UserID:           request.UserID,
		UserName:         request.UserName,
		CreatedAt:        time.Now(),
		MovementDate:     time.Now(),
	}

	// Update inventory items
	sourceItem.QuantityAvailable -= request.Quantity
	sourceItem.QuantityTotal -= request.Quantity
	sourceItem.UpdatedAt = time.Now()
	sourceItem.UpdateStatus()

	destItem.QuantityAvailable += request.Quantity
	destItem.QuantityTotal += request.Quantity
	destItem.UpdatedAt = time.Now()
	destItem.UpdateStatus()

	// Save everything
	if err := s.inventoryRepo.UpdateInventoryItem(ctx, sourceItem); err != nil {
		return nil, fmt.Errorf("failed to update source inventory: %w", err)
	}

	if err := s.inventoryRepo.UpdateInventoryItem(ctx, destItem); err != nil {
		return nil, fmt.Errorf("failed to update destination inventory: %w", err)
	}

	if err := s.movementRepo.CreateStockMovement(ctx, outboundMovement); err != nil {
		return nil, fmt.Errorf("failed to create outbound movement: %w", err)
	}

	if err := s.movementRepo.CreateStockMovement(ctx, inboundMovement); err != nil {
		return nil, fmt.Errorf("failed to create inbound movement: %w", err)
	}

	return []*StockMovementResponse{
		s.mapMovementToResponse(outboundMovement),
		s.mapMovementToResponse(inboundMovement),
	}, nil
}

// Helper methods
func (s *inventoryService) mapInventoryItemToResponse(item *entity.InventoryItem, warehouse *entity.Warehouse) *InventoryItemResponse {
	response := &InventoryItemResponse{
		ID:                item.ID,
		ProductID:         item.ProductID,
		VariantID:         item.VariantID,
		WarehouseID:       item.WarehouseID,
		QuantityAvailable: item.QuantityAvailable,
		QuantityReserved:  item.QuantityReserved,
		QuantityTotal:     item.QuantityTotal,
		MinStockLevel:     item.MinStockLevel,
		MaxStockLevel:     item.MaxStockLevel,
		ReorderPoint:      item.ReorderPoint,
		Status:            item.Status,
		Location:          item.Location,
		SKU:               item.SKU,
		Barcode:           item.Barcode,
		CostPrice:         item.CostPrice,
		LastCostPrice:     item.LastCostPrice,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		LastRestockedAt:   item.LastRestockedAt,
		LastSoldAt:        item.LastSoldAt,
		IsLowStock:        item.IsLowStock(),
		IsOutOfStock:      item.IsOutOfStock(),
		TurnoverRate:      item.GetTurnoverRate(30), // 30-day turnover
	}

	if warehouse != nil {
		response.WarehouseName = warehouse.Name
		response.WarehouseCode = warehouse.Code
	}

	return response
}

func (s *inventoryService) mapReservationToResponse(reservation *entity.StockReservation) *StockReservationResponse {
	return &StockReservationResponse{
		ID:              reservation.ID,
		InventoryItemID: reservation.InventoryItemID,
		OrderID:         reservation.OrderID,
		Quantity:        reservation.Quantity,
		ReservedPrice:   reservation.ReservedPrice,
		IsActive:        reservation.IsActive,
		ReservationCode: reservation.ReservationCode,
		CreatedAt:       reservation.CreatedAt,
		UpdatedAt:       reservation.UpdatedAt,
		ExpiresAt:       reservation.ExpiresAt,
		ReleasedAt:      reservation.ReleasedAt,
		IsExpired:       reservation.IsExpired(),
	}
}

func (s *inventoryService) mapMovementToResponse(movement *entity.StockMovement) *StockMovementResponse {
	return &StockMovementResponse{
		ID:               movement.ID,
		InventoryItemID:  movement.InventoryItemID,
		Type:             movement.Type,
		Quantity:         movement.Quantity,
		PreviousQuantity: movement.PreviousQuantity,
		NewQuantity:      movement.NewQuantity,
		ReferenceType:    movement.ReferenceType,
		ReferenceID:      movement.ReferenceID,
		Reason:           movement.Reason,
		Notes:            movement.Notes,
		UnitCost:         movement.UnitCost,
		TotalCost:        movement.TotalCost,
		UserID:           movement.UserID,
		UserName:         movement.UserName,
		CreatedAt:        movement.CreatedAt,
		MovementDate:     movement.MovementDate,
	}
}

// Additional methods for warehouse management, alerts, and analytics would go here
// This is a comprehensive foundation for the inventory service

func (s *inventoryService) GenerateStockAlerts(ctx context.Context) (*AlertGenerationResponse, error) {
	// Get all low stock and out of stock items
	lowStockItems, _, err := s.inventoryRepo.GetLowStockItems(ctx, nil, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}

	outOfStockItems, _, err := s.inventoryRepo.GetOutOfStockItems(ctx, nil, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get out of stock items: %w", err)
	}

	alertsCreated := 0

	// Create alerts for low stock items
	for _, item := range lowStockItems {
		alert := &entity.StockAlert{
			ID:              uuid.New(),
			InventoryItemID: item.ID,
			Type:            "low_stock",
			Message:         fmt.Sprintf("Low stock alert: %s (SKU: %s) has only %d units remaining", item.SKU, item.SKU, item.QuantityTotal),
			Severity:        "medium",
			CreatedAt:       time.Now(),
		}

		if err := s.alertRepo.CreateStockAlert(ctx, alert); err == nil {
			alertsCreated++
		}
	}

	// Create alerts for out of stock items
	for _, item := range outOfStockItems {
		alert := &entity.StockAlert{
			ID:              uuid.New(),
			InventoryItemID: item.ID,
			Type:            "out_of_stock",
			Message:         fmt.Sprintf("Out of stock alert: %s (SKU: %s) is completely out of stock", item.SKU, item.SKU),
			Severity:        "high",
			CreatedAt:       time.Now(),
		}

		if err := s.alertRepo.CreateStockAlert(ctx, alert); err == nil {
			alertsCreated++
		}
	}

	return &AlertGenerationResponse{
		AlertsCreated:     alertsCreated,
		LowStockItems:     len(lowStockItems),
		OutOfStockItems:   len(outOfStockItems),
		GeneratedAt:       time.Now(),
	}, nil
}