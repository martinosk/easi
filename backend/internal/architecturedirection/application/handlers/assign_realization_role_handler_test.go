package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationRolesRepository struct {
	saved     []*aggregates.RealizationRoles
	loaded    *aggregates.RealizationRoles
	getErr    error
	saveErr   error
	getCalled bool
}

func (m *mockRealizationRolesRepository) Save(_ context.Context, rr *aggregates.RealizationRoles) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, rr)
	return nil
}

func (m *mockRealizationRolesRepository) GetByID(_ context.Context, _ string) (*aggregates.RealizationRoles, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

type mockExistingRealizationRolesLookup struct {
	id     string
	exists bool
	err    error
}

func (m *mockExistingRealizationRolesLookup) FindAggregateIDForCapability(_ context.Context, _ string) (string, bool, error) {
	if m.err != nil {
		return "", false, m.err
	}
	return m.id, m.exists, nil
}

func validAssignRoleCmd() *commands.AssignRealizationRole {
	return &commands.AssignRealizationRole{
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		Role:         valueobjects.RealizationRoleStandard,
		AssignedBy:   "architect@example.com",
	}
}

func TestAssignRealizationRoleHandler_FirstAssignment_CreatesAggregateWithFreshID(t *testing.T) {
	repo := &mockRealizationRolesRepository{}
	lookup := &mockExistingRealizationRolesLookup{exists: false}

	handler := NewAssignRealizationRoleHandler(repo, lookup, alwaysDirect(uuid.New().String()))
	cmd := validAssignRoleCmd()
	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, repo.saved[0].ID(), result.CreatedID)
	assert.False(t, repo.getCalled, "first assignment must not load an existing aggregate")
	role, ok := repo.saved[0].RoleFor(mustApplicationRef(t, cmd.ComponentID))
	require.True(t, ok)
	assert.Equal(t, valueobjects.RealizationRoleStandard, role.Value())
}

func TestAssignRealizationRoleHandler_ExistingCapabilityAggregate_LoadsAndAssigns(t *testing.T) {
	capID := uuid.New().String()
	compID := uuid.New().String()
	existing := buildExistingRealizationRoles(t, capID, uuid.New().String(), valueobjects.RealizationRoleLegacy)
	repo := &mockRealizationRolesRepository{loaded: existing}
	lookup := &mockExistingRealizationRolesLookup{id: existing.ID(), exists: true}

	handler := NewAssignRealizationRoleHandler(repo, lookup, alwaysDirect(uuid.New().String()))
	cmd := &commands.AssignRealizationRole{
		CapabilityID: capID,
		ComponentID:  compID,
		Role:         valueobjects.RealizationRoleStandard,
		AssignedBy:   "b@example.com",
	}
	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	require.True(t, repo.getCalled)
	require.Len(t, repo.saved, 1)
	assert.Equal(t, existing.ID(), result.CreatedID)
	role, ok := repo.saved[0].RoleFor(mustApplicationRef(t, compID))
	require.True(t, ok)
	assert.Equal(t, valueobjects.RealizationRoleStandard, role.Value())
}

func TestAssignRealizationRoleHandler_NoDirectRealization_Fails(t *testing.T) {
	repo := &mockRealizationRolesRepository{}
	handler := NewAssignRealizationRoleHandler(repo, &mockExistingRealizationRolesLookup{}, neverDirect())

	_, err := handler.Handle(context.Background(), validAssignRoleCmd())

	assert.ErrorIs(t, err, services.ErrReferencedEntityNotFound)
	assert.Empty(t, repo.saved)
}

func TestAssignRealizationRoleHandler_DirectRealizationCheckErrors_Fails(t *testing.T) {
	repo := &mockRealizationRolesRepository{}
	handler := NewAssignRealizationRoleHandler(repo, &mockExistingRealizationRolesLookup{}, failingDirectLookup(errors.New("db down")))

	_, err := handler.Handle(context.Background(), validAssignRoleCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestAssignRealizationRoleHandler_InvalidInputs_FailWithoutSaving(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*commands.AssignRealizationRole)
	}{
		{"invalid capability id", func(c *commands.AssignRealizationRole) { c.CapabilityID = "not-a-uuid" }},
		{"invalid component id", func(c *commands.AssignRealizationRole) { c.ComponentID = "not-a-uuid" }},
		{"invalid role", func(c *commands.AssignRealizationRole) { c.Role = "Standard" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockRealizationRolesRepository{}
			handler := NewAssignRealizationRoleHandler(repo, &mockExistingRealizationRolesLookup{}, alwaysDirect(uuid.New().String()))
			cmd := validAssignRoleCmd()
			tc.mutate(cmd)

			_, err := handler.Handle(context.Background(), cmd)

			assert.Error(t, err)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestAssignRealizationRoleHandler_LookupError_Fails(t *testing.T) {
	repo := &mockRealizationRolesRepository{}
	lookup := &mockExistingRealizationRolesLookup{err: errors.New("db down")}

	handler := NewAssignRealizationRoleHandler(repo, lookup, alwaysDirect(uuid.New().String()))
	_, err := handler.Handle(context.Background(), validAssignRoleCmd())

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func buildExistingRealizationRoles(t *testing.T, capID, compID, role string) *aggregates.RealizationRoles {
	t.Helper()
	capRef, err := valueobjects.NewPhysicalCapabilityRef(capID)
	require.NoError(t, err)
	roleVO, err := valueobjects.NewRealizationRole(role)
	require.NoError(t, err)
	rr, err := aggregates.NewRealizationRoles(aggregates.RealizationRolesFacts{
		CapabilityID:  capRef,
		ComponentID:   mustApplicationRef(t, compID),
		RealizationID: uuid.New().String(),
		Role:          roleVO,
		AssignedBy:    "a@example.com",
	})
	require.NoError(t, err)
	rr.MarkChangesAsCommitted()
	return rr
}

func mustApplicationRef(t *testing.T, id string) valueobjects.ApplicationRef {
	t.Helper()
	ref, err := valueobjects.NewApplicationRef(id)
	require.NoError(t, err)
	return ref
}
