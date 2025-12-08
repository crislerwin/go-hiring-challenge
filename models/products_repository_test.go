package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllProducts_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProductsRepository(db)

	t.Run("returns products with pagination", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(0, 10)

		assert.NoError(t, err)
		assert.Greater(t, total, int64(0), "Should have products in database")
		assert.LessOrEqual(t, len(products), 10, "Should respect limit")
		if len(products) > 0 {
			assert.NotNil(t, products[0].Category, "Should preload category")
		}
	})

	t.Run("handles large offset", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(1000, 10)

		assert.NoError(t, err)
		assert.Greater(t, total, int64(0))
		assert.Equal(t, 0, len(products), "Should return empty array for offset beyond total")
	})

	t.Run("returns ErrInvalidPagination for negative offset", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(-1, 10)

		assert.ErrorIs(t, err, ErrInvalidPagination)
		assert.Nil(t, products)
		assert.Equal(t, int64(0), total)
	})

	t.Run("returns ErrInvalidPagination for negative limit", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(0, -1)

		assert.ErrorIs(t, err, ErrInvalidPagination)
		assert.Nil(t, products)
		assert.Equal(t, int64(0), total)
	})

	t.Run("returns ErrInvalidPagination for zero limit", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(0, 0)

		assert.ErrorIs(t, err, ErrInvalidPagination)
		assert.Nil(t, products)
		assert.Equal(t, int64(0), total)
	})
}

func TestGetProductByCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProductsRepository(db)

	t.Run("returns ErrProductNotFound for non-existent product", func(t *testing.T) {
		product, err := repo.GetProductByCode("NONEXISTENT")

		assert.ErrorIs(t, err, ErrProductNotFound)
		assert.Nil(t, product)
	})

	t.Run("returns product successfully", func(t *testing.T) {
		product, err := repo.GetProductByCode("PROD001")

		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "PROD001", product.Code)
		assert.NotNil(t, product.Category)
	})
}

func TestGetProductsWithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProductsRepository(db)

	t.Run("filters by category code", func(t *testing.T) {
		filters := ProductFilters{
			Offset:       0,
			Limit:        10,
			CategoryCode: "CLOTHING",
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		assert.Greater(t, len(products), 0, "Should have CLOTHING products")
		assert.Equal(t, total, int64(len(products)))

		// Verify all products are in CLOTHING category
		for _, p := range products {
			assert.Equal(t, "CLOTHING", p.Category.Code)
		}
	})

	t.Run("filters by price less than", func(t *testing.T) {
		price := mustDecimal("15.00")
		filters := ProductFilters{
			Offset:        0,
			Limit:         10,
			PriceLessThan: &price,
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		// Verify all products are less than $15
		for _, p := range products {
			assert.True(t, p.Price.LessThan(price), "Product price should be less than filter")
		}
		assert.Equal(t, int64(len(products)), total)
	})

	t.Run("filters by category and price combined", func(t *testing.T) {
		price := mustDecimal("20.00")
		filters := ProductFilters{
			Offset:        0,
			Limit:         10,
			CategoryCode:  "SHOES",
			PriceLessThan: &price,
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		// Verify all products match both filters
		for _, p := range products {
			assert.Equal(t, "SHOES", p.Category.Code)
			assert.True(t, p.Price.LessThan(price))
		}
		assert.Equal(t, int64(len(products)), total)
	})

	t.Run("respects pagination with filters", func(t *testing.T) {
		filters := ProductFilters{
			Offset:       1,
			Limit:        2,
			CategoryCode: "CLOTHING",
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		assert.LessOrEqual(t, len(products), 2, "Should respect limit")
		assert.GreaterOrEqual(t, total, int64(len(products)), "Total should be >= returned count")
	})

	t.Run("returns empty for non-existent category", func(t *testing.T) {
		filters := ProductFilters{
			Offset:       0,
			Limit:        10,
			CategoryCode: "NONEXISTENT",
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(products))
		assert.Equal(t, int64(0), total)
	})

	t.Run("returns empty for price filter with no matches", func(t *testing.T) {
		price := mustDecimal("0.01")
		filters := ProductFilters{
			Offset:        0,
			Limit:         10,
			PriceLessThan: &price,
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(products))
		assert.Equal(t, int64(0), total)
	})

	t.Run("returns ErrInvalidPagination for negative offset", func(t *testing.T) {
		filters := ProductFilters{
			Offset: -1,
			Limit:  10,
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.ErrorIs(t, err, ErrInvalidPagination)
		assert.Nil(t, products)
		assert.Equal(t, int64(0), total)
	})

	t.Run("returns ErrInvalidPagination for zero limit", func(t *testing.T) {
		filters := ProductFilters{
			Offset: 0,
			Limit:  0,
		}

		products, total, err := repo.GetProductsWithFilters(filters)

		assert.ErrorIs(t, err, ErrInvalidPagination)
		assert.Nil(t, products)
		assert.Equal(t, int64(0), total)
	})

	t.Run("preloads category and variants", func(t *testing.T) {
		filters := ProductFilters{
			Offset:       0,
			Limit:        1,
			CategoryCode: "CLOTHING",
		}

		products, _, err := repo.GetProductsWithFilters(filters)

		assert.NoError(t, err)
		if len(products) > 0 {
			assert.NotNil(t, products[0].Category, "Should preload category")
			assert.NotEmpty(t, products[0].Category.Code)
			// Variants should be preloaded (even if empty array)
			assert.NotNil(t, products[0].Variants)
		}
	})
}
