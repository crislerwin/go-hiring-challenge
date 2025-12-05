package models

import "gorm.io/gorm"

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
	return r.db.Create(category).Error
}
