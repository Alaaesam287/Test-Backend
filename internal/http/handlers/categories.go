package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Secure-Website-Builder/Backend/internal/services/category"
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
