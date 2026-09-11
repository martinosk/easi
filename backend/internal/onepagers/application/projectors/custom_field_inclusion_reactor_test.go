package projectors

import (
	"context"
	"encoding/json"
	"testing"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeConfigLookup struct {
	records map[string]*readmodels.ConfigurationRecord
}

func (f *fakeConfigLookup) GetBySubjectType(_ context.Context, subjectType string) (*readmodels.ConfigurationRecord, error) {
	return f.records[subjectType], nil
}

type recordingDispatcher struct {
	dispatched []cqrs.Command
	onDispatch func(cmd cqrs.Command) (cqrs.CommandResult, error)
}

func (d *recordingDispatcher) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	d.dispatched = append(d.dispatched, cmd)
	if d.onDispatch != nil {
		return d.onDispatch(cmd)
	}
	return cqrs.EmptyResult(), nil
}

func reactTo(t *testing.T, reactor *CustomFieldInclusionReactor, eventType string, payload any) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return reactor.ProjectEvent(context.Background(), eventType, data)
}

func TestInclusionReactor_DefinedIncludesFieldOnExistingConfiguration(t *testing.T) {
	lookup := &fakeConfigLookup{records: map[string]*readmodels.ConfigurationRecord{"application": {ID: "config-1", SubjectType: "application"}}}
	dispatcher := &recordingDispatcher{}
	reactor := NewCustomFieldInclusionReactor(lookup, dispatcher)

	require.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeDefinedPayload{SubjectAttributeEventBase: schemaBase("attr-1"), Name: "Region", AttributeType: "text"}))

	require.Len(t, dispatcher.dispatched, 1)
	assert.Equal(t, &commands.IncludeCustomField{ConfigID: "config-1", FieldID: "attr-1", ModifiedBy: "steward@example.com"}, dispatcher.dispatched[0])
}

func TestInclusionReactor_DefinedCreatesMissingConfigurationFirst(t *testing.T) {
	lookup := &fakeConfigLookup{records: map[string]*readmodels.ConfigurationRecord{}}
	dispatcher := &recordingDispatcher{}
	dispatcher.onDispatch = func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		if create, ok := cmd.(*commands.CreateOnePagerConfiguration); ok {
			assert.Equal(t, "tenant-123", create.TenantID)
			assert.Equal(t, "application", create.SubjectType)
			assert.Equal(t, "steward@example.com", create.CreatedBy)
			lookup.records["application"] = &readmodels.ConfigurationRecord{ID: "config-new", SubjectType: "application"}
			return cqrs.NewResult("config-new"), nil
		}
		return cqrs.EmptyResult(), nil
	}
	reactor := NewCustomFieldInclusionReactor(lookup, dispatcher)

	require.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeDefinedPayload{SubjectAttributeEventBase: schemaBase("attr-1"), Name: "Region", AttributeType: "text"}))

	require.Len(t, dispatcher.dispatched, 2)
	assert.Equal(t, &commands.IncludeCustomField{ConfigID: "config-new", FieldID: "attr-1", ModifiedBy: "steward@example.com"}, dispatcher.dispatched[1])
}

func TestInclusionReactor_RetiredExcludesAndReactivatedIncludes(t *testing.T) {
	lookup := &fakeConfigLookup{records: map[string]*readmodels.ConfigurationRecord{"application": {ID: "config-1", SubjectType: "application"}}}
	dispatcher := &recordingDispatcher{}
	reactor := NewCustomFieldInclusionReactor(lookup, dispatcher)

	require.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeRetiredPayload{SubjectAttributeEventBase: schemaBase("attr-1")}))
	require.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeReactivated, mmPL.SubjectAttributeReactivatedPayload{SubjectAttributeEventBase: schemaBase("attr-1")}))

	require.Len(t, dispatcher.dispatched, 2)
	assert.Equal(t, &commands.ExcludeCustomField{ConfigID: "config-1", FieldID: "attr-1", ModifiedBy: "steward@example.com"}, dispatcher.dispatched[0])
	assert.Equal(t, &commands.IncludeCustomField{ConfigID: "config-1", FieldID: "attr-1", ModifiedBy: "steward@example.com"}, dispatcher.dispatched[1])
}

func TestInclusionReactor_RetiredWithoutConfigurationDoesNothing(t *testing.T) {
	dispatcher := &recordingDispatcher{}
	reactor := NewCustomFieldInclusionReactor(&fakeConfigLookup{records: map[string]*readmodels.ConfigurationRecord{}}, dispatcher)

	require.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeRetiredPayload{SubjectAttributeEventBase: schemaBase("attr-1")}))

	assert.Empty(t, dispatcher.dispatched)
}

func TestInclusionReactor_ToleratesInclusionStateAlreadyInPlace(t *testing.T) {
	lookup := &fakeConfigLookup{records: map[string]*readmodels.ConfigurationRecord{"application": {ID: "config-1", SubjectType: "application"}}}
	dispatcher := &recordingDispatcher{onDispatch: func(cmd cqrs.Command) (cqrs.CommandResult, error) {
		switch cmd.(type) {
		case *commands.IncludeCustomField:
			return cqrs.EmptyResult(), aggregates.ErrCustomFieldAlreadyIncluded
		case *commands.ExcludeCustomField:
			return cqrs.EmptyResult(), aggregates.ErrCustomFieldNotIncluded
		}
		return cqrs.EmptyResult(), nil
	}}
	reactor := NewCustomFieldInclusionReactor(lookup, dispatcher)

	assert.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeDefinedPayload{SubjectAttributeEventBase: schemaBase("attr-1"), Name: "Region", AttributeType: "text"}))
	assert.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeRetiredPayload{SubjectAttributeEventBase: schemaBase("attr-1")}))
}

func TestInclusionReactor_IgnoresNonInclusionSchemaEvents(t *testing.T) {
	dispatcher := &recordingDispatcher{}
	reactor := NewCustomFieldInclusionReactor(&fakeConfigLookup{records: map[string]*readmodels.ConfigurationRecord{}}, dispatcher)

	require.NoError(t, reactTo(t, reactor, mmPL.SubjectAttributeRenamed, mmPL.SubjectAttributeRenamedPayload{SubjectAttributeEventBase: schemaBase("attr-1"), NewName: "x"}))
	require.NoError(t, reactor.ProjectEvent(context.Background(), mmPL.StrategyPillarAdded, []byte(`{}`)))

	assert.Empty(t, dispatcher.dispatched)
	assert.ElementsMatch(t, []string{mmPL.SubjectAttributeDefined, mmPL.SubjectAttributeRetired, mmPL.SubjectAttributeReactivated}, CustomFieldInclusionEventTypes())
}
