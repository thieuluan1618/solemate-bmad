package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/product-service/internal/domain/service"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

// CreateReview handles POST /api/v1/products/:id/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	// Get product ID from URL
	productIDParam := c.Param("id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	// Bind request body
	var req service.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	// Set product ID from URL (override any product_id in request body)
	req.ProductID = productID.String()

	// Create review
	review, err := h.reviewService.CreateReview(c.Request.Context(), userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create review", err.Error())
		return
	}

	utils.CreatedResponse(c, "Review created successfully", review)
}

// GetReviewsByProductID handles GET /api/v1/products/:id/reviews
func (h *ReviewHandler) GetReviewsByProductID(c *gin.Context) {
	// Get product ID from URL
	productIDParam := c.Param("id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	// Parse query parameters
	var req service.GetReviewsRequest
	req.Status = c.Query("status")
	req.SortBy = c.DefaultQuery("sort_by", "newest")
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get reviews
	reviews, total, err := h.reviewService.GetReviewsByProductID(c.Request.Context(), productID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve reviews", err.Error())
		return
	}

	pagination := utils.CalculatePagination(req.Page, req.Limit, total)
	utils.PaginatedSuccessResponse(c, "Reviews retrieved successfully", reviews, pagination)
}

// GetReview handles GET /api/v1/reviews/:id
func (h *ReviewHandler) GetReview(c *gin.Context) {
	reviewIDParam := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid review ID", err.Error())
		return
	}

	review, err := h.reviewService.GetReviewByID(c.Request.Context(), reviewID)
	if err != nil {
		utils.NotFoundResponse(c, "Review not found")
		return
	}

	utils.SuccessResponse(c, "Review retrieved successfully", review)
}

// UpdateReview handles PUT /api/v1/reviews/:id
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	// Get review ID from URL
	reviewIDParam := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid review ID", err.Error())
		return
	}

	// Bind request body
	var req service.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	// Update review
	review, err := h.reviewService.UpdateReview(c.Request.Context(), userID, reviewID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to update review", err.Error())
		return
	}

	utils.SuccessResponse(c, "Review updated successfully", review)
}

// DeleteReview handles DELETE /api/v1/reviews/:id
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	// Get user ID from context
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		utils.UnauthorizedResponse(c, "Authentication required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", err.Error())
		return
	}

	// Get review ID from URL
	reviewIDParam := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid review ID", err.Error())
		return
	}

	// Delete review
	err = h.reviewService.DeleteReview(c.Request.Context(), userID, reviewID)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to delete review", err.Error())
		return
	}

	utils.SuccessResponse(c, "Review deleted successfully", nil)
}
