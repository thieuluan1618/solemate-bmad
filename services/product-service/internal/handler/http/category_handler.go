package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/product-service/internal/domain/service"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	category, err := h.categoryService.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create category", err.Error())
		return
	}

	utils.CreatedResponse(c, "Category created successfully", category)
}

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	categoryIDParam := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category ID", err.Error())
		return
	}

	category, err := h.categoryService.GetCategoryByID(c.Request.Context(), categoryID)
	if err != nil {
		utils.NotFoundResponse(c, "Category not found")
		return
	}

	utils.SuccessResponse(c, "Category retrieved successfully", category)
}

func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.BadRequestResponse(c, "Category slug is required", "")
		return
	}

	category, err := h.categoryService.GetCategoryBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.NotFoundResponse(c, "Category not found")
		return
	}

	utils.SuccessResponse(c, "Category retrieved successfully", category)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	categoryIDParam := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category ID", err.Error())
		return
	}

	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	category, err := h.categoryService.UpdateCategory(c.Request.Context(), categoryID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to update category", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category updated successfully", category)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	categoryIDParam := c.Param("id")
	categoryID, err := uuid.Parse(categoryIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category ID", err.Error())
		return
	}

	err = h.categoryService.DeleteCategory(c.Request.Context(), categoryID)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to delete category", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category deleted successfully", nil)
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	categories, total, err := h.categoryService.ListCategories(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve categories", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.PaginatedSuccessResponse(c, "Categories retrieved successfully", categories, pagination)
}

func (h *CategoryHandler) GetCategoryTree(c *gin.Context) {
	categories, err := h.categoryService.GetCategoryTree(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve category tree", err.Error())
		return
	}

	utils.SuccessResponse(c, "Category tree retrieved successfully", categories)
}