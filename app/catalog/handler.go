package catalog

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type Response struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}

type Product struct {
	Code     string   `json:"code"`
	Price    float64  `json:"price"`
	Category Category `json:"category"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ProductDetailsResponse represents a single product with full details
type ProductDetailsResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category Category          `json:"category"`
	Variants []VariantResponse `json:"variants"`
}

// VariantResponse represents a product variant
type VariantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type CatalogHandler struct {
	repo *models.ProductsRepository
}

func NewCatalogHandler(r *models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offset := parseIntParam(r, "offset", 0)
	limit := parseIntParam(r, "limit", 10)

	// Validate and normalize limit (min: 1, max: 100)
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	// Parse filter parameters
	categoryCode := r.URL.Query().Get("category")
	var priceLessThan *decimal.Decimal
	if priceStr := r.URL.Query().Get("priceLessThan"); priceStr != "" {
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid priceLessThan format: must be a valid number")
			return
		}
		if price.IsNegative() {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid priceLessThan: must be a positive number")
			return
		}
		priceLessThan = &price
	}

	// Build filters
	filters := models.ProductFilters{
		Offset:        offset,
		Limit:         limit,
		CategoryCode:  categoryCode,
		PriceLessThan: priceLessThan,
	}

	// Fetch products with filters
	products, total, err := h.repo.GetProductsWithFilters(filters)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map response
	response := mapProductsResponse(products, total)
	api.OKResponse(w, response)
}

// parseIntParam parses an integer query parameter with a default value
func parseIntParam(r *http.Request, key string, defaultValue int) int {
	if valueStr := r.URL.Query().Get(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// mapProductsResponse maps domain models to API response
func mapProductsResponse(products []models.Product, total int64) Response {
	responseProducts := make([]Product, len(products))
	for i, p := range products {
		responseProducts[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: Category{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	return Response{
		Products: responseProducts,
		Total:    total,
	}
}

// HandleGetDetails handles GET /catalog/{code} - returns product details with variants
func (h *CatalogHandler) HandleGetDetails(w http.ResponseWriter, r *http.Request) {
	// Extract product code from URL path parameter
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Product code is required")
		return
	}

	// Fetch product by code from repository
	product, err := h.repo.GetProductByCode(code)
	if err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, models.ErrProductNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "Product not found")
			return
		}
		// Other errors are internal server errors
		api.ErrorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Map to response with variant price inheritance
	response := mapProductDetailsResponse(product)
	api.OKResponse(w, response)
}

// mapProductDetailsResponse maps product model to details response
// Implements variant price inheritance: variants with zero/null price inherit from product
func mapProductDetailsResponse(product *models.Product) ProductDetailsResponse {
	variants := make([]VariantResponse, len(product.Variants))

	for i, v := range product.Variants {
		price := v.Price

		// Price inheritance logic: if variant price is zero (NULL in DB), inherit from product
		if price.IsZero() {
			price = product.Price
		}

		variants[i] = VariantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	return ProductDetailsResponse{
		Code:  product.Code,
		Price: product.Price.InexactFloat64(),
		Category: Category{
			Code: product.Category.Code,
			Name: product.Category.Name,
		},
		Variants: variants,
	}
}
