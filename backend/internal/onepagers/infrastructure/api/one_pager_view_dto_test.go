package api

import (
	"encoding/json"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOnePagerDTO_RendersReferenceListAsReferencesVariant(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "capability",
		SubjectID:   "cap-1",
		SubjectName: "Billing",
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "realizing-applications", Label: "Realizing Applications", Value: ports.ReferenceListValue{References: []ports.Reference{
				{ID: "app-1", Label: "Billing Service", SubjectType: "application"},
				{ID: "dom-1", Label: "Payments", SubjectType: ""},
			}}}},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	require.Len(t, dto.Fields, 1)
	value := dto.Fields[0].BuiltIn.Value
	require.NotNil(t, value)
	assert.Equal(t, "references", value.Type)
	require.Len(t, value.References, 2)
	assert.Equal(t, ReferenceDTO{ID: "app-1", Label: "Billing Service", SubjectType: "application"}, value.References[0])
	assert.Equal(t, ReferenceDTO{ID: "dom-1", Label: "Payments"}, value.References[1])
}

func TestBuildOnePagerDTO_EmptyReferenceListRendersNullValue(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "capability",
		SubjectID:   "cap-1",
		SubjectName: "Billing",
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "depends-on", Label: "Depends On", Value: ports.ReferenceListValue{References: nil}}},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	require.Len(t, dto.Fields, 1)
	assert.Nil(t, dto.Fields[0].BuiltIn.Value)
}

func TestBuildOnePagerDTO_ReferencesJSONShape(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "capability",
		SubjectID:   "cap-1",
		SubjectName: "Billing",
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "business-domains", Label: "Business Domains", Value: ports.ReferenceListValue{References: []ports.Reference{
				{ID: "dom-1", Label: "Payments"},
			}}}},
		},
	}

	body, err := json.Marshal(BuildOnePagerDTO(onePager, testLinks(), stakeholderActor()))
	require.NoError(t, err)
	assert.Contains(t, string(body), `"type":"references"`)
	assert.Contains(t, string(body), `"references":[{"id":"dom-1","label":"Payments"}]`)
}

func TestBuildOnePagerDTO_SkipsZeroValueField(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   "app-1",
		SubjectName: "Payments",
		Fields: []queries.Field{
			{BuiltIn: &queries.BuiltInField{ID: "description", Label: "Description"}},
			{},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	require.Len(t, dto.Fields, 1)
	assert.Equal(t, "builtIn", dto.Fields[0].Kind)
}

func TestBuildOnePagerDTO_MapsCompletenessCounts(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   "app-1",
		SubjectName: "Payments",
		Completeness: queries.Completeness{
			RequiredCount: 2,
			FilledCount:   1,
			MissingFields: []queries.MissingField{{FieldID: "contract-link", Name: "Contract link"}},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	assert.Equal(t, 2, dto.Completeness.RequiredCount)
	assert.Equal(t, 1, dto.Completeness.FilledCount)
	require.Len(t, dto.Completeness.MissingFields, 1)
	assert.Equal(t, "contract-link", dto.Completeness.MissingFields[0].FieldID)
	assert.Equal(t, "Contract link", dto.Completeness.MissingFields[0].Name)
}

func TestBuildOnePagerDTO_CompletenessJSONFieldNames(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   "app-1",
		SubjectName: "Payments",
		Completeness: queries.Completeness{
			RequiredCount: 2,
			FilledCount:   1,
			MissingFields: []queries.MissingField{{FieldID: "contract-link", Name: "Contract link"}},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	body, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &decoded))

	completeness, ok := decoded["completeness"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), completeness["requiredCount"])
	assert.Equal(t, float64(1), completeness["filledCount"])

	missingFields, ok := completeness["missingFields"].([]interface{})
	require.True(t, ok)
	require.Len(t, missingFields, 1)
	missingField, ok := missingFields[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "contract-link", missingField["fieldId"])
	assert.Equal(t, "Contract link", missingField["name"])
}

func TestBuildOnePagerDTO_CustomFieldCarriesOutOfBoundsFlag(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType: "application",
		SubjectID:   "app-1",
		SubjectName: "Payments",
		Fields: []queries.Field{
			{Custom: &queries.CustomField{FieldID: "maturity", Name: "Maturity score", FieldType: "number", OutOfBounds: true}},
		},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	require.Len(t, dto.Fields, 1)
	require.NotNil(t, dto.Fields[0].Custom)
	assert.True(t, dto.Fields[0].Custom.OutOfBounds)
}

func TestBuildOnePagerDTO_EmptyMissingFieldsSerializesAsEmptyArrayNotNull(t *testing.T) {
	onePager := &queries.OnePager{
		SubjectType:  "application",
		SubjectID:    "app-1",
		SubjectName:  "Payments",
		Completeness: queries.Completeness{RequiredCount: 2, FilledCount: 2},
	}

	dto := BuildOnePagerDTO(onePager, testLinks(), stakeholderActor())

	body, err := json.Marshal(dto)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"missingFields":[]`)
	assert.NotContains(t, string(body), `"missingFields":null`)
}
