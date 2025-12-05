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
