package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemorySchemaStore struct {
	records map[string]readmodels.SubjectAttributeSchemaRecord
}

func newInMemorySchemaStore() *inMemorySchemaStore {
	return &inMemorySchemaStore{records: map[string]readmodels.SubjectAttributeSchemaRecord{}}
}

func (s *inMemorySchemaStore) Insert(_ context.Context, record readmodels.SubjectAttributeSchemaRecord) error {
	s.records[record.ID] = record
	return nil
}

func (s *inMemorySchemaStore) GetByID(_ context.Context, id string) (*readmodels.SubjectAttributeSchemaRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, nil
	}
	return &record, nil
}

func (s *inMemorySchemaStore) Update(_ context.Context, params readmodels.UpdateSchemaParams) error {
	record := s.records[params.ID]
	record.Attributes = params.Attributes
	record.Version = params.Version
	record.ModifiedAt = params.ModifiedAt
	record.ModifiedBy = params.ModifiedBy
	s.records[params.ID] = record
	return nil
}

func newSchemaAggregate(t *testing.T) (*aggregates.SubjectAttributeSchema, valueobjects.UserEmail) {
	t.Helper()
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	subjectType, _ := valueobjects.NewSubjectType("application")
	steward, _ := valueobjects.NewUserEmail("steward@example.com")
	schema, err := aggregates.NewSubjectAttributeSchema(tenantID, subjectType, steward)
	require.NoError(t, err)
	return schema, steward
}

func project(t *testing.T, projector *SubjectAttributeSchemaProjector, schema *aggregates.SubjectAttributeSchema) {
	t.Helper()
	for _, event := range schema.GetUncommittedChanges() {
		require.NoError(t, projector.Handle(context.Background(), event))
	}
	schema.MarkChangesAsCommitted()
}

func TestSubjectAttributeSchemaProjector_ProjectsFullLifecycle(t *testing.T) {
	store := newInMemorySchemaStore()
	projector := NewSubjectAttributeSchemaProjector(store)
	schema, steward := newSchemaAggregate(t)
	name, _ := valueobjects.NewAttributeName("Region")
	selection, _ := valueobjects.NewAttributeType("selection")
	eu, _ := valueobjects.NewOptionLabel("EU")
	us, _ := valueobjects.NewOptionLabel("US")
	attributeID, err := schema.DefineAttribute(aggregates.DefineAttributeParams{Name: name, Type: selection, OptionLabels: []valueobjects.OptionLabel{eu}}, steward)
	require.NoError(t, err)
	project(t, projector, schema)

	record, _ := store.GetByID(context.Background(), schema.ID())
	require.NotNil(t, record)
	assert.Equal(t, "application", record.SubjectType)
	assert.Equal(t, 2, record.Version)
	require.Len(t, record.Attributes, 1)
	assert.Equal(t, readmodels.SubjectAttributeRecord{
		ID: attributeID.Value(), Name: "Region", Type: "selection", Active: true,
		Options: []readmodels.AttributeOptionRecord{{ID: record.Attributes[0].Options[0].ID, Label: "EU", Active: true}},
	}, record.Attributes[0])

	optionID, err := schema.AddOption(attributeID, us, steward)
	require.NoError(t, err)
	require.NoError(t, schema.RetireOption(attributeID, optionID, steward))
	help, _ := valueobjects.NewHelpText("Where it runs")
	renamed, _ := valueobjects.NewAttributeName("Hosting Region")
	require.NoError(t, schema.RenameAttribute(aggregates.RenameAttributeParams{AttributeID: attributeID, Name: renamed, HelpText: help}, steward))
	require.NoError(t, schema.RetireAttribute(attributeID, steward))
	project(t, projector, schema)

	record, _ = store.GetByID(context.Background(), schema.ID())
	attribute := record.Attributes[0]
	assert.Equal(t, "Hosting Region", attribute.Name)
	assert.Equal(t, "Where it runs", attribute.HelpText)
	assert.False(t, attribute.Active)
	require.Len(t, attribute.Options, 2)
	assert.False(t, attribute.Options[1].Active)
	assert.Equal(t, schema.Version(), record.Version)
	assert.Equal(t, "steward@example.com", record.ModifiedBy)

	require.NoError(t, schema.ReactivateAttribute(attributeID, steward))
	project(t, projector, schema)
	record, _ = store.GetByID(context.Background(), schema.ID())
	assert.True(t, record.Attributes[0].Active)
}

func TestSubjectAttributeSchemaProjector_ProjectsBounds(t *testing.T) {
	store := newInMemorySchemaStore()
	projector := NewSubjectAttributeSchemaProjector(store)
	schema, steward := newSchemaAggregate(t)
	name, _ := valueobjects.NewAttributeName("Score")
	number, _ := valueobjects.NewAttributeType("number")
	attributeID, err := schema.DefineAttribute(aggregates.DefineAttributeParams{Name: name, Type: number}, steward)
	require.NoError(t, err)
	min, max := 1.0, 5.0
	require.NoError(t, schema.SetBounds(attributeID, &min, &max, steward))
	project(t, projector, schema)

	record, _ := store.GetByID(context.Background(), schema.ID())
	assert.Equal(t, 1.0, *record.Attributes[0].Min)
	assert.Equal(t, 5.0, *record.Attributes[0].Max)
}

func TestSubjectAttributeSchemaProjector_IgnoresUnrelatedEvents(t *testing.T) {
	store := newInMemorySchemaStore()
	projector := NewSubjectAttributeSchemaProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "StrategyPillarAdded", []byte(`{}`)))

	assert.Empty(t, store.records)
}
