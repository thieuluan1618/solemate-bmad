package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/cart-service/internal/domain/service"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

type AddItemRequest struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id,omitempty"`
	Quantity  int        `json:"quantity" binding:"required,min=1"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

type ApplyDiscountRequest struct {
	Discount float64 `json:"discount" binding:"min=0"`
}

type ExtendExpirationRequest struct {
	Hours int `json:"hours" binding:"required,min=1,max=168"` // Max 1 week
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	cart, err := h.cartService.GetCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart", err.Error())
		return
	}

	utils.SuccessResponse(c, "Cart retrieved successfully", cart)
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err := h.cartService.ValidateAndAddItem(c.Request.Context(), userUUID, req.ProductID, req.VariantID, req.Quantity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to add item to cart", err.Error())
		return
	}

	// Return updated cart
	cart, err := h.cartService.GetCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Item added but failed to retrieve cart", "cart_retrieval_error")
		return
	}

	utils.SuccessResponse(c, "Item added to cart successfully", cart)
}

func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID format", "invalid_item_id")
		return
	}

	var req UpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err = h.cartService.UpdateItemQuantity(c.Request.Context(), userUUID, itemID, req.Quantity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update item quantity", err.Error())
		return
	}

	// Return updated cart
	cart, err := h.cartService.GetCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Quantity updated but failed to retrieve cart", "cart_retrieval_error")
		return
	}

	utils.SuccessResponse(c, "Item quantity updated successfully", cart)
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID format", "invalid_item_id")
		return
	}

	err = h.cartService.RemoveItem(c.Request.Context(), userUUID, itemID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to remove item from cart", err.Error())
		return
	}

	// Return updated cart
	cart, err := h.cartService.GetCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Item removed but failed to retrieve cart", "cart_retrieval_error")
		return
	}

	utils.SuccessResponse(c, "Item removed from cart successfully", cart)
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	err := h.cartService.ClearCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear cart", err.Error())
		return
	}

	// Return empty cart
	cart, err := h.cartService.GetCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Cart cleared but failed to retrieve cart", "cart_retrieval_error")
		return
	}

	utils.SuccessResponse(c, "Cart cleared successfully", cart)
}

func (h *CartHandler) GetCartSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	summary, err := h.cartService.GetCartSummary(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart summary", err.Error())
		return
	}

	utils.SuccessResponse(c, "Cart summary retrieved successfully", summary)
}

func (h *CartHandler) GetItemCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	count, err := h.cartService.GetItemCount(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get item count", err.Error())
		return
	}

	utils.SuccessResponse(c, "Item count retrieved successfully", gin.H{
		"item_count": count,
	})
}

func (h *CartHandler) ExtendExpiration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	var req ExtendExpirationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	duration := time.Duration(req.Hours) * time.Hour
	err := h.cartService.ExtendCartExpiration(c.Request.Context(), userUUID, duration)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to extend cart expiration", err.Error())
		return
	}

	utils.SuccessResponse(c, "Cart expiration extended successfully", gin.H{
		"message": "Cart expiration extended successfully",
		"hours":   req.Hours,
	})
}

func (h *CartHandler) ApplyDiscount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", "invalid_user_id")
		return
	}

	itemIDStr := c.Param("item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID format", "invalid_item_id")
		return
	}

	var req ApplyDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	err = h.cartService.ApplyDiscount(c.Request.Context(), userUUID, itemID, req.Discount)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to apply discount", err.Error())
		return
	}

	// Return updated cart
	cart, err := h.cartService.GetCart(c.Request.Context(), userUUID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Discount applied but failed to retrieve cart", "cart_retrieval_error")
		return
	}

	utils.SuccessResponse(c, "Discount applied successfully", cart)
}

func (h *CartHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	cartRoutes := router.Group("/cart")
	cartRoutes.Use(authMiddleware)

	cartRoutes.GET("", h.GetCart)
	cartRoutes.POST("/items", h.AddItem)
	cartRoutes.PATCH("/items/:item_id/quantity", h.UpdateItemQuantity)
	cartRoutes.DELETE("/items/:item_id", h.RemoveItem)
	cartRoutes.DELETE("", h.ClearCart)
	cartRoutes.GET("/summary", h.GetCartSummary)
	cartRoutes.GET("/count", h.GetItemCount)
	cartRoutes.POST("/extend", h.ExtendExpiration)
	cartRoutes.POST("/items/:item_id/discount", h.ApplyDiscount)
}