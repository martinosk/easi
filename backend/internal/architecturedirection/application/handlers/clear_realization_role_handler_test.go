package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validClearRoleCmd(capID, compID string) *commands.ClearRealizationRole {
	return &commands.ClearRealizationRole{
		CapabilityID: capID,
		ComponentID:  compID,
		ClearedBy:    "architect@example.com",
	}
}

func TestClearRealizationRoleHandler_ExistingRole_ClearsAndSaves(t *testing.T) {
	capID := uuid.New().String()
	compID := uuid.New().String()
	existing := buildExistingRealizationRoles(t, capID, compID, valueobjects.RealizationRoleLegacy)
	repo := &mockRealizationRolesRepository{loaded: existing}
	lookup := &mockExistingRealizationRolesLookup{id: existing.ID(), exists: true}

	handler := NewClearRealizationRoleHandler(repo, lookup)
	result, err := handler.Handle(context.Background(), validClearRoleCmd(capID, compID))

	require.NoError(t, err)
	require.True(t, repo.getCalled)
	require.Len(t, repo.saved, 1)
	_, ok := repo.saved[0].RoleFor(mustApplicationRef(t, compID))
	assert.False(t, ok)
	assert.Equal(t, existing.ID(), result.CreatedID)
}

func TestClearRealizationRoleHandler_NoAggregateForCapability_Fails(t *testing.T) {
	repo := &mockRealizationRolesRepository{}
	lookup := &mockExistingRealizationRolesLookup{exists: false}

	handler := NewClearRealizationRoleHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validClearRoleCmd(uuid.New().String(), uuid.New().String()))

	assert.ErrorIs(t, err, ErrRealizationRoleNotFoundForPair)
	assert.Empty(t, repo.saved)
	assert.False(t, repo.getCalled)
}

func TestClearRealizationRoleHandler_ComponentHoldsNoRole_Fails(t *testing.T) {
	capID := uuid.New().String()
	holder := uuid.New().String()
	other := uuid.New().String()
	existing := buildExistingRealizationRoles(t, capID, holder, valueobjects.RealizationRoleLegacy)
	repo := &mockRealizationRolesRepository{loaded: existing}
	lookup := &mockExistingRealizationRolesLookup{id: existing.ID(), exists: true}

	handler := NewClearRealizationRoleHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validClearRoleCmd(capID, other))

	assert.ErrorIs(t, err, aggregates.ErrNoRoleToClear)
	assert.Empty(t, repo.saved)
}

func TestClearRealizationRoleHandler_LookupError_Fails(t *testing.T) {
	repo := &mockRealizationRolesRepository{}
	lookup := &mockExistingRealizationRolesLookup{err: errors.New("db down")}

	handler := NewClearRealizationRoleHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validClearRoleCmd(uuid.New().String(), uuid.New().String()))

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestClearRealizationRoleHandler_LoadError_Fails(t *testing.T) {
	repo := &mockRealizationRolesRepository{getErr: errors.New("db down")}
	lookup := &mockExistingRealizationRolesLookup{id: uuid.New().String(), exists: true}

	handler := NewClearRealizationRoleHandler(repo, lookup)
	_, err := handler.Handle(context.Background(), validClearRoleCmd(uuid.New().String(), uuid.New().String()))

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}
