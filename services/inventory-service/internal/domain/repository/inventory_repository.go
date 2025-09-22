package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"solemate/services/inventory-service/internal/domain/entity"
)

type InventoryRepository interface {
	// Inventory item CRUD operations
	CreateInventoryItem(ctx context.Context, item *entity.InventoryItem) error
	GetInventoryItemByID(ctx context.Context, itemID uuid.UUID) (*entity.InventoryItem, error)
	GetInventoryItemByProductAndWarehouse(ctx context.Context, productID, warehouseID uuid.UUID, variantID *uuid.UUID) (*entity.InventoryItem, error)
	GetInventoryItemBySKU(ctx context.Context, sku string) (*entity.InventoryItem, error)
	GetInventoryItemByBarcode(ctx context.Context, barcode string) (*entity.InventoryItem, error)
	UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) error
	DeleteInventoryItem(ctx context.Context, itemID uuid.UUID) error

	// Inventory querying and filtering
	GetInventoryItemsByProduct(ctx context.Context, productID uuid.UUID) ([]*entity.InventoryItem, error)
	GetInventoryItemsByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]*entity.InventoryItem, int64, error)
	GetInventoryItemsByStatus(ctx context.Context, status entity.StockStatus, limit, offset int) ([]*entity.InventoryItem, int64, error)
	GetLowStockItems(ctx context.Context, warehouseID *uuid.UUID, limit, offset int) ([]*entity.InventoryItem, int64, error)
	GetOutOfStockItems(ctx context.Context, warehouseID *uuid.UUID, limit, offset int) ([]*entity.InventoryItem, int64, error)
	SearchInventoryItems(ctx context.Context, filters *InventoryFilters) ([]*entity.InventoryItem, int64, error)

	// Stock operations
	CheckAvailability(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, quantity int, warehouseID *uuid.UUID) (*StockAvailability, error)
	ReserveStock(ctx context.Context, request *StockReservationRequest) (*entity.StockReservation, error)
	ReleaseStock(ctx context.Context, reservationID uuid.UUID) error
	FulfillStock(ctx context.Context, reservationID uuid.UUID) error
	AdjustStock(ctx context.Context, request *StockAdjustmentRequest) error

	// Bulk operations
	BulkUpdateStock(ctx context.Context, updates []*BulkStockUpdate) error
	BulkReserveStock(ctx context.Context, reservations []*StockReservationRequest) ([]*entity.StockReservation, error)
}

type WarehouseRepository interface {
	// Warehouse CRUD operations
	CreateWarehouse(ctx context.Context, warehouse *entity.Warehouse) error
	GetWarehouseByID(ctx context.Context, warehouseID uuid.UUID) (*entity.Warehouse, error)
	GetWarehouseByCode(ctx context.Context, code string) (*entity.Warehouse, error)
	GetDefaultWarehouse(ctx context.Context) (*entity.Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse *entity.Warehouse) error
	DeleteWarehouse(ctx context.Context, warehouseID uuid.UUID) error

	// Warehouse querying
	GetAllWarehouses(ctx context.Context, activeOnly bool) ([]*entity.Warehouse, error)
	GetWarehousesByPriority(ctx context.Context) ([]*entity.Warehouse, error)
	GetNearestWarehouses(ctx context.Context, address entity.Address, limit int) ([]*entity.Warehouse, error)

	// Warehouse analytics
	GetWarehouseCapacityReport(ctx context.Context, warehouseID uuid.UUID) (*WarehouseCapacityReport, error)
	GetWarehouseInventorySummary(ctx context.Context, warehouseID uuid.UUID) (*WarehouseInventorySummary, error)
}

type StockMovementRepository interface {
	// Stock movement operations
	CreateStockMovement(ctx context.Context, movement *entity.StockMovement) error
	GetStockMovementByID(ctx context.Context, movementID uuid.UUID) (*entity.StockMovement, error)
	GetStockMovementsByItem(ctx context.Context, itemID uuid.UUID, limit, offset int) ([]*entity.StockMovement, int64, error)
	GetStockMovementsByType(ctx context.Context, movementType entity.MovementType, startDate, endDate time.Time, limit, offset int) ([]*entity.StockMovement, int64, error)
	GetStockMovementsByReference(ctx context.Context, referenceType string, referenceID uuid.UUID) ([]*entity.StockMovement, error)
	GetStockMovementsByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*entity.StockMovement, int64, error)

	// Movement analytics
	GetMovementSummary(ctx context.Context, startDate, endDate time.Time, warehouseID *uuid.UUID) (*MovementSummary, error)
	GetTopMovedProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]*ProductMovementStats, error)
}

type StockReservationRepository interface {
	// Reservation CRUD operations
	CreateStockReservation(ctx context.Context, reservation *entity.StockReservation) error
	GetStockReservationByID(ctx context.Context, reservationID uuid.UUID) (*entity.StockReservation, error)
	GetStockReservationByCode(ctx context.Context, code string) (*entity.StockReservation, error)
	GetStockReservationsByOrder(ctx context.Context, orderID uuid.UUID) ([]*entity.StockReservation, error)
	GetStockReservationsByItem(ctx context.Context, itemID uuid.UUID, activeOnly bool) ([]*entity.StockReservation, error)
	UpdateStockReservation(ctx context.Context, reservation *entity.StockReservation) error
	DeleteStockReservation(ctx context.Context, reservationID uuid.UUID) error

	// Reservation management
	GetExpiredReservations(ctx context.Context, limit int) ([]*entity.StockReservation, error)
	GetActiveReservations(ctx context.Context, warehouseID *uuid.UUID, limit, offset int) ([]*entity.StockReservation, int64, error)
	ReleaseExpiredReservations(ctx context.Context) (int, error)
}

type StockAlertRepository interface {
	// Alert CRUD operations
	CreateStockAlert(ctx context.Context, alert *entity.StockAlert) error
	GetStockAlertByID(ctx context.Context, alertID uuid.UUID) (*entity.StockAlert, error)
	GetStockAlertsByItem(ctx context.Context, itemID uuid.UUID, unreadOnly bool) ([]*entity.StockAlert, error)
	UpdateStockAlert(ctx context.Context, alert *entity.StockAlert) error
	DeleteStockAlert(ctx context.Context, alertID uuid.UUID) error

	// Alert management
	GetUnreadAlerts(ctx context.Context, severity string, limit, offset int) ([]*entity.StockAlert, int64, error)
	GetAlertsByType(ctx context.Context, alertType string, limit, offset int) ([]*entity.StockAlert, int64, error)
	MarkAlertAsRead(ctx context.Context, alertID uuid.UUID) error
	MarkAlertAsResolved(ctx context.Context, alertID uuid.UUID) error
	BulkMarkAlertsAsRead(ctx context.Context, alertIDs []uuid.UUID) error
}

// Supporting types for repository operations
type InventoryFilters struct {
	ProductID     *uuid.UUID
	VariantID     *uuid.UUID
	WarehouseID   *uuid.UUID
	Status        *entity.StockStatus
	SKU           string
	Barcode       string
	MinQuantity   *int
	MaxQuantity   *int
	LowStock      bool
	OutOfStock    bool
	Location      string
	SearchTerm    string
	SortBy        string
	SortOrder     string
	Limit         int
	Offset        int
}

type StockAvailability struct {
	ProductID         uuid.UUID                    `json:"product_id"`
	VariantID         *uuid.UUID                   `json:"variant_id"`
	TotalAvailable    int                          `json:"total_available"`
	TotalReserved     int                          `json:"total_reserved"`
	WarehouseStock    []WarehouseStockInfo         `json:"warehouse_stock"`
	IsAvailable       bool                         `json:"is_available"`
	AllocationSuggestion []StockAllocation         `json:"allocation_suggestion"`
}

type WarehouseStockInfo struct {
	WarehouseID       uuid.UUID `json:"warehouse_id"`
	WarehouseName     string    `json:"warehouse_name"`
	QuantityAvailable int       `json:"quantity_available"`
	QuantityReserved  int       `json:"quantity_reserved"`
	QuantityTotal     int       `json:"quantity_total"`
	Location          string    `json:"location"`
}

type StockAllocation struct {
	WarehouseID       uuid.UUID `json:"warehouse_id"`
	WarehouseName     string    `json:"warehouse_name"`
	AllocatedQuantity int       `json:"allocated_quantity"`
}

type StockReservationRequest struct {
	ProductID       uuid.UUID  `json:"product_id" validate:"required"`
	VariantID       *uuid.UUID `json:"variant_id"`
	OrderID         uuid.UUID  `json:"order_id" validate:"required"`
	Quantity        int        `json:"quantity" validate:"required,gt=0"`
	PreferredWarehouse *uuid.UUID `json:"preferred_warehouse"`
	ExpirationHours int        `json:"expiration_hours,omitempty"`
	ReservedPrice   float64    `json:"reserved_price"`
}

type StockAdjustmentRequest struct {
	InventoryItemID uuid.UUID            `json:"inventory_item_id" validate:"required"`
	Quantity        int                  `json:"quantity" validate:"required"` // Positive for increase, negative for decrease
	Type            entity.MovementType  `json:"type" validate:"required"`
	Reason          string               `json:"reason" validate:"required"`
	Notes           string               `json:"notes"`
	UnitCost        float64              `json:"unit_cost"`
	UserID          *uuid.UUID           `json:"user_id"`
	UserName        string               `json:"user_name"`
}

type BulkStockUpdate struct {
	InventoryItemID uuid.UUID `json:"inventory_item_id"`
	Quantity        int       `json:"quantity"`
	UnitCost        float64   `json:"unit_cost"`
	Reason          string    `json:"reason"`
}

type WarehouseCapacityReport struct {
	WarehouseID         uuid.UUID `json:"warehouse_id"`
	WarehouseName       string    `json:"warehouse_name"`
	TotalCapacity       int       `json:"total_capacity"`
	UsedCapacity        int       `json:"used_capacity"`
	AvailableCapacity   int       `json:"available_capacity"`
	CapacityUtilization float64   `json:"capacity_utilization_percent"`
	TotalItems          int       `json:"total_items"`
	UniqueProducts      int       `json:"unique_products"`
	TotalValue          float64   `json:"total_value"`
}

type WarehouseInventorySummary struct {
	WarehouseID       uuid.UUID                      `json:"warehouse_id"`
	WarehouseName     string                         `json:"warehouse_name"`
	TotalItems        int                            `json:"total_items"`
	TotalValue        float64                        `json:"total_value"`
	LowStockItems     int                            `json:"low_stock_items"`
	OutOfStockItems   int                            `json:"out_of_stock_items"`
	StatusBreakdown   map[entity.StockStatus]int     `json:"status_breakdown"`
	TopProducts       []ProductInventorySummary      `json:"top_products"`
	RecentMovements   []entity.StockMovement         `json:"recent_movements"`
}

type ProductInventorySummary struct {
	ProductID         uuid.UUID `json:"product_id"`
	VariantID         *uuid.UUID `json:"variant_id"`
	SKU               string    `json:"sku"`
	QuantityTotal     int       `json:"quantity_total"`
	QuantityAvailable int       `json:"quantity_available"`
	QuantityReserved  int       `json:"quantity_reserved"`
	Value             float64   `json:"value"`
	TurnoverRate      float64   `json:"turnover_rate"`
}

type MovementSummary struct {
	StartDate         time.Time                              `json:"start_date"`
	EndDate           time.Time                              `json:"end_date"`
	TotalMovements    int                                    `json:"total_movements"`
	InboundQuantity   int                                    `json:"inbound_quantity"`
	OutboundQuantity  int                                    `json:"outbound_quantity"`
	NetMovement       int                                    `json:"net_movement"`
	MovementsByType   map[entity.MovementType]int            `json:"movements_by_type"`
	DailyMovements    []DailyMovementStats                   `json:"daily_movements"`
	WarehouseMovements map[uuid.UUID]WarehouseMovementStats  `json:"warehouse_movements"`
}

type DailyMovementStats struct {
	Date              time.Time `json:"date"`
	InboundQuantity   int       `json:"inbound_quantity"`
	OutboundQuantity  int       `json:"outbound_quantity"`
	NetMovement       int       `json:"net_movement"`
	TotalMovements    int       `json:"total_movements"`
}

type WarehouseMovementStats struct {
	WarehouseID       uuid.UUID `json:"warehouse_id"`
	WarehouseName     string    `json:"warehouse_name"`
	InboundQuantity   int       `json:"inbound_quantity"`
	OutboundQuantity  int       `json:"outbound_quantity"`
	NetMovement       int       `json:"net_movement"`
	TotalMovements    int       `json:"total_movements"`
}

type ProductMovementStats struct {
	ProductID         uuid.UUID `json:"product_id"`
	VariantID         *uuid.UUID `json:"variant_id"`
	SKU               string    `json:"sku"`
	TotalMovements    int       `json:"total_movements"`
	InboundQuantity   int       `json:"inbound_quantity"`
	OutboundQuantity  int       `json:"outbound_quantity"`
	NetMovement       int       `json:"net_movement"`
	MovementValue     float64   `json:"movement_value"`
}

// Integration with external services
type ProductRepository interface {
	// Integration with product service
	GetProductByID(ctx context.Context, productID uuid.UUID) (*ProductData, error)
	GetProductsBatch(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]*ProductData, error)
	ValidateProductExists(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (bool, error)
}

type OrderRepository interface {
	// Integration with order service
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*OrderData, error)
	UpdateOrderStockStatus(ctx context.Context, orderID uuid.UUID, status string, details map[string]interface{}) error
	NotifyStockAllocation(ctx context.Context, orderID uuid.UUID, allocations []StockAllocation) error
}

// External service data types
type ProductData struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Category    string    `json:"category"`
	Brand       string    `json:"brand"`
	IsActive    bool      `json:"is_active"`
	Variants    []VariantData `json:"variants"`
}

type VariantData struct {
	ID          uuid.UUID `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Size        string    `json:"size"`
	Color       string    `json:"color"`
	IsActive    bool      `json:"is_active"`
}

type OrderData struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Status      string    `json:"status"`
	Items       []OrderItemData `json:"items"`
}

type OrderItemData struct {
	ProductID   uuid.UUID  `json:"product_id"`
	VariantID   *uuid.UUID `json:"variant_id"`
	Quantity    int        `json:"quantity"`
	Price       float64    `json:"price"`
}