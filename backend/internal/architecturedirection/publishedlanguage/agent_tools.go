package publishedlanguage

import (
	"easi/backend/internal/shared/agenttools"
)

func AgentTools() []agenttools.AgentToolSpec {
	tools := timeAssessmentTools()
	tools = append(tools, realizationRoleTools()...)
	tools = append(tools, capabilityJourneyTools()...)
	return tools
}

func timeAssessmentTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name:        "get_time_assessment_for_realization",
			Description: "Get the current TIME grade (Invest / Tolerate / Migrate / Eliminate) an architect has recorded for a direct realisation — the pairing of a domain capability and the application component that realises it — alongside the TIME suggestion computed from the pair's fit gaps. Returns 404 if the pair has never been assessed.",
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
			Description: "Bulk-fetch the TIME picture for every realisation — one entry per (capability, component) pair, carrying the grade an architect recorded (null when unassessed) and the computed TIME suggestion with its confidence and technical/functional gaps, where the fit data yields one. Narrow to a set of domain capabilities with capabilityIds, or omit it for the whole collection.",
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
