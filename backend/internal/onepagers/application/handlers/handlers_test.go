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

func includeField(t *testing.T, repo *repositories.OnePagerConfigurationRepository, configID string) string {
	t.Helper()
	fieldID := valueobjects.NewFieldID().Value()
	_, err := NewIncludeCustomFieldHandler(repo).Handle(context.Background(), &commands.IncludeCustomField{
		ConfigID:   configID,
		FieldID:    fieldID,
		ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	return fieldID
}

func mustFieldID(t *testing.T, raw string) valueobjects.FieldID {
	t.Helper()
	fieldID, err := valueobjects.NewFieldIDFromString(raw)
	require.NoError(t, err)
	return fieldID
}

func TestCreateOnePagerConfigurationHandler_CreatesWithDefaults(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)

	config := loadConfig(t, repo, configID)
	assert.Equal(t, "application", config.SubjectType().Value())
	assert.Len(t, config.DisplayOrder(), 3)
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

	_, err := handler.Handle(context.Background(), &commands.IncludeCustomField{})

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestIncludeCustomFieldHandler_AppendsFieldToDisplayOrder(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)

	fieldID := includeField(t, repo, configID)

	config := loadConfig(t, repo, configID)
	assert.True(t, config.IsCustomFieldIncluded(mustFieldID(t, fieldID)))
	assert.Len(t, config.DisplayOrder(), 4)
}

func TestIncludeCustomFieldHandler_RejectsInvalidFieldID(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)

	_, err := NewIncludeCustomFieldHandler(repo).Handle(context.Background(), &commands.IncludeCustomField{
		ConfigID: configID, FieldID: "not-a-uuid", ModifiedBy: "admin@example.com",
	})

	assert.ErrorIs(t, err, valueobjects.ErrInvalidFieldID)
}

func TestChangeRequirementAndExcludeHandlers(t *testing.T) {
	repo := newTestRepo()
	configID := createConfiguration(t, repo)
	fieldID := includeField(t, repo, configID)
	ctx := context.Background()

	_, err := NewChangeCustomFieldRequirementHandler(repo).Handle(ctx, &commands.ChangeCustomFieldRequirement{
		ConfigID: configID, FieldID: fieldID, Required: true, ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.True(t, loadConfig(t, repo, configID).IsCustomFieldRequired(mustFieldID(t, fieldID)))

	_, err = NewExcludeCustomFieldHandler(repo).Handle(ctx, &commands.ExcludeCustomField{
		ConfigID: configID, FieldID: fieldID, ModifiedBy: "admin@example.com",
	})
	require.NoError(t, err)
	assert.False(t, loadConfig(t, repo, configID).IsCustomFieldIncluded(mustFieldID(t, fieldID)))

	_, err = NewExcludeCustomFieldHandler(repo).Handle(ctx, &commands.ExcludeCustomField{
		ConfigID: configID, FieldID: fieldID, ModifiedBy: "admin@example.com",
	})
	assert.ErrorIs(t, err, aggregates.ErrCustomFieldNotIncluded)
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
	fieldID := includeField(t, repo, configID)

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

func TestModifyHandlers_RejectWrongCommandType(t *testing.T) {
	repo := newTestRepo()
	_, err := NewChangeCustomFieldRequirementHandler(repo).Handle(context.Background(), &commands.CreateOnePagerConfiguration{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
