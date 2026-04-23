package mcp

import (
	"github.com/go-chi/chi/v5"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MountRoutes creates the MCP server with read-only analytics tools
// and mounts SSE transport routes on the given Chi router under /mcp/.
func MountRoutes(r chi.Router) {
	mcpServer := server.NewMCPServer(
		"Stockii Analytics",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Register all read-only tools
	for _, t := range allTools() {
		mcpServer.AddTool(t.Tool, t.Handler)
	}

	sseServer := server.NewSSEServer(mcpServer,
		server.WithStaticBasePath("/mcp"),
	)

	r.Route("/mcp", func(r chi.Router) {
		r.Handle("/*", sseServer)
	})
}

// allTools returns all read-only MCP tool definitions.
func allTools() []server.ServerTool {
	return []server.ServerTool{
		toolListProducts(),
		toolListConventions(),
		toolListCategories(),
		toolListFandoms(),
		toolListConventionSeries(),
		toolGetConventionPnL(),
		toolGetAllConventionPnLs(),
		toolGetConventionInventory(),
		toolGetProductHistory(),
		toolGetCategoryReport(),
		toolGetFandomReport(),
		toolGetDashboardStats(),
	}
}

// helper to build a no-arg tool
func noArgTool(name, desc string, handler server.ToolHandlerFunc) server.ServerTool {
	return server.ServerTool{
		Tool:    mcp.NewTool(name, mcp.WithDescription(desc)),
		Handler: handler,
	}
}
