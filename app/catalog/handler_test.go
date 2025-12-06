package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/internal/testutil"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestServer() (*http.ServeMux, *gorm.DB) {
	db := testutil.SetupTestDB()

	repo := models.NewProductsRepository(db)
	handler := NewCatalogHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", handler.HandleGet)

	return mux, db
}

func TestCatalogEndpoint_DefaultPagination(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog with default pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Default should be limit=10
		assert.LessOrEqual(t, len(response.Products), 10, "Should return at most 10 products by default")
		assert.Greater(t, response.Total, int64(0), "Total should be greater than 0")

		// Verify products have required fields including category
		if len(response.Products) > 0 {
			product := response.Products[0]
			assert.NotEmpty(t, product.Code, "Product should have code")
			assert.Greater(t, product.Price, 0.0, "Product should have price")
			assert.NotEmpty(t, product.Category.Code, "Product should have category code")
			assert.NotEmpty(t, product.Category.Name, "Product should have category name")
		}
	})
}

func TestCatalogEndpoint_CustomPagination(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog with custom offset and limit", func(t *testing.T) {
		// First, get total count
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var response Response
		json.NewDecoder(w.Body).Decode(&response)
		total := response.Total

		// Now test with offset=2, limit=3
		req = httptest.NewRequest(http.MethodGet, "/catalog?offset=2&limit=3", nil)
		w = httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var paginatedResponse Response
		err := json.NewDecoder(w.Body).Decode(&paginatedResponse)
		assert.NoError(t, err)

		// Should return at most 3 products
		assert.LessOrEqual(t, len(paginatedResponse.Products), 3, "Should return at most 3 products")

		// Total should remain the same
		assert.Equal(t, total, paginatedResponse.Total, "Total count should be consistent")
	})

	t.Run("GET /catalog with large offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=1000&limit=10", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should return empty array but still have total count
		assert.Equal(t, 0, len(response.Products), "Should return empty products array")
		assert.Greater(t, response.Total, int64(0), "Total should still be greater than 0")
	})
}

func TestCatalogEndpoint_LimitValidation(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog with limit less than 1 should use 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=0", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should return at least 1 product if available
		if response.Total > 0 {
			assert.LessOrEqual(t, len(response.Products), 1, "Should apply minimum limit of 1")
		}
	})

	t.Run("GET /catalog with limit greater than 100 should use 100", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=500", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should return at most 100 products
		assert.LessOrEqual(t, len(response.Products), 100, "Should apply maximum limit of 100")
	})

	t.Run("GET /catalog with negative limit should use 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=-5", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should return at most 1 product
		if response.Total > 0 {
			assert.LessOrEqual(t, len(response.Products), 1, "Should apply minimum limit of 1")
		}
	})
}

func TestCatalogEndpoint_CategoryFilter(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog filtered by CLOTHING category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=CLOTHING", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Greater(t, len(response.Products), 0, "Should have CLOTHING products")

		// Verify all returned products are in CLOTHING category
		for _, product := range response.Products {
			assert.Equal(t, "CLOTHING", product.Category.Code, "All products should be in CLOTHING category")
		}
	})

	t.Run("GET /catalog filtered by SHOES category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=SHOES", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Greater(t, len(response.Products), 0, "Should have SHOES products")

		// Verify all returned products are in SHOES category
		for _, product := range response.Products {
			assert.Equal(t, "SHOES", product.Category.Code, "All products should be in SHOES category")
		}
	})

	t.Run("GET /catalog filtered by non-existent category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=NONEXISTENT", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should return empty array
		assert.Equal(t, 0, len(response.Products), "Should return no products for non-existent category")
		assert.Equal(t, int64(0), response.Total, "Total should be 0 for non-existent category")
	})
}

func TestCatalogEndpoint_PriceFilter(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog filtered by priceLessThan", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=15.00", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Greater(t, len(response.Products), 0, "Should have products under $15")

		// Verify all returned products are less than $15
		for _, product := range response.Products {
			assert.Less(t, product.Price, 15.00, "All products should be less than $15")
		}
	})

	t.Run("GET /catalog filtered by very low price", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=1.00", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should return empty or very few products
		assert.Equal(t, 0, len(response.Products), "Should return no products under $1")
	})
}

func TestCatalogEndpoint_CombinedFilters(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog with category and price filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=SHOES&priceLessThan=10.00", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify all returned products match both filters
		for _, product := range response.Products {
			assert.Equal(t, "SHOES", product.Category.Code, "All products should be in SHOES category")
			assert.Less(t, product.Price, 10.00, "All products should be less than $10")
		}
	})

	t.Run("GET /catalog with all filters and pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog?category=CLOTHING&priceLessThan=20.00&offset=0&limit=5", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Should respect all filters
		assert.LessOrEqual(t, len(response.Products), 5, "Should return at most 5 products")

		for _, product := range response.Products {
			assert.Equal(t, "CLOTHING", product.Category.Code, "All products should be in CLOTHING category")
			assert.Less(t, product.Price, 20.00, "All products should be less than $20")
		}
	})
}

func TestCatalogEndpoint_ResponseFormat(t *testing.T) {
	mux, _ := setupTestServer()

	t.Run("GET /catalog returns correct response structure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response Response
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response has both products and total
		assert.NotNil(t, response.Products, "Response should have products array")
		assert.Greater(t, response.Total, int64(0), "Response should have total count")

		// Verify product structure
		if len(response.Products) > 0 {
			product := response.Products[0]

			// Product fields
			assert.NotEmpty(t, product.Code, "Product should have code")
			assert.Greater(t, product.Price, 0.0, "Product should have price")

			// Category nested structure
			assert.NotEmpty(t, product.Category.Code, "Product should have category code")
			assert.NotEmpty(t, product.Category.Name, "Product should have category name")
		}
	})
}
