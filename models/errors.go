package models

import "errors"

// Domain errors for the models package
var (
	// Product errors
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProduct  = errors.New("invalid product data")

	// Category errors
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryCodeExists  = errors.New("category code already exists")
	ErrInvalidCategory     = errors.New("invalid category data")

	// Validation errors
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)
