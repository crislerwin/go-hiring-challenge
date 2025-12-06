package categories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/internal/testutil"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) (*http.ServeMux, *gorm.DB) {
	db := testutil.SetupTestDB()

	repo := models.NewCategoriesRepository(db)
	handler := NewCategoriesHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /categories", handler.HandleList)
	mux.HandleFunc("POST /categories", handler.HandleCreate)

	return mux, db
}

func TestCategoriesEndpoint_List(t *testing.T) {
	mux, _ := setupTestServer(t)

	t.Run("GET /categories returns all categories", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response []CategoryResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Greater(t, len(response), 0, "Should have categories from seed data")

		// Verify we have the expected seed categories
		codes := make(map[string]bool)
		for _, cat := range response {
			codes[cat.Code] = true
		}
		assert.True(t, codes["CLOTHING"], "Should have CLOTHING category")
		assert.True(t, codes["SHOES"], "Should have SHOES category")
		assert.True(t, codes["ACCESSORIES"], "Should have ACCESSORIES category")
	})
}

func TestCategoriesEndpoint_Create(t *testing.T) {
	mux, db := setupTestServer(t)

	t.Run("POST /categories creates new category", func(t *testing.T) {
		testCode := "TEST_CREATE"

		// Cleanup before and after test
		db.Where("code = ?", testCode).Delete(&models.Category{})
		t.Cleanup(func() {
			db.Where("code = ?", testCode).Delete(&models.Category{})
		})

		requestBody := CreateCategoryRequest{
			Code: testCode,
			Name: "Test Category",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response CategoryResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, testCode, response.Code)
		assert.Equal(t, "Test Category", response.Name)

		// Verify it was actually created in database
		var created models.Category
		err = db.Where("code = ?", testCode).First(&created).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Category", created.Name)
	})

	t.Run("POST /categories returns 400 for invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /categories returns 400 for missing code", func(t *testing.T) {
		requestBody := CreateCategoryRequest{
			Code: "",
			Name: "Test",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /categories returns 400 for missing name", func(t *testing.T) {
		requestBody := CreateCategoryRequest{
			Code: "TEST",
			Name: "",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /categories returns 409 for duplicate code", func(t *testing.T) {
		requestBody := CreateCategoryRequest{
			Code: "CLOTHING", // Already exists in seed data
			Name: "Duplicate",
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var errorResponse map[string]string
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "already exists")
	})
}
