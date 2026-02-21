package toolimpls

import (
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
)

func RegisterAllTools(registry *tools.Registry, client *agenthttp.Client) {
	RegisterSpecTools(registry, client)
	registerCompositeTools(registry, client)
}

func registerCompositeTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name: "list_application_relations", Description: "List all relations (incoming and outgoing) for an application",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Application ID (UUID)", Required: true},
		},
	}, &listApplicationRelationsTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "search_architecture", Description: "Search across applications, capabilities, and business domains by name",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "query", Type: "string", Description: "Search query (name to search for)", Required: true},
		},
	}, &searchArchitectureTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_portfolio_summary", Description: "Get aggregate statistics across the architecture portfolio",
		Permission: "components:read", Access: tools.AccessRead,
	}, &getPortfolioSummaryTool{client: client})
}
