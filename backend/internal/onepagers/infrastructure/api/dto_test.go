package api

import (
	"testing"
	"time"

	"easi/backend/internal/onepagers/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func adminActor() sharedctx.Actor {
	return sharedctx.NewActor("user-1", "admin@example.com", sharedctx.RoleAdmin)
}

func stakeholderActor() sharedctx.Actor {
	return sharedctx.NewActor("user-2", "viewer@example.com", sharedctx.RoleStakeholder)
}

func testLinks() *OnePagerLinks {
	return NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
}

func floatPtr(v float64) *float64 {
	return &v
}

const (
	selectionFieldID = "9f0d5e69-0000-0000-0000-000000000001"
	retiredFieldID   = "9f0d5e69-0000-0000-0000-000000000002"
	numberFieldID    = "9f0d5e69-0000-0000-0000-000000000004"
)

func applicationRecord() *readmodels.ConfigurationRecord {
	now := time.Now().UTC()
	return &readmodels.ConfigurationRecord{
		ID:          "config-1",
		TenantID:    "tenant-123",
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.FieldRequirementRecord{
				{ID: selectionFieldID, Required: true},
				{ID: retiredFieldID, Required: true},
			},
			BuiltInFields: []readmodels.FieldRequirementRecord{
				{ID: "description", Required: true},
				{ID: "experts", Required: true},
			},
			DisplayOrder: []readmodels.FieldRefRecord{
				{Kind: "builtIn", ID: "name"},
				{Kind: "custom", ID: selectionFieldID},
				{Kind: "builtIn", ID: "description"},
				{Kind: "custom", ID: numberFieldID},
			},
		},
		Version:    4,
		CreatedAt:  now,
		ModifiedAt: now,
		ModifiedBy: "admin@example.com",
	}
}

func applicationDefinitions() readmodels.CustomFieldDefinitions {
	return readmodels.CustomFieldDefinitions{
		{
			ID: selectionFieldID, Name: "Hosting model", Type: "selection", HelpText: "Where it runs", Active: true,
			Options: []readmodels.OptionRecord{
				{ID: "9f0d5e69-0000-0000-0000-00000000000a", Label: "On-prem", Active: false},
				{ID: "9f0d5e69-0000-0000-0000-00000000000b", Label: "Cloud", Active: true},
			},
		},
		{ID: retiredFieldID, Name: "Old field", Type: "text", Active: false},
		{ID: numberFieldID, Name: "Maturity score", Type: "number", Active: true, Min: floatPtr(0), Max: floatPtr(5)},
	}
}

func applicationView() ConfigurationView {
	return ConfigurationView{Record: applicationRecord(), Definitions: applicationDefinitions()}
}

func TestBuildConfigurationDTO_MergesCatalogWithInclusionState(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), stakeholderActor())

	assert.Equal(t, "config-1", dto.ID)
	assert.Equal(t, "application", dto.SubjectType)
	assert.Equal(t, 4, dto.Version)

	require.Len(t, dto.BuiltInFields, 8)
	inclusion := map[string]bool{}
	for _, field := range dto.BuiltInFields {
		inclusion[field.ID] = field.Included
	}
	assert.True(t, inclusion["name"])
	assert.True(t, inclusion["description"])
	assert.False(t, inclusion["experts"])
	assert.False(t, inclusion["realized-capabilities"], "relation built-ins are listed but excluded by default")
	assert.False(t, inclusion["built-by"])

	require.Len(t, dto.DisplayOrder, 4)
	assert.Equal(t, "custom", dto.DisplayOrder[1].Kind)
}

func TestBuildConfigurationDTO_CustomFieldsMergeDefinitionsWithPresentationPolicy(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), stakeholderActor())

	require.Len(t, dto.CustomFields, 3)
	selection := dto.CustomFields[0]
	assert.Equal(t, "Hosting model", selection.Name)
	assert.Equal(t, "selection", selection.Type)
	assert.Equal(t, "Where it runs", selection.HelpText)
	assert.True(t, selection.Required, "required-ness comes from the configuration document")
	assert.True(t, selection.Included)
	assert.True(t, selection.Active)
	require.Len(t, selection.Options, 2)
	assert.False(t, selection.Options[0].Active)

	retired := dto.CustomFields[1]
	assert.False(t, retired.Active)
	assert.False(t, retired.Included)
	assert.True(t, retired.Required, "dormant requirement is reported as recorded")

	number := dto.CustomFields[2]
	assert.False(t, number.Required, "definition without a requirement record defaults to optional")
	assert.True(t, number.Included)
	assert.Equal(t, floatPtr(0), number.Min)
	assert.Equal(t, floatPtr(5), number.Max)
}

func TestBuildConfigurationDTO_SelfAndAttributeSchemaLinksAlwaysPresent(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), stakeholderActor())

	self, ok := dto.Links["self"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application", self.Href)
	assert.Equal(t, "GET", self.Method)

	schema, ok := dto.Links["x-attribute-schema"]
	require.True(t, ok, "the configuration names the MetaModel schema surface that owns the custom fields")
	assert.Equal(t, "/api/v1/meta-model/subject-types/application/attributes", schema.Href)
	assert.Equal(t, "GET", schema.Method)
}

func TestBuildConfigurationDTO_NoWriteAffordancesWithoutWritePermission(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), stakeholderActor())

	assert.NotContains(t, dto.Links, "x-reorder")
	assert.NotContains(t, dto.Links, "x-impact-preview")
	for _, field := range dto.BuiltInFields {
		assert.Empty(t, field.Links, field.ID)
	}
	for _, field := range dto.CustomFields {
		assert.Empty(t, field.Links, field.ID)
	}
}

func TestBuildConfigurationDTO_WriteAffordancesForAdmin(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), adminActor())

	assert.NotContains(t, dto.Links, "x-define-custom-field", "schema definition moved to MetaModel")

	reorder, ok := dto.Links["x-reorder"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/display-order", reorder.Href)
	assert.Equal(t, "PUT", reorder.Method)

	preview, ok := dto.Links["x-impact-preview"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/impact-preview", preview.Href)
	assert.Equal(t, "GET", preview.Method)
}

func TestBuildConfigurationDTO_BuiltInFieldLinksReflectInclusion(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), adminActor())

	for _, field := range dto.BuiltInFields {
		if field.Included {
			exclude, ok := field.Links["x-exclude"]
			require.True(t, ok, field.ID)
			assert.Equal(t, "/api/v1/one-pagers/configurations/application/built-in-fields/"+field.ID+"/exclude", exclude.Href)
			assert.NotContains(t, field.Links, "x-include")
		} else {
			include, ok := field.Links["x-include"]
			require.True(t, ok, field.ID)
			assert.Equal(t, "POST", include.Method)
			assert.NotContains(t, field.Links, "x-exclude")
		}
	}
}

func builtInFieldByID(dto OnePagerConfigurationDTO, id string) (BuiltInFieldDTO, bool) {
	for _, field := range dto.BuiltInFields {
		if field.ID == id {
			return field, true
		}
	}
	return BuiltInFieldDTO{}, false
}

func TestBuildConfigurationDTO_BuiltInRequiredFlagReflectsDocument(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), adminActor())

	description, ok := builtInFieldByID(dto, "description")
	require.True(t, ok)
	assert.True(t, description.Required, "included required built-in reports required")

	name, ok := builtInFieldByID(dto, "name")
	require.True(t, ok)
	assert.False(t, name.Required, "built-in with no requirement record defaults to optional")
}

func TestBuildConfigurationDTO_IncludedBuiltInHasSetRequirementLinkForAdmin(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), adminActor())

	description, ok := builtInFieldByID(dto, "description")
	require.True(t, ok)
	link, ok := description.Links["x-set-requirement"]
	require.True(t, ok, "included built-in exposes x-set-requirement")
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/built-in-fields/description/requirement", link.Href)
	assert.Equal(t, "PUT", link.Method)
}

func TestBuildConfigurationDTO_ExcludedBuiltInHasNoSetRequirementLink(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), adminActor())

	experts, ok := builtInFieldByID(dto, "experts")
	require.True(t, ok)
	assert.NotContains(t, experts.Links, "x-set-requirement", "excluded built-in exposes no set-requirement affordance")
}

func TestBuildConfigurationDTO_CustomFieldLinksCarryOnlyPresentationAffordances(t *testing.T) {
	dto := BuildConfigurationDTO(applicationView(), testLinks(), adminActor())

	included := dto.CustomFields[0]
	assert.Equal(t, map[string]bool{"x-set-requirement": true}, linkRelations(included.Links))
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/custom-fields/"+included.ID+"/requirement", included.Links["x-set-requirement"].Href)
	assert.Equal(t, "PUT", included.Links["x-set-requirement"].Method)

	retired := dto.CustomFields[1]
	assert.Empty(t, retired.Links, "retired attributes are managed in MetaModel, not on the one-pager configuration")
}

func linkRelations(links map[string]sharedAPI.Link) map[string]bool {
	relations := map[string]bool{}
	for relation := range links {
		relations[relation] = true
	}
	return relations
}

func TestBuildConfigurationDTO_ExcludedCustomFieldHasNoSetRequirementLink(t *testing.T) {
	view := applicationView()
	view.Record.Document.DisplayOrder = view.Record.Document.DisplayOrder[:1]

	dto := BuildConfigurationDTO(view, testLinks(), adminActor())

	for _, field := range dto.CustomFields {
		assert.False(t, field.Included, field.ID)
		assert.Empty(t, field.Links, field.ID)
	}
}
