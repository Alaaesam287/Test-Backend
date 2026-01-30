package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Secure-Website-Builder/Backend/internal/database"
	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/services/product"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	Service *product.Service
}

func NewProductHandler(s *product.Service) *ProductHandler {
	return &ProductHandler{Service: s}
}

func (h *ProductHandler) GetProduct(c *gin.Context) {

	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 64)
	productID, _ := strconv.ParseInt(c.Param("product_id"), 10, 64)

	product, err := h.Service.GetFullProduct(c.Request.Context(), storeID, productID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// ListProducts handles GET /stores/:store_id/products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	ctx := c.Request.Context()

	// parse store_id
	storeIDStr := c.Param("store_id")
	storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	// pagination
	page := 1
	limit := 20

	if p := c.DefaultQuery("page", "1"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.DefaultQuery("limit", "20"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}

	// category
	var categoryPtr *string
	if v := c.Query("category"); v != "" {
		categoryPtr = &v
	}

	var categoryID *int64
	if categoryPtr == nil {
		categoryID = nil
	} else {
		id, err := h.Service.ResolveCategoryNameToID(ctx, storeID, *categoryPtr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
			return
		}
		categoryID = &id
	}

	// price filters
	var minPricePtr *float64
	var maxPricePtr *float64

	if v := c.Query("min-price"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min-price"})
			return
		}
		minPricePtr = &f
	}

	if v := c.Query("max-price"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max-price"})
			return
		}
		maxPricePtr = &f
	}

	// brand
	var brandPtr *string
	if v := c.Query("brand"); v != "" {
		brandPtr = &v
	}

	// instock
	var instock *bool
	if v := c.Query("instock"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instock"})
			return
		}
		instock = &b
	}

	// reserved params
	reserved := map[string]bool{
		"page":      true,
		"limit":     true,
		"category":  true,
		"min-price": true,
		"max-price": true,
		"brand":     true,
		"instock":   true,
	}

	// ---------------------------------------
	// Attribute filters (grouped for IN logic)
	// ---------------------------------------
	attrMap := make(map[int64][]string)
	q := c.Request.URL.Query()

	for key, values := range q {
		if reserved[key] {
			continue
		}

		attrID, err := h.Service.ResolveAttributeNameToID(ctx, storeID, key)
		if err != nil {
			// unknown attribute name -> skip
			continue
		}

		// append all values for this attribute
		attrMap[attrID] = append(attrMap[attrID], values...)
	}

	attrFilters := make([]database.AttributeFilter, 0, len(attrMap))
	for attrID, vals := range attrMap {
		attrFilters = append(attrFilters, database.AttributeFilter{
			AttributeID: attrID,
			Values:      vals,
		})
	}

	// build final filters object
	filters := product.ListProductFilters{
		Page:       page,
		Limit:      limit,
		CategoryID: categoryID,
		MinPrice:   minPricePtr,
		MaxPrice:   maxPricePtr,
		Brand:      brandPtr,
		InStock:    instock,
		Attributes: attrFilters,
	}

	results, err := h.Service.ListProducts(ctx, storeID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "failed to load products",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 64)

	// 1. Read JSON part
	jsonPart := c.PostForm("data")
	if jsonPart == "" {
		c.JSON(400, gin.H{"error": "missing data field"})
		return
	}

	var req models.CreateProductInput
	if err := json.Unmarshal([]byte(jsonPart), &req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	// 2. Read file (optional)
	file, _, _ := c.Request.FormFile("primary_image")

	product, variant, err := h.Service.CreateProduct(
		c.Request.Context(),
		storeID,
		req,
		file,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"product":         product,
		"default_variant": variant,
	})
}

func (h *ProductHandler) AddVariant(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 64)
	productID, _ := strconv.ParseInt(c.Param("product_id"), 10, 64)

	jsonPart := c.PostForm("data")
	var req models.VariantInput
	if err := json.Unmarshal([]byte(jsonPart), &req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	file, _, _ := c.Request.FormFile("primary_image")

	variant, err := h.Service.AddVariant(
		c.Request.Context(),
		storeID,
		productID,
		req,
		file,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"variant": variant})
}

func (h *ProductHandler) UploadVariantImage(c *gin.Context) {
	storeID, _ := strconv.ParseInt(c.Param("store_id"), 10, 64)
	productID, _ := strconv.ParseInt(c.Param("product_id"), 10, 64)
	variantID, _ := strconv.ParseInt(c.Param("variant_id"), 10, 64)

	isPrimary := c.PostForm("is_primary") == "true"

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "cannot open file"})
		return
	}
	defer file.Close()

	url, err := h.Service.UploadVariantImage(
		c.Request.Context(),
		storeID,
		productID,
		variantID,
		file,
		isPrimary,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"image_url":  url,
		"is_primary": isPrimary,
	})
}
