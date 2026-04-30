package models

import "time"

type ExportCategory struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ExportFandom struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ExportConventionSeries struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type ExportProduct struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	CategoryID uint      `json:"category_id"`
	Type       string    `json:"type"`
	FandomID   *uint     `json:"fandom_id"`
	ImagePath  string    `json:"image_path"`
	Archived   bool      `json:"archived"`
	CreatedAt  time.Time `json:"created_at"`
}

type ExportConvention struct {
	ID                 uint      `json:"id"`
	ConventionSeriesID uint      `json:"convention_series_id"`
	Name               string    `json:"name"`
	Location           string    `json:"location"`
	DateStart          time.Time `json:"date_start"`
	DateEnd            time.Time `json:"date_end"`
	Type               string    `json:"type"`
	TableCost          float64   `json:"table_cost"`
	PrepCost           float64   `json:"prep_cost"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
}

type ExportConventionProduct struct {
	ID           uint    `json:"id"`
	ConventionID uint    `json:"convention_id"`
	ProductID    uint    `json:"product_id"`
	QtyBrought   int     `json:"qty_brought"`
	SalePrice    float64 `json:"sale_price"`
}

type ExportSale struct {
	ID                  uint      `json:"id"`
	ConventionProductID uint      `json:"convention_product_id"`
	Quantity            int       `json:"quantity"`
	CreatedAt           time.Time `json:"created_at"`
}

type ExportData struct {
	ExportedAt         time.Time                 `json:"exported_at"`
	Categories         []ExportCategory          `json:"categories"`
	Fandoms            []ExportFandom            `json:"fandoms"`
	ConventionSeries   []ExportConventionSeries  `json:"convention_series"`
	Products           []ExportProduct           `json:"products"`
	Conventions        []ExportConvention        `json:"conventions"`
	ConventionProducts []ExportConventionProduct `json:"convention_products"`
	Sales              []ExportSale              `json:"sales"`
}
