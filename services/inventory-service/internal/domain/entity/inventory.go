package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StockStatus string

const (
	StockStatusInStock    StockStatus = "in_stock"
	StockStatusLowStock   StockStatus = "low_stock"
	StockStatusOutOfStock StockStatus = "out_of_stock"
	StockStatusBackorder  StockStatus = "backorder"
)

type MovementType string

const (
	MovementTypeInbound    MovementType = "inbound"     // Stock receiving/restocking
	MovementTypeOutbound   MovementType = "outbound"    // Sales, orders
	MovementTypeAdjustment MovementType = "adjustment"  // Inventory corrections
	MovementTypeReserved   MovementType = "reserved"    // Pending orders
	MovementTypeReleased   MovementType = "released"    // Cancelled reservations
	MovementTypeTransfer   MovementType = "transfer"    // Between warehouses
	MovementTypeDamaged    MovementType = "damaged"     // Damaged goods removal
	MovementTypeReturned   MovementType = "returned"    // Customer returns
)

// InventoryItem represents stock levels for a specific product variant at a specific location
type InventoryItem struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID       uuid.UUID   `json:"product_id" gorm:"type:uuid;not null;index"`
	VariantID       *uuid.UUID  `json:"variant_id" gorm:"type:uuid;index"`
	WarehouseID     uuid.UUID   `json:"warehouse_id" gorm:"type:uuid;not null;index"`

	// Stock levels
	QuantityAvailable int         `json:"quantity_available" gorm:"not null;default:0"`
	QuantityReserved  int         `json:"quantity_reserved" gorm:"not null;default:0"`
	QuantityTotal     int         `json:"quantity_total" gorm:"not null;default:0"`
	MinStockLevel     int         `json:"min_stock_level" gorm:"not null;default:0"`
	MaxStockLevel     int         `json:"max_stock_level" gorm:"not null;default:1000"`
	ReorderPoint      int         `json:"reorder_point" gorm:"not null;default:10"`

	// Status and metadata
	Status            StockStatus `json:"status" gorm:"type:varchar(20);not null;default:'in_stock'"`
	Location          string      `json:"location" gorm:"type:varchar(50)"` // Aisle, Shelf, Bin location
	SKU               string      `json:"sku" gorm:"type:varchar(100);index"`
	Barcode           string      `json:"barcode" gorm:"type:varchar(100);index"`

	// Cost and pricing
	CostPrice         float64     `json:"cost_price" gorm:"type:decimal(10,2);default:0"`
	LastCostPrice     float64     `json:"last_cost_price" gorm:"type:decimal(10,2);default:0"`

	// Timestamps
	CreatedAt         time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	LastRestockedAt   *time.Time  `json:"last_restocked_at"`
	LastSoldAt        *time.Time  `json:"last_sold_at"`

	// Relationships
	Warehouse         *Warehouse          `json:"warehouse,omitempty" gorm:"foreignKey:WarehouseID"`
	Movements         []StockMovement     `json:"movements,omitempty" gorm:"foreignKey:InventoryItemID"`
	Reservations      []StockReservation  `json:"reservations,omitempty" gorm:"foreignKey:InventoryItemID"`
}

// Warehouse represents a storage location/facility
type Warehouse struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null"`
	Code        string    `json:"code" gorm:"type:varchar(20);unique;not null;index"`
	Description string    `json:"description" gorm:"type:text"`

	// Address information
	Address     Address   `json:"address" gorm:"embedded;embeddedPrefix:address_"`

	// Warehouse details
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	Priority    int       `json:"priority" gorm:"default:0"` // For fulfillment prioritization
	Capacity    int       `json:"capacity" gorm:"default:10000"`

	// Contact information
	ManagerName  string   `json:"manager_name" gorm:"type:varchar(100)"`
	Phone        string   `json:"phone" gorm:"type:varchar(20)"`
	Email        string   `json:"email" gorm:"type:varchar(100)"`

	// Timestamps
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	InventoryItems []InventoryItem `json:"inventory_items,omitempty" gorm:"foreignKey:WarehouseID"`
}

// StockMovement tracks all inventory movements (in, out, adjustments)
type StockMovement struct {
	ID               uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	InventoryItemID  uuid.UUID    `json:"inventory_item_id" gorm:"type:uuid;not null;index"`

	// Movement details
	Type             MovementType `json:"type" gorm:"type:varchar(20);not null;index"`
	Quantity         int          `json:"quantity" gorm:"not null"` // Positive for inbound, negative for outbound
	PreviousQuantity int          `json:"previous_quantity" gorm:"not null"`
	NewQuantity      int          `json:"new_quantity" gorm:"not null"`

	// Reference information
	ReferenceType    string       `json:"reference_type" gorm:"type:varchar(50)"` // order, purchase_order, adjustment, etc.
	ReferenceID      *uuid.UUID   `json:"reference_id" gorm:"type:uuid;index"`

	// Additional details
	Reason           string       `json:"reason" gorm:"type:varchar(255)"`
	Notes            string       `json:"notes" gorm:"type:text"`
	UnitCost         float64      `json:"unit_cost" gorm:"type:decimal(10,2);default:0"`
	TotalCost        float64      `json:"total_cost" gorm:"type:decimal(10,2);default:0"`

	// User tracking
	UserID           *uuid.UUID   `json:"user_id" gorm:"type:uuid;index"`
	UserName         string       `json:"user_name" gorm:"type:varchar(100)"`

	// Timestamps
	CreatedAt        time.Time    `json:"created_at" gorm:"autoCreateTime"`
	MovementDate     time.Time    `json:"movement_date" gorm:"not null"`

	// Relationships
	InventoryItem    *InventoryItem `json:"inventory_item,omitempty" gorm:"foreignKey:InventoryItemID"`
}

// StockReservation tracks reserved stock for pending orders
type StockReservation struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	InventoryItemID uuid.UUID `json:"inventory_item_id" gorm:"type:uuid;not null;index"`

	// Reservation details
	OrderID         uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	Quantity        int       `json:"quantity" gorm:"not null"`
	ReservedPrice   float64   `json:"reserved_price" gorm:"type:decimal(10,2);default:0"`

	// Status and lifecycle
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	ReservationCode string    `json:"reservation_code" gorm:"type:varchar(50);unique;index"`

	// Timestamps
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ExpiresAt       *time.Time `json:"expires_at"`
	ReleasedAt      *time.Time `json:"released_at"`

	// Relationships
	InventoryItem   *InventoryItem `json:"inventory_item,omitempty" gorm:"foreignKey:InventoryItemID"`
}

// StockAlert represents low stock alerts and notifications
type StockAlert struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	InventoryItemID uuid.UUID `json:"inventory_item_id" gorm:"type:uuid;not null;index"`

	// Alert details
	Type            string    `json:"type" gorm:"type:varchar(50);not null"` // low_stock, out_of_stock, overstock
	Message         string    `json:"message" gorm:"type:text;not null"`
	Severity        string    `json:"severity" gorm:"type:varchar(20);not null;default:'medium'"` // low, medium, high, critical

	// Status
	IsRead          bool      `json:"is_read" gorm:"default:false"`
	IsResolved      bool      `json:"is_resolved" gorm:"default:false"`

	// Timestamps
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ReadAt          *time.Time `json:"read_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`

	// Relationships
	InventoryItem   *InventoryItem `json:"inventory_item,omitempty" gorm:"foreignKey:InventoryItemID"`
}

// Address represents a physical address
type Address struct {
	AddressLine1 string `json:"address_line_1" gorm:"type:varchar(255)"`
	AddressLine2 string `json:"address_line_2" gorm:"type:varchar(255)"`
	City         string `json:"city" gorm:"type:varchar(100)"`
	State        string `json:"state" gorm:"type:varchar(100)"`
	PostalCode   string `json:"postal_code" gorm:"type:varchar(20)"`
	Country      string `json:"country" gorm:"type:varchar(100);default:'US'"`
}

// Business logic methods
func (item *InventoryItem) IsAvailable(quantity int) bool {
	return item.QuantityAvailable >= quantity
}

func (item *InventoryItem) IsLowStock() bool {
	return item.QuantityTotal <= item.ReorderPoint
}

func (item *InventoryItem) IsOutOfStock() bool {
	return item.QuantityTotal <= 0
}

func (item *InventoryItem) UpdateStatus() {
	if item.IsOutOfStock() {
		item.Status = StockStatusOutOfStock
	} else if item.IsLowStock() {
		item.Status = StockStatusLowStock
	} else {
		item.Status = StockStatusInStock
	}
}

func (item *InventoryItem) ReserveStock(quantity int) error {
	if !item.IsAvailable(quantity) {
		return fmt.Errorf("insufficient stock available: requested %d, available %d", quantity, item.QuantityAvailable)
	}

	item.QuantityAvailable -= quantity
	item.QuantityReserved += quantity
	item.UpdatedAt = time.Now()
	item.UpdateStatus()

	return nil
}

func (item *InventoryItem) ReleaseStock(quantity int) error {
	if item.QuantityReserved < quantity {
		return fmt.Errorf("insufficient reserved stock: requested %d, reserved %d", quantity, item.QuantityReserved)
	}

	item.QuantityReserved -= quantity
	item.QuantityAvailable += quantity
	item.UpdatedAt = time.Now()
	item.UpdateStatus()

	return nil
}

func (item *InventoryItem) FulfillStock(quantity int) error {
	if item.QuantityReserved < quantity {
		return fmt.Errorf("insufficient reserved stock: requested %d, reserved %d", quantity, item.QuantityReserved)
	}

	item.QuantityReserved -= quantity
	item.QuantityTotal -= quantity
	item.LastSoldAt = &time.Time{}
	*item.LastSoldAt = time.Now()
	item.UpdatedAt = time.Now()
	item.UpdateStatus()

	return nil
}

func (item *InventoryItem) AddStock(quantity int, costPrice float64) {
	item.QuantityAvailable += quantity
	item.QuantityTotal += quantity
	if costPrice > 0 {
		item.LastCostPrice = item.CostPrice
		item.CostPrice = costPrice
	}
	now := time.Now()
	item.LastRestockedAt = &now
	item.UpdatedAt = now
	item.UpdateStatus()
}

func (item *InventoryItem) GetTurnoverRate(days int) float64 {
	if item.LastSoldAt == nil || days <= 0 {
		return 0
	}

	// Simple turnover calculation - would need actual sales data in production
	averageStock := float64(item.QuantityTotal + item.QuantityReserved) / 2
	if averageStock == 0 {
		return 0
	}

	// Mock calculation - in real implementation, would calculate based on actual sales
	return float64(item.QuantityTotal) / averageStock
}

// Warehouse business logic
func (w *Warehouse) GetTotalCapacityUsed() int {
	totalItems := 0
	for _, item := range w.InventoryItems {
		totalItems += item.QuantityTotal
	}
	return totalItems
}

func (w *Warehouse) GetCapacityUtilization() float64 {
	if w.Capacity == 0 {
		return 0
	}
	used := float64(w.GetTotalCapacityUsed())
	return (used / float64(w.Capacity)) * 100
}

func (w *Warehouse) IsNearCapacity(threshold float64) bool {
	return w.GetCapacityUtilization() >= threshold
}

// StockReservation business logic
func (r *StockReservation) IsExpired() bool {
	return r.ExpiresAt != nil && time.Now().After(*r.ExpiresAt)
}

func (r *StockReservation) Expire() {
	r.IsActive = false
	now := time.Now()
	r.ReleasedAt = &now
	r.UpdatedAt = now
}

func (r *StockReservation) GenerateReservationCode() string {
	return fmt.Sprintf("RSV-%s", r.ID.String()[:8])
}

// Custom errors
type InventoryError struct {
	Message string
	Code    string
}

func (e InventoryError) Error() string {
	return e.Message
}

var (
	ErrInsufficientStock    = InventoryError{Message: "insufficient stock available", Code: "insufficient_stock"}
	ErrInvalidQuantity      = InventoryError{Message: "invalid quantity specified", Code: "invalid_quantity"}
	ErrInventoryNotFound    = InventoryError{Message: "inventory item not found", Code: "inventory_not_found"}
	ErrWarehouseNotFound    = InventoryError{Message: "warehouse not found", Code: "warehouse_not_found"}
	ErrReservationNotFound  = InventoryError{Message: "stock reservation not found", Code: "reservation_not_found"}
	ErrReservationExpired   = InventoryError{Message: "stock reservation has expired", Code: "reservation_expired"}
	ErrDuplicateReservation = InventoryError{Message: "duplicate reservation exists", Code: "duplicate_reservation"}
)