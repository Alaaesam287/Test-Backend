package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RequireSameStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")

		// Only enforce for customers
		if role != "customer" {
			c.Next()
			return
		}

		// Store ID from URL
		storeIDParam := c.Param("store_id")
		storeIDFromURL, err := strconv.ParseInt(storeIDParam, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid store_id",
			})
			return
		}

		// Store ID from token
		storeIDFromToken, exists := c.Get("store_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "store context missing in token",
			})
			return
		}

		if storeIDFromURL != storeIDFromToken.(int64) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "accessing another store is forbidden",
			})
			return
		}

		c.Next()
	}
}
