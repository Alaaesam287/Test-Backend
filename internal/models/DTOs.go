package models

import (
	"database/sql"
)

type ProductFullDetailsDTO struct {
	ProductID      int64          `json:"product_id"`
	StoreID        int64          `json:"store_id"`
	ProductName    string         `json:"product_name"`
	Slug           sql.NullString `json:"slug"`
	Description    sql.NullString `json:"description"`
	Brand          sql.NullString `json:"brand"`
	TotalStock     int32          `json:"total_stock"`
	CategoryID     int64          `json:"category_id"`
	CategoryName   string         `json:"category_name"`
	Price          string         `json:"price"`
	InStock        bool           `json:"in_stock"`
	PrimaryImage   string         `json:"primary_image"`
	DefaultVariant VariantDTO     `json:"default_variant"`
	Variants       []VariantDTO   `json:"variants"`
}

type ProductDTO struct {
	ProductID   int64          `json:"product_id"`
	Name        string         `json:"name"`
	Slug        sql.NullString `json:"slug"`
	Brand       sql.NullString `json:"brand"`
	Description sql.NullString `json:"description"`
	CategoryID  int64          `json:"category_id"`
	TotalStock  int32          `json:"total_stock"`
	ItemStock   int32          `json:"item_stock"`
	Price       string         `json:"price"`
	ImageURL    string         `json:"image_url"`
	InStock     bool           `json:"in_stock"`
}

type AttributeDTO struct {
	AttributeID    int64  `json:"attribute_id"`
	AttributeName  string `json:"attribute_name"`
	AttributeValue string `json:"attribute_value"`
}

type VariantDTO struct {
	VariantID     int64          `json:"variant_id"`
	SKU           string         `json:"sku"`
	Price         string         `json:"price"`
	StockQuantity int32          `json:"stock_quantity"`
	ImageURL      string         `json:"image_url"`
	Attributes    []AttributeDTO `json:"attributes"`
}

type CartItemDTO struct {
	CartItemID int64  `json:"cart_item_id"`
	VariantID  int64  `json:"variant_id"`
	ProductID  int64  `json:"product_id"`
	Product    string `json:"product_name"`
	SKU        string `json:"sku"`
	ImageURL   string `json:"image_url"`
	Price      string `json:"price"`
	Quantity   int32  `json:"quantity"`
	Subtotal   string `json:"subtotal"`
}

type CartDTO struct {
	CartID    int64         `json:"cart_id"`
	StoreID   int64         `json:"store_id"`
	Items     []CartItemDTO `json:"items"`
	Total     float64       `json:"total"`
	UpdatedAt sql.NullTime  `json:"updated_at"`
}


type VariantAttributeInput struct {
	AttributeID int64  `json:"attribute_id"`
	Value       string `json:"value"`
}

type VariantInput struct {
	SKU        string                  `json:"sku"`
	Price      float64                 `json:"price"`
	Stock      int32                   `json:"stock"`
	ImageURL   string                  `json:"image_url"`
	Attributes []VariantAttributeInput `json:"attributes"`
}

type CreateProductInput struct {
	CategoryID int64       `json:"category_id"`
	Name       string      `json:"name"`
	Slug       string      `json:"slug"`
	Description string     `json:"description"`
	Brand      string      `json:"brand"`
	Variant    VariantInput `json:"variant"`
}
