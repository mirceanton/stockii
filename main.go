package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mirceanton/stockii/internal/db"
	"github.com/mirceanton/stockii/internal/handlers"
	stockiimcp "github.com/mirceanton/stockii/internal/mcp"
)

func main() {
	port := os.Getenv("STOCKII_PORT")
	if port == "" {
		port = "8080"
	}

	driver := os.Getenv("STOCKII_DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	var dsn string
	switch driver {
	case "sqlite":
		dsn = os.Getenv("STOCKII_DB_URI")
		if dsn == "" {
			dsn = os.Getenv("STOCKII_CONFIG_PATH")
		}
		if dsn == "" {
			dsn = "/config/stockii.db"
		}
	case "postgres":
		dsn = os.Getenv("STOCKII_DB_URI")
		if dsn == "" {
			host := os.Getenv("STOCKII_DB_HOST")
			if host == "" {
				host = "localhost"
			}
			port := os.Getenv("STOCKII_DB_PORT")
			if port == "" {
				port = "5432"
			}
			user := os.Getenv("STOCKII_DB_USER")
			if user == "" {
				user = "postgres"
			}
			password := os.Getenv("STOCKII_DB_PASSWORD")
			dbName := os.Getenv("STOCKII_DB_NAME")
			if dbName == "" {
				dbName = "stockii"
			}
			sslMode := os.Getenv("STOCKII_DB_SSLMODE")
			if sslMode == "" {
				sslMode = "disable"
			}
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				host, port, user, password, dbName, sslMode)
		}
	}

	dataPath := os.Getenv("STOCKII_DATA_PATH")
	if dataPath == "" {
		dataPath = "/data"
	}

	templatesPath := os.Getenv("STOCKII_TEMPLATES_PATH")
	if templatesPath == "" {
		templatesPath = "templates"
	}

	staticPath := os.Getenv("STOCKII_STATIC_PATH")
	if staticPath == "" {
		staticPath = "static"
	}

	handlers.SetDataPath(dataPath)

	// Initialize database
	if err := db.Init(driver, dsn); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize templates
	if err := handlers.InitTemplates(templatesPath); err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	// Ensure images directory exists
	if err := os.MkdirAll(filepath.Join(dataPath, "images"), 0o755); err != nil {
		log.Fatalf("Failed to create images directory: %v", err)
	}

	// Set up router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Static files
	imagesDir := http.Dir(filepath.Join(dataPath, "images"))
	r.Handle("/images/*", http.StripPrefix("/images/", http.FileServer(imagesDir)))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Pages
	r.Get("/", handlers.DashboardHandler)
	r.Get("/conventions", handlers.ConventionsHandler)
	r.Get("/conventions/new", handlers.ConventionFormHandler)
	r.Get("/conventions/{id}", handlers.ConventionDetailHandler)
	r.Get("/conventions/{id}/edit", handlers.ConventionFormHandler)
	r.Get("/conventions/{id}/sales", handlers.ConventionSalesHandler)
	r.Get("/conventions/{id}/sell", handlers.SellHandler)
	r.Get("/products", handlers.ProductsHandler)
	r.Get("/products/new", handlers.ProductFormHandler)
	r.Get("/products/{id}", handlers.ProductDetailHandler)
	r.Get("/products/{id}/edit", handlers.ProductFormHandler)
	r.Get("/reports", handlers.ReportsHandler)
	r.Get("/settings", handlers.SettingsHandler)

	// API - Conventions
	r.Post("/api/conventions", handlers.CreateConventionHandler)
	r.Put("/api/conventions/{id}", handlers.UpdateConventionHandler)
	r.Delete("/api/conventions/{id}", handlers.DeleteConventionHandler)

	// API - Convention Products
	r.Post("/api/conventions/{id}/products", handlers.AddProductToConventionHandler)
	r.Put("/api/convention-products/{id}", handlers.UpdateConventionProductHandler)
	r.Post("/api/convention-products/{id}/restock", handlers.RestockConventionProductHandler)
	r.Delete("/api/convention-products/{id}", handlers.RemoveProductFromConventionHandler)

	// API - Sales
	r.Post("/api/conventions/{id}/sales", handlers.RecordSaleHandler)
	r.Post("/api/conventions/{id}/sales/bulk", handlers.BulkSyncSalesHandler)
	r.Put("/api/sales/{id}", handlers.UpdateSaleHandler)
	r.Delete("/api/sales/{id}", handlers.UndoSaleHandler)

	// API - Products
	r.Post("/api/products", handlers.CreateProductHandler)
	r.Put("/api/products/{id}", handlers.UpdateProductHandler)
	r.Delete("/api/products/{id}", handlers.DeleteProductHandler)
	r.Post("/api/products/{id}/archive", handlers.ArchiveProductHandler)
	r.Post("/api/products/{id}/unarchive", handlers.UnarchiveProductHandler)

	// MCP SSE endpoint (read-only analytics)
	stockiimcp.MountRoutes(r)

	// API - Data Management
	r.Get("/api/export", handlers.ExportDataHandler)
	r.Post("/api/import", handlers.ImportDataHandler)

	// API - Settings
	r.Post("/api/categories", handlers.CreateCategoryHandler)
	r.Put("/api/categories/{id}", handlers.UpdateCategoryHandler)
	r.Delete("/api/categories/{id}", handlers.DeleteCategoryHandler)
	r.Post("/api/fandoms", handlers.CreateFandomHandler)
	r.Put("/api/fandoms/{id}", handlers.UpdateFandomHandler)
	r.Delete("/api/fandoms/{id}", handlers.DeleteFandomHandler)
	r.Post("/api/convention-series", handlers.CreateConventionSeriesHandler)
	r.Put("/api/convention-series/{id}", handlers.UpdateConventionSeriesHandler)
	r.Delete("/api/convention-series/{id}", handlers.DeleteConventionSeriesHandler)

	log.Printf("Stockii starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
