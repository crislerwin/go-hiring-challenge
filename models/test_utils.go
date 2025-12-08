package models

import (
	"os"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// setupTestDB initializes a test database connection using PostgreSQL
func setupTestDB(t *testing.T) *gorm.DB {
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "password")
	dbname := getEnv("POSTGRES_DB", "challenge")
	port := getEnv("POSTGRES_PORT", "5432")

	db, _ := database.New(user, password, dbname, port)
	return db
}

// getEnv retrieves an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// cleanupCategory removes a category by code after test completes
func cleanupCategory(t *testing.T, db *gorm.DB, code string) {
	t.Cleanup(func() {
		db.Where("code = ?", code).Delete(&Category{})
	})
}

// cleanupProduct removes a product by code after test completes
func cleanupProduct(t *testing.T, db *gorm.DB, code string) {
	t.Cleanup(func() {
		db.Where("code = ?", code).Delete(&Product{})
	})
}

// mustDecimal creates a decimal from string, panics on error (for test fixtures)
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}
