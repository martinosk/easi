package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
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

	existing := s.events[aggregateID]
	if len(existing) != expectedVersion {
		return domain.ErrConcurrencyConflict
	}
	for _, evt := range events {
		jsonData, _ := json.Marshal(evt.EventData())
		stored := domain.NewGenericDomainEvent(aggregateID, evt.EventType(), jsonData, evt.OccurredAt())
		s.events[aggregateID] = append(s.events[aggregateID], stored)
	}
	return nil
}

func (s *inMemoryEventStore) GetEvents(_ context.Context, aggregateID string) ([]domain.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[aggregateID]
	result := make([]domain.DomainEvent, len(events))
	copy(result, events)
	return result, nil
}

type fakeLookup struct {
	exists bool
	err    error
}

func (f *fakeLookup) ConfigurationExists(_ context.Context, _ string) (bool, error) {
	return f.exists, f.err
}

func newTestRepo() *repositories.OnePagerConfigurationRepository {
	return repositories.NewOnePagerConfigurationRepository(newInMemoryEventStore())
}

func createConfiguration(t *testing.T, repo *repositories.OnePagerConfigurationRepository) string {
	t.Helper()
	handler := NewCreateOnePagerConfigurationHandler(repo, &fakeLookup{})
	result, err := handler.Handle(context.Background(), &commands.CreateOnePagerConfiguration{
		TenantID:    "tenant-123",
		SubjectType: "application",
		CreatedBy:   "admin@example.com",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.CreatedID)
	return result.CreatedID
}

func loadConfig(t *testing.T, repo *repositories.OnePagerConfigurationRepository, id string) *aggregates.OnePagerConfiguration {
	t.Helper()
	config, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	return config
}

func defineTextField(t *testing.T, repo *repositories.OnePagerConfigurationRepository, configID, name string) string {
	t.Helper()
	handler := NewDefineCustomFieldHandler(repo)
	result, err := handler.Handle(context.Background(), &commands.DefineCustomField{
		ConfigID:   configID,
		Name:       name,
		FieldType:  "text",
		ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	return result.CreatedID
}

func TestCreateOnePagerConfigurationHandler_CreatesWithDefaults(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)

	config := loadConfig(t, repo, configID)
	assert.Equal(t, "application", config.SubjectType().Value())
	assert.Len(t, config.DisplayOrder(), 3)
	assert.Empty(t, config.CustomFields())
}

func TestCreateOnePagerConfigurationHandler_ErrorPaths(t *testing.T) {
	lookupErr := errors.New("boom")
	cases := []struct {
		name        string
		lookup      *fakeLookup
		subjectType string
		wantErr     error
	}{
		{"rejects existing subject type configuration", &fakeLookup{exists: true}, "application", ErrConfigurationAlreadyExists},
		{"propagates lookup error", &fakeLookup{err: lookupErr}, "application", lookupErr},
		{"rejects invalid subject type", &fakeLookup{}, "starship", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewCreateOnePagerConfigurationHandler(newTestRepo(), tc.lookup)

			_, err := handler.Handle(context.Background(), &commands.CreateOnePagerConfiguration{
				TenantID:    "tenant-123",
				SubjectType: tc.subjectType,
				CreatedBy:   "admin@example.com",
			})

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCreateOnePagerConfigurationHandler_RejectsWrongCommandType(t *testing.T) {
	repo := newTestRepo()
	handler := NewCreateOnePagerConfigurationHandler(repo, &fakeLookup{})

	_, err := handler.Handle(context.Background(), &commands.DefineCustomField{})

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDefineCustomFieldHandler_AddsFieldAndReturnsFieldID(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)

	handler := NewDefineCustomFieldHandler(repo)
	result, err := handler.Handle(context.Background(), &commands.DefineCustomField{
		ConfigID:     configID,
		Name:         "Hosting model",
		FieldType:    "selection",
		Required:     true,
		HelpText:     "Where it runs",
		OptionLabels: []string{"On-prem", "Cloud"},
		ModifiedBy:   "admin@example.com",
	})

	require.NoError(t, err)
	require.NotEmpty(t, result.CreatedID)

	config := loadConfig(t, repo, configID)
	fields := config.CustomFields()
	require.Len(t, fields, 1)
	assert.Equal(t, result.CreatedID, fields[0].ID().Value())
	assert.Equal(t, "selection", fields[0].Type().Value())
	assert.True(t, fields[0].IsRequired())
	assert.Len(t, fields[0].Options(), 2)
}

func TestDefineCustomFieldHandler_RejectsInvalidFieldType(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)

	handler := NewDefineCustomFieldHandler(repo)
	_, err := handler.Handle(context.Background(), &commands.DefineCustomField{
		ConfigID:   configID,
		Name:       "Broken",
		FieldType:  "checkbox",
		ModifiedBy: "admin@example.com",
	})

	assert.Error(t, err)
}

func TestRenameCustomFieldHandler_RenamesField(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := defineTextField(t, repo, configID, "Contract")

	handler := NewRenameCustomFieldHandler(repo)
	_, err := handler.Handle(context.Background(), &commands.RenameCustomField{
		ConfigID:   configID,
		FieldID:    fieldID,
		Name:       "Contract link",
		HelpText:   "URL of the contract",
		ModifiedBy: "admin@example.com",
	})

	require.NoError(t, err)
	config := loadConfig(t, repo, configID)
	assert.Equal(t, "Contract link", config.CustomFields()[0].Name().Value())
}

func TestRenameCustomFieldHandler_RejectsTypeChange(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := defineTextField(t, repo, configID, "Contract")

	handler := NewRenameCustomFieldHandler(repo)
	_, err := handler.Handle(context.Background(), &commands.RenameCustomField{
		ConfigID:      configID,
		FieldID:       fieldID,
		Name:          "Contract link",
		RequestedType: "link",
		ModifiedBy:    "admin@example.com",
	})

	assert.ErrorIs(t, err, aggregates.ErrFieldTypeImmutable)
}

func TestChangeRequirementAndRetireReactivateHandlers(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := defineTextField(t, repo, configID, "Product owner")
	ctx := context.Background()

	_, err := NewChangeCustomFieldRequirementHandler(repo).Handle(ctx, &commands.ChangeCustomFieldRequirement{
		ConfigID: configID, FieldID: fieldID, Required: true, ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.True(t, loadConfig(t, repo, configID).CustomFields()[0].IsRequired())

	_, err = NewRetireCustomFieldHandler(repo).Handle(ctx, &commands.RetireCustomField{
		ConfigID: configID, FieldID: fieldID, ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.False(t, loadConfig(t, repo, configID).CustomFields()[0].IsActive())

	_, err = NewReactivateCustomFieldHandler(repo).Handle(ctx, &commands.ReactivateCustomField{
		ConfigID: configID, FieldID: fieldID, ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.True(t, loadConfig(t, repo, configID).CustomFields()[0].IsActive())
}

func TestBuiltInFieldHandlers_ExcludeAndInclude(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	ctx := context.Background()

	_, err := NewExcludeBuiltInFieldHandler(repo).Handle(ctx, &commands.ExcludeBuiltInField{
		ConfigID: configID, EntryID: "experts", ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.Len(t, loadConfig(t, repo, configID).DisplayOrder(), 2)

	_, err = NewIncludeBuiltInFieldHandler(repo).Handle(ctx, &commands.IncludeBuiltInField{
		ConfigID: configID, EntryID: "experts", ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.Len(t, loadConfig(t, repo, configID).DisplayOrder(), 3)
}

func TestReorderFieldsHandler_AppliesNewOrder(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := defineTextField(t, repo, configID, "Contract link")

	handler := NewReorderOnePagerFieldsHandler(repo)
	_, err := handler.Handle(context.Background(), &commands.ReorderOnePagerFields{
		ConfigID: configID,
		Order: []commands.FieldRefInput{
			{Kind: "builtIn", ID: "name"},
			{Kind: "custom", ID: fieldID},
			{Kind: "builtIn", ID: "description"},
			{Kind: "builtIn", ID: "experts"},
		},
		ModifiedBy: "admin@example.com",
	})

	require.NoError(t, err)
	order := loadConfig(t, repo, configID).DisplayOrder()
	assert.Equal(t, fieldID, order[1].RefID())
}

func TestSelectionOptionHandlers_AddAndRetire(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	ctx := context.Background()

	defineHandler := NewDefineCustomFieldHandler(repo)
	defineResult, err := defineHandler.Handle(ctx, &commands.DefineCustomField{
		ConfigID:     configID,
		Name:         "Hosting model",
		FieldType:    "selection",
		OptionLabels: []string{"On-prem", "Cloud"},
		ModifiedBy:   "admin@example.com",
	})
	require.NoError(t, err)
	fieldID := defineResult.CreatedID

	addResult, err := NewAddSelectionOptionHandler(repo).Handle(ctx, &commands.AddSelectionOption{
		ConfigID: configID, FieldID: fieldID, Label: "Hybrid", ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	require.NotEmpty(t, addResult.CreatedID)

	config := loadConfig(t, repo, configID)
	options := config.CustomFields()[0].Options()
	require.Len(t, options, 3)

	_, err = NewRetireSelectionOptionHandler(repo).Handle(ctx, &commands.RetireSelectionOption{
		ConfigID: configID, FieldID: fieldID, OptionID: options[0].ID().Value(), ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)

	config = loadConfig(t, repo, configID)
	assert.False(t, config.CustomFields()[0].Options()[0].IsActive())
}

func floatPtr(v float64) *float64 {
	return &v
}

func defineNumberField(t *testing.T, repo *repositories.OnePagerConfigurationRepository, configID, name string) string {
	t.Helper()
	handler := NewDefineCustomFieldHandler(repo)
	result, err := handler.Handle(context.Background(), &commands.DefineCustomField{
		ConfigID:   configID,
		Name:       name,
		FieldType:  "number",
		ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	return result.CreatedID
}

func TestSetNumberFieldBoundsHandler_SetsBounds(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := defineNumberField(t, repo, configID, "Maturity score")

	handler := NewSetNumberFieldBoundsHandler(repo)
	_, err := handler.Handle(context.Background(), &commands.SetNumberFieldBounds{
		ConfigID: configID, FieldID: fieldID, Min: floatPtr(0), Max: floatPtr(5), ModifiedBy: "admin@example.com",
	})

	require.NoError(t, err)
	field := loadConfig(t, repo, configID).CustomFields()[0]
	assert.Equal(t, floatPtr(0), field.Min())
	assert.Equal(t, floatPtr(5), field.Max())
}

func TestSetNumberFieldBoundsHandler_RejectsMinimumGreaterThanMaximum(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := defineNumberField(t, repo, configID, "Maturity score")

	handler := NewSetNumberFieldBoundsHandler(repo)
	_, err := handler.Handle(context.Background(), &commands.SetNumberFieldBounds{
		ConfigID: configID, FieldID: fieldID, Min: floatPtr(10), Max: floatPtr(5), ModifiedBy: "admin@example.com",
	})

	assert.ErrorIs(t, err, valueobjects.ErrMinExceedsMax)
}

func TestSetNumberFieldBoundsHandler_RejectsWrongCommandType(t *testing.T) {
	repo := newTestRepo()
	_, err := NewSetNumberFieldBoundsHandler(repo).Handle(context.Background(), &commands.CreateOnePagerConfiguration{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestModifyHandlers_RejectWrongCommandType(t *testing.T) {
	repo := newTestRepo()
	_, err := NewRenameCustomFieldHandler(repo).Handle(context.Background(), &commands.CreateOnePagerConfiguration{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
