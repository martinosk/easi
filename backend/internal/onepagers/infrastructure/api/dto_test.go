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

func applicationRecord() *readmodels.ConfigurationRecord {
	now := time.Now().UTC()
	return &readmodels.ConfigurationRecord{
		ID:          "config-1",
		TenantID:    "tenant-123",
		SubjectType: "application",
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{
				{
					ID: "9f0d5e69-0000-0000-0000-000000000001", Name: "Hosting model", Type: "selection",
					Required: true, HelpText: "Where it runs", Active: true,
					Options: []readmodels.OptionRecord{
						{ID: "9f0d5e69-0000-0000-0000-00000000000a", Label: "On-prem", Active: false},
						{ID: "9f0d5e69-0000-0000-0000-00000000000b", Label: "Cloud", Active: true},
					},
				},
				{
					ID: "9f0d5e69-0000-0000-0000-000000000002", Name: "Old field", Type: "text", Active: false,
				},
				{
					ID: "9f0d5e69-0000-0000-0000-000000000004", Name: "Maturity score", Type: "number", Active: true,
					Min: floatPtr(0), Max: floatPtr(5),
				},
			},
			BuiltInFields: []readmodels.BuiltInFieldRecord{
				{ID: "description", Required: true},
				{ID: "experts", Required: true},
			},
			DisplayOrder: []readmodels.FieldRefRecord{
				{Kind: "builtIn", ID: "name"},
				{Kind: "custom", ID: "9f0d5e69-0000-0000-0000-000000000001"},
				{Kind: "builtIn", ID: "description"},
			},
		},
		Version:    4,
		CreatedAt:  now,
		ModifiedAt: now,
		ModifiedBy: "admin@example.com",
	}
}

func TestBuildConfigurationDTO_MergesCatalogWithInclusionState(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), stakeholderActor())

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

	require.Len(t, dto.CustomFields, 3)
	assert.Equal(t, "Hosting model", dto.CustomFields[0].Name)
	assert.False(t, dto.CustomFields[1].Active)
	assert.Equal(t, floatPtr(0), dto.CustomFields[2].Min)
	assert.Equal(t, floatPtr(5), dto.CustomFields[2].Max)

	require.Len(t, dto.DisplayOrder, 3)
	assert.Equal(t, "custom", dto.DisplayOrder[1].Kind)
}

func TestBuildConfigurationDTO_SelfLinkAlwaysPresent(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), stakeholderActor())

	self, ok := dto.Links["self"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application", self.Href)
	assert.Equal(t, "GET", self.Method)
}

func TestBuildConfigurationDTO_NoWriteAffordancesWithoutWritePermission(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), stakeholderActor())

	assert.NotContains(t, dto.Links, "x-define-custom-field")
	assert.NotContains(t, dto.Links, "x-reorder")
	assert.NotContains(t, dto.Links, "x-impact-preview")
	for _, field := range dto.BuiltInFields {
		assert.Empty(t, field.Links, field.ID)
	}
	for _, field := range dto.CustomFields {
		assert.Empty(t, field.Links, field.ID)
		for _, option := range field.Options {
			assert.Empty(t, option.Links)
		}
	}
}

func TestBuildConfigurationDTO_WriteAffordancesForAdmin(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	define, ok := dto.Links["x-define-custom-field"]
	require.True(t, ok)
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/custom-fields", define.Href)
	assert.Equal(t, "POST", define.Method)

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
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

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
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	description, ok := builtInFieldByID(dto, "description")
	require.True(t, ok)
	assert.True(t, description.Required, "included required built-in reports required")

	name, ok := builtInFieldByID(dto, "name")
	require.True(t, ok)
	assert.False(t, name.Required, "built-in with no requirement record defaults to optional")
}

func TestBuildConfigurationDTO_IncludedBuiltInHasSetRequirementLinkForAdmin(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	description, ok := builtInFieldByID(dto, "description")
	require.True(t, ok)
	link, ok := description.Links["x-set-requirement"]
	require.True(t, ok, "included built-in exposes x-set-requirement")
	assert.Equal(t, "/api/v1/one-pagers/configurations/application/built-in-fields/description/requirement", link.Href)
	assert.Equal(t, "PUT", link.Method)
}

func TestBuildConfigurationDTO_ExcludedBuiltInHasNoSetRequirementLink(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	experts, ok := builtInFieldByID(dto, "experts")
	require.True(t, ok)
	assert.NotContains(t, experts.Links, "x-set-requirement", "excluded built-in exposes no set-requirement affordance")
}

func TestBuildConfigurationDTO_NoSetRequirementLinkWithoutWritePermission(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), stakeholderActor())

	description, ok := builtInFieldByID(dto, "description")
	require.True(t, ok)
	assert.NotContains(t, description.Links, "x-set-requirement")
}

func TestBuildConfigurationDTO_ActiveCustomFieldLinks(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	active := dto.CustomFields[0]
	base := "/api/v1/one-pagers/configurations/application/custom-fields/" + active.ID
	assert.Equal(t, base, active.Links["x-rename"].Href)
	assert.Equal(t, "PUT", active.Links["x-rename"].Method)
	assert.Equal(t, base+"/requirement", active.Links["x-set-requirement"].Href)
	assert.Equal(t, base+"/retire", active.Links["x-retire"].Href)
	assert.Equal(t, base+"/options", active.Links["x-add-option"].Href)
	assert.NotContains(t, active.Links, "x-reactivate")

	retiredOption := active.Options[0]
	assert.Empty(t, retiredOption.Links)
	activeOption := active.Options[1]
	assert.Equal(t, base+"/options/"+activeOption.ID+"/retire", activeOption.Links["x-retire"].Href)
}

func TestBuildConfigurationDTO_RetiredCustomFieldOnlyReactivate(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	retired := dto.CustomFields[1]
	reactivate, ok := retired.Links["x-reactivate"]
	require.True(t, ok)
	assert.Equal(t, "POST", reactivate.Method)
	assert.NotContains(t, retired.Links, "x-rename")
	assert.NotContains(t, retired.Links, "x-retire")
	assert.NotContains(t, retired.Links, "x-add-option")
}

func TestBuildConfigurationDTO_NonSelectionFieldHasNoAddOption(t *testing.T) {
	record := applicationRecord()
	record.Document.CustomFields[0].Type = "text"
	record.Document.CustomFields[0].Options = nil

	dto := BuildConfigurationDTO(record, testLinks(), adminActor())

	assert.NotContains(t, dto.CustomFields[0].Links, "x-add-option")
}

func TestBuildConfigurationDTO_NumberFieldHasSetBoundsLink(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	numberField := dto.CustomFields[2]
	base := "/api/v1/one-pagers/configurations/application/custom-fields/" + numberField.ID
	assert.Equal(t, base+"/bounds", numberField.Links["x-set-bounds"].Href)
	assert.Equal(t, "PUT", numberField.Links["x-set-bounds"].Method)
}

func TestBuildConfigurationDTO_NonNumberFieldHasNoSetBoundsLink(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), adminActor())

	assert.NotContains(t, dto.CustomFields[0].Links, "x-set-bounds")
}

func TestBuildConfigurationDTO_NoSetBoundsLinkWithoutWritePermission(t *testing.T) {
	dto := BuildConfigurationDTO(applicationRecord(), testLinks(), stakeholderActor())

	assert.NotContains(t, dto.CustomFields[2].Links, "x-set-bounds")
}
