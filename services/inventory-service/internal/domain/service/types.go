package service

import (
	"time"

	"github.com/google/uuid"
	"solemate/services/inventory-service/internal/domain/entity"
	"solemate/services/inventory-service/internal/domain/repository"
)

// Request types
type CreateInventoryItemRequest struct {
	ProductID       uuid.UUID  `json:"product_id" validate:"required"`
	VariantID       *uuid.UUID `json:"variant_id"`
	WarehouseID     uuid.UUID  `json:"warehouse_id" validate:"required"`
	InitialQuantity int        `json:"initial_quantity" validate:"min=0"`
	MinStockLevel   int        `json:"min_stock_level" validate:"min=0"`
	MaxStockLevel   int        `json:"max_stock_level" validate:"min=1"`
	ReorderPoint    int        `json:"reorder_point" validate:"min=0"`
	Location        string     `json:"location"`
	SKU             string     `json:"sku" validate:"required"`
	Barcode         string     `json:"barcode"`
	CostPrice       float64    `json:"cost_price" validate:"min=0"`
	UserName        string     `json:"user_name"`
}

type UpdateInventoryItemRequest struct {
	MinStockLevel *int     `json:"min_stock_level,omitempty"`
	MaxStockLevel *int     `json:"max_stock_level,omitempty"`
	ReorderPoint  *int     `json:"reorder_point,omitempty"`
	Location      *string  `json:"location,omitempty"`
	CostPrice     *float64 `json:"cost_price,omitempty"`
}

type SearchInventoryRequest struct {
	ProductID   *uuid.UUID           `json:"product_id,omitempty"`
	VariantID   *uuid.UUID           `json:"variant_id,omitempty"`
	WarehouseID *uuid.UUID           `json:"warehouse_id,omitempty"`
	Status      *entity.StockStatus  `json:"status,omitempty"`
	SKU         string               `json:"sku,omitempty"`
	Barcode     string               `json:"barcode,omitempty"`
	MinQuantity *int                 `json:"min_quantity,omitempty"`
	MaxQuantity *int                 `json:"max_quantity,omitempty"`
	LowStock    bool                 `json:"low_stock,omitempty"`
	OutOfStock  bool                 `json:"out_of_stock,omitempty"`
	Location    string               `json:"location,omitempty"`
	SearchTerm  string               `json:"search_term,omitempty"`
	SortBy      string               `json:"sort_by,omitempty"`
	SortOrder   string               `json:"sort_order,omitempty"`
	Limit       int                  `json:"limit" validate:"min=1,max=100"`
	Offset      int                  `json:"offset" validate:"min=0"`
}

type StockAvailabilityRequest struct {
	ProductID   uuid.UUID  `json:"product_id" validate:"required"`
	VariantID   *uuid.UUID `json:"variant_id"`
	Quantity    int        `json:"quantity" validate:"required,gt=0"`
	WarehouseID *uuid.UUID `json:"warehouse_id"`
}

type ReserveStockRequest struct {
	ProductID          uuid.UUID  `json:"product_id" validate:"required"`
	VariantID          *uuid.UUID `json:"variant_id"`
	OrderID            uuid.UUID  `json:"order_id" validate:"required"`
	Quantity           int        `json:"quantity" validate:"required,gt=0"`
	PreferredWarehouse *uuid.UUID `json:"preferred_warehouse"`
	ExpirationHours    int        `json:"expiration_hours" validate:"min=1,max=168"`
	ReservedPrice      float64    `json:"reserved_price" validate:"min=0"`
}

type AdjustStockRequest struct {
	InventoryItemID uuid.UUID            `json:"inventory_item_id" validate:"required"`
	Quantity        int                  `json:"quantity" validate:"required"`
	Type            entity.MovementType  `json:"type" validate:"required"`
	Reason          string               `json:"reason" validate:"required"`
	Notes           string               `json:"notes"`
	UnitCost        float64              `json:"unit_cost" validate:"min=0"`
	UserID          *uuid.UUID           `json:"user_id"`
	UserName        string               `json:"user_name" validate:"required"`
}

type TransferStockRequest struct {
	ProductID       uuid.UUID  `json:"product_id" validate:"required"`
	VariantID       *uuid.UUID `json:"variant_id"`
	FromWarehouseID uuid.UUID  `json:"from_warehouse_id" validate:"required"`
	ToWarehouseID   uuid.UUID  `json:"to_warehouse_id" validate:"required"`
	Quantity        int        `json:"quantity" validate:"required,gt=0"`
	Reason          string     `json:"reason" validate:"required"`
	UserID          *uuid.UUID `json:"user_id"`
	UserName        string     `json:"user_name" validate:"required"`
}

type BulkStockUpdateRequest struct {
	Updates  []BulkStockUpdateItem `json:"updates" validate:"required,dive"`
	Reason   string                `json:"reason" validate:"required"`
	UserID   *uuid.UUID            `json:"user_id"`
	UserName string                `json:"user_name" validate:"required"`
}

type BulkStockUpdateItem struct {
	InventoryItemID uuid.UUID `json:"inventory_item_id" validate:"required"`
	Quantity        int       `json:"quantity" validate:"required"`
	UnitCost        float64   `json:"unit_cost" validate:"min=0"`
}

type BulkReserveStockRequest struct {
	OrderID      uuid.UUID                `json:"order_id" validate:"required"`
	Reservations []BulkReservationItem    `json:"reservations" validate:"required,dive"`
	ExpirationHours int                   `json:"expiration_hours" validate:"min=1,max=168"`
}

type BulkReservationItem struct {
	ProductID   uuid.UUID  `json:"product_id" validate:"required"`
	VariantID   *uuid.UUID `json:"variant_id"`
	Quantity    int        `json:"quantity" validate:"required,gt=0"`
	ReservedPrice float64  `json:"reserved_price" validate:"min=0"`
}

type CreateWarehouseRequest struct {
	Name        string         `json:"name" validate:"required"`
	Code        string         `json:"code" validate:"required"`
	Description string         `json:"description"`
	Address     entity.Address `json:"address"`
	IsActive    bool           `json:"is_active"`
	IsDefault   bool           `json:"is_default"`
	Priority    int            `json:"priority"`
	Capacity    int            `json:"capacity" validate:"min=1"`
	ManagerName string         `json:"manager_name"`
	Phone       string         `json:"phone"`
	Email       string         `json:"email"`
}

type UpdateWarehouseRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Address     *entity.Address `json:"address,omitempty"`
	IsActive    *bool           `json:"is_active,omitempty"`
	IsDefault   *bool           `json:"is_default,omitempty"`
	Priority    *int            `json:"priority,omitempty"`
	Capacity    *int            `json:"capacity,omitempty"`
	ManagerName *string         `json:"manager_name,omitempty"`
	Phone       *string         `json:"phone,omitempty"`
	Email       *string         `json:"email,omitempty"`
}

type StockMovementHistoryRequest struct {
	InventoryItemID *uuid.UUID            `json:"inventory_item_id,omitempty"`
	WarehouseID     *uuid.UUID            `json:"warehouse_id,omitempty"`
	Type            *entity.MovementType  `json:"type,omitempty"`
	ReferenceType   string                `json:"reference_type,omitempty"`
	ReferenceID     *uuid.UUID            `json:"reference_id,omitempty"`
	StartDate       *time.Time            `json:"start_date,omitempty"`
	EndDate         *time.Time            `json:"end_date,omitempty"`
	UserID          *uuid.UUID            `json:"user_id,omitempty"`
	Limit           int                   `json:"limit" validate:"min=1,max=100"`
	Offset          int                   `json:"offset" validate:"min=0"`
}

type MovementSummaryRequest struct {
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	StartDate   time.Time  `json:"start_date" validate:"required"`
	EndDate     time.Time  `json:"end_date" validate:"required"`
}

type StockAlertsRequest struct {
	WarehouseID  *uuid.UUID `json:"warehouse_id,omitempty"`
	AlertType    string     `json:"alert_type,omitempty"`
	Severity     string     `json:"severity,omitempty"`
	UnreadOnly   bool       `json:"unread_only,omitempty"`
	UnresolvedOnly bool     `json:"unresolved_only,omitempty"`
	Limit        int        `json:"limit" validate:"min=1,max=100"`
	Offset       int        `json:"offset" validate:"min=0"`
}

type InventoryAnalyticsRequest struct {
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	StartDate   time.Time  `json:"start_date" validate:"required"`
	EndDate     time.Time  `json:"end_date" validate:"required"`
}

type TurnoverReportRequest struct {
	WarehouseID *uuid.UUID `json:"warehouse_id,omitempty"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Days        int        `json:"days" validate:"min=1,max=365"`
	Limit       int        `json:"limit" validate:"min=1,max=100"`
}

// Response types
type InventoryItemResponse struct {
	ID                uuid.UUID        `json:"id"`
	ProductID         uuid.UUID        `json:"product_id"`
	VariantID         *uuid.UUID       `json:"variant_id"`
	WarehouseID       uuid.UUID        `json:"warehouse_id"`
	WarehouseName     string           `json:"warehouse_name"`
	WarehouseCode     string           `json:"warehouse_code"`
	QuantityAvailable int              `json:"quantity_available"`
	QuantityReserved  int              `json:"quantity_reserved"`
	QuantityTotal     int              `json:"quantity_total"`
	MinStockLevel     int              `json:"min_stock_level"`
	MaxStockLevel     int              `json:"max_stock_level"`
	ReorderPoint      int              `json:"reorder_point"`
	Status            entity.StockStatus `json:"status"`
	Location          string           `json:"location"`
	SKU               string           `json:"sku"`
	Barcode           string           `json:"barcode"`
	CostPrice         float64          `json:"cost_price"`
	LastCostPrice     float64          `json:"last_cost_price"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	LastRestockedAt   *time.Time       `json:"last_restocked_at"`
	LastSoldAt        *time.Time       `json:"last_sold_at"`
	IsLowStock        bool             `json:"is_low_stock"`
	IsOutOfStock      bool             `json:"is_out_of_stock"`
	TurnoverRate      float64          `json:"turnover_rate"`
}

type StockAvailabilityResponse struct {
	ProductID             uuid.UUID                          `json:"product_id"`
	VariantID             *uuid.UUID                         `json:"variant_id"`
	RequestedQuantity     int                                `json:"requested_quantity"`
	IsAvailable           bool                               `json:"is_available"`
	TotalAvailable        int                                `json:"total_available"`
	WarehouseStock        []repository.WarehouseStockInfo    `json:"warehouse_stock"`
	AllocationSuggestion  []repository.StockAllocation       `json:"allocation_suggestion"`
}

type StockReservationResponse struct {
	ID              uuid.UUID  `json:"id"`
	InventoryItemID uuid.UUID  `json:"inventory_item_id"`
	OrderID         uuid.UUID  `json:"order_id"`
	Quantity        int        `json:"quantity"`
	ReservedPrice   float64    `json:"reserved_price"`
	IsActive        bool       `json:"is_active"`
	ReservationCode string     `json:"reservation_code"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	ReleasedAt      *time.Time `json:"released_at"`
	IsExpired       bool       `json:"is_expired"`
}

type StockMovementResponse struct {
	ID               uuid.UUID            `json:"id"`
	InventoryItemID  uuid.UUID            `json:"inventory_item_id"`
	Type             entity.MovementType  `json:"type"`
	Quantity         int                  `json:"quantity"`
	PreviousQuantity int                  `json:"previous_quantity"`
	NewQuantity      int                  `json:"new_quantity"`
	ReferenceType    string               `json:"reference_type"`
	ReferenceID      *uuid.UUID           `json:"reference_id"`
	Reason           string               `json:"reason"`
	Notes            string               `json:"notes"`
	UnitCost         float64              `json:"unit_cost"`
	TotalCost        float64              `json:"total_cost"`
	UserID           *uuid.UUID           `json:"user_id"`
	UserName         string               `json:"user_name"`
	CreatedAt        time.Time            `json:"created_at"`
	MovementDate     time.Time            `json:"movement_date"`
}

type WarehouseResponse struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Description string         `json:"description"`
	Address     entity.Address `json:"address"`
	IsActive    bool           `json:"is_active"`
	IsDefault   bool           `json:"is_default"`
	Priority    int            `json:"priority"`
	Capacity    int            `json:"capacity"`
	ManagerName string         `json:"manager_name"`
	Phone       string         `json:"phone"`
	Email       string         `json:"email"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	TotalItems  int            `json:"total_items"`
	CapacityUsed int           `json:"capacity_used"`
	CapacityUtilization float64 `json:"capacity_utilization_percent"`
}

type WarehouseInventorySummaryResponse struct {
	WarehouseID       uuid.UUID                               `json:"warehouse_id"`
	WarehouseName     string                                  `json:"warehouse_name"`
	TotalItems        int                                     `json:"total_items"`
	TotalValue        float64                                 `json:"total_value"`
	LowStockItems     int                                     `json:"low_stock_items"`
	OutOfStockItems   int                                     `json:"out_of_stock_items"`
	StatusBreakdown   map[entity.StockStatus]int              `json:"status_breakdown"`
	TopProducts       []repository.ProductInventorySummary    `json:"top_products"`
	RecentMovements   []*StockMovementResponse                `json:"recent_movements"`
}

type BulkOperationResponse struct {
	SuccessfulUpdates int                      `json:"successful_updates"`
	FailedUpdates     int                      `json:"failed_updates"`
	TotalRequested    int                      `json:"total_requested"`
	Errors            []BulkOperationError     `json:"errors,omitempty"`
	ProcessedAt       time.Time                `json:"processed_at"`
}

type BulkOperationError struct {
	ItemIndex int    `json:"item_index"`
	ItemID    string `json:"item_id"`
	Error     string `json:"error"`
}

type BulkReservationResponse struct {
	SuccessfulReservations int                      `json:"successful_reservations"`
	FailedReservations     int                      `json:"failed_reservations"`
	TotalRequested         int                      `json:"total_requested"`
	Reservations           []*StockReservationResponse `json:"reservations"`
	Errors                 []BulkOperationError     `json:"errors,omitempty"`
	ProcessedAt            time.Time                `json:"processed_at"`
}

type AlertGenerationResponse struct {
	AlertsCreated     int       `json:"alerts_created"`
	LowStockItems     int       `json:"low_stock_items"`
	OutOfStockItems   int       `json:"out_of_stock_items"`
	GeneratedAt       time.Time `json:"generated_at"`
}

type MovementSummaryResponse struct {
	StartDate            time.Time                                          `json:"start_date"`
	EndDate              time.Time                                          `json:"end_date"`
	TotalMovements       int                                                `json:"total_movements"`
	InboundQuantity      int                                                `json:"inbound_quantity"`
	OutboundQuantity     int                                                `json:"outbound_quantity"`
	NetMovement          int                                                `json:"net_movement"`
	MovementsByType      map[entity.MovementType]int                        `json:"movements_by_type"`
	DailyMovements       []repository.DailyMovementStats                    `json:"daily_movements"`
	WarehouseMovements   map[uuid.UUID]repository.WarehouseMovementStats    `json:"warehouse_movements"`
}

type InventoryAnalyticsResponse struct {
	TotalItems          int                                        `json:"total_items"`
	TotalValue          float64                                    `json:"total_value"`
	AverageValue        float64                                    `json:"average_value"`
	LowStockItems       int                                        `json:"low_stock_items"`
	OutOfStockItems     int                                        `json:"out_of_stock_items"`
	StatusBreakdown     map[entity.StockStatus]int                 `json:"status_breakdown"`
	WarehouseBreakdown  map[uuid.UUID]WarehouseAnalytics           `json:"warehouse_breakdown"`
	TopProducts         []repository.ProductInventorySummary       `json:"top_products"`
	SlowMovingProducts  []repository.ProductInventorySummary       `json:"slow_moving_products"`
	MovementTrends      []repository.DailyMovementStats            `json:"movement_trends"`
}

type WarehouseAnalytics struct {
	WarehouseID         uuid.UUID `json:"warehouse_id"`
	WarehouseName       string    `json:"warehouse_name"`
	TotalItems          int       `json:"total_items"`
	TotalValue          float64   `json:"total_value"`
	CapacityUtilization float64   `json:"capacity_utilization_percent"`
	LowStockItems       int       `json:"low_stock_items"`
	OutOfStockItems     int       `json:"out_of_stock_items"`
}

type TurnoverReportResponse struct {
	ReportPeriodDays    int                                   `json:"report_period_days"`
	TotalProducts       int                                   `json:"total_products"`
	AverageTurnover     float64                               `json:"average_turnover"`
	HighTurnoverProducts []ProductTurnoverInfo                `json:"high_turnover_products"`
	LowTurnoverProducts  []ProductTurnoverInfo                `json:"low_turnover_products"`
	GeneratedAt         time.Time                             `json:"generated_at"`
}

type ProductTurnoverInfo struct {
	ProductID       uuid.UUID  `json:"product_id"`
	VariantID       *uuid.UUID `json:"variant_id"`
	SKU             string     `json:"sku"`
	QuantityTotal   int        `json:"quantity_total"`
	TurnoverRate    float64    `json:"turnover_rate"`
	DaysSinceLastSale int      `json:"days_since_last_sale"`
	RecommendedAction string   `json:"recommended_action"`
}

type StockValuationResponse struct {
	WarehouseID         *uuid.UUID                 `json:"warehouse_id"`
	WarehouseName       string                     `json:"warehouse_name,omitempty"`
	TotalValue          float64                    `json:"total_value"`
	TotalQuantity       int                        `json:"total_quantity"`
	AverageUnitValue    float64                    `json:"average_unit_value"`
	ValuationByStatus   map[entity.StockStatus]float64 `json:"valuation_by_status"`
	TopValueProducts    []ProductValuationInfo     `json:"top_value_products"`
	GeneratedAt         time.Time                  `json:"generated_at"`
}

type ProductValuationInfo struct {
	ProductID       uuid.UUID  `json:"product_id"`
	VariantID       *uuid.UUID `json:"variant_id"`
	SKU             string     `json:"sku"`
	Quantity        int        `json:"quantity"`
	UnitValue       float64    `json:"unit_value"`
	TotalValue      float64    `json:"total_value"`
	PercentOfTotal  float64    `json:"percent_of_total"`
}

type OrderStockAllocationResponse struct {
	OrderID             uuid.UUID                      `json:"order_id"`
	TotalItemsRequested int                            `json:"total_items_requested"`
	TotalItemsAllocated int                            `json:"total_items_allocated"`
	FullyAllocated      bool                           `json:"fully_allocated"`
	Allocations         []OrderItemAllocation          `json:"allocations"`
	Reservations        []*StockReservationResponse    `json:"reservations"`
	ProcessedAt         time.Time                      `json:"processed_at"`
}

type OrderItemAllocation struct {
	ProductID         uuid.UUID                    `json:"product_id"`
	VariantID         *uuid.UUID                   `json:"variant_id"`
	RequestedQuantity int                          `json:"requested_quantity"`
	AllocatedQuantity int                          `json:"allocated_quantity"`
	IsFullyAllocated  bool                         `json:"is_fully_allocated"`
	Allocations       []repository.StockAllocation `json:"allocations"`
}

// Pagination types
type PaginatedInventoryResponse struct {
	Items   []*InventoryItemResponse `json:"items"`
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"per_page"`
	Pages   int                      `json:"pages"`
}

type PaginatedStockMovementResponse struct {
	Movements []*StockMovementResponse `json:"movements"`
	Total     int64                    `json:"total"`
	Page      int                      `json:"page"`
	PerPage   int                      `json:"per_page"`
	Pages     int                      `json:"pages"`
}

type PaginatedStockAlertResponse struct {
	Alerts  []*StockAlertResponse `json:"alerts"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	PerPage int                   `json:"per_page"`
	Pages   int                   `json:"pages"`
}

type StockAlertResponse struct {
	ID              uuid.UUID  `json:"id"`
	InventoryItemID uuid.UUID  `json:"inventory_item_id"`
	Type            string     `json:"type"`
	Message         string     `json:"message"`
	Severity        string     `json:"severity"`
	IsRead          bool       `json:"is_read"`
	IsResolved      bool       `json:"is_resolved"`
	CreatedAt       time.Time  `json:"created_at"`
	ReadAt          *time.Time `json:"read_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`
}