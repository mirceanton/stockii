package handlers

import (
	"net/http"

	"github.com/mirceanton/stockii/internal/db"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := db.GetDashboardStats()
	if err != nil {
		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
		return
	}

	upcoming, _ := db.GetUpcomingConventions()
	recent, _ := db.GetRecentConventions()

	// Get P&L for recent conventions
	type recentPnL struct {
		Name   string
		Profit float64
		ROI    float64
	}
	var recentPnLs []recentPnL
	for _, conv := range recent {
		pnl, err := db.GetConventionPnL(conv.ID)
		if err != nil {
			continue
		}
		recentPnLs = append(recentPnLs, recentPnL{
			Name:   conv.Name,
			Profit: pnl.Profit,
			ROI:    pnl.ROI,
		})
	}

	render(w, "dashboard.html", map[string]interface{}{
		"Page":       "dashboard",
		"Stats":      stats,
		"Upcoming":   upcoming,
		"Recent":     recent,
		"RecentPnLs": recentPnLs,
	})
}
