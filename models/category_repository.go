package models

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	GetAllCategories() ([]Category, error)
	CreateCategory(category *Category) error
}

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoriesRepository) CreateCategory(category *Category) error {
	// Validate input
	if category == nil {
		return ErrInvalidCategory
	}

	if strings.TrimSpace(category.Code) == "" || strings.TrimSpace(category.Name) == "" {
		return ErrInvalidCategory
	}

	// Attempt to create
	if err := r.db.Create(category).Error; err != nil {
		// Check for PostgreSQL unique violation error (code 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrCategoryCodeExists
		}
		return err
	}

	return nil
}
