package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_applications", Description: "List applications in the architecture portfolio. Optionally filter by name.",
			Access: pl.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("name", "Filter by application name (partial match)", false),
				pl.IntParam("limit", "Max results (1-50, default 20)"),
			},
		},
		{
			Name: "get_application_details", Description: "Get full details of an application by ID",
			Access: pl.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application ID (UUID)")},
		},
		{
			Name: "create_application", Description: "Create a new application in the architecture portfolio",
			Access: pl.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/components",
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "Application name", true),
				pl.StringParam("description", "Application description", false),
			},
		},
		{
			Name: "update_application", Description: "Update an existing application's properties",
			Access: pl.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New application name", false),
				pl.StringParam("description", "New application description", false),
			},
		},
		{
			Name: "delete_application", Description: "Delete an application from the portfolio",
			Access: pl.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/components/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application ID (UUID)")},
		},
		{
			Name: "create_application_relation", Description: "Create a relation between two applications",
			Access: pl.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/relations",
			BodyParams: []pl.ParamSpec{
				{Name: "sourceComponentId", Type: "uuid", Description: "Source application ID (UUID)", Required: true},
				{Name: "targetComponentId", Type: "uuid", Description: "Target application ID (UUID)", Required: true},
				pl.StringParam("relationType", "Relation type (e.g. depends_on)", true),
				pl.StringParam("description", "Relation description", false),
			},
		},
		{
			Name: "delete_application_relation", Description: "Delete a relation between applications",
			Access: pl.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/relations/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Relation ID (UUID)")},
		},
	}
}
