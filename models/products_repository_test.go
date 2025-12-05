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

	t.Run("returns paginated results with offset and limit", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(0, 3)

		assert.NoError(t, err)
		assert.Greater(t, total, int64(0))
		assert.LessOrEqual(t, len(products), 3)
	})

	t.Run("handles large offset", func(t *testing.T) {
		products, total, err := repo.GetAllProducts(1000, 10)

		assert.NoError(t, err)
		assert.Greater(t, total, int64(0))
		assert.Equal(t, 0, len(products), "Should return empty array for offset beyond total")
	})
}
