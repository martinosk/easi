package toolimpls_test

import (
	"testing"

	"easi/backend/internal/archassistant/infrastructure/toolimpls"
	pl "easi/backend/internal/archassistant/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"

	"github.com/stretchr/testify/assert"
)

func TestCollectToolSpecs_GathersFromMultipleProviders(t *testing.T) {
	providerA := func() []pl.AgentToolSpec {
		return []pl.AgentToolSpec{
			{Name: "tool_a1"},
			{Name: "tool_a2"},
		}
	}
	providerB := func() []pl.AgentToolSpec {
		return []pl.AgentToolSpec{
			{Name: "tool_b1"},
		}
	}

	specs := toolimpls.CollectToolSpecs(providerA, providerB)

	names := make([]string, len(specs))
	for i, s := range specs {
		names[i] = s.Name
	}
	assert.Equal(t, []string{"tool_a1", "tool_a2", "tool_b1"}, names)
}

func TestCollectToolSpecs_EmptyProviders(t *testing.T) {
	empty := func() []pl.AgentToolSpec { return nil }
	specs := toolimpls.CollectToolSpecs(empty)
	assert.Empty(t, specs)
}

func TestCollectToolSpecs_NoProviders(t *testing.T) {
	specs := toolimpls.CollectToolSpecs()
	assert.Empty(t, specs)
}

func TestContextOwnedCatalogs_ContainAllTools(t *testing.T) {
	specs := toolimpls.CollectToolSpecs(
		amPL.AgentTools,
		cmPL.AgentTools,
		eaPL.AgentTools,
		vsPL.AgentTools,
		mmPL.AgentTools,
	)

	names := make([]string, len(specs))
	for i, s := range specs {
		names[i] = s.Name
	}

	assert.ElementsMatch(t, allExpectedSpecToolNames, names)
}

func TestContextOwnedCatalogs_ToolCounts(t *testing.T) {
	assert.Len(t, amPL.AgentTools(), 26, "architecturemodeling")
	assert.Len(t, cmPL.AgentTools(), 34, "capabilitymapping")
	assert.Len(t, vsPL.AgentTools(), 9, "valuestreams")
	assert.Len(t, eaPL.AgentTools(), 12, "enterprisearchitecture")
	assert.Len(t, mmPL.AgentTools(), 2, "metamodel")
}
