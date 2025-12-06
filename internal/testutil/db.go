// Package testutil provides shared testing utilities for integration tests.
// This package is internal and cannot be imported by external packages.
package testutil

import (
	"os"

	"github.com/mytheresa/go-hiring-challenge/app/database"
	"gorm.io/gorm"
)

// SetupTestDB initializes a test database connection using PostgreSQL.
// It reads configuration from environment variables with sensible defaults.
func SetupTestDB() *gorm.DB {
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "password")
	dbname := getEnv("POSTGRES_DB", "challenge")
	port := getEnv("POSTGRES_PORT", "5432")

	db, _ := database.New(user, password, dbname, port)
	return db
}

// getEnv retrieves an environment variable with a fallback value.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
