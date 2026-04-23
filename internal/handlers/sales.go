package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mirceanton/stockii/internal/db"
	"github.com/mirceanton/stockii/internal/models"
)

func SellHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	conv, err := db.GetConvention(uint(id))
	if err != nil {
		http.Error(w, "Convention not found", http.StatusNotFound)
		return
	}

	views, err := db.GetConventionProductViews(uint(id))
	if err != nil {
		http.Error(w, "Failed to load products", http.StatusInternalServerError)
		return
	}

	// Group by category
	type categoryGroup struct {
		Name     string
		Products []models.ConventionProductView
	}
	groups := make(map[string]*categoryGroup)
	var groupOrder []string
	for _, v := range views {
		catName := v.Product.Category.Name
		if _, ok := groups[catName]; !ok {
			groups[catName] = &categoryGroup{Name: catName}
			groupOrder = append(groupOrder, catName)
		}
		groups[catName].Products = append(groups[catName].Products, v)
	}

	var orderedGroups []categoryGroup
	for _, name := range groupOrder {
		orderedGroups = append(orderedGroups, *groups[name])
	}

	renderStandalone(w, "sell.html", map[string]interface{}{
		"Convention": conv,
		"Groups":     orderedGroups,
		"Products":   views,
	})
}

func RecordSaleHandler(w http.ResponseWriter, r *http.Request) {
	// Support both URL-encoded and multipart form data
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}
	}

	cpID, _ := strconv.ParseUint(r.FormValue("convention_product_id"), 10, 64)
	quantity, _ := strconv.Atoi(r.FormValue("quantity"))
	if quantity <= 0 {
		quantity = 1
	}

	sale := &models.Sale{
		ConventionProductID: uint(cpID),
		Quantity:            quantity,
	}

	if err := db.RecordSale(sale); err != nil {
		http.Error(w, "Failed to record sale", http.StatusInternalServerError)
		return
	}

	// Return JSON with sale ID for undo support
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sale_id":  sale.ID,
		"cp_id":    cpID,
		"quantity": quantity,
	})
}

func UndoSaleHandler(w http.ResponseWriter, r *http.Request) {
	saleID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := db.DeleteSale(uint(saleID)); err != nil {
		http.Error(w, "Failed to undo sale", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// BulkSyncSalesHandler handles offline sale sync
func BulkSyncSalesHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Sales []struct {
			ConventionProductID uint   `json:"convention_product_id"`
			Quantity            int    `json:"quantity"`
			Timestamp           string `json:"timestamp"`
		} `json:"sales"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var sales []models.Sale
	for _, s := range payload.Sales {
		createdAt := time.Now()
		if s.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, s.Timestamp); err == nil {
				createdAt = t
			}
		}
		sales = append(sales, models.Sale{
			ConventionProductID: s.ConventionProductID,
			Quantity:            s.Quantity,
			CreatedAt:           createdAt,
		})
	}

	if err := db.RecordSalesBulk(sales); err != nil {
		http.Error(w, "Failed to sync sales", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"synced": len(sales),
	})
}
