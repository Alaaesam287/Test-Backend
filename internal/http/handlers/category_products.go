package handlers

import (
	"net/http"
	"strconv"

	"github.com/Secure-Website-Builder/Backend/internal/services/product"
	"github.com/gin-gonic/gin"
)

type CategoryProductHandler struct {
	service *product.Service
}

func NewCategoryProductHandler(s *product.Service) *CategoryProductHandler {
	return &CategoryProductHandler{service: s}
}

func (h *CategoryProductHandler) GetTopProducts(c *gin.Context) {

	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 64)
	categoryID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Default limit = 5
	limitStr := c.DefaultQuery("limit", "5")
	limit, _ := strconv.ParseInt(limitStr, 10, 32)

	products, err := h.service.GetTopProductsByCategory(
		c.Request.Context(),
		storeID,
		categoryID,
		int32(limit),
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load top products",
		})
		return
	}

	c.JSON(http.StatusOK, products)
}
