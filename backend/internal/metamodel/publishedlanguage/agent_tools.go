package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "get_strategy_pillars", Description: "Get the configured strategy pillars. Strategy pillars are the strategic dimensions (e.g. Business Agility, Cost Efficiency, Security) against which capabilities are rated for importance and applications are scored for fit. Defined in the MetaModel by enterprise architects.",
			Access: pl.AccessRead, Permission: "metamodel:read",
			Method: "GET", Path: "/meta-model/strategy-pillars",
		},
		{
			Name: "get_maturity_scale", Description: "Get the configured maturity scale. The maturity scale defines the levels (e.g. Initial, Managed, Defined, Optimized) used to assess capability maturity. Each level has a numeric value and description. Defined in the MetaModel by enterprise architects.",
			Access: pl.AccessRead, Permission: "metamodel:read",
			Method: "GET", Path: "/meta-model/maturity-scale",
		},
	}
}
