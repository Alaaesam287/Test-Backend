package models

import "database/sql"

type ProductDTO struct {
	ProductID        int64            `json:"product_id"`
	StoreID          int64            `json:"store_id"`
	Name             string           `json:"name"`
	Slug             sql.NullString   `json:"slug"`
	Description      sql.NullString   `json:"description"`
	Brand            sql.NullString   `json:"brand"`
	CategoryID       sql.NullInt64    `json:"category_id"`
	InStock          bool             `json:"in_stock"`
	PrimaryImage     sql.NullString   `json:"primary_image"`
	Attributes       []AttributeDTO   `json:"attributes"`
	Variants         []VariantDTO     `json:"variants"`
	DefaultVariantID sql.NullInt64    `json:"default_variant_id"`
}

type AttributeDTO struct {
	AttributeID int64   `json:"attribute_id"`
	Name        string  `json:"name"`
	DataType    string  `json:"data_type"`
	Value       any     `json:"value"`
}

type VariantDTO struct {
	VariantID     int64             `json:"variant_id"`
	SKU           string           `json:"sku"`
	Price         string           `json:"price"`
	StockQuantity int32             `json:"stock_quantity"`
	ImageURL      string           `json:"image_url"`
	Options       []VariantOptionDTO `json:"options"`
}

type VariantOptionDTO struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
