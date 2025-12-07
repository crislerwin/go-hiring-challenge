package categories

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

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
	slog.Info("Fetching all categories")

	categories, err := h.repo.GetAllCategories()
	if err != nil {
		slog.Error("Failed to fetch categories", "error", err)
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Info("Successfully fetched categories", "count", len(categories))

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
		slog.Warn("Invalid request body", "error", err)
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Code == "" || req.Name == "" {
		slog.Warn("Missing required fields", "code", req.Code, "name", req.Name)
		api.ErrorResponse(w, http.StatusBadRequest, "Code and name are required")
		return
	}

	// Validate non-whitespace
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		slog.Warn("Whitespace-only fields", "code", req.Code, "name", req.Name)
		api.ErrorResponse(w, http.StatusBadRequest, "Code and name cannot be empty or whitespace only")
		return
	}

	// Validate max length (code: 50 chars, name: 255 chars)
	if len(req.Code) > 50 {
		slog.Warn("Code too long", "code", req.Code, "length", len(req.Code))
		api.ErrorResponse(w, http.StatusBadRequest, "Code too long: maximum 50 characters")
		return
	}
	if len(req.Name) > 255 {
		slog.Warn("Name too long", "name", req.Name, "length", len(req.Name))
		api.ErrorResponse(w, http.StatusBadRequest, "Name too long: maximum 255 characters")
		return
	}

	slog.Info("Creating category", "code", req.Code, "name", req.Name)

	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.repo.CreateCategory(category); err != nil {
		if errors.Is(err, models.ErrCategoryCodeExists) {
			slog.Warn("Duplicate category code", "code", req.Code)
			api.ErrorResponse(w, http.StatusConflict, "Category code already exists")
			return
		}
		if errors.Is(err, models.ErrInvalidCategory) {
			slog.Warn("Invalid category", "code", req.Code, "error", err)
			api.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("Failed to create category", "code", req.Code, "error", err)
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Info("Successfully created category", "code", category.Code)

	response := CategoryResponse{
		Code: category.Code,
		Name: category.Name,
	}

	// Return 201 Created with JSON response
	api.CreatedResponse(w, response)
}
