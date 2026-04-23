package handlers

import (
	"net/http"

	"github.com/mirceanton/stockii/internal/db"
)

func ReportsHandler(w http.ResponseWriter, r *http.Request) {
	pnls, _ := db.GetAllConventionPnLs()
	categoryReport, _ := db.GetCategorySalesReport()
	fandomReport, _ := db.GetFandomSalesReport()

	render(w, "reports.html", map[string]interface{}{
		"Page":           "reports",
		"PnLs":           pnls,
		"CategoryReport": categoryReport,
		"FandomReport":   fandomReport,
	})
}
