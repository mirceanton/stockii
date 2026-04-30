package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mirceanton/stockii/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var Driver string

func Init(driver, dsn string) error {
	Driver = driver

	var dialector gorm.Dialector

	switch driver {
	case "sqlite":
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create db directory: %w", err)
		}
		dialector = sqlite.Open(dsn + "?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000")
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver %q: must be \"sqlite\" or \"postgres\"", driver)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

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

	seedCategories()

	log.Printf("Database initialized (driver: %s)", driver)
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
