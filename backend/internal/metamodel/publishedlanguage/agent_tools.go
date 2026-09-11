package publishedlanguage

import (
	"easi/backend/internal/shared/agenttools"
)

func AgentTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "get_strategy_pillars", Description: "Get the configured strategy pillars. Strategy pillars are the strategic dimensions (e.g. Business Agility, Cost Efficiency, Security) against which capabilities are rated for importance and applications are scored for fit. Defined in the MetaModel by enterprise architects.",
			Access: agenttools.AccessRead, Permission: "metamodel:read",
			Method: "GET", Path: "/meta-model/strategy-pillars",
		},
		{
			Name: "get_maturity_scale", Description: "Get the configured maturity scale. The maturity scale defines the levels (e.g. Initial, Managed, Defined, Optimized) used to assess capability maturity. Each level has a numeric value and description. Defined in the MetaModel by enterprise architects.",
			Access: agenttools.AccessRead, Permission: "metamodel:read",
			Method: "GET", Path: "/meta-model/maturity-scale",
		},
		{
			Name: "get_subject_attributes", Description: "Get the custom attribute schema MetaModel defines for a subject type (capability, application, acquired-entity, vendor or internal-team): each attribute's name, type (text, number, date, link, selection, contact-person), help text, selection options, number bounds and whether it is active. One-pagers show these attributes as custom fields; which of them are required is one-pager configuration, not part of the schema.",
			Access: agenttools.AccessRead, Permission: "metamodel:read",
			Method: "GET", Path: "/meta-model/subject-types/{subjectType}/attributes",
			PathParams: []agenttools.ParamSpec{agenttools.StringParam("subjectType", "Subject type: capability, application, acquired-entity, vendor or internal-team", true)},
		},
	}
}
