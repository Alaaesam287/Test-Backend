package handlers

import (
	"net/http"
	"strconv"

	"github.com/Secure-Website-Builder/Backend/internal/services/category"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service *category.Service
}

func NewCategoryHandler(service *category.Service) *CategoryHandler {
	return &CategoryHandler{service}
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	storeParam := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	categories, err := h.service.ListCategoriesByStore(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) ListAttributes(c *gin.Context) {
	storeParam := c.Param("store_id")
	if _, err := strconv.ParseInt(storeParam, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	categoryParam := c.Param("category_id")
	categoryID, err := strconv.ParseInt(categoryParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}

	attributes, err := h.service.ListAttributesByCategory(c.Request.Context(), categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch attributes"})
		return
	}

	c.JSON(http.StatusOK, attributes)
}
