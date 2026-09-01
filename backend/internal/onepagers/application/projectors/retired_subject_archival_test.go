package projectors

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTenantDirectory struct {
	ids []string
	err error
}

func (f *fakeTenantDirectory) TenantIDs(_ context.Context) ([]string, error) {
	return f.ids, f.err
}

type fakeSubjectIDLister struct {
	byTenantAndType map[string][]string
}

func (f *fakeSubjectIDLister) SubjectIDs(ctx context.Context, subjectType string) ([]string, error) {
	tenant, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	return f.byTenantAndType[tenant.Value()+"/"+subjectType], nil
}

type tenantAwareFactsFinder struct {
	ids map[string]string
}

func (f *tenantAwareFactsFinder) FactsIDForSubject(ctx context.Context, subject readmodels.SubjectKey) (string, error) {
	tenant, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return f.ids[tenant.Value()+"/"+subject.SubjectType+"/"+subject.SubjectID], nil
}

type tenantCapturingDispatcher struct {
	dispatched []cqrs.Command
	tenants    []string
	err        error
}

func (d *tenantCapturingDispatcher) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	tenant, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	d.dispatched = append(d.dispatched, cmd)
	d.tenants = append(d.tenants, tenant.Value())
	return cqrs.EmptyResult(), d.err
}

func newArchivalUnderTest(tenants *fakeTenantDirectory, subjects *fakeSubjectIDLister, facts *tenantAwareFactsFinder, dispatcher *tenantCapturingDispatcher) *RetiredSubjectArchival {
	return NewRetiredSubjectArchival(tenants, subjects, facts, dispatcher)
}

func TestRetiredSubjectArchival_ArchivesFactsForEverySubjectAcrossTenants(t *testing.T) {
	tenants := &fakeTenantDirectory{ids: []string{"tenant-a", "tenant-b"}}
	subjects := &fakeSubjectIDLister{byTenantAndType: map[string][]string{
		"tenant-a/enterprise-capability": {"ec-1", "ec-2"},
		"tenant-b/enterprise-capability": {"ec-3"},
	}}
	facts := &tenantAwareFactsFinder{ids: map[string]string{
		"tenant-a/enterprise-capability/ec-1": "facts-1",
		"tenant-b/enterprise-capability/ec-3": "facts-3",
	}}
	dispatcher := &tenantCapturingDispatcher{}

	err := newArchivalUnderTest(tenants, subjects, facts, dispatcher).Run(context.Background())

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 2, "ec-2 holds no facts and is skipped")
	first, ok := dispatcher.dispatched[0].(*commands.ArchiveOnePagerFacts)
	require.True(t, ok)
	assert.Equal(t, "facts-1", first.FactsID)
	assert.Equal(t, "subject deleted", first.Reason)
	second, ok := dispatcher.dispatched[1].(*commands.ArchiveOnePagerFacts)
	require.True(t, ok)
	assert.Equal(t, "facts-3", second.FactsID)
	assert.Equal(t, []string{"tenant-a", "tenant-b"}, dispatcher.tenants, "each archive is dispatched in its own tenant's context")
}

func TestRetiredSubjectArchival_NoSubjects_NoDispatch(t *testing.T) {
	tenants := &fakeTenantDirectory{ids: []string{"tenant-a"}}
	dispatcher := &tenantCapturingDispatcher{}

	err := newArchivalUnderTest(tenants, &fakeSubjectIDLister{}, &tenantAwareFactsFinder{}, dispatcher).Run(context.Background())

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestRetiredSubjectArchival_TenantDirectoryErrorStopsTheSweep(t *testing.T) {
	tenants := &fakeTenantDirectory{err: errors.New("tenants unavailable")}

	err := newArchivalUnderTest(tenants, &fakeSubjectIDLister{}, &tenantAwareFactsFinder{}, &tenantCapturingDispatcher{}).Run(context.Background())

	assert.ErrorContains(t, err, "tenants unavailable")
}

func TestRetiredSubjectArchival_DispatchErrorSurfaces(t *testing.T) {
	tenants := &fakeTenantDirectory{ids: []string{"tenant-a"}}
	subjects := &fakeSubjectIDLister{byTenantAndType: map[string][]string{
		"tenant-a/enterprise-capability": {"ec-1"},
	}}
	facts := &tenantAwareFactsFinder{ids: map[string]string{"tenant-a/enterprise-capability/ec-1": "facts-1"}}
	dispatcher := &tenantCapturingDispatcher{err: errors.New("archive refused")}

	err := newArchivalUnderTest(tenants, subjects, facts, dispatcher).Run(context.Background())

	assert.ErrorContains(t, err, "archive refused")
}
