package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/domain/aggregates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubInvitationStore struct {
	saved []*aggregates.Invitation
}

func (s *stubInvitationStore) Save(_ context.Context, invitation *aggregates.Invitation) error {
	s.saved = append(s.saved, invitation)
	return nil
}

type stubPredicate struct {
	result bool
}

func (p stubPredicate) GetByEmail(_ context.Context, _ string) (*readmodels.UserDTO, error) {
	if p.result {
		return &readmodels.UserDTO{Email: "newcomer@acme.com"}, nil
	}
	return nil, nil
}

func (p stubPredicate) ExistsPendingForEmail(_ context.Context, _ string) (bool, error) {
	return p.result, nil
}

func (p stubPredicate) IsDomainAllowed(_ context.Context, _ string) (bool, error) {
	return p.result, nil
}

type ensureInvitationSetup struct {
	userExists     bool
	pendingExists  bool
	domainAllowed  bool
	invitationRole string
}

func runEnsureInvitation(t *testing.T, setup ensureInvitationSetup) (*stubInvitationStore, string, error) {
	t.Helper()
	store := &stubInvitationStore{}
	handler := NewEnsureInvitationHandler(
		store,
		stubPredicate{result: setup.userExists},
		stubPredicate{result: setup.pendingExists},
		stubPredicate{result: setup.domainAllowed},
	)

	role := setup.invitationRole
	if role == "" {
		role = "stakeholder"
	}
	result, err := handler.Handle(context.Background(), &commands.EnsureInvitation{
		Email: "newcomer@acme.com",
		Role:  role,
	})
	return store, result.CreatedID, err
}

func TestEnsureInvitationHandler_CreatesInvitationWhenNothingBlocksIt(t *testing.T) {
	store, createdID, err := runEnsureInvitation(t, ensureInvitationSetup{domainAllowed: true})

	require.NoError(t, err)
	require.Len(t, store.saved, 1)
	assert.Equal(t, store.saved[0].ID(), createdID)
	assert.NotEmpty(t, createdID)
}

func TestEnsureInvitationHandler_DoesNothingWhenTheEmailAlreadyHasAUser(t *testing.T) {
	store, createdID, err := runEnsureInvitation(t, ensureInvitationSetup{userExists: true, domainAllowed: true})

	require.NoError(t, err)
	assert.Empty(t, store.saved)
	assert.Empty(t, createdID)
}

func TestEnsureInvitationHandler_DoesNothingWhenAPendingInvitationExists(t *testing.T) {
	store, createdID, err := runEnsureInvitation(t, ensureInvitationSetup{pendingExists: true, domainAllowed: true})

	require.NoError(t, err)
	assert.Empty(t, store.saved)
	assert.Empty(t, createdID)
}

func TestEnsureInvitationHandler_RejectsAnEmailDomainTheTenantDoesNotOwn(t *testing.T) {
	store, createdID, err := runEnsureInvitation(t, ensureInvitationSetup{domainAllowed: false})

	assert.ErrorIs(t, err, ErrEmailDomainNotAllowed)
	assert.Empty(t, store.saved)
	assert.Empty(t, createdID)
}

func TestEnsureInvitationHandler_RejectsAnUnknownCommand(t *testing.T) {
	handler := NewEnsureInvitationHandler(&stubInvitationStore{}, stubPredicate{}, stubPredicate{}, stubPredicate{})

	_, err := handler.Handle(context.Background(), &commands.CreateInvitation{Email: "a@b.com", Role: "admin"})

	assert.Error(t, err)
}
