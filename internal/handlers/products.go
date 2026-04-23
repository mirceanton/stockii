package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mirceanton/stockii/internal/db"
	"github.com/mirceanton/stockii/internal/models"
)

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	showArchived := r.URL.Query().Get("archived") == "1"

	products, err := db.GetAllProducts()
	if err != nil {
		http.Error(w, "Failed to load products", http.StatusInternalServerError)
		return
	}

	// Filter based on archived toggle
	var filtered []models.Product
	for _, p := range products {
		if showArchived || !p.Archived {
			filtered = append(filtered, p)
		}
	}

	categories, _ := db.GetAllCategories()
	fandoms, _ := db.GetAllFandoms()

	render(w, "products.html", map[string]interface{}{
		"Page":         "products",
		"Products":     filtered,
		"Categories":   categories,
		"Fandoms":      fandoms,
		"ShowArchived": showArchived,
	})
}

func ProductDetailHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	product, err := db.GetProduct(uint(id))
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	history, _ := db.GetProductHistory(uint(id))

	// Calculate averages
	var totalSellThrough float64
	for _, h := range history {
		totalSellThrough += h.SellThroughRate
	}
	avgSellThrough := 0.0
	if len(history) > 0 {
		avgSellThrough = totalSellThrough / float64(len(history))
	}

	render(w, "product_detail.html", map[string]interface{}{
		"Page":           "products",
		"Product":        product,
		"History":        history,
		"AvgSellThrough": avgSellThrough,
	})
}

func ProductFormHandler(w http.ResponseWriter, r *http.Request) {
	categories, _ := db.GetAllCategories()
	fandoms, _ := db.GetAllFandoms()

	var product *models.Product
	idParam := chi.URLParam(r, "id")
	if idParam != "" {
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err == nil {
			product, _ = db.GetProduct(uint(id))
		}
	}

	render(w, "product_form.html", map[string]interface{}{
		"Page":       "products",
		"Product":    product,
		"Categories": categories,
		"Fandoms":    fandoms,
	})
}

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		http.Error(w, "Form too large", http.StatusBadRequest)
		return
	}

	categoryID, _ := strconv.ParseUint(r.FormValue("category_id"), 10, 64)
	productType := r.FormValue("type")
	if productType == "" {
		productType = "original"
	}

	product := &models.Product{
		Name:       r.FormValue("name"),
		CategoryID: uint(categoryID),
		Type:       productType,
	}

	if productType == "fanart" {
		fandomID, err := strconv.ParseUint(r.FormValue("fandom_id"), 10, 64)
		if err == nil {
			fid := uint(fandomID)
			product.FandomID = &fid
		}
	}

	// Handle image upload
	imagePath, err := handleImageUpload(r, "image")
	if err == nil && imagePath != "" {
		product.ImagePath = imagePath
	}

	if err := db.CreateProduct(product); err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/products/%d", product.ID))
	w.WriteHeader(http.StatusOK)
}

func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	product, err := db.GetProduct(uint(id))
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Form too large", http.StatusBadRequest)
		return
	}

	categoryID, _ := strconv.ParseUint(r.FormValue("category_id"), 10, 64)
	product.Name = r.FormValue("name")
	product.CategoryID = uint(categoryID)
	product.Type = r.FormValue("type")

	if product.Type == "fanart" {
		fandomID, err := strconv.ParseUint(r.FormValue("fandom_id"), 10, 64)
		if err == nil {
			fid := uint(fandomID)
			product.FandomID = &fid
		}
	} else {
		product.FandomID = nil
	}

	// Handle new image upload
	imagePath, err := handleImageUpload(r, "image")
	if err == nil && imagePath != "" {
		// Delete old image if exists
		if product.ImagePath != "" {
			os.Remove(filepath.Join(GetDataPath(), "images", product.ImagePath))
		}
		product.ImagePath = imagePath
	}

	if err := db.UpdateProduct(product); err != nil {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/products/%d", product.ID))
	w.WriteHeader(http.StatusOK)
}

func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	product, err := db.GetProduct(uint(id))
	if err == nil && product.ImagePath != "" {
		os.Remove(filepath.Join(GetDataPath(), "images", product.ImagePath))
	}

	if err := db.DeleteProduct(uint(id)); err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/products")
	w.WriteHeader(http.StatusOK)
}

func ArchiveProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := db.ArchiveProduct(uint(id)); err != nil {
		http.Error(w, "Failed to archive product", http.StatusInternalServerError)
		return
	}

	// If called from the product detail/edit page, redirect back there; otherwise go to list.
	if strings.Contains(r.Header.Get("HX-Current-URL"), fmt.Sprintf("/products/%d", id)) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/products/%d", id))
	} else {
		w.Header().Set("HX-Redirect", "/products")
	}
	w.WriteHeader(http.StatusOK)
}

func UnarchiveProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := db.UnarchiveProduct(uint(id)); err != nil {
		http.Error(w, "Failed to unarchive product", http.StatusInternalServerError)
		return
	}

	// If called from the product detail/edit page, redirect back there; otherwise go to list.
	if strings.Contains(r.Header.Get("HX-Current-URL"), fmt.Sprintf("/products/%d", id)) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/products/%d", id))
	} else {
		w.Header().Set("HX-Redirect", "/products")
	}
	w.WriteHeader(http.StatusOK)
}

func handleImageUpload(r *http.Request, fieldName string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err // No file uploaded
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid file type: %s", contentType)
	}

	// Create images directory
	imagesDir := filepath.Join(GetDataPath(), "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		return "", fmt.Errorf("create images dir: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// Save file
	dst, err := os.Create(filepath.Join(imagesDir, filename))
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	return filename, nil
}
