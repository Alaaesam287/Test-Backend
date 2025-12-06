package router

import (
	"github.com/Secure-Website-Builder/Backend/internal/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(categoryHandler *handlers.CategoryHandler) *gin.Engine {
	r := gin.Default()

	stores := r.Group("/stores/:store_id")
	{
		stores.GET("/categories", categoryHandler.ListCategories)
	}

	return r
}
