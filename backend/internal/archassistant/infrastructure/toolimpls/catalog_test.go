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

	expectedTools := []string{
		"list_applications", "get_application_details",
		"create_application", "update_application", "delete_application",
		"create_application_relation", "delete_application_relation",
		"list_vendors", "get_vendor_details",
		"list_acquired_entities", "get_acquired_entity_details",
		"list_internal_teams", "get_internal_team_details",
		"get_component_origin",
		"set_acquired_via_origin", "clear_acquired_via_origin",
		"set_purchased_from_origin", "clear_purchased_from_origin",
		"set_built_by_origin", "clear_built_by_origin",
		"create_acquired_entity", "update_acquired_entity",
		"create_vendor", "update_vendor",
		"create_internal_team", "update_internal_team",

		"list_capabilities", "get_capability_details",
		"create_capability", "update_capability", "delete_capability",
		"realize_capability", "unrealize_capability",
		"list_business_domains", "get_business_domain_details",
		"create_business_domain", "update_business_domain",
		"assign_capability_to_domain", "remove_capability_from_domain",
		"list_capability_dependencies", "create_capability_dependency", "delete_capability_dependency",
		"get_capability_children",
		"get_strategy_importance", "set_strategy_importance",
		"get_application_fit_scores", "set_application_fit_score",
		"get_strategic_fit_analysis",
		"get_capability_metadata_index", "get_capability_maturity_levels",
		"get_capability_statuses", "get_capability_ownership_models",
		"get_capability_expert_roles",
		"update_capability_metadata",
		"get_capability_realizations", "get_capability_business_domains",
		"get_domain_importance_overview", "get_fit_scores_by_pillar",

		"list_enterprise_capabilities", "get_enterprise_capability_details",
		"create_enterprise_capability", "update_enterprise_capability", "delete_enterprise_capability",
		"link_capability_to_enterprise", "unlink_capability_from_enterprise",
		"get_enterprise_strategic_importance", "set_enterprise_strategic_importance",
		"get_time_suggestions",
		"get_maturity_analysis", "get_maturity_gap_detail",

		"list_value_streams", "get_value_stream_details",
		"create_value_stream", "update_value_stream",
		"get_value_stream_capabilities",
		"create_value_stream_stage", "update_value_stream_stage",
		"reorder_value_stream_stages", "add_stage_capability",

		"get_strategy_pillars", "get_maturity_scale",
	}

	assert.ElementsMatch(t, expectedTools, names)
}

func TestContextOwnedCatalogs_ToolCounts(t *testing.T) {
	assert.Len(t, amPL.AgentTools(), 26, "architecturemodeling")
	assert.Len(t, cmPL.AgentTools(), 32, "capabilitymapping")
	assert.Len(t, vsPL.AgentTools(), 9, "valuestreams")
	assert.Len(t, eaPL.AgentTools(), 12, "enterprisearchitecture")
	assert.Len(t, mmPL.AgentTools(), 2, "metamodel")
}
