package middleware

import (
	"net/http"
	"strconv"

	"github.com/Secure-Website-Builder/Backend/internal/services/store"
	"github.com/gin-gonic/gin"
)

type StoreOwnerChecker struct {
	Service *store.Service
}

func NewStoreOwnerChecker(service *store.Service) *StoreOwnerChecker {
	return &StoreOwnerChecker{Service: service}
}

func (s *StoreOwnerChecker) IsOwner(c *gin.Context) {
	role := c.GetString("role")
	// Only enforce for store owners
	if role != "store_owner" {
		c.Next()
		return
	}

	userID := c.GetInt64("user_id")

	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	ok, err := s.Service.IsOwner(c.Request.Context(), userID, storeID)
	if err != nil || !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not store owner"})
		return
	}

	c.Next()
}

func RequireStoreOwner(checker *StoreOwnerChecker) gin.HandlerFunc {
	return checker.IsOwner
}
