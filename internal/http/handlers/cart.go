package handlers

import (
	"net/http"
	"strconv"

	"github.com/Secure-Website-Builder/Backend/internal/services/cart"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CartHandler struct {
	Service *cart.Service
}

func NewCartHandler(s *cart.Service) *CartHandler {
	return &CartHandler{Service: s}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	ctx := c.Request.Context()

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	rawSessionID := c.GetHeader("X-Session-ID")
	if rawSessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Session-ID header"})
		return
	}

	sessionID, err := uuid.Parse(rawSessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}

	cartDTO, err := h.Service.GetCart(ctx, storeID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart"})
		return
	}

	c.JSON(http.StatusOK, cartDTO)
}
