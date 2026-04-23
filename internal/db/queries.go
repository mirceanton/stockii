package db

import (
	"github.com/mirceanton/stockii/internal/models"
)

// --- Categories ---

func GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := DB.Order("name ASC").Find(&categories).Error
	return categories, err
}

func CreateCategory(name string) (*models.Category, error) {
	cat := models.Category{Name: name}
	err := DB.Create(&cat).Error
	return &cat, err
}

func DeleteCategory(id uint) error {
	return DB.Delete(&models.Category{}, id).Error
}

// --- Fandoms ---

func GetAllFandoms() ([]models.Fandom, error) {
	var fandoms []models.Fandom
	err := DB.Order("name ASC").Find(&fandoms).Error
	return fandoms, err
}

func CreateFandom(name string) (*models.Fandom, error) {
	f := models.Fandom{Name: name}
	err := DB.Create(&f).Error
	return &f, err
}

func DeleteFandom(id uint) error {
	return DB.Delete(&models.Fandom{}, id).Error
}

// --- Convention Series ---

func GetAllConventionSeries() ([]models.ConventionSeries, error) {
	var series []models.ConventionSeries
	err := DB.Order("name ASC").Find(&series).Error
	return series, err
}

func CreateConventionSeries(name, notes string) (*models.ConventionSeries, error) {
	cs := models.ConventionSeries{Name: name, Notes: notes}
	err := DB.Create(&cs).Error
	return &cs, err
}

func UpdateConventionSeries(id uint, name, notes string) error {
	return DB.Model(&models.ConventionSeries{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":  name,
		"notes": notes,
	}).Error
}

func DeleteConventionSeries(id uint) error {
	return DB.Delete(&models.ConventionSeries{}, id).Error
}

// --- Products ---

func GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	err := DB.Preload("Category").Preload("Fandom").Order("name ASC").Find(&products).Error
	return products, err
}

// GetActiveProducts returns only non-archived products
func GetActiveProducts() ([]models.Product, error) {
	var products []models.Product
	err := DB.Preload("Category").Preload("Fandom").Where("archived = ?", false).Order("name ASC").Find(&products).Error
	return products, err
}

func GetProduct(id uint) (*models.Product, error) {
	var product models.Product
	err := DB.Preload("Category").Preload("Fandom").First(&product, id).Error
	return &product, err
}

func CreateProduct(p *models.Product) error {
	return DB.Create(p).Error
}

func UpdateProduct(p *models.Product) error {
	return DB.Save(p).Error
}

func DeleteProduct(id uint) error {
	return DB.Delete(&models.Product{}, id).Error
}

// ArchiveProduct sets a product's archived flag to true
func ArchiveProduct(id uint) error {
	return DB.Model(&models.Product{}).Where("id = ?", id).Update("archived", true).Error
}

// UnarchiveProduct sets a product's archived flag to false
func UnarchiveProduct(id uint) error {
	return DB.Model(&models.Product{}).Where("id = ?", id).Update("archived", false).Error
}

// --- Conventions ---

func GetAllConventions() ([]models.Convention, error) {
	var conventions []models.Convention
	err := DB.Preload("ConventionSeries").Order("date_start DESC").Find(&conventions).Error
	return conventions, err
}

func GetConvention(id uint) (*models.Convention, error) {
	var conv models.Convention
	err := DB.Preload("ConventionSeries").First(&conv, id).Error
	return &conv, err
}

func CreateConvention(c *models.Convention) error {
	return DB.Create(c).Error
}

func UpdateConvention(c *models.Convention) error {
	return DB.Save(c).Error
}

func DeleteConvention(id uint) error {
	// Delete associated convention products and sales first
	DB.Where("convention_product_id IN (?)",
		DB.Model(&models.ConventionProduct{}).Select("id").Where("convention_id = ?", id),
	).Delete(&models.Sale{})
	DB.Where("convention_id = ?", id).Delete(&models.ConventionProduct{})
	return DB.Delete(&models.Convention{}, id).Error
}

// --- Convention Products ---

func GetConventionProducts(conventionID uint) ([]models.ConventionProduct, error) {
	var cps []models.ConventionProduct
	err := DB.Preload("Product").Preload("Product.Category").Preload("Product.Fandom").
		Where("convention_id = ?", conventionID).Find(&cps).Error
	return cps, err
}

func GetConventionProduct(id uint) (*models.ConventionProduct, error) {
	var cp models.ConventionProduct
	err := DB.Preload("Product").Preload("Product.Category").First(&cp, id).Error
	return &cp, err
}

func AddProductToConvention(cp *models.ConventionProduct) error {
	return DB.Create(cp).Error
}

func UpdateConventionProduct(cp *models.ConventionProduct) error {
	return DB.Save(cp).Error
}

func RemoveProductFromConvention(id uint) error {
	DB.Where("convention_product_id = ?", id).Delete(&models.Sale{})
	return DB.Delete(&models.ConventionProduct{}, id).Error
}

// --- Sales ---

func RecordSale(sale *models.Sale) error {
	return DB.Create(sale).Error
}

func RecordSalesBulk(sales []models.Sale) error {
	if len(sales) == 0 {
		return nil
	}
	return DB.CreateInBatches(sales, 100).Error
}

func DeleteSale(id uint) error {
	return DB.Delete(&models.Sale{}, id).Error
}

func GetSalesForConventionProduct(cpID uint) ([]models.Sale, error) {
	var sales []models.Sale
	err := DB.Where("convention_product_id = ?", cpID).Order("created_at DESC").Find(&sales).Error
	return sales, err
}

func GetTotalSold(cpID uint) int {
	var total int64
	DB.Model(&models.Sale{}).Where("convention_product_id = ?", cpID).
		Select("COALESCE(SUM(quantity), 0)").Scan(&total)
	return int(total)
}

// --- Computed Views ---

func GetConventionProductViews(conventionID uint) ([]models.ConventionProductView, error) {
	cps, err := GetConventionProducts(conventionID)
	if err != nil {
		return nil, err
	}

	views := make([]models.ConventionProductView, len(cps))
	for i, cp := range cps {
		qtySold := GetTotalSold(cp.ID)
		revenue := float64(qtySold) * cp.SalePrice
		var sellThrough float64
		if cp.QtyBrought > 0 {
			sellThrough = float64(qtySold) / float64(cp.QtyBrought) * 100
		}
		stockLevel := "plenty"
		if cp.QtyBrought > 0 {
			pct := float64(qtySold) / float64(cp.QtyBrought)
			switch {
			case pct >= 1.0:
				stockLevel = "sold_out"
			case pct >= 0.9:
				stockLevel = "critical"
			case pct >= 0.7:
				stockLevel = "low"
			}
		}
		views[i] = models.ConventionProductView{
			ConventionProduct: cp,
			QtySold:           qtySold,
			Revenue:           revenue,
			SellThroughRate:   sellThrough,
			StockLevel:        stockLevel,
		}
	}
	return views, nil
}

func GetConventionPnL(conventionID uint) (*models.ConventionPnL, error) {
	conv, err := GetConvention(conventionID)
	if err != nil {
		return nil, err
	}

	views, err := GetConventionProductViews(conventionID)
	if err != nil {
		return nil, err
	}

	pnl := &models.ConventionPnL{Convention: *conv}
	for _, v := range views {
		pnl.TotalRevenue += v.Revenue
		pnl.TotalSold += v.QtySold
		pnl.TotalBrought += v.QtyBrought
		pnl.ProductCount++
	}
	pnl.TotalCost = conv.TableCost + conv.PrepCost
	pnl.Profit = pnl.TotalRevenue - pnl.TotalCost
	if pnl.TotalCost > 0 {
		pnl.ROI = pnl.Profit / pnl.TotalCost * 100
	}
	return pnl, nil
}

func GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	DB.Model(&models.Product{}).Count(new(int64))
	var productCount, convCount int64
	DB.Model(&models.Product{}).Count(&productCount)
	DB.Model(&models.Convention{}).Count(&convCount)
	stats.TotalProducts = int(productCount)
	stats.TotalConventions = int(convCount)

	// Calculate totals across all conventions
	var conventions []models.Convention
	DB.Find(&conventions)

	var totalBrought, totalSold int
	for _, conv := range conventions {
		pnl, err := GetConventionPnL(conv.ID)
		if err != nil {
			continue
		}
		stats.TotalRevenue += pnl.TotalRevenue
		stats.TotalProfit += pnl.Profit
		totalBrought += pnl.TotalBrought
		totalSold += pnl.TotalSold
	}

	if totalBrought > 0 {
		stats.AvgSellThroughRate = float64(totalSold) / float64(totalBrought) * 100
	}

	return stats, nil
}

func GetProductHistory(productID uint) ([]models.ProductHistory, error) {
	var cps []models.ConventionProduct
	err := DB.Preload("Convention").Preload("Product").Preload("Product.Category").Preload("Product.Fandom").
		Where("product_id = ?", productID).Find(&cps).Error
	if err != nil {
		return nil, err
	}

	history := make([]models.ProductHistory, len(cps))
	for i, cp := range cps {
		qtySold := GetTotalSold(cp.ID)
		revenue := float64(qtySold) * cp.SalePrice
		var sellThrough float64
		if cp.QtyBrought > 0 {
			sellThrough = float64(qtySold) / float64(cp.QtyBrought) * 100
		}
		history[i] = models.ProductHistory{
			Product:         cp.Product,
			ConventionName:  cp.Convention.Name,
			ConventionDate:  cp.Convention.DateStart,
			QtyBrought:      cp.QtyBrought,
			QtySold:         qtySold,
			SalePrice:       cp.SalePrice,
			Revenue:         revenue,
			SellThroughRate: sellThrough,
		}
	}
	return history, nil
}

// GetProductsNotInConvention returns products that haven't been added to a convention yet
func GetProductsNotInConvention(conventionID uint) ([]models.Product, error) {
	var products []models.Product
	subQuery := DB.Model(&models.ConventionProduct{}).Select("product_id").Where("convention_id = ?", conventionID)
	err := DB.Preload("Category").Preload("Fandom").
		Where("id NOT IN (?)", subQuery).Where("archived = ?", false).Order("name ASC").Find(&products).Error
	return products, err
}

// GetUpcomingConventions returns conventions with date_start in the future
func GetUpcomingConventions() ([]models.Convention, error) {
	var conventions []models.Convention
	err := DB.Preload("ConventionSeries").Where("date_start >= date('now')").
		Order("date_start ASC").Limit(5).Find(&conventions).Error
	return conventions, err
}

// GetRecentConventions returns past conventions ordered by most recent
func GetRecentConventions() ([]models.Convention, error) {
	var conventions []models.Convention
	err := DB.Preload("ConventionSeries").Where("date_start < date('now')").
		Order("date_start DESC").Limit(5).Find(&conventions).Error
	return conventions, err
}

// GetAllConventionPnLs returns P&L data for all conventions
func GetAllConventionPnLs() ([]models.ConventionPnL, error) {
	conventions, err := GetAllConventions()
	if err != nil {
		return nil, err
	}
	pnls := make([]models.ConventionPnL, 0, len(conventions))
	for _, conv := range conventions {
		pnl, err := GetConventionPnL(conv.ID)
		if err != nil {
			continue
		}
		pnls = append(pnls, *pnl)
	}
	return pnls, nil
}

// GetCategorySalesReport returns aggregated sell-through by category across all conventions
func GetCategorySalesReport() ([]map[string]interface{}, error) {
	var results []struct {
		CategoryName string
		TotalBrought int
		TotalSold    int
		TotalRevenue float64
	}

	// Get all convention products grouped by category
	categories, err := GetAllCategories()
	if err != nil {
		return nil, err
	}

	report := make([]map[string]interface{}, 0)
	for _, cat := range categories {
		var cps []models.ConventionProduct
		DB.Joins("JOIN products ON products.id = convention_products.product_id").
			Where("products.category_id = ?", cat.ID).Find(&cps)

		if len(cps) == 0 {
			continue
		}

		totalBrought := 0
		totalSold := 0
		totalRevenue := 0.0
		for _, cp := range cps {
			sold := GetTotalSold(cp.ID)
			totalBrought += cp.QtyBrought
			totalSold += sold
			totalRevenue += float64(sold) * cp.SalePrice
		}

		sellThrough := 0.0
		if totalBrought > 0 {
			sellThrough = float64(totalSold) / float64(totalBrought) * 100
		}

		report = append(report, map[string]interface{}{
			"category":     cat.Name,
			"totalBrought": totalBrought,
			"totalSold":    totalSold,
			"totalRevenue": totalRevenue,
			"sellThrough":  sellThrough,
		})
	}

	_ = results
	return report, nil
}

// GetFandomSalesReport returns aggregated performance by fandom
func GetFandomSalesReport() ([]map[string]interface{}, error) {
	fandoms, err := GetAllFandoms()
	if err != nil {
		return nil, err
	}

	report := make([]map[string]interface{}, 0)

	// Also include "Original" as a pseudo-fandom
	var originalCPs []models.ConventionProduct
	DB.Joins("JOIN products ON products.id = convention_products.product_id").
		Where("products.type = ?", "original").Find(&originalCPs)
	if len(originalCPs) > 0 {
		totalBrought, totalSold := 0, 0
		totalRevenue := 0.0
		for _, cp := range originalCPs {
			sold := GetTotalSold(cp.ID)
			totalBrought += cp.QtyBrought
			totalSold += sold
			totalRevenue += float64(sold) * cp.SalePrice
		}
		sellThrough := 0.0
		if totalBrought > 0 {
			sellThrough = float64(totalSold) / float64(totalBrought) * 100
		}
		report = append(report, map[string]interface{}{
			"fandom":       "Original",
			"totalBrought": totalBrought,
			"totalSold":    totalSold,
			"totalRevenue": totalRevenue,
			"sellThrough":  sellThrough,
		})
	}

	for _, fandom := range fandoms {
		var cps []models.ConventionProduct
		DB.Joins("JOIN products ON products.id = convention_products.product_id").
			Where("products.fandom_id = ?", fandom.ID).Find(&cps)
		if len(cps) == 0 {
			continue
		}
		totalBrought, totalSold := 0, 0
		totalRevenue := 0.0
		for _, cp := range cps {
			sold := GetTotalSold(cp.ID)
			totalBrought += cp.QtyBrought
			totalSold += sold
			totalRevenue += float64(sold) * cp.SalePrice
		}
		sellThrough := 0.0
		if totalBrought > 0 {
			sellThrough = float64(totalSold) / float64(totalBrought) * 100
		}
		report = append(report, map[string]interface{}{
			"fandom":       fandom.Name,
			"totalBrought": totalBrought,
			"totalSold":    totalSold,
			"totalRevenue": totalRevenue,
			"sellThrough":  sellThrough,
		})
	}

	return report, nil
}
