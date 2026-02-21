package toolimpls_test

import (
	"testing"

	pl "easi/backend/internal/archassistant/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	vsPL "easi/backend/internal/valuestreams/publishedlanguage"
	"easi/backend/internal/archassistant/infrastructure/toolimpls"

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

func TestContextOwnedCatalogs_ContainAll21Tools(t *testing.T) {
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

	expectedTools := []string{
		"list_applications", "get_application_details",
		"create_application", "update_application", "delete_application",
		"create_application_relation", "delete_application_relation",
		"list_capabilities", "get_capability_details",
		"create_capability", "update_capability", "delete_capability",
		"realize_capability", "unrealize_capability",
		"list_business_domains", "get_business_domain_details",
		"create_business_domain", "update_business_domain",
		"assign_capability_to_domain", "remove_capability_from_domain",
		"list_value_streams", "get_value_stream_details",
	}

	assert.ElementsMatch(t, expectedTools, names)
}

func TestContextOwnedCatalogs_ToolCounts(t *testing.T) {
	assert.Len(t, amPL.AgentTools(), 7, "architecturemodeling")
	assert.Len(t, cmPL.AgentTools(), 13, "capabilitymapping")
	assert.Len(t, vsPL.AgentTools(), 2, "valuestreams")
	assert.Empty(t, eaPL.AgentTools(), "enterprisearchitecture")
	assert.Empty(t, mmPL.AgentTools(), "metamodel")
}
