package models

import "time"

type Category struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Fandom struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Name       string    `gorm:"not null" json:"name"`
	CategoryID uint      `gorm:"not null" json:"category_id"`
	Category   Category  `json:"category"`
	Type       string    `gorm:"not null;default:'original'" json:"type"` // "fanart" or "original"
	FandomID   *uint     `json:"fandom_id"`                               // nullable for originals
	Fandom     *Fandom   `json:"fandom"`
	ImagePath  string    `json:"image_path"` // relative path under data/images/
	Archived   bool      `gorm:"not null;default:false" json:"archived"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConventionSeries struct {
	ID          uint         `gorm:"primarykey" json:"id"`
	Name        string       `gorm:"uniqueIndex;not null" json:"name"`
	Notes       string       `json:"notes"`
	CreatedAt   time.Time    `json:"created_at"`
	Conventions []Convention `json:"conventions,omitempty"`
}

type Convention struct {
	ID                 uint                `gorm:"primarykey" json:"id"`
	ConventionSeriesID uint                `gorm:"not null" json:"convention_series_id"`
	ConventionSeries   ConventionSeries    `json:"convention_series"`
	Name               string              `gorm:"not null" json:"name"`
	Location           string              `json:"location"`
	DateStart          time.Time           `gorm:"not null" json:"date_start"`
	DateEnd            time.Time           `json:"date_end"`
	Type               string              `gorm:"not null;default:'mixed'" json:"type"` // "fanart", "original", "mixed"
	TableCost          float64             `json:"table_cost"`
	PrepCost           float64             `json:"prep_cost"`
	Notes              string              `json:"notes"`
	CreatedAt          time.Time           `json:"created_at"`
	ConventionProducts []ConventionProduct `json:"convention_products,omitempty"`
}

type ConventionProduct struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	ConventionID uint       `gorm:"not null;uniqueIndex:idx_conv_prod" json:"convention_id"`
	Convention   Convention `json:"-"`
	ProductID    uint       `gorm:"not null;uniqueIndex:idx_conv_prod" json:"product_id"`
	Product      Product    `json:"product"`
	QtyBrought   int        `gorm:"not null;default:0" json:"qty_brought"`
	SalePrice    float64    `gorm:"not null" json:"sale_price"`
	Sales        []Sale     `json:"sales,omitempty"`
}

type Sale struct {
	ID                  uint              `gorm:"primarykey" json:"id"`
	ConventionProductID uint              `gorm:"not null;index" json:"convention_product_id"`
	ConventionProduct   ConventionProduct `json:"-"`
	Quantity            int               `gorm:"not null;default:1" json:"quantity"`
	CreatedAt           time.Time         `json:"created_at"`
}

// Computed view models (not DB tables)

type ConventionProductView struct {
	ConventionProduct
	QtySold         int     `json:"qty_sold"`
	Revenue         float64 `json:"revenue"`
	SellThroughRate float64 `json:"sell_through_rate"`
	StockLevel      string  `json:"stock_level"` // "plenty", "low", "critical", "sold_out"
}

type ConventionPnL struct {
	Convention   Convention `json:"convention"`
	TotalRevenue float64    `json:"total_revenue"`
	TotalCost    float64    `json:"total_cost"`
	Profit       float64    `json:"profit"`
	ROI          float64    `json:"roi"`
	ProductCount int        `json:"product_count"`
	TotalSold    int        `json:"total_sold"`
	TotalBrought int        `json:"total_brought"`
}

type ProductHistory struct {
	Product         Product   `json:"product"`
	ConventionName  string    `json:"convention_name"`
	ConventionDate  time.Time `json:"convention_date"`
	QtyBrought      int       `json:"qty_brought"`
	QtySold         int       `json:"qty_sold"`
	SalePrice       float64   `json:"sale_price"`
	Revenue         float64   `json:"revenue"`
	SellThroughRate float64   `json:"sell_through_rate"`
}

type DashboardStats struct {
	TotalProducts      int     `json:"total_products"`
	TotalConventions   int     `json:"total_conventions"`
	TotalRevenue       float64 `json:"total_revenue"`
	TotalProfit        float64 `json:"total_profit"`
	AvgSellThroughRate float64 `json:"avg_sell_through_rate"`
}
