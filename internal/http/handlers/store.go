package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Secure-Website-Builder/Backend/internal/services/store"
	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	Service *store.Service
}

func NewStoreHandler(s *store.Service) *StoreHandler {
	return &StoreHandler{Service: s}
}

type CreateStoreRequest struct {
	Name       string          `json:"name" binding:"required"`
	Domain     string          `json:"domain"`
	Currency   string          `json:"currency"`
	Timezone   string          `json:"timezone"`
	SiteConfig json.RawMessage `json:"site_config" binding:"required"`
}

func (h *StoreHandler) CreateStore(c *gin.Context) {

	var req CreateStoreRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storeOwnerID := c.GetInt64("store_owner_id")

	storeID, err := h.Service.CreateStore(
		c.Request.Context(),
		storeOwnerID,
		req.Name,
		req.Domain,
		req.Currency,
		req.Timezone,
		req.SiteConfig,
	)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"store_id": storeID,
	})
}
