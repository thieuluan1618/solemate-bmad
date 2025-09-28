package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/pkg/utils"
	"solemate/services/product-service/internal/domain/service"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	product, err := h.productService.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to create product", err.Error())
		return
	}

	utils.CreatedResponse(c, "Product created successfully", product)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	productIDParam := c.Param("id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	product, err := h.productService.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.SuccessResponse(c, "Product retrieved successfully", product)
}

func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.BadRequestResponse(c, "Product slug is required", "")
		return
	}

	product, err := h.productService.GetProductBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.SuccessResponse(c, "Product retrieved successfully", product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productIDParam := c.Param("id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), productID, &req)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to update product", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product updated successfully", product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productIDParam := c.Param("id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	err = h.productService.DeleteProduct(c.Request.Context(), productID)
	if err != nil {
		utils.BadRequestResponse(c, "Failed to delete product", err.Error())
		return
	}

	utils.SuccessResponse(c, "Product deleted successfully", nil)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	products, total, err := h.productService.ListProducts(c.Request.Context(), page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve products", err.Error())
		return
	}

	pagination := utils.CalculatePagination(page, limit, total)
	utils.PaginatedSuccessResponse(c, "Products retrieved successfully", products, pagination)
}

func (h *ProductHandler) SearchProducts(c *gin.Context) {
	var req service.ProductSearchRequest

	// Parse query parameters
	req.Query = c.Query("q")
	req.CategoryID = getStringPtr(c.Query("category_id"))
	req.BrandID = getStringPtr(c.Query("brand_id"))
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.SortOrder = c.DefaultQuery("sort_order", "desc")

	// Parse numeric parameters
	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			req.MinPrice = &price
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			req.MaxPrice = &price
		}
	}

	if inStock := c.Query("in_stock"); inStock != "" {
		if stock, err := strconv.ParseBool(inStock); err == nil {
			req.InStock = &stock
		}
	}

	// Parse pagination
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Parse tags
	if tagsParam := c.Query("tags"); tagsParam != "" {
		req.Tags = []string{tagsParam} // For simplicity, accepting single tag
	}

	products, total, err := h.productService.SearchProducts(c.Request.Context(), &req)
	if err != nil {
		utils.BadRequestResponse(c, "Search failed", err.Error())
		return
	}

	pagination := utils.CalculatePagination(req.Page, req.Limit, total)
	result := map[string]interface{}{
		"products": products,
		"query":    req.Query,
		"filters": map[string]interface{}{
			"category_id": req.CategoryID,
			"brand_id":    req.BrandID,
			"min_price":   req.MinPrice,
			"max_price":   req.MaxPrice,
			"tags":        req.Tags,
			"in_stock":    req.InStock,
		},
	}

	utils.PaginatedSuccessResponse(c, "Product search completed", result, pagination)
}

func (h *ProductHandler) GetRelatedProducts(c *gin.Context) {
	productIDParam := c.Param("id")
	productID, err := uuid.Parse(productIDParam)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err.Error())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "4"))

	products, err := h.productService.GetRelatedProducts(c.Request.Context(), productID, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to retrieve related products", err.Error())
		return
	}

	utils.SuccessResponse(c, "Related products retrieved successfully", products)
}

// Helper function
func getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}