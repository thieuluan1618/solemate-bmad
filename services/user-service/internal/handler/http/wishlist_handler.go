package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/user-service/internal/domain/service"
)

type WishlistHandler struct {
	wishlistService *service.WishlistService
}

func NewWishlistHandler(wishlistService *service.WishlistService) *WishlistHandler {
	return &WishlistHandler{
		wishlistService: wishlistService,
	}
}

// GetWishlist retrieves user's wishlist
// GET /api/v1/wishlist
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	wishlist, err := h.wishlistService.GetWishlist(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get wishlist", err.Error())
		return
	}

	utils.SuccessResponse(c, "Wishlist retrieved successfully", wishlist)
}

// AddItem adds a product to wishlist
// POST /api/v1/wishlist/items
// Body: { "product_id": "uuid" }
func (h *WishlistHandler) AddItem(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	item, err := h.wishlistService.AddToWishlist(c.Request.Context(), id, productID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to add to wishlist", err.Error())
		return
	}

	utils.CreatedResponse(c, "Product added to wishlist", item)
}

// RemoveItem removes a product from wishlist
// DELETE /api/v1/wishlist/items/:product_id
func (h *WishlistHandler) RemoveItem(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	productIDParam := c.Param("product_id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	err = h.wishlistService.RemoveFromWishlist(c.Request.Context(), id, productID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to remove from wishlist", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product removed from wishlist", nil)
}

// ClearWishlist removes all items from wishlist
// DELETE /api/v1/wishlist
func (h *WishlistHandler) ClearWishlist(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	err = h.wishlistService.ClearWishlist(c.Request.Context(), id)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to clear wishlist", err.Error())
		return
	}

	utils.SuccessResponse(c, "Wishlist cleared successfully", nil)
}

// MoveToCart endpoint is handled by frontend calling cart API directly
// This handler is included for completeness but not implemented here
// Frontend should:
// 1. Get product from wishlist
// 2. Call cart service to add product
// 3. Call this service to remove from wishlist
func (h *WishlistHandler) MoveToCart(c *gin.Context) {
	utils.BadRequestResponse(c, "Move to cart should be handled by frontend",
		"Call cart service to add item, then remove from wishlist")
}
