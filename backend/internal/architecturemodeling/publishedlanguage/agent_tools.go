package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_applications", Description: "List application components (IT systems) in the architecture portfolio. Applications can realize business capabilities, have relations to other applications, and carry fit scores per strategy pillar. Filter by name substring. Returns up to limit results.",
			Access: pl.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("name", "Filter by application name (partial match)", false),
				pl.IntParam("limit", "Max results (1-50, default 20)"),
			},
		},
		{
			Name: "get_application_details", Description: "Get full details of an application component by ID, including its description, origin, and metadata. Use get_application_relations for its links to other systems.",
			Access: pl.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application ID (UUID)")},
		},
		{
			Name: "create_application", Description: "Register a new application component (IT system) in the architecture portfolio. The application can then be linked to capabilities via realizations, related to other applications, and scored against strategy pillars.",
			Access: pl.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/components",
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "Application name", true),
				pl.StringParam("description", "Application description", false),
			},
		},
		{
			Name: "update_application", Description: "Update an existing application component's name or description. Does not affect its realizations, relations, or scores.",
			Access: pl.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New application name", false),
				pl.StringParam("description", "New application description", false),
			},
		},
		{
			Name: "delete_application", Description: "Remove an application component from the portfolio. This also removes its realizations, relations, and fit scores.",
			Access: pl.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/components/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application ID (UUID)")},
		},
		{
			Name: "create_application_relation", Description: "Create a directed relation between two application components (e.g. depends_on, uses, sends_data_to). Relations model integration dependencies and data flows between systems.",
			Access: pl.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/relations",
			BodyParams: []pl.ParamSpec{
				{Name: "sourceComponentId", Type: "uuid", Description: "Source application ID (UUID)", Required: true},
				{Name: "targetComponentId", Type: "uuid", Description: "Target application ID (UUID)", Required: true},
				pl.StringParam("relationType", "Relation type (e.g. depends_on, uses, sends_data_to)", true),
				pl.StringParam("description", "Relation description", false),
			},
		},
		{
			Name: "delete_application_relation", Description: "Delete a relation between two application components. Does not affect the applications themselves.",
			Access: pl.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/relations/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Relation ID (UUID)")},
		},
	}
}
