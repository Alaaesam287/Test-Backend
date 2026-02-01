package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/services/cart"
	"github.com/Secure-Website-Builder/Backend/internal/errorx"
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

type AddCartItemRequest struct {
	VariantID int64 `json:"variant_id" binding:"required"`
	Quantity  int32 `json:"quantity" binding:"required,min=1"`
}

func (h *CartHandler) AddItem(c *gin.Context) {
	ctx := c.Request.Context()

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	rawSessionID := c.GetHeader("X-Session-ID")
	if rawSessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-Session-ID header"})
		return
	}

	sessionID, err := uuid.Parse(rawSessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session id format"})
		return
	}

	var req AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = h.Service.AddItem(
		ctx,
		storeID,
		sessionID,
		req.VariantID,
		req.Quantity,
	)

	if err != nil {
		h.handleAddItemError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

// TO DO: refactor error handling into a middleware or utility
// helper to handle errors from AddItem
func (h *CartHandler) handleAddItemError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})

	case strings.Contains(err.Error(), "invalid session"):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})

	case strings.Contains(err.Error(), "invalid variant"):
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid variant"})

	case strings.Contains(err.Error(), "insufficient stock"):
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient stock"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item to cart"})
	}
}


type CheckoutRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required"`
}

func (h *CartHandler) Checkout(c *gin.Context) {
	ctx := c.Request.Context()

	storeID, err := strconv.ParseInt(c.Param("store_id"), 10, 64)
	if err != nil {
		c.Error(errorx.ErrInvalidStoreID)
		return 
	}

	rawSessionID := c.GetHeader("X-Session-ID")
	if rawSessionID == "" {
		c.Error(errorx.ErrMissingSessionID)
		return
	}

	sessionID, err := uuid.Parse(rawSessionID)
	if err != nil {
		c.Error(errorx.ErrInvalidSessionID)
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errorx.ErrInvalidRequestBody)
		return
	}

	err = h.Service.Checkout(ctx, storeID, sessionID, req.PaymentMethod)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusCreated)
}
