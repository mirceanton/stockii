package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mirceanton/stockii/internal/db"
	"github.com/mirceanton/stockii/internal/models"
)

func ConventionsHandler(w http.ResponseWriter, r *http.Request) {
	conventions, err := db.GetAllConventions()
	if err != nil {
		http.Error(w, "Failed to load conventions", http.StatusInternalServerError)
		return
	}

	// Build P&L data for each
	type convRow struct {
		models.Convention
		Revenue float64
		Profit  float64
	}
	var rows []convRow
	for _, c := range conventions {
		pnl, _ := db.GetConventionPnL(c.ID)
		row := convRow{Convention: c}
		if pnl != nil {
			row.Revenue = pnl.TotalRevenue
			row.Profit = pnl.Profit
		}
		rows = append(rows, row)
	}

	render(w, "conventions.html", map[string]interface{}{
		"Page":        "conventions",
		"Conventions": rows,
	})
}

func ConventionDetailHandler(w http.ResponseWriter, r *http.Request) {
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

	pnl, _ := db.GetConventionPnL(uint(id))

	// Get products not yet added to this convention (for the add form)
	availableProducts, _ := db.GetProductsNotInConvention(uint(id))

	render(w, "convention_detail.html", map[string]interface{}{
		"Page":              "conventions",
		"Convention":        conv,
		"Products":          views,
		"PnL":               pnl,
		"AvailableProducts": availableProducts,
	})
}

func ConventionFormHandler(w http.ResponseWriter, r *http.Request) {
	series, _ := db.GetAllConventionSeries()

	// Check if editing
	idParam := chi.URLParam(r, "id")
	var conv *models.Convention
	if idParam != "" {
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err == nil {
			conv, _ = db.GetConvention(uint(id))
		}
	}

	render(w, "convention_form.html", map[string]interface{}{
		"Page":       "conventions",
		"Series":     series,
		"Convention": conv,
	})
}

func CreateConventionHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	seriesID, _ := strconv.ParseUint(r.FormValue("series_id"), 10, 64)
	dateStart, _ := time.Parse("2006-01-02", r.FormValue("date_start"))
	dateEnd, _ := time.Parse("2006-01-02", r.FormValue("date_end"))
	tableCost, _ := strconv.ParseFloat(r.FormValue("table_cost"), 64)
	prepCost, _ := strconv.ParseFloat(r.FormValue("prep_cost"), 64)
	convType := r.FormValue("type")
	if convType == "" {
		convType = "mixed"
	}

	// Auto-generate name if not provided
	name := r.FormValue("name")
	if name == "" {
		// Get series name for auto-generation
		series, _ := db.GetAllConventionSeries()
		for _, s := range series {
			if s.ID == uint(seriesID) {
				name = fmt.Sprintf("%s - %s %d", s.Name, dateStart.Month().String()[:3], dateStart.Year())
				break
			}
		}
	}

	conv := &models.Convention{
		ConventionSeriesID: uint(seriesID),
		Name:               name,
		Location:           r.FormValue("location"),
		DateStart:          dateStart,
		DateEnd:            dateEnd,
		Type:               convType,
		TableCost:          tableCost,
		PrepCost:           prepCost,
		Notes:              r.FormValue("notes"),
	}

	if err := db.CreateConvention(conv); err != nil {
		http.Error(w, "Failed to create convention", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/conventions/%d", conv.ID))
	w.WriteHeader(http.StatusOK)
}

func UpdateConventionHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	seriesID, _ := strconv.ParseUint(r.FormValue("series_id"), 10, 64)
	dateStart, _ := time.Parse("2006-01-02", r.FormValue("date_start"))
	dateEnd, _ := time.Parse("2006-01-02", r.FormValue("date_end"))
	tableCost, _ := strconv.ParseFloat(r.FormValue("table_cost"), 64)
	prepCost, _ := strconv.ParseFloat(r.FormValue("prep_cost"), 64)

	conv.ConventionSeriesID = uint(seriesID)
	conv.Name = r.FormValue("name")
	conv.Location = r.FormValue("location")
	conv.DateStart = dateStart
	conv.DateEnd = dateEnd
	conv.Type = r.FormValue("type")
	conv.TableCost = tableCost
	conv.PrepCost = prepCost
	conv.Notes = r.FormValue("notes")

	if err := db.UpdateConvention(conv); err != nil {
		http.Error(w, "Failed to update convention", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/conventions/%d", conv.ID))
	w.WriteHeader(http.StatusOK)
}

func DeleteConventionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := db.DeleteConvention(uint(id)); err != nil {
		http.Error(w, "Failed to delete convention", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/conventions")
	w.WriteHeader(http.StatusOK)
}

func AddProductToConventionHandler(w http.ResponseWriter, r *http.Request) {
	convID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid convention ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	productID, _ := strconv.ParseUint(r.FormValue("product_id"), 10, 64)
	qtyBrought, _ := strconv.Atoi(r.FormValue("qty_brought"))
	salePrice, _ := strconv.ParseFloat(r.FormValue("sale_price"), 64)

	cp := &models.ConventionProduct{
		ConventionID: uint(convID),
		ProductID:    uint(productID),
		QtyBrought:   qtyBrought,
		SalePrice:    salePrice,
	}

	if err := db.AddProductToConvention(cp); err != nil {
		http.Error(w, "Failed to add product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/conventions/%d", convID))
	w.WriteHeader(http.StatusOK)
}

func UpdateConventionProductHandler(w http.ResponseWriter, r *http.Request) {
	cpID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	cp, err := db.GetConventionProduct(uint(cpID))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	qtyBrought, _ := strconv.Atoi(r.FormValue("qty_brought"))
	salePrice, _ := strconv.ParseFloat(r.FormValue("sale_price"), 64)

	cp.QtyBrought = qtyBrought
	cp.SalePrice = salePrice

	if err := db.UpdateConventionProduct(cp); err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/conventions/%d", cp.ConventionID))
	w.WriteHeader(http.StatusOK)
}

func RestockConventionProductHandler(w http.ResponseWriter, r *http.Request) {
	cpID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	cp, err := db.GetConventionProduct(uint(cpID))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	qtyAdd, _ := strconv.Atoi(r.FormValue("qty_add"))
	if qtyAdd <= 0 {
		http.Error(w, "Quantity must be positive", http.StatusBadRequest)
		return
	}

	if err := db.RestockConventionProduct(uint(cpID), qtyAdd); err != nil {
		http.Error(w, "Failed to restock", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/conventions/%d", cp.ConventionID))
	w.WriteHeader(http.StatusOK)
}

func RemoveProductFromConventionHandler(w http.ResponseWriter, r *http.Request) {
	cpID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	cp, err := db.GetConventionProduct(uint(cpID))
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	convID := cp.ConventionID
	if err := db.RemoveProductFromConvention(uint(cpID)); err != nil {
		http.Error(w, "Failed to remove", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/conventions/%d", convID))
	w.WriteHeader(http.StatusOK)
}
