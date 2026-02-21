package toolimpls

import (
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"
)

var allowedContexts = []func() []AgentToolSpec{
	amPL.AgentTools,
	cmPL.AgentTools,
	eaPL.AgentTools,
	vsPL.AgentTools,
	mmPL.AgentTools,
}

func CollectToolSpecs(providers ...func() []AgentToolSpec) []AgentToolSpec {
	var all []AgentToolSpec
	for _, provider := range providers {
		all = append(all, provider()...)
	}
	return all
}

func AllContextToolSpecs() []AgentToolSpec {
	return CollectToolSpecs(allowedContexts...)
}
