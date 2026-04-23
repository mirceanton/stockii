package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mirceanton/stockii/internal/db"
)

func jsonResult(data interface{}, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	result, jsonErr := mcp.NewToolResultJSON(data)
	if jsonErr != nil {
		return nil, fmt.Errorf("json marshal failed: %w", jsonErr)
	}
	return result, nil
}

func toolListProducts() server.ServerTool {
	return noArgTool("list_products",
		"List all products with their category and fandom information.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetAllProducts())
		},
	)
}

func toolListConventions() server.ServerTool {
	return noArgTool("list_conventions",
		"List all conventions ordered by date (most recent first), with their series info.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetAllConventions())
		},
	)
}

func toolListCategories() server.ServerTool {
	return noArgTool("list_categories",
		"List all product categories.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetAllCategories())
		},
	)
}

func toolListFandoms() server.ServerTool {
	return noArgTool("list_fandoms",
		"List all fandoms.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetAllFandoms())
		},
	)
}

func toolListConventionSeries() server.ServerTool {
	return noArgTool("list_convention_series",
		"List all convention series.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetAllConventionSeries())
		},
	)
}

func toolGetConventionPnL() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_convention_pnl",
			mcp.WithDescription("Get profit & loss data for a specific convention, including revenue, costs, profit, ROI, and total items sold/brought."),
			mcp.WithNumber("convention_id", mcp.Required(), mcp.Description("The convention ID")),
		),
		Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("convention_id", 0)
			if id == 0 {
				return mcp.NewToolResultText("error: convention_id is required"), nil
			}
			return jsonResult(db.GetConventionPnL(uint(id)))
		},
	}
}

func toolGetAllConventionPnLs() server.ServerTool {
	return noArgTool("get_all_convention_pnls",
		"Get profit & loss data for all conventions. Useful for comparing performance across events.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetAllConventionPnLs())
		},
	)
}

func toolGetConventionInventory() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_convention_inventory",
			mcp.WithDescription("Get detailed inventory for a convention: each product with qty brought, qty sold, revenue, sell-through rate, and stock level."),
			mcp.WithNumber("convention_id", mcp.Required(), mcp.Description("The convention ID")),
		),
		Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("convention_id", 0)
			if id == 0 {
				return mcp.NewToolResultText("error: convention_id is required"), nil
			}
			return jsonResult(db.GetConventionProductViews(uint(id)))
		},
	}
}

func toolGetProductHistory() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_product_history",
			mcp.WithDescription("Get sales history for a product across all conventions it has been brought to, including qty brought, qty sold, revenue, and sell-through rate per convention."),
			mcp.WithNumber("product_id", mcp.Required(), mcp.Description("The product ID")),
		),
		Handler: func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("product_id", 0)
			if id == 0 {
				return mcp.NewToolResultText("error: product_id is required"), nil
			}
			return jsonResult(db.GetProductHistory(uint(id)))
		},
	}
}

func toolGetCategoryReport() server.ServerTool {
	return noArgTool("get_category_report",
		"Get aggregated sales performance by product category across all conventions: total brought, sold, revenue, and sell-through rate per category.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetCategorySalesReport())
		},
	)
}

func toolGetFandomReport() server.ServerTool {
	return noArgTool("get_fandom_report",
		"Get aggregated sales performance by fandom across all conventions: total brought, sold, revenue, and sell-through rate per fandom. Includes 'Original' as a pseudo-fandom.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetFandomSalesReport())
		},
	)
}

func toolGetDashboardStats() server.ServerTool {
	return noArgTool("get_dashboard_stats",
		"Get high-level dashboard statistics: total products, total conventions, total revenue, total profit, and average sell-through rate.",
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return jsonResult(db.GetDashboardStats())
		},
	)
}
