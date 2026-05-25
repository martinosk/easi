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
	}
}
