package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemoryDefinitionCache struct {
	saved       map[string]readmodels.CustomFieldRecord
	subjectType map[string]string
}

func newInMemoryDefinitionCache() *inMemoryDefinitionCache {
	return &inMemoryDefinitionCache{saved: map[string]readmodels.CustomFieldRecord{}, subjectType: map[string]string{}}
}

func (c *inMemoryDefinitionCache) Save(_ context.Context, definition readmodels.SubjectDefinition) error {
	c.saved[definition.Field.ID] = definition.Field
	c.subjectType[definition.Field.ID] = definition.SubjectType
	return nil
}

func (c *inMemoryDefinitionCache) Get(_ context.Context, fieldID string) (*readmodels.CustomFieldRecord, error) {
	field, ok := c.saved[fieldID]
	if !ok {
		return nil, nil
	}
	return &field, nil
}

func schemaBase(attributeID string) mmPL.SubjectAttributeEventBase {
	return mmPL.SubjectAttributeEventBase{
		ID: "schema-1", TenantID: "tenant-123", SubjectType: "application", Version: 2,
		AttributeID: attributeID, ModifiedAt: time.Now(), ModifiedBy: "steward@example.com",
	}
}

func projectPayload(t *testing.T, projector *CustomFieldDefinitionProjector, eventType string, payload any) {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, projector.ProjectEvent(context.Background(), eventType, data))
}

func TestCustomFieldDefinitionProjector_DefinedCachesDefinitionUnderSubjectType(t *testing.T) {
	cache := newInMemoryDefinitionCache()
	projector := NewCustomFieldDefinitionProjector(cache)
	min := 1.0

	projectPayload(t, projector, mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeDefinedPayload{
		SubjectAttributeEventBase: schemaBase("attr-1"),
		Name:                      "Hosting Region",
		AttributeType:             "selection",
		HelpText:                  "Where it runs",
		Options:                   []mmPL.SubjectAttributeOptionData{{ID: "opt-eu", Label: "EU", Active: true}, {ID: "opt-mars", Label: "Mars", Active: false}},
		Min:                       &min,
	})

	assert.Equal(t, "application", cache.subjectType["attr-1"])
	assert.Equal(t, readmodels.CustomFieldRecord{
		ID: "attr-1", Name: "Hosting Region", Type: "selection", HelpText: "Where it runs", Active: true,
		Options: []readmodels.OptionRecord{{ID: "opt-eu", Label: "EU", Active: true}, {ID: "opt-mars", Label: "Mars", Active: false}},
		Min:     &min,
	}, cache.saved["attr-1"])
}

func TestCustomFieldDefinitionProjector_MutationsUpdateCachedDefinition(t *testing.T) {
	cache := newInMemoryDefinitionCache()
	projector := NewCustomFieldDefinitionProjector(cache)
	projectPayload(t, projector, mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeDefinedPayload{
		SubjectAttributeEventBase: schemaBase("attr-1"), Name: "Region", AttributeType: "selection",
		Options: []mmPL.SubjectAttributeOptionData{{ID: "opt-eu", Label: "EU", Active: true}},
	})

	projectPayload(t, projector, mmPL.SubjectAttributeRenamed, mmPL.SubjectAttributeRenamedPayload{SubjectAttributeEventBase: schemaBase("attr-1"), NewName: "Hosting Region", NewHelpText: "help"})
	projectPayload(t, projector, mmPL.SubjectAttributeOptionAdded, mmPL.SubjectAttributeOptionAddedPayload{SubjectAttributeEventBase: schemaBase("attr-1"), OptionID: "opt-us", Label: "US"})
	projectPayload(t, projector, mmPL.SubjectAttributeOptionRetired, mmPL.SubjectAttributeOptionRetiredPayload{SubjectAttributeEventBase: schemaBase("attr-1"), OptionID: "opt-eu"})
	projectPayload(t, projector, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeRetiredPayload{SubjectAttributeEventBase: schemaBase("attr-1")})

	field := cache.saved["attr-1"]
	assert.Equal(t, "Hosting Region", field.Name)
	assert.Equal(t, "help", field.HelpText)
	assert.False(t, field.Active)
	assert.Equal(t, []readmodels.OptionRecord{{ID: "opt-eu", Label: "EU", Active: false}, {ID: "opt-us", Label: "US", Active: true}}, field.Options)

	projectPayload(t, projector, mmPL.SubjectAttributeReactivated, mmPL.SubjectAttributeReactivatedPayload{SubjectAttributeEventBase: schemaBase("attr-1")})
	assert.True(t, cache.saved["attr-1"].Active)
}

func TestCustomFieldDefinitionProjector_BoundsChanged(t *testing.T) {
	cache := newInMemoryDefinitionCache()
	projector := NewCustomFieldDefinitionProjector(cache)
	projectPayload(t, projector, mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeDefinedPayload{SubjectAttributeEventBase: schemaBase("attr-n"), Name: "Score", AttributeType: "number"})
	max := 5.0

	projectPayload(t, projector, mmPL.SubjectAttributeBoundsChanged, mmPL.SubjectAttributeBoundsChangedPayload{SubjectAttributeEventBase: schemaBase("attr-n"), Max: &max})

	assert.Nil(t, cache.saved["attr-n"].Min)
	assert.Equal(t, 5.0, *cache.saved["attr-n"].Max)
}

func TestCustomFieldDefinitionProjector_MutationOnUnknownFieldIsSkipped(t *testing.T) {
	cache := newInMemoryDefinitionCache()
	projector := NewCustomFieldDefinitionProjector(cache)

	projectPayload(t, projector, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeRetiredPayload{SubjectAttributeEventBase: schemaBase("ghost")})

	assert.Empty(t, cache.saved)
}

func TestCustomFieldDefinitionProjector_IgnoresUnrelatedEvents(t *testing.T) {
	cache := newInMemoryDefinitionCache()
	projector := NewCustomFieldDefinitionProjector(cache)

	require.NoError(t, projector.ProjectEvent(context.Background(), mmPL.StrategyPillarAdded, []byte(`{}`)))

	assert.Empty(t, cache.saved)
}

func TestCustomFieldDefinitionEventTypes_CoverEveryPublishedSchemaEvent(t *testing.T) {
	assert.ElementsMatch(t, mmPL.SubjectAttributeEventTypes(), CustomFieldDefinitionEventTypes())
}
