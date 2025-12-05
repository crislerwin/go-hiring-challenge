package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductFilters contains filtering options for product queries
type ProductFilters struct {
	Offset        int
	Limit         int
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	GetAllProducts(offset, limit int) ([]Product, int64, error)
	GetProductByCode(code string) (*Product, error)
	GetProductsWithFilters(filters ProductFilters) ([]Product, int64, error)
}

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

// GetAllProducts retrieves products with pagination
func (r *ProductsRepository) GetAllProducts(offset, limit int) ([]Product, int64, error) {
	var products []Product
	var total int64

	// Count total products
	if err := r.db.Model(&Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated products with relationships
	if err := r.db.Preload("Category").Preload("Variants").
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetProductByCode retrieves a single product by its code
func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Preload("Category").Preload("Variants").
		Where("code = ?", code).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// GetProductsWithFilters retrieves products with filtering and pagination
func (r *ProductsRepository) GetProductsWithFilters(filters ProductFilters) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.Model(&Product{})

	// Apply category filter
	if filters.CategoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", filters.CategoryCode)
	}

	// Apply price filter
	if filters.PriceLessThan != nil {
		query = query.Where("products.price < ?", filters.PriceLessThan)
	}

	// Count total with filters
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination and preload
	if err := query.Preload("Category").Preload("Variants").
		Offset(filters.Offset).Limit(filters.Limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
