package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/services/inventory-service/internal/domain/service"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

func (h *InventoryHandler) RegisterRoutes(router *gin.RouterGroup, jwtMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	inventory := router.Group("/inventory")
	inventory.Use(jwtMiddleware)
	{
		// Inventory item operations
		inventory.POST("/items", h.CreateInventoryItem)
		inventory.GET("/items/:id", h.GetInventoryItem)
		inventory.PUT("/items/:id", h.UpdateInventoryItem)
		inventory.DELETE("/items/:id", h.DeleteInventoryItem)
		inventory.GET("/items", h.SearchInventoryItems)

		// Stock operations
		inventory.POST("/check-availability", h.CheckStockAvailability)
		inventory.POST("/reserve", h.ReserveStock)
		inventory.DELETE("/reservations/:id", h.ReleaseStockReservation)
		inventory.POST("/reservations/:id/fulfill", h.FulfillStockReservation)
		inventory.POST("/adjust", h.AdjustStock)
		inventory.POST("/transfer", h.TransferStock)

		// Warehouse operations
		inventory.POST("/warehouses", h.CreateWarehouse)
		inventory.GET("/warehouses/:id", h.GetWarehouse)
		inventory.PUT("/warehouses/:id", h.UpdateWarehouse)
		inventory.GET("/warehouses", h.GetAllWarehouses)
		inventory.GET("/warehouses/:id/summary", h.GetWarehouseInventorySummary)

		// Stock movements
		inventory.GET("/movements", h.GetStockMovements)

		// Admin operations
		admin := inventory.Group("/admin")
		admin.Use(adminMiddleware)
		{
			admin.POST("/bulk-update", h.BulkStockUpdate)
			admin.POST("/bulk-reserve", h.BulkReserveStock)
			admin.GET("/analytics", h.GetInventoryAnalytics)
			admin.GET("/alerts", h.GetStockAlerts)
			admin.POST("/alerts/generate", h.GenerateStockAlerts)
			admin.POST("/alerts/:id/read", h.MarkAlertAsRead)
			admin.POST("/alerts/:id/resolve", h.MarkAlertAsResolved)
		}
	}
}

func (h *InventoryHandler) CreateInventoryItem(c *gin.Context) {
	var request service.CreateInventoryItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.CreateInventoryItem(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *InventoryHandler) GetInventoryItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	response, err := h.inventoryService.GetInventoryItem(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory item not found"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) UpdateInventoryItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var request service.UpdateInventoryItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.UpdateInventoryItem(c.Request.Context(), itemID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) DeleteInventoryItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := h.inventoryService.DeleteInventoryItem(c.Request.Context(), itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory item deleted successfully"})
}

func (h *InventoryHandler) SearchInventoryItems(c *gin.Context) {
	var request service.SearchInventoryRequest

	// Parse query parameters
	if productID := c.Query("product_id"); productID != "" {
		if id, err := uuid.Parse(productID); err == nil {
			request.ProductID = &id
		}
	}

	if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
		if id, err := uuid.Parse(warehouseID); err == nil {
			request.WarehouseID = &id
		}
	}

	request.SKU = c.Query("sku")
	request.Barcode = c.Query("barcode")
	request.Location = c.Query("location")
	request.SearchTerm = c.Query("search")
	request.SortBy = c.DefaultQuery("sort_by", "updated_at")
	request.SortOrder = c.DefaultQuery("sort_order", "desc")

	if lowStock := c.Query("low_stock"); lowStock == "true" {
		request.LowStock = true
	}

	if outOfStock := c.Query("out_of_stock"); outOfStock == "true" {
		request.OutOfStock = true
	}

	// Parse pagination
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	request.Limit = limit

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	request.Offset = offset

	response, err := h.inventoryService.SearchInventoryItems(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) CheckStockAvailability(c *gin.Context) {
	var request service.StockAvailabilityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.CheckStockAvailability(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) ReserveStock(c *gin.Context) {
	var request service.ReserveStockRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.ReserveStock(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *InventoryHandler) ReleaseStockReservation(c *gin.Context) {
	reservationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	if err := h.inventoryService.ReleaseStockReservation(c.Request.Context(), reservationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock reservation released successfully"})
}

func (h *InventoryHandler) FulfillStockReservation(c *gin.Context) {
	reservationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	if err := h.inventoryService.FulfillStockReservation(c.Request.Context(), reservationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock reservation fulfilled successfully"})
}

func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	var request service.AdjustStockRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.AdjustStock(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) TransferStock(c *gin.Context) {
	var request service.TransferStockRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.TransferStock(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movements": response})
}

func (h *InventoryHandler) CreateWarehouse(c *gin.Context) {
	var request service.CreateWarehouseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.CreateWarehouse(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *InventoryHandler) GetWarehouse(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse ID"})
		return
	}

	response, err := h.inventoryService.GetWarehouse(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) UpdateWarehouse(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse ID"})
		return
	}

	var request service.UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.UpdateWarehouse(c.Request.Context(), warehouseID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) GetAllWarehouses(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	response, err := h.inventoryService.GetAllWarehouses(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"warehouses": response})
}

func (h *InventoryHandler) GetWarehouseInventorySummary(c *gin.Context) {
	warehouseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warehouse ID"})
		return
	}

	response, err := h.inventoryService.GetWarehouseInventorySummary(c.Request.Context(), warehouseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) GetStockMovements(c *gin.Context) {
	var request service.StockMovementHistoryRequest

	// Parse query parameters
	if itemID := c.Query("item_id"); itemID != "" {
		if id, err := uuid.Parse(itemID); err == nil {
			request.InventoryItemID = &id
		}
	}

	if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
		if id, err := uuid.Parse(warehouseID); err == nil {
			request.WarehouseID = &id
		}
	}

	request.ReferenceType = c.Query("reference_type")

	// Parse pagination
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	request.Limit = limit

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	request.Offset = offset

	response, err := h.inventoryService.GetStockMovements(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Admin operations
func (h *InventoryHandler) BulkStockUpdate(c *gin.Context) {
	var request service.BulkStockUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.BulkStockUpdate(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) BulkReserveStock(c *gin.Context) {
	var request service.BulkReserveStockRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.inventoryService.BulkReserveStock(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) GetInventoryAnalytics(c *gin.Context) {
	// Implementation would parse date ranges and call analytics service
	c.JSON(http.StatusOK, gin.H{"message": "Analytics endpoint - implementation pending"})
}

func (h *InventoryHandler) GetStockAlerts(c *gin.Context) {
	var request service.StockAlertsRequest

	// Parse query parameters
	request.AlertType = c.Query("type")
	request.Severity = c.Query("severity")
	request.UnreadOnly = c.Query("unread") == "true"
	request.UnresolvedOnly = c.Query("unresolved") == "true"

	// Parse pagination
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	request.Limit = limit

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	request.Offset = offset

	response, err := h.inventoryService.GetStockAlerts(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) GenerateStockAlerts(c *gin.Context) {
	response, err := h.inventoryService.GenerateStockAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InventoryHandler) MarkAlertAsRead(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.inventoryService.MarkAlertAsRead(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert marked as read"})
}

func (h *InventoryHandler) MarkAlertAsResolved(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.inventoryService.MarkAlertAsResolved(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert marked as resolved"})
}