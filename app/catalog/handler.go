package catalog

import (
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
		if price, err := decimal.NewFromString(priceStr); err == nil {
			priceLessThan = &price
		}
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
