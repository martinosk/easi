package publishedlanguage

import (
	"easi/backend/internal/shared/agenttools"
)

func AgentTools() []agenttools.AgentToolSpec {
	var specs []agenttools.AgentToolSpec
	specs = append(specs, applicationTools()...)
	specs = append(specs, originEntityTools()...)
	specs = append(specs, originLinkTools()...)
	specs = append(specs, originEntityCRUDTools()...)
	return specs
}

func applicationTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "list_applications", Description: "List application components (IT systems) in the architecture portfolio. Applications can realize business capabilities, have relations to other applications, and carry fit scores per strategy pillar. Filter by name substring. Returns up to limit results.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components",
			QueryParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Filter by application name (partial match)", false),
				agenttools.IntParam("limit", "Max results (1-50, default 20)"),
			},
		},
		{
			Name: "get_application_details", Description: "Get full details of an application component by ID, including its description, origin, and metadata. Use get_application_relations for its links to other systems.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Application ID (UUID)")},
		},
		{
			Name: "create_application", Description: "Register a new application component (IT system) in the architecture portfolio. The application can then be linked to capabilities via realizations, related to other applications, and scored against strategy pillars.",
			Access: agenttools.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/components",
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Application name", true),
				agenttools.StringParam("description", "Application description", false),
			},
		},
		{
			Name: "update_application", Description: "Update an existing application component's name or description. Does not affect its realizations, relations, or scores.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Application ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "New application name", false),
				agenttools.StringParam("description", "New application description", false),
			},
		},
		{
			Name: "delete_application", Description: "Remove an application component from the portfolio. This also removes its realizations, relations, and fit scores.",
			Access: agenttools.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/components/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Application ID (UUID)")},
		},
		{
			Name: "create_application_relation", Description: "Create a directed relation between two application components (e.g. depends_on, uses, sends_data_to). Relations model integration dependencies and data flows between systems.",
			Access: agenttools.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/relations",
			BodyParams: []agenttools.ParamSpec{
				{Name: "sourceComponentId", Type: "uuid", Description: "Source application ID (UUID)", Required: true},
				{Name: "targetComponentId", Type: "uuid", Description: "Target application ID (UUID)", Required: true},
				agenttools.StringParam("relationType", "Relation type (e.g. depends_on, uses, sends_data_to)", true),
				agenttools.StringParam("description", "Relation description", false),
			},
		},
		{
			Name: "delete_application_relation", Description: "Delete a relation between two application components. Does not affect the applications themselves.",
			Access: agenttools.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/relations/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Relation ID (UUID)")},
		},
	}
}

func originEntityTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "list_vendors", Description: "List all vendors. Vendors are external companies that sell software products. Applications can be linked to a vendor via 'purchased from' origin to track where commercial software comes from.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/vendors",
		},
		{
			Name: "get_vendor_details", Description: "Get details of a vendor including name, description, and which applications are purchased from them. Use to understand commercial software sourcing.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/vendors/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Vendor ID (UUID)")},
		},
		{
			Name: "list_acquired_entities", Description: "List all acquired entities. Acquired entities are companies or business units that were acquired (M&A) and brought their own IT systems. Applications can be linked to an acquired entity via 'acquired via' origin.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/acquired-entities",
		},
		{
			Name: "get_acquired_entity_details", Description: "Get details of an acquired entity including name, description, and which applications came through this acquisition. Use to trace M&A-originated systems.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/acquired-entities/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Acquired entity ID (UUID)")},
		},
		{
			Name: "list_internal_teams", Description: "List all internal teams. Internal teams are development groups that build and maintain in-house applications. Applications can be linked to a team via 'built by' origin.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/internal-teams",
		},
		{
			Name: "get_internal_team_details", Description: "Get details of an internal team including name, description, and which applications they build and maintain. Use to understand in-house software ownership.",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/internal-teams/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Internal team ID (UUID)")},
		},
		{
			Name: "get_component_origin", Description: "Get all origin relationships for an application component — whether it was purchased from a vendor, acquired via an M&A entity, or built by an internal team. One application can have multiple origins (e.g. acquired then maintained by internal team).",
			Access: agenttools.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components/{componentId}/origins",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
		},
	}
}

func originLinkTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "set_acquired_via_origin", Description: "Link an application component to an acquired entity, recording that it was acquired through an M&A transaction. Replaces any existing acquired-via link for this component.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{componentId}/origin/acquired-via",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				{Name: "acquiredEntityId", Type: "uuid", Description: "Acquired entity ID (UUID)", Required: true},
				agenttools.StringParam("notes", "Additional notes about the acquisition origin", false),
			},
		},
		{
			Name: "clear_acquired_via_origin", Description: "Remove the acquired-via origin link from an application component. Does not delete the acquired entity or the application.",
			Access: agenttools.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/components/{componentId}/origin/acquired-via",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
		},
		{
			Name: "set_purchased_from_origin", Description: "Link an application component to a vendor, recording that it was purchased from that vendor. Replaces any existing purchased-from link for this component.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{componentId}/origin/purchased-from",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				{Name: "vendorId", Type: "uuid", Description: "Vendor ID (UUID)", Required: true},
				agenttools.StringParam("notes", "Additional notes about the purchase origin", false),
			},
		},
		{
			Name: "clear_purchased_from_origin", Description: "Remove the purchased-from origin link from an application component. Does not delete the vendor or the application.",
			Access: agenttools.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/components/{componentId}/origin/purchased-from",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
		},
		{
			Name: "set_built_by_origin", Description: "Link an application component to an internal team, recording that it was built by that team. Replaces any existing built-by link for this component.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{componentId}/origin/built-by",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				{Name: "internalTeamId", Type: "uuid", Description: "Internal team ID (UUID)", Required: true},
				agenttools.StringParam("notes", "Additional notes about the internal development origin", false),
			},
		},
		{
			Name: "clear_built_by_origin", Description: "Remove the built-by origin link from an application component. Does not delete the internal team or the application.",
			Access: agenttools.AccessDelete, Permission: "components:write",
			Method: "DELETE", Path: "/components/{componentId}/origin/built-by",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("componentId", "Application component ID (UUID)")},
		},
	}
}

func originEntityCRUDTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "create_acquired_entity", Description: "Register a new acquired entity (M&A company or business unit). After creation, use set_acquired_via_origin to link application components that came through this acquisition.",
			Access: agenttools.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/acquired-entities",
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Acquired entity name", true),
				agenttools.StringParam("acquisitionDate", "Acquisition date (YYYY-MM-DD format)", false),
				agenttools.StringParam("integrationStatus", "Integration status", false),
				agenttools.StringParam("notes", "Additional notes", false),
			},
		},
		{
			Name: "update_acquired_entity", Description: "Update an acquired entity's details. Does not affect origin links to application components.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/acquired-entities/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Acquired entity ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Acquired entity name", true),
				agenttools.StringParam("acquisitionDate", "Acquisition date (YYYY-MM-DD format)", false),
				agenttools.StringParam("integrationStatus", "Integration status", false),
				agenttools.StringParam("notes", "Additional notes", false),
			},
		},
		{
			Name: "create_vendor", Description: "Register a new vendor (external software supplier). After creation, use set_purchased_from_origin to link application components purchased from this vendor.",
			Access: agenttools.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/vendors",
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Vendor name", true),
				agenttools.StringParam("implementationPartner", "Implementation partner name", false),
				agenttools.StringParam("notes", "Additional notes", false),
			},
		},
		{
			Name: "update_vendor", Description: "Update a vendor's details. Does not affect origin links to application components.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/vendors/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Vendor ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Vendor name", true),
				agenttools.StringParam("implementationPartner", "Implementation partner name", false),
				agenttools.StringParam("notes", "Additional notes", false),
			},
		},
		{
			Name: "create_internal_team", Description: "Register a new internal development team. After creation, use set_built_by_origin to link application components built by this team.",
			Access: agenttools.AccessCreate, Permission: "components:write",
			Method: "POST", Path: "/internal-teams",
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Team name", true),
				agenttools.StringParam("department", "Department name", false),
				agenttools.StringParam("contactPerson", "Contact person name", false),
				agenttools.StringParam("notes", "Additional notes", false),
			},
		},
		{
			Name: "update_internal_team", Description: "Update an internal team's details. Does not affect origin links to application components.",
			Access: agenttools.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/internal-teams/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Internal team ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Team name", true),
				agenttools.StringParam("department", "Department name", false),
				agenttools.StringParam("contactPerson", "Contact person name", false),
				agenttools.StringParam("notes", "Additional notes", false),
			},
		},
	}
}
