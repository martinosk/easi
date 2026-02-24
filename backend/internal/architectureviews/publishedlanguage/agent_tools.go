package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_views", Description: "List architecture views. Views are curated perspectives on the architecture — each view groups a subset of application components, capabilities, and origin entities into a named, describable layout. Views can be default (shown first) or private. Returns view names, descriptions, component/capability membership, and visibility status.",
			Access: pl.AccessRead, Permission: "views:read",
			Method: "GET", Path: "/views",
		},
		{
			Name: "get_view_details", Description: "Get a specific architecture view by ID with full details including component positions, capabilities, origin entities, layout direction, edge type, and color scheme. Use to understand which applications and capabilities are grouped in a particular architectural perspective.",
			Access: pl.AccessRead, Permission: "views:read",
			Method: "GET", Path: "/views/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "View ID (UUID)")},
		},
	}
}
