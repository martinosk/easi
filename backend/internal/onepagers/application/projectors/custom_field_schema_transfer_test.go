package projectors

import (
	"context"
	"testing"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTenants struct{ ids []string }

func (f *fakeTenants) TenantIDs(_ context.Context) ([]string, error) { return f.ids, nil }

type fakePendingSource struct {
	byTenant    map[string][]readmodels.SubjectDefinition
	transferred []string
}

func (f *fakePendingSource) PendingTransfers(ctx context.Context) ([]readmodels.SubjectDefinition, error) {
	tenant, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	return f.byTenant[tenant.Value()], nil
}

func (f *fakePendingSource) MarkTransferred(ctx context.Context, fieldID string) error {
	tenant, _ := sharedctx.GetTenant(ctx)
	f.transferred = append(f.transferred, tenant.Value()+"/"+fieldID)
	return nil
}

type tenantRecordingDispatcher struct {
	recordingDispatcher
	tenants []string
}

func (d *tenantRecordingDispatcher) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	tenant, _ := sharedctx.GetTenant(ctx)
	d.tenants = append(d.tenants, tenant.Value())
	return d.recordingDispatcher.Dispatch(ctx, cmd)
}

func TestSchemaTransfer_ImportsEveryPendingDefinitionPerTenantWithPreservedIdentity(t *testing.T) {
	min := 0.0
	pending := &fakePendingSource{byTenant: map[string][]readmodels.SubjectDefinition{
		"tenant-a": {
			{SubjectType: "application", Field: readmodels.CustomFieldRecord{
				ID: "field-1", Name: "Hosting", Type: "selection", HelpText: "h", Active: false,
				Options: []readmodels.OptionRecord{{ID: "opt-1", Label: "Cloud", Active: true}, {ID: "opt-2", Label: "Mars", Active: false}},
			}},
			{SubjectType: "vendor", Field: readmodels.CustomFieldRecord{ID: "field-2", Name: "Score", Type: "number", Active: true, Min: &min}},
		},
		"tenant-b": {},
	}}
	dispatcher := &tenantRecordingDispatcher{}
	transfer := NewCustomFieldSchemaTransfer(&fakeTenants{ids: []string{"tenant-a", "tenant-b"}}, pending, dispatcher)

	require.NoError(t, transfer.Run(context.Background()))

	require.Len(t, dispatcher.dispatched, 2)
	assert.Equal(t, []string{"tenant-a", "tenant-a"}, dispatcher.tenants)
	assert.Equal(t, &mmPL.ImportSubjectAttribute{
		TenantID: "tenant-a", SubjectType: "application", ImportedBy: transferActor,
		Attribute: mmPL.ImportedAttribute{
			ID: "field-1", Name: "Hosting", Type: "selection", HelpText: "h", Active: false,
			Options: []mmPL.ImportedOption{{ID: "opt-1", Label: "Cloud", Active: true}, {ID: "opt-2", Label: "Mars", Active: false}},
		},
	}, dispatcher.dispatched[0])
	assert.Equal(t, &mmPL.ImportSubjectAttribute{
		TenantID: "tenant-a", SubjectType: "vendor", ImportedBy: transferActor,
		Attribute: mmPL.ImportedAttribute{ID: "field-2", Name: "Score", Type: "number", Active: true, Min: &min},
	}, dispatcher.dispatched[1])
	assert.Equal(t, []string{"tenant-a/field-1", "tenant-a/field-2"}, pending.transferred)
}

func TestSchemaTransfer_StopsOnDispatchFailureWithoutMarking(t *testing.T) {
	pending := &fakePendingSource{byTenant: map[string][]readmodels.SubjectDefinition{
		"tenant-a": {{SubjectType: "application", Field: readmodels.CustomFieldRecord{ID: "field-1", Name: "Hosting", Type: "text", Active: true}}},
	}}
	dispatcher := &tenantRecordingDispatcher{recordingDispatcher: recordingDispatcher{onDispatch: func(cqrs.Command) (cqrs.CommandResult, error) {
		return cqrs.EmptyResult(), assert.AnError
	}}}
	transfer := NewCustomFieldSchemaTransfer(&fakeTenants{ids: []string{"tenant-a"}}, pending, dispatcher)

	err := transfer.Run(context.Background())

	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, pending.transferred)
}
