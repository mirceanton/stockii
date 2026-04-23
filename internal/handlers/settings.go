package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mirceanton/stockii/internal/db"
)

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	categories, _ := db.GetAllCategories()
	fandoms, _ := db.GetAllFandoms()
	series, _ := db.GetAllConventionSeries()

	render(w, "settings.html", map[string]interface{}{
		"Page":       "settings",
		"Categories": categories,
		"Fandoms":    fandoms,
		"Series":     series,
	})
}

// Category CRUD

func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if _, err := db.CreateCategory(name); err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if err := db.UpdateCategory(uint(id), name); err != nil {
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := db.DeleteCategory(uint(id)); err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

// Fandom CRUD

func CreateFandomHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if _, err := db.CreateFandom(name); err != nil {
		http.Error(w, "Failed to create fandom", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func UpdateFandomHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if err := db.UpdateFandom(uint(id), name); err != nil {
		http.Error(w, "Failed to update fandom", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func DeleteFandomHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := db.DeleteFandom(uint(id)); err != nil {
		http.Error(w, "Failed to delete fandom", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

// Convention Series CRUD

func CreateConventionSeriesHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if _, err := db.CreateConventionSeries(name, r.FormValue("notes")); err != nil {
		http.Error(w, "Failed to create series", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func UpdateConventionSeriesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	if err := db.UpdateConventionSeries(uint(id), r.FormValue("name"), r.FormValue("notes")); err != nil {
		http.Error(w, "Failed to update series", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}

func DeleteConventionSeriesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := db.DeleteConventionSeries(uint(id)); err != nil {
		http.Error(w, "Failed to delete series", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/settings")
	w.WriteHeader(http.StatusOK)
}
