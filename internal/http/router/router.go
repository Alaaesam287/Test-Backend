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
	storeHandler *handlers.StoreHandler,
	rateLimiter *middleware.RateLimiter,
	storeOwnerChecker *middleware.StoreOwnerChecker,
	jwtSecret string,
) *gin.Engine {

	r := gin.Default()
	r.Use(middleware.ErrorMiddleware())
	r.Use(rateLimiter.Middleware())

	// Auth routes (public)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)
	r.POST("/auth/refresh", authHandler.RefreshToken)
	r.POST("/admin/auth/login", authHandler.AdminLogin)

	auth := r.Group("/")
	auth.Use(middleware.JWTAuth(jwtSecret))

	// Create store (store owner only)
	stores := auth.Group("/stores")
	stores.Use(middleware.RequireRole("store_owner"))
	{
		stores.POST("", storeHandler.CreateStore)
	}

	// Public / customer-facing store routes
	storeRoutes := auth.Group("/stores/:store_id")
	{
		// Shared middlewares for customer/store_owner/admin
		storeRoutes.Use(
			middleware.RequireRole("customer", "store_owner", "admin"),
			middleware.RequireSameStore(),
			middleware.RequireStoreOwner(storeOwnerChecker),
		)

		storeRoutes.GET("", storeHandler.GetStore)
		storeRoutes.GET("/categories", categoryHandler.ListCategories)
		storeRoutes.GET("/categories/:category_id/attributes", categoryHandler.ListAttributes)
		storeRoutes.GET("/categories/:category_id/top-products", categoryProductHandler.GetTopProducts)
		storeRoutes.GET("/products", productHandler.ListProducts)
		storeRoutes.GET("/products/:product_id", productHandler.GetProduct)
	}

	// Cart endpoints
	cartGroup := auth.Group("/stores/:store_id/cart")
	cartGroup.Use(
		middleware.RequireRole("customer", "admin"),
		middleware.RequireSameStore(),
	)
	cartGroup.GET("", cartHandler.GetCart)
	cartGroup.POST("/items", cartHandler.AddItem)
	cartGroup.POST("/checkout", cartHandler.Checkout)

	// Store owner dashboard routes
	dashboard := auth.Group("/dashboard/stores/:store_id")
	dashboard.Use(
		middleware.RequireRole("store_owner"),
		middleware.RequireStoreOwner(storeOwnerChecker),
	)
	{
		dashboard.POST("/products", productHandler.CreateProduct)
		dashboard.POST("/products/:product_id/variants", productHandler.AddVariant)
		dashboard.POST("/products/:product_id/variants/:variant_id/images", productHandler.UploadVariantImage)
	}

	// Admin-only routes
	admin := r.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	{
		// admin endpoints here
	}

	return r
}
