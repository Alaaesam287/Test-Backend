package middleware

import (
	"net/http"
	"strings"

	"github.com/Secure-Website-Builder/Backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, claims, err := utils.ParseJWT(tokenStr, secret)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", int64(claims["user_id"].(float64)))
		c.Set("role", claims["role"].(string))

		if storeID, ok := claims["store_id"]; ok {
			id := int64(storeID.(float64))
			c.Set("store_id", &id)
		}

		c.Next()
	}
}

// RequireRole middleware
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		for _, r := range roles {
			if r == role || role == "admin" { // admin bypass
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
