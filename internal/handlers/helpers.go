package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

var pageTemplates map[string]*template.Template

var funcMap = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	"title": func(s string) string {
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	},
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mul": func(a, b float64) float64 { return a * b },
	"formatMoney": func(f float64) string {
		return fmt.Sprintf("%.2f", f)
	},
	"formatPct": func(f float64) string {
		return fmt.Sprintf("%.1f", f)
	},
	"stockClass": func(level string) string {
		switch level {
		case "sold_out":
			return "text-secondary"
		case "critical":
			return "text-danger"
		case "low":
			return "text-warning"
		default:
			return "text-success"
		}
	},
	"stockBgClass": func(level string) string {
		switch level {
		case "sold_out":
			return "bg-secondary"
		case "critical":
			return "bg-danger"
		case "low":
			return "bg-warning"
		default:
			return "bg-success"
		}
	},
	"conventionTypeLabel": func(t string) string {
		switch t {
		case "fanart":
			return "Fanart"
		case "original":
			return "Original"
		case "mixed":
			return "Mixed"
		default:
			return t
		}
	},
	"productTypeLabel": func(t string) string {
		switch t {
		case "fanart":
			return "Fanart"
		case "original":
			return "Original"
		default:
			return t
		}
	},
	"seq": func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i
		}
		return s
	},
	"derefUint": func(p *uint) uint {
		if p != nil {
			return *p
		}
		return 0
	},
	"hasImage": func(path string) bool {
		return path != ""
	},
}

var dataPath string

func SetDataPath(path string) {
	dataPath = path
}

func GetDataPath() string {
	return dataPath
}

func InitTemplates(templatesPath string) error {
	pageTemplates = make(map[string]*template.Template)

	pages := []string{
		"dashboard.html",
		"conventions.html",
		"convention_detail.html",
		"convention_sales.html",
		"convention_form.html",
		"products.html",
		"product_detail.html",
		"product_form.html",
		"sell.html",
		"reports.html",
		"settings.html",
	}

	for _, page := range pages {
		t, err := template.New("").Funcs(funcMap).ParseFiles(filepath.Join(templatesPath, "base.html"), filepath.Join(templatesPath, page))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", page, err)
		}
		pageTemplates[page] = t
	}

	// Parse sell.html independently since it has its own layout
	sellT, err := template.New("").Funcs(funcMap).ParseFiles(filepath.Join(templatesPath, "sell.html"))
	if err != nil {
		return fmt.Errorf("parse sell template: %w", err)
	}
	pageTemplates["sell.html"] = sellT

	return nil
}

func render(w http.ResponseWriter, name string, data interface{}) {
	t, ok := pageTemplates[name]
	if !ok {
		log.Printf("template %s not found", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("render template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func renderStandalone(w http.ResponseWriter, name string, data interface{}) {
	t, ok := pageTemplates[name]
	if !ok {
		log.Printf("template %s not found", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "sell_page", data); err != nil {
		log.Printf("render standalone template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
