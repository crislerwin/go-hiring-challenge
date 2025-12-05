package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllCategories(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCategoriesRepository(db)

	t.Run("returns all categories from database", func(t *testing.T) {
		categories, err := repo.GetAllCategories()

		assert.NoError(t, err)
		assert.Greater(t, len(categories), 0, "Should have categories in database")

		// Verify we have the expected categories
		codes := make(map[string]bool)
		for _, cat := range categories {
			codes[cat.Code] = true
		}

		assert.True(t, codes["CLOTHING"], "Should have CLOTHING category")
		assert.True(t, codes["SHOES"], "Should have SHOES category")
		assert.True(t, codes["ACCESSORIES"], "Should have ACCESSORIES category")
	})
}

func TestCreateCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCategoriesRepository(db)

	t.Run("creates a new category successfully", func(t *testing.T) {
		// Clean up before and after test
		db.Where("code = ?", "ELECTRONICS").Delete(&Category{})
		cleanupCategory(t, db, "ELECTRONICS")

		newCategory := &Category{
			Code: "ELECTRONICS",
			Name: "Electronics",
		}

		err := repo.CreateCategory(newCategory)

		assert.NoError(t, err)
		assert.NotZero(t, newCategory.ID, "Should set ID after creation")

		// Verify it was created
		var found Category
		err = db.Where("code = ?", "ELECTRONICS").First(&found).Error
		assert.NoError(t, err)
		assert.Equal(t, "Electronics", found.Name)
	})

	t.Run("returns ErrCategoryCodeExists for duplicate code", func(t *testing.T) {
		duplicateCategory := &Category{
			Code: "CLOTHING", // Already exists
			Name: "Duplicate Clothing",
		}

		err := repo.CreateCategory(duplicateCategory)

		assert.ErrorIs(t, err, ErrCategoryCodeExists)
	})

	t.Run("returns ErrInvalidCategory for nil category", func(t *testing.T) {
		err := repo.CreateCategory(nil)

		assert.ErrorIs(t, err, ErrInvalidCategory)
	})

	t.Run("returns ErrInvalidCategory for empty code", func(t *testing.T) {
		invalidCategory := &Category{
			Code: "",
			Name: "Invalid",
		}

		err := repo.CreateCategory(invalidCategory)

		assert.ErrorIs(t, err, ErrInvalidCategory)
	})

	t.Run("returns ErrInvalidCategory for empty name", func(t *testing.T) {
		invalidCategory := &Category{
			Code: "VALID",
			Name: "",
		}

		err := repo.CreateCategory(invalidCategory)

		assert.ErrorIs(t, err, ErrInvalidCategory)
	})
}
