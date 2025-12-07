package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Secure-Website-Builder/Backend/internal/services/product"
)

type ProductHandler struct {
	service *product.Service
}

func NewProductHandler(s *product.Service) *ProductHandler {
	return &ProductHandler{service: s}
}

func (h *ProductHandler) GetProduct(c *gin.Context) {

	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 64)
	productID, _ := strconv.ParseInt(c.Param("product_id"), 10, 64)

	product, err := h.service.GetFullProduct(c, storeID, productID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}
