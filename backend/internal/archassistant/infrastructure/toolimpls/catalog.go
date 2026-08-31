package toolimpls

import (
	adPL "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	avPL "easi/backend/internal/architectureviews/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"
)

var allowedContexts = []func() []AgentToolSpec{
	amPL.AgentTools,
	avPL.AgentTools,
	cmPL.AgentTools,
	vsPL.AgentTools,
	mmPL.AgentTools,
	adPL.AgentTools,
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
