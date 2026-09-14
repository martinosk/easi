package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockContainmentsRepository struct {
	loaded  *aggregates.ComponentContainments
	saved   []*aggregates.ComponentContainments
	getErr  error
	saveErr error
}

func (m *mockContainmentsRepository) GetByID(_ context.Context, _ string) (*aggregates.ComponentContainments, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

func (m *mockContainmentsRepository) Save(_ context.Context, containments *aggregates.ComponentContainments) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, containments)
	return nil
}

type fakeContainmentsLookup struct {
	id    string
	found bool
	err   error
}

func (f *fakeContainmentsLookup) FindContainmentsAggregateID(_ context.Context) (string, bool, error) {
	return f.id, f.found, f.err
}

type fakeComponentRecordReader struct {
	existing map[string]bool
	err      error
}

func (f *fakeComponentRecordReader) GetByID(_ context.Context, id string) (*readmodels.ApplicationComponentDTO, error) {
	if f.err != nil {
		return nil, f.err
	}
	if !f.existing[id] {
		return nil, nil
	}
	return &readmodels.ApplicationComponentDTO{ID: id}, nil
}

func existingComponents(ids ...string) *fakeComponentRecordReader {
	existing := map[string]bool{}
	for _, id := range ids {
		existing[id] = true
	}
	return &fakeComponentRecordReader{existing: existing}
}

func registeredContainments(t *testing.T, arrange func(*aggregates.ComponentContainments) error) (*mockContainmentsRepository, *fakeContainmentsLookup) {
	t.Helper()
	containments := aggregates.NewComponentContainments()
	if arrange != nil {
		require.NoError(t, arrange(containments))
	}
	containments.MarkChangesAsCommitted()
	return &mockContainmentsRepository{loaded: containments}, &fakeContainmentsLookup{id: containments.ID(), found: true}
}

func attachCommand(part, parent valueobjects.ComponentID, kind string) *commands.AttachComponent {
	return &commands.AttachComponent{ComponentID: part.Value(), ParentID: parent.Value(), Kind: kind}
}

func TestAttachComponentHandler_ProvisionsContainmentsOnFirstAttach(t *testing.T) {
	part, parent := valueobjects.NewComponentID(), valueobjects.NewComponentID()
	repo := &mockContainmentsRepository{}
	handler := NewAttachComponentHandler(repo, &fakeContainmentsLookup{}, existingComponents(part.Value(), parent.Value()))

	_, err := handler.Handle(context.Background(), attachCommand(part, parent, valueobjects.ContainmentComposition))

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	containment, ok := repo.saved[0].ContainmentOf(part)
	require.True(t, ok)
	assert.True(t, containment.Parent.Equals(parent))
	assert.Equal(t, valueobjects.ContainmentComposition, containment.Kind.String())
}

type crmSuiteFixture struct {
	repo    *mockContainmentsRepository
	lookup  *fakeContainmentsLookup
	quoting valueobjects.ComponentID
	crm     valueobjects.ComponentID
}

func crmSuiteWithQuoting(t *testing.T) crmSuiteFixture {
	t.Helper()
	quoting, crm := valueobjects.NewComponentID(), valueobjects.NewComponentID()
	repo, lookup := registeredContainments(t, func(c *aggregates.ComponentContainments) error {
		kind, _ := valueobjects.NewContainmentKind(valueobjects.ContainmentComposition)
		return c.Attach(quoting, crm, kind)
	})
	return crmSuiteFixture{repo: repo, lookup: lookup, quoting: quoting, crm: crm}
}

func TestAttachComponentHandler_UsesRegisteredContainments(t *testing.T) {
	suite := crmSuiteWithQuoting(t)
	billing := valueobjects.NewComponentID()
	handler := NewAttachComponentHandler(suite.repo, suite.lookup, existingComponents(billing.Value(), suite.crm.Value()))

	_, err := handler.Handle(context.Background(), attachCommand(billing, suite.crm, valueobjects.ContainmentAggregation))

	require.NoError(t, err)
	require.Len(t, suite.repo.saved, 1)
	assert.Equal(t, suite.lookup.id, suite.repo.saved[0].ID())
	assert.True(t, suite.repo.saved[0].IsPart(billing))
}

func TestAttachComponentHandler_RejectsRuleViolationsWithoutSaving(t *testing.T) {
	suite := crmSuiteWithQuoting(t)
	erp := valueobjects.NewComponentID()
	handler := NewAttachComponentHandler(suite.repo, suite.lookup, existingComponents(suite.quoting.Value(), erp.Value()))

	_, err := handler.Handle(context.Background(), attachCommand(suite.quoting, erp, valueobjects.ContainmentAggregation))

	assert.ErrorIs(t, err, aggregates.ErrPartAlreadyAttached)
	assert.Empty(t, suite.repo.saved)
}

func TestAttachComponentHandler_RejectsInvalidInputs(t *testing.T) {
	part, parent := valueobjects.NewComponentID(), valueobjects.NewComponentID()
	cases := []struct {
		name       string
		command    *commands.AttachComponent
		components *fakeComponentRecordReader
		expected   error
	}{
		{"unknown kind", attachCommand(part, parent, "nesting"), existingComponents(part.Value(), parent.Value()), valueobjects.ErrInvalidContainmentKind},
		{"missing part", attachCommand(part, parent, valueobjects.ContainmentComposition), existingComponents(parent.Value()), ErrPartComponentNotFound},
		{"missing parent", attachCommand(part, parent, valueobjects.ContainmentComposition), existingComponents(part.Value()), ErrParentComponentNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockContainmentsRepository{}
			handler := NewAttachComponentHandler(repo, &fakeContainmentsLookup{}, tc.components)

			_, err := handler.Handle(context.Background(), tc.command)

			assert.ErrorIs(t, err, tc.expected)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestAttachComponentHandler_RejectsMalformedIDs(t *testing.T) {
	handler := NewAttachComponentHandler(&mockContainmentsRepository{}, &fakeContainmentsLookup{}, existingComponents())

	_, err := handler.Handle(context.Background(), &commands.AttachComponent{ComponentID: "not-a-uuid", ParentID: valueobjects.NewComponentID().Value(), Kind: valueobjects.ContainmentComposition})

	assert.Error(t, err)
}

func TestAttachComponentHandler_PropagatesInfrastructureErrors(t *testing.T) {
	part, parent := valueobjects.NewComponentID(), valueobjects.NewComponentID()
	lookupErr := errors.New("registry unavailable")
	handler := NewAttachComponentHandler(&mockContainmentsRepository{}, &fakeContainmentsLookup{err: lookupErr}, existingComponents(part.Value(), parent.Value()))

	_, err := handler.Handle(context.Background(), attachCommand(part, parent, valueobjects.ContainmentComposition))

	assert.ErrorIs(t, err, lookupErr)
}

func TestAttachComponentHandler_RejectsWrongCommandType(t *testing.T) {
	handler := NewAttachComponentHandler(&mockContainmentsRepository{}, &fakeContainmentsLookup{}, existingComponents())

	_, err := handler.Handle(context.Background(), &commands.DetachComponent{ComponentID: "c1"})

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDetachComponentHandler_ReleasesPart(t *testing.T) {
	suite := crmSuiteWithQuoting(t)
	handler := NewDetachComponentHandler(suite.repo, suite.lookup)

	_, err := handler.Handle(context.Background(), &commands.DetachComponent{ComponentID: suite.quoting.Value()})

	require.NoError(t, err)
	require.Len(t, suite.repo.saved, 1)
	assert.False(t, suite.repo.saved[0].IsPart(suite.quoting))
}

func TestDetachComponentHandler_RejectsStandaloneComponent(t *testing.T) {
	repo, lookup := registeredContainments(t, nil)
	handler := NewDetachComponentHandler(repo, lookup)

	_, err := handler.Handle(context.Background(), &commands.DetachComponent{ComponentID: valueobjects.NewComponentID().Value()})

	assert.ErrorIs(t, err, aggregates.ErrNotAPart)
	assert.Empty(t, repo.saved)
}

func TestDetachComponentHandler_RejectsWhenTenantHasNoContainments(t *testing.T) {
	repo := &mockContainmentsRepository{}
	handler := NewDetachComponentHandler(repo, &fakeContainmentsLookup{})

	_, err := handler.Handle(context.Background(), &commands.DetachComponent{ComponentID: valueobjects.NewComponentID().Value()})

	assert.ErrorIs(t, err, aggregates.ErrNotAPart)
	assert.Empty(t, repo.saved)
}

func TestDetachComponentHandler_RejectsWrongCommandType(t *testing.T) {
	handler := NewDetachComponentHandler(&mockContainmentsRepository{}, &fakeContainmentsLookup{})

	_, err := handler.Handle(context.Background(), &commands.ClearApplicationComponentOwnership{ComponentID: "c1"})

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
