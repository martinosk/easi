package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/domain/aggregates"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRevokeHandler() (*RevokeEditGrantHandler, *CreateEditGrantHandler) {
	es := newInMemoryEventStore()
	repo := repositories.NewEditGrantRepository(es)
	return NewRevokeEditGrantHandler(repo), NewCreateEditGrantHandler(repo)
}

func TestRevokeEditGrantHandler_ValidRevoke_Succeeds(t *testing.T) {
	revokeHandler, createHandler := newTestRevokeHandler()

	createResult, err := createHandler.Handle(context.Background(), validCreateCommand())
	require.NoError(t, err)

	cmd := &commands.RevokeEditGrant{ID: createResult.CreatedID, RevokedBy: "admin-id"}
	_, err = revokeHandler.Handle(context.Background(), cmd)
	assert.NoError(t, err)
}

func TestRevokeEditGrantHandler_AlreadyRevoked_ReturnsError(t *testing.T) {
	revokeHandler, createHandler := newTestRevokeHandler()

	createResult, err := createHandler.Handle(context.Background(), validCreateCommand())
	require.NoError(t, err)

	cmd := &commands.RevokeEditGrant{ID: createResult.CreatedID, RevokedBy: "admin-id"}
	_, err = revokeHandler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	_, err = revokeHandler.Handle(context.Background(), cmd)
	assert.Equal(t, aggregates.ErrGrantAlreadyRevoked, err)
}

func TestRevokeEditGrantHandler_NonExistentGrant_ReturnsError(t *testing.T) {
	revokeHandler, _ := newTestRevokeHandler()

	cmd := &commands.RevokeEditGrant{ID: "nonexistent-id", RevokedBy: "admin-id"}
	_, err := revokeHandler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
