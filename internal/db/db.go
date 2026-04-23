package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mirceanton/stockii/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dbPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Auto-migrate all models
	if err := DB.AutoMigrate(
		&models.Category{},
		&models.Fandom{},
		&models.Product{},
		&models.ConventionSeries{},
		&models.Convention{},
		&models.ConventionProduct{},
		&models.Sale{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	// Seed default categories
	seedCategories()

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Failed to get underlying DB: %v", err)
		return
	}
	sqlDB.Close()
}

func seedCategories() {
	defaults := []string{
		"Print A3", "Print A4", "Print A5",
		"Sticker", "Sticker Pack",
		"Keychain", "Enamel Pin",
		"Poster", "Bookmark", "Other",
	}
	for _, name := range defaults {
		DB.FirstOrCreate(&models.Category{}, models.Category{Name: name})
	}
}
