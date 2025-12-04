package models

// Category represents a product category in the catalog.
// It includes a unique code and a human-readable name.
type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Code string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null"`
}

func (c *Category) TableName() string {
	return "categories"
}
