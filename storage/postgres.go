package storage

import (
	"fmt"
	"log"

	// "os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

// Config represents the database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var db *gorm.DB

// NewConnection establishes a new database connection and returns it
func NewConnection(config *Config) (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbURI := fmt.Sprintf("host=%s port=%s password=%s user=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Password, config.User, config.DBName)

	var err error
	db, err = gorm.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	return db, nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
