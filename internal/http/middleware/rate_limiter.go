package middleware

import (
	"net/http"

	"github.com/Secure-Website-Builder/Backend/internal/limiter"
	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	manager *limiter.Manager
}

func NewRateLimiter(m *limiter.Manager) *RateLimiter {
	return &RateLimiter{manager: m}
}

func clientKey(c *gin.Context) string {
	ip := c.ClientIP()

	storeID := c.GetHeader("X-Store-ID")
	if storeID == "" {
		storeID = c.Param("storeId")
	}

	if storeID == "" {
		return ip
	}

	return ip + ":" + storeID
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := clientKey(c)
		bucket := rl.manager.GetBucket(key)

		if !bucket.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
