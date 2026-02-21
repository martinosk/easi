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
		Name: "list_application_relations", Description: "List all relations (incoming and outgoing) for an application component. Relations show dependencies, data flows, and integration links between systems.",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Application ID (UUID)", Required: true},
		},
	}, &listApplicationRelationsTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "search_architecture", Description: "Search across applications, capabilities, and business domains by name. Use for discovery when you know a partial name but not the entity type or ID.",
		Permission: "components:read", Access: tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "query", Type: "string", Description: "Search query (name to search for)", Required: true},
		},
	}, &searchArchitectureTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name: "get_portfolio_summary", Description: "Get aggregate counts across the architecture portfolio: applications, capabilities, business domains, value streams, and relations. Use for a quick landscape overview.",
		Permission: "components:read", Access: tools.AccessRead,
	}, &getPortfolioSummaryTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "query_domain_model",
		Description: "Get detailed information about the EASI domain model structure, relationships, and business rules. Use when you need to understand how concepts relate before performing operations.",
		Permission:  "assistant:use",
		Access:      tools.AccessRead,
		Parameters: []tools.ParameterDef{
			{Name: "topic", Type: "string", Description: "Topic to query: capability-hierarchy, business-domains, realizations, strategy, enterprise-capabilities, time-classification, value-streams, component-origins, overview", Required: true},
		},
	}, &domainKnowledgeTool{})
}
