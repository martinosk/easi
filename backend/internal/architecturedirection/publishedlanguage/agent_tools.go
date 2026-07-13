package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name:        "get_direction_for_enterprise_capability",
			Description: "Get the active architecture direction on an enterprise capability — what the architecture group intends to do with it (consolidate / decompose / stay), where it is on the agenda (draft / proposed / agreed), the narrative, and the affected physical capabilities. Returns null if no direction has been captured.",
			Access:      pl.AccessRead,
			Permission:  "architecture-direction:read",
			Method:      "GET",
			Path:        "/enterprise-capabilities/{id}/direction",
			PathParams:  []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name:        "get_standard_application_for_enterprise_capability",
			Description: "Get the standard application for an enterprise capability — the architecture group's recorded answer to which application should realise this capability, with the narrative that explains the choice. Returns null in the standard envelope if no standard has been set.",
			Access:      pl.AccessRead,
			Permission:  "architecture-direction:read",
			Method:      "GET",
			Path:        "/enterprise-capabilities/{id}/standard-application",
			PathParams:  []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name:        "get_time_assessment_for_realization",
			Description: "Get the current TIME grade (Invest / Tolerate / Migrate / Eliminate) an architect has recorded for a direct realisation — the pairing of a domain capability and the application component that realises it. Returns 404 if the pair has never been assessed.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/components/{componentId}/time-assessment",
			PathParams: []pl.ParamSpec{
				pl.UUIDParam("id", "Domain capability ID (UUID)"),
				pl.UUIDParam("componentId", "Application component ID (UUID)"),
			},
		},
		{
			Name:        "list_time_assessments",
			Description: "Bulk-fetch the current TIME assessment for every realisation among a set of domain capabilities — one entry per assessed (capability, component) pair.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/time-assessments",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("capabilityIds", "Comma-separated domain capability IDs (UUIDs)", true),
			},
		},
		{
			Name:        "get_time_assessment_rollups",
			Description: "Get, for each given application component, the count of current TIME assessments per grade across the landscape — how many capabilities grade this application Invest / Tolerate / Migrate / Eliminate. Useful for spotting carve-out and rationalisation candidates.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/time-assessments/rollups",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("componentIds", "Comma-separated application component IDs (UUIDs)", true),
			},
		},
		{
			Name:        "get_realization_role_for_capability_component",
			Description: "Get the current realization role (standard / legacy) an architect has assigned to a direct realisation — the pairing of a domain capability and the application component that realises it. Absence of a role means unclassified. Returns 404 if the pair has never been assigned a role.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/components/{componentId}/realization-role",
			PathParams: []pl.ParamSpec{
				pl.UUIDParam("id", "Domain capability ID (UUID)"),
				pl.UUIDParam("componentId", "Application component ID (UUID)"),
			},
		},
		{
			Name:        "list_realization_roles",
			Description: "Bulk-fetch the current realization role for every realisation among a set of domain capabilities — one entry per (capability, component) pair that has been assigned standard or legacy.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/realization-roles",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("capabilityIds", "Comma-separated domain capability IDs (UUIDs)", true),
			},
		},
		{
			Name:        "get_capability_journey",
			Description: "Get the active journey on a domain capability — the recorded change story (migration / consolidation / carve-out / move), its status, progress, target period, note, and milestones. Returns null in the journey envelope if no journey is active.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/journey",
			PathParams:  []pl.ParamSpec{pl.UUIDParam("id", "Domain capability ID (UUID)")},
		},
		{
			Name:        "get_capability_journey_history",
			Description: "Get every journey ever captured on a domain capability, newest first — the full record of how the capability's realisation has changed over time, including completed and abandoned journeys.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capabilities/{id}/journey/history",
			PathParams:  []pl.ParamSpec{pl.UUIDParam("id", "Domain capability ID (UUID)")},
		},
		{
			Name:        "list_capability_journeys",
			Description: "Bulk-fetch the current journeys (active plus most recent terminal) for a set of domain capabilities — one entry per journey found.",
			Access:      pl.AccessRead,
			Permission:  "domains:read",
			Method:      "GET",
			Path:        "/capability-journeys",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("capabilityIds", "Comma-separated domain capability IDs (UUIDs)", true),
			},
		},
	}
}
