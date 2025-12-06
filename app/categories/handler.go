package categories

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoriesHandler struct {
	repo models.CategoryRepository
}

func NewCategoriesHandler(repo models.CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{repo: repo}
}

func (h *CategoriesHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		response[i] = CategoryResponse{
			Code: cat.Code,
			Name: cat.Name,
		}
	}

	api.OKResponse(w, response)
}

func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Code and name are required")
		return
	}

	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.repo.CreateCategory(category); err != nil {
		if errors.Is(err, models.ErrCategoryCodeExists) {
			api.ErrorResponse(w, http.StatusConflict, "Category code already exists")
			return
		}
		if errors.Is(err, models.ErrInvalidCategory) {
			api.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := CategoryResponse{
		Code: category.Code,
		Name: category.Name,
	}

	// Return 201 Created with JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
