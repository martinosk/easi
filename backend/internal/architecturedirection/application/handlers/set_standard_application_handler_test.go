package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStandardApplicationRepository struct {
	saved     []*aggregates.StandardApplication
	loaded    *aggregates.StandardApplication
	getErr    error
	saveErr   error
	getCalled bool
}

func (m *mockStandardApplicationRepository) Save(_ context.Context, sa *aggregates.StandardApplication) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, sa)
	return nil
}

func (m *mockStandardApplicationRepository) GetByID(_ context.Context, _ string) (*aggregates.StandardApplication, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

type mockExistingStandardLookup struct {
	id     string
	exists bool
	err    error
}

func (m *mockExistingStandardLookup) FindAggregateIDForEnterpriseCapability(_ context.Context, _ string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	return m.id, m.exists, nil
}

func validSetCmd() *commands.SetStandardApplication {
	return &commands.SetStandardApplication{
		EnterpriseCapabilityID: uuid.New().String(),
		ApplicationID:          uuid.New().String(),
		Narrative:              "covers operational and reporting layers",
	}
}

func TestSetStandardApplicationHandler_FirstSet_CreatesAggregateWithFreshID(t *testing.T) {
	repo := &mockStandardApplicationRepository{}
	lookup := &mockExistingStandardLookup{exists: false}

	handler := NewSetStandardApplicationHandler(repo, lookup, allReferencesExist())
	cmd := validSetCmd()
	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, repo.saved[0].ID(), result.CreatedID)
	assert.NotEqual(t, cmd.EnterpriseCapabilityID, result.CreatedID,
		"aggregate ID must be its own identity, not derived from the EC's ID")
	assert.False(t, repo.getCalled, "first set must not load an existing aggregate")
}

func TestSetStandardApplicationHandler_Replacement_LoadsExistingAndChanges(t *testing.T) {
	ec := uuid.New().String()
	originalApp := uuid.New().String()
	existing := buildExistingAggregate(t, ec, originalApp)
	repo := &mockStandardApplicationRepository{loaded: existing}
	lookup := &mockExistingStandardLookup{id: existing.ID(), exists: true}

	handler := NewSetStandardApplicationHandler(repo, lookup, allReferencesExist())
	cmd := &commands.SetStandardApplication{
		EnterpriseCapabilityID: ec,
		ApplicationID:          uuid.New().String(),
		Narrative:              "replacement narrative",
	}
	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.True(t, repo.getCalled)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, existing.ID(), result.CreatedID)
	assert.Equal(t, cmd.ApplicationID, repo.saved[0].CurrentApplication().Value())
}

func TestSetStandardApplicationHandler_InvalidInputs_FailWithoutSaving(t *testing.T) {
	cases := []struct {
		name      string
		mutate    func(*commands.SetStandardApplication)
		expectErr error
	}{
		{
			name:      "blank narrative",
			mutate:    func(c *commands.SetStandardApplication) { c.Narrative = "   " },
			expectErr: aggregates.ErrNarrativeRequiredForStandardApplication,
		},
		{
			name:   "invalid enterprise capability id",
			mutate: func(c *commands.SetStandardApplication) { c.EnterpriseCapabilityID = "not-a-uuid" },
		},
		{
			name:   "invalid application id",
			mutate: func(c *commands.SetStandardApplication) { c.ApplicationID = "not-a-uuid" },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockStandardApplicationRepository{}
			handler := NewSetStandardApplicationHandler(repo, &mockExistingStandardLookup{}, allReferencesExist())
			cmd := validSetCmd()
			tc.mutate(cmd)

			_, err := handler.Handle(context.Background(), cmd)

			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			} else {
				assert.Error(t, err)
			}
			assert.Empty(t, repo.saved)
		})
	}
}

func TestSetStandardApplicationHandler_UnknownEnterpriseCapability_Fails(t *testing.T) {
	repo := &mockStandardApplicationRepository{}
	refs := allReferencesExist()
	refs.EnterpriseCapabilityExists = constantExists(false)

	handler := NewSetStandardApplicationHandler(repo, &mockExistingStandardLookup{}, refs)
	_, err := handler.Handle(context.Background(), validSetCmd())

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}

func TestSetStandardApplicationHandler_LookupError_Fails(t *testing.T) {
	repo := &mockStandardApplicationRepository{}
	lookup := &mockExistingStandardLookup{err: errors.New("db down")}

	handler := NewSetStandardApplicationHandler(repo, lookup, allReferencesExist())
	_, err := handler.Handle(context.Background(), validSetCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func buildExistingAggregate(t *testing.T, ec, app string) *aggregates.StandardApplication {
	t.Helper()
	ecRef := mustNewEnterpriseCapabilityRef(t, ec)
	appRef := mustNewApplicationRef(t, app)
	narrative := mustNewNarrative(t, "first")
	sa, err := aggregates.NewStandardApplication(ecRef, appRef, narrative)
	require.NoError(t, err)
	sa.MarkChangesAsCommitted()
	return sa
}
