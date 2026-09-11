package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/aggregates"
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]domain.DomainEvent
}

func newInMemoryEventStore() *inMemoryEventStore {
	return &inMemoryEventStore{events: make(map[string][]domain.DomainEvent)}
}

func (s *inMemoryEventStore) SaveEvents(_ context.Context, aggregateID string, events []domain.DomainEvent, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.events[aggregateID]) != expectedVersion {
		return domain.ErrConcurrencyConflict
	}
	for _, evt := range events {
		jsonData, _ := json.Marshal(evt.EventData())
		s.events[aggregateID] = append(s.events[aggregateID], domain.NewGenericDomainEvent(aggregateID, evt.EventType(), jsonData, evt.OccurredAt()))
	}
	return nil
}

func (s *inMemoryEventStore) GetEvents(_ context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.DomainEvent, len(s.events[aggregateID]))
	copy(result, s.events[aggregateID])
	return result, nil
}

type fakeSchemaLookup struct {
	records map[string]*readmodels.SubjectAttributeSchemaRecord
}

func (f *fakeSchemaLookup) GetBySubjectType(_ context.Context, subjectType string) (*readmodels.SubjectAttributeSchemaRecord, error) {
	return f.records[subjectType], nil
}

func newSchemaRepo() *repositories.SubjectAttributeSchemaRepository {
	return repositories.NewSubjectAttributeSchemaRepository(newInMemoryEventStore())
}

func createSchema(t *testing.T, repo *repositories.SubjectAttributeSchemaRepository, lookup *fakeSchemaLookup) string {
	t.Helper()
	result, err := NewCreateSubjectAttributeSchemaHandler(repo, lookup).Handle(context.Background(), &commands.CreateSubjectAttributeSchema{
		TenantID: "tenant-123", SubjectType: "application", CreatedBy: "steward@example.com",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.CreatedID)
	lookup.records["application"] = &readmodels.SubjectAttributeSchemaRecord{ID: result.CreatedID, SubjectType: "application"}
	return result.CreatedID
}

func TestCreateSubjectAttributeSchemaHandler_RejectsDuplicate(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	createSchema(t, repo, lookup)

	_, err := NewCreateSubjectAttributeSchemaHandler(repo, lookup).Handle(context.Background(), &commands.CreateSubjectAttributeSchema{
		TenantID: "tenant-123", SubjectType: "application", CreatedBy: "steward@example.com",
	})

	assert.ErrorIs(t, err, ErrSchemaAlreadyExists)
}

func TestDefineSubjectAttributeHandler_DefinesAndReturnsAttributeID(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	schemaID := createSchema(t, repo, lookup)
	min := 1.0

	result, err := NewDefineSubjectAttributeHandler(repo).Handle(context.Background(), &commands.DefineSubjectAttribute{
		SchemaID: schemaID, Name: "Score", AttributeType: "number", Min: &min, ModifiedBy: "steward@example.com",
	})

	require.NoError(t, err)
	schema, err := repo.GetByID(context.Background(), schemaID)
	require.NoError(t, err)
	require.Len(t, schema.Attributes(), 1)
	assert.Equal(t, result.CreatedID, schema.Attributes()[0].ID().Value())
	assert.Equal(t, 1.0, *schema.Attributes()[0].Min())
}

func TestSchemaModificationHandlers_FullLifecycle(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	schemaID := createSchema(t, repo, lookup)
	ctx := context.Background()
	by := "steward@example.com"

	defined, err := NewDefineSubjectAttributeHandler(repo).Handle(ctx, &commands.DefineSubjectAttribute{
		SchemaID: schemaID, Name: "Region", AttributeType: "selection", OptionLabels: []string{"EU"}, ModifiedBy: by,
	})
	require.NoError(t, err)
	attributeID := defined.CreatedID

	_, err = NewRenameSubjectAttributeHandler(repo).Handle(ctx, &commands.RenameSubjectAttribute{SchemaID: schemaID, AttributeID: attributeID, Name: "Hosting Region", HelpText: "help", ModifiedBy: by})
	require.NoError(t, err)
	added, err := NewAddSubjectAttributeOptionHandler(repo).Handle(ctx, &commands.AddSubjectAttributeOption{SchemaID: schemaID, AttributeID: attributeID, Label: "US", ModifiedBy: by})
	require.NoError(t, err)
	_, err = NewRetireSubjectAttributeOptionHandler(repo).Handle(ctx, &commands.RetireSubjectAttributeOption{SchemaID: schemaID, AttributeID: attributeID, OptionID: added.CreatedID, ModifiedBy: by})
	require.NoError(t, err)
	_, err = NewRetireSubjectAttributeHandler(repo).Handle(ctx, &commands.RetireSubjectAttribute{SchemaID: schemaID, AttributeID: attributeID, ModifiedBy: by})
	require.NoError(t, err)
	_, err = NewReactivateSubjectAttributeHandler(repo).Handle(ctx, &commands.ReactivateSubjectAttribute{SchemaID: schemaID, AttributeID: attributeID, ModifiedBy: by})
	require.NoError(t, err)
	min := 1.0
	_, err = NewSetSubjectAttributeBoundsHandler(repo).Handle(ctx, &commands.SetSubjectAttributeBounds{SchemaID: schemaID, AttributeID: attributeID, Min: &min, ModifiedBy: by})
	assert.ErrorIs(t, err, valueobjects.ErrBoundsNotAllowed, "bounds on a selection attribute are rejected")
	_, err = NewSetSubjectAttributeBoundsHandler(repo).Handle(ctx, &commands.SetSubjectAttributeBounds{SchemaID: schemaID, AttributeID: "cccccccc-cccc-cccc-cccc-cccccccccccc", Min: &min, ModifiedBy: by})
	assert.ErrorIs(t, err, aggregates.ErrAttributeNotFound, "bounds on an unknown attribute id are rejected")

	schema, err := repo.GetByID(ctx, schemaID)
	require.NoError(t, err)
	attribute := schema.Attributes()[0]
	assert.Equal(t, "Hosting Region", attribute.Name().Value())
	assert.True(t, attribute.IsActive())
	assert.Len(t, attribute.Options(), 2)
	assert.False(t, attribute.Options()[1].IsActive())
}

func TestSetSubjectAttributeBoundsHandler(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	schemaID := createSchema(t, repo, lookup)
	ctx := context.Background()
	defined, err := NewDefineSubjectAttributeHandler(repo).Handle(ctx, &commands.DefineSubjectAttribute{SchemaID: schemaID, Name: "Score", AttributeType: "number", ModifiedBy: "steward@example.com"})
	require.NoError(t, err)
	max := 9.0

	_, err = NewSetSubjectAttributeBoundsHandler(repo).Handle(ctx, &commands.SetSubjectAttributeBounds{SchemaID: schemaID, AttributeID: defined.CreatedID, Max: &max, ModifiedBy: "steward@example.com"})

	require.NoError(t, err)
	schema, _ := repo.GetByID(ctx, schemaID)
	assert.Equal(t, 9.0, *schema.Attributes()[0].Max())
}

func importCommand(attributeID string) *mmPL.ImportSubjectAttribute {
	return &mmPL.ImportSubjectAttribute{
		TenantID:    "tenant-123",
		SubjectType: "vendor",
		ImportedBy:  "system@example.com",
		Attribute: mmPL.ImportedAttribute{
			ID:       attributeID,
			Name:     "Hosting Region",
			Type:     "selection",
			HelpText: "Where it runs",
			Options: []mmPL.ImportedOption{
				{ID: "11111111-1111-1111-1111-111111111111", Label: "EU", Active: true},
				{ID: "22222222-2222-2222-2222-222222222222", Label: "Mars", Active: false},
			},
			Active: false,
		},
	}
}

func TestImportSubjectAttributeHandler_CreatesSchemaAndPreservesIdentity(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	handler := NewImportSubjectAttributeHandler(repo, lookup)
	attributeID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	result, err := handler.Handle(context.Background(), importCommand(attributeID))

	require.NoError(t, err)
	require.NotEmpty(t, result.CreatedID)
	schema, err := repo.GetByID(context.Background(), result.CreatedID)
	require.NoError(t, err)
	assert.Equal(t, "vendor", schema.SubjectType().Value())
	require.Len(t, schema.Attributes(), 1)
	imported := schema.Attributes()[0]
	assert.Equal(t, attributeID, imported.ID().Value())
	assert.False(t, imported.IsActive())
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", imported.Options()[0].ID().Value())
	assert.False(t, imported.Options()[1].IsActive())
	assert.Equal(t, "Where it runs", imported.HelpText().Value())
}

func TestImportSubjectAttributeHandler_IsIdempotentOnExistingSchema(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	handler := NewImportSubjectAttributeHandler(repo, lookup)
	attributeID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	first, err := handler.Handle(context.Background(), importCommand(attributeID))
	require.NoError(t, err)
	lookup.records["vendor"] = &readmodels.SubjectAttributeSchemaRecord{ID: first.CreatedID, SubjectType: "vendor"}

	second, err := handler.Handle(context.Background(), importCommand(attributeID))

	require.NoError(t, err)
	assert.Equal(t, first.CreatedID, second.CreatedID)
	schema, _ := repo.GetByID(context.Background(), first.CreatedID)
	assert.Len(t, schema.Attributes(), 1)
	assert.Equal(t, 3, schema.Version(), "created, defined, retired — no further events on re-import")
}

func TestImportSubjectAttributeHandler_ImportsBounds(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	min, max := 0.0, 5.0
	cmd := &mmPL.ImportSubjectAttribute{
		TenantID: "tenant-123", SubjectType: "application", ImportedBy: "system@example.com",
		Attribute: mmPL.ImportedAttribute{ID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", Name: "Score", Type: "number", Min: &min, Max: &max, Active: true},
	}

	result, err := NewImportSubjectAttributeHandler(repo, lookup).Handle(context.Background(), cmd)

	require.NoError(t, err)
	schema, _ := repo.GetByID(context.Background(), result.CreatedID)
	assert.Equal(t, 5.0, *schema.Attributes()[0].Max())
	assert.True(t, schema.Attributes()[0].IsActive())
}

func TestHandlers_RejectWrongCommandType(t *testing.T) {
	repo := newSchemaRepo()
	lookup := &fakeSchemaLookup{records: map[string]*readmodels.SubjectAttributeSchemaRecord{}}
	for _, handler := range []cqrs.CommandHandler{
		NewCreateSubjectAttributeSchemaHandler(repo, lookup),
		NewDefineSubjectAttributeHandler(repo),
		NewImportSubjectAttributeHandler(repo, lookup),
	} {
		_, err := handler.Handle(context.Background(), &commands.AddStrategyPillar{})
		assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
	}
}
