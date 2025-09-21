package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/product-service/internal/domain/service"
)

type BrandHandler struct {
	brandService *service.BrandService
}

func NewBrandHandler(brandService *service.BrandService) *BrandHandler {
	return &BrandHandler{
		brandService: brandService,
	}
}

func (h *BrandHandler) CreateBrand(c *gin.Context) {
	var req service.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	brand, err := h.brandService.CreateBrand(c.Request.Context(), &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create brand", err.Error())
		return
	}

	utils.CreatedResponse(c, "Brand created successfully", brand)
}

func (h *BrandHandler) GetBrand(c *gin.Context) {
	brandIDParam := c.Param("id")
	brandID, err := uuid.Parse(brandIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid brand ID", err.Error())
		return
	}

	brand, err := h.brandService.GetBrandByID(c.Request.Context(), brandID)
	if err != nil {
		utils.NotFoundResponse(c, "Brand not found")
		return
	}

	utils.SuccessResponse(c, "Brand retrieved successfully", brand)
}

func (h *BrandHandler) GetBrandBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.BadRequestResponse(c, "Brand slug is required", "")
		return
	}

	brand, err := h.brandService.GetBrandBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.NotFoundResponse(c, "Brand not found")
		return
	}

	utils.SuccessResponse(c, "Brand retrieved successfully", brand)
}

func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	brandIDParam := c.Param("id")
	brandID, err := uuid.Parse(brandIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid brand ID", err.Error())
		return
	}

	var req service.UpdateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	brand, err := h.brandService.UpdateBrand(c.Request.Context(), brandID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to update brand", err.Error())
		return
	}

	utils.SuccessResponse(c, "Brand updated successfully", brand)
}

func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	brandIDParam := c.Param("id")
	brandID, err := uuid.Parse(brandIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid brand ID", err.Error())
		return
	}

	err = h.brandService.DeleteBrand(c.Request.Context(), brandID)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to delete brand", err.Error())
		return
	}

	utils.SuccessResponse(c, "Brand deleted successfully", nil)
}

func (h *BrandHandler) ListBrands(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	brands, total, err := h.brandService.ListBrands(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve brands", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.PaginatedSuccessResponse(c, "Brands retrieved successfully", brands, pagination)
}