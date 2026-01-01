package router

import (
	"github.com/Secure-Website-Builder/Backend/internal/http/handlers"
	"github.com/Secure-Website-Builder/Backend/internal/http/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	categoryHandler *handlers.CategoryHandler,
	productHandler *handlers.ProductHandler,
	categoryProductHandler *handlers.CategoryProductHandler,
	cartHandler *handlers.CartHandler,
	authHandler *handlers.AuthHandler,
	storeOwnerChecker *middleware.StoreOwnerChecker,
	jwtSecret string,
) *gin.Engine {

	r := gin.Default()

	// Auth routes (public)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)
	r.POST("/auth/refresh", authHandler.RefreshToken)
	r.POST("/admin/auth/login", authHandler.AdminLogin)

	auth := r.Group("/")
	auth.Use(middleware.JWTAuth(jwtSecret))

	stores := auth.Group("/stores/:store_id")
	{
		// Shared middlewares for customer/store_owner/admin
		stores.Use(
			middleware.RequireRole("customer", "store_owner", "admin"),
			middleware.RequireSameStore(),
			middleware.RequireStoreOwner(storeOwnerChecker),
		)

		stores.GET("/categories", categoryHandler.ListCategories)
		stores.GET("/categories/:category_id/attributes", categoryHandler.ListAttributes)
		stores.GET("/categories/:category_id/top-products", categoryProductHandler.GetTopProducts)
		stores.GET("/products", productHandler.ListProducts)
		stores.GET("/products/:product_id", productHandler.GetProduct)
	}

	// Cart endpoints
	cartGroup := auth.Group("/stores/:store_id/cart")
	cartGroup.Use(
		middleware.RequireRole("customer", "admin"),
		middleware.RequireSameStore(),
	)
	cartGroup.GET("", cartHandler.GetCart)

	// Admin-only routes
	admin := r.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	{
		// admin endpoints here
	}

	return r
}
