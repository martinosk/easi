package publishedlanguage

import (
	"easi/backend/internal/shared/agenttools"
)

func AgentTools() []agenttools.AgentToolSpec {
	tools := directionTools()
	tools = append(tools, timeAssessmentTools()...)
	tools = append(tools, realizationRoleTools()...)
	tools = append(tools, capabilityJourneyTools()...)
	tools = append(tools, compositionTools()...)
	return tools
}

func directionTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name:        "get_direction_for_enterprise_capability",
			Description: "Get the active architecture direction on an enterprise capability — what the architecture group intends to do with it (consolidate / decompose / stay), where it is on the agenda (draft / proposed / agreed), the narrative, and the affected physical capabilities. Returns null if no direction has been captured.",
			Access:      agenttools.AccessRead,
			Permission:  "architecture-direction:read",
			Method:      "GET",
			Path:        "/enterprise-capabilities/{id}/direction",
			PathParams:  []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name:        "get_standard_application_for_enterprise_capability",
			Description: "Get the standard application for an enterprise capability — the architecture group's recorded answer to which application should realise this capability, with the narrative that explains the choice. Returns null in the standard envelope if no standard has been set.",
			Access:      agenttools.AccessRead,
			Permission:  "architecture-direction:read",
			Method:      "GET",
			Path:        "/enterprise-capabilities/{id}/standard-application",
			PathParams:  []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
	}
}

func timeAssessmentTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name:        "get_time_assessment_for_realization",
			Description: "Get the current TIME grade (Invest / Tolerate / Migrate / Eliminate) an architect has recorded for a direct realisation — the pairing of a domain capability and the application component that realises it. Returns 404 if the pair has never been assessed.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/components/{componentId}/time-assessment",
			PathParams: []agenttools.ParamSpec{
				agenttools.UUIDParam("id", "Domain capability ID (UUID)"),
				agenttools.UUIDParam("componentId", "Application component ID (UUID)"),
			},
		},
		{
			Name:        "list_time_assessments",
			Description: "Bulk-fetch the current TIME assessment for every realisation — one entry per assessed (capability, component) pair. Narrow to a set of domain capabilities with capabilityIds, or omit it for the whole collection.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/time-assessments",
			QueryParams: []agenttools.ParamSpec{
				agenttools.StringParam("capabilityIds", "Comma-separated domain capability IDs (UUIDs); omit to fetch the whole collection", false),
			},
		},
		{
			Name:        "get_time_assessment_rollups",
			Description: "Get, for each given application component, the count of current TIME assessments per grade across the landscape — how many capabilities grade this application Invest / Tolerate / Migrate / Eliminate. Useful for spotting carve-out and rationalisation candidates.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/time-assessments/rollups",
			QueryParams: []agenttools.ParamSpec{
				agenttools.StringParam("componentIds", "Comma-separated application component IDs (UUIDs)", true),
			},
		},
	}
}

func realizationRoleTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name:        "get_realization_role_for_capability_component",
			Description: "Get the current realization role (standard / legacy) an architect has assigned to a direct realisation — the pairing of a domain capability and the application component that realises it. Absence of a role means unclassified. Returns 404 if the pair has never been assigned a role.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/components/{componentId}/realization-role",
			PathParams: []agenttools.ParamSpec{
				agenttools.UUIDParam("id", "Domain capability ID (UUID)"),
				agenttools.UUIDParam("componentId", "Application component ID (UUID)"),
			},
		},
		{
			Name:        "list_realization_roles",
			Description: "Bulk-fetch the current realization role for every realisation — one entry per (capability, component) pair that has been assigned standard or legacy. Narrow to a set of domain capabilities with capabilityIds, or omit it for the whole collection.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/realization-roles",
			QueryParams: []agenttools.ParamSpec{
				agenttools.StringParam("capabilityIds", "Comma-separated domain capability IDs (UUIDs); omit to fetch the whole collection", false),
			},
		},
	}
}

func capabilityJourneyTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name:        "get_capability_journey",
			Description: "Get the active journey on a domain capability — the recorded change story (migration / consolidation / carve-out / move), its status, progress, target period, note, and milestones. Returns null in the journey envelope if no journey is active.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/journey",
			PathParams:  []agenttools.ParamSpec{agenttools.UUIDParam("id", "Domain capability ID (UUID)")},
		},
		{
			Name:        "get_capability_journey_history",
			Description: "Get every journey ever captured on a domain capability, newest first — the full record of how the capability's realisation has changed over time, including completed and abandoned journeys.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/journey/history",
			PathParams:  []agenttools.ParamSpec{agenttools.UUIDParam("id", "Domain capability ID (UUID)")},
		},
		{
			Name:        "list_capability_journeys",
			Description: "Bulk-fetch the current journeys (active plus most recent terminal) — one entry per journey found. Narrow to a set of domain capabilities with capabilityIds, or omit it for the whole collection.",
			Access:      agenttools.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capability-journeys",
			QueryParams: []agenttools.ParamSpec{
				agenttools.StringParam("capabilityIds", "Comma-separated domain capability IDs (UUIDs); omit to fetch the whole collection", false),
			},
		},
	}
}

func compositionTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name:        "list_enterprise_capability_compositions",
			Description: "List composition summaries for every enterprise capability: source, included, carved-out and business-domain counts derived from each active direction, plus the direction status. Enterprise capabilities without an active direction report zero counts. Use get_enterprise_capability_composition for the detailed breakdown of one enterprise capability.",
			Access:      agenttools.AccessRead,
			Permission:  "enterprise-arch:read",
			Method:      "GET",
			Path:        "/enterprise-capability-compositions",
		},
		{
			Name: "get_enterprise_capability_composition", Description: "Get the composition of an enterprise capability: every domain capability included via its active direction's sources and their subtrees, grouped by business domain, with carve-out attribution where a more specific source on another enterprise capability owns a subtree.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}/composition",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "search_direction_source_candidates", Description: "Search domain capabilities by name as candidate sources for an enterprise capability's direction, with per-candidate eligibility (a capability may be the explicit source of at most one active direction).",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/capabilities/source-candidates",
			QueryParams: []agenttools.ParamSpec{
				agenttools.StringParam("q", "Search term (case-insensitive substring match on capability name)", true),
				{Name: "ecId", Type: "uuid", Description: "Enterprise capability the sources are searched for", Required: true},
				{Name: "domainId", Type: "uuid", Description: "Filter to capabilities in this business domain"},
				agenttools.IntParam("limit", "Max results to return (default 20)"),
			},
		},
		{
			Name: "get_maturity_analysis", Description: "Get maturity analysis candidates — enterprise capabilities where the current maturity of included domain capabilities falls below the target maturity level. Use to identify strategic themes that need maturity investment.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/maturity-analysis",
		},
		{
			Name: "get_maturity_gap_detail", Description: "Get detailed maturity gap analysis for a specific enterprise capability. Shows each included domain capability's current maturity versus the enterprise capability's target maturity, highlighting where gaps exist.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}/maturity-gap",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
	}
}
