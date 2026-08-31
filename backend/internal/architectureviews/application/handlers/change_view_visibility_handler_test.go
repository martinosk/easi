package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanActorEditView_OwnerCanEdit(t *testing.T) {
	ownership := ownershipContext{isOwner: true, actorRole: sharedctx.RoleStakeholder}
	assert.True(t, canActorEditView(ownership))
}

func TestCanActorEditView_ViewWithNoOwnerCanBeEditedByAnyone(t *testing.T) {
	ownership := ownershipContext{hasNoOwner: true, actorRole: sharedctx.RoleStakeholder}
	assert.True(t, canActorEditView(ownership))
}

func TestCanActorEditView_AdminCanEditNonOwnedView(t *testing.T) {
	ownership := ownershipContext{isOwner: false, actorRole: sharedctx.RoleAdmin}
	assert.True(t, canActorEditView(ownership))
}

func TestCanActorEditView_NonAdminNonOwnerCannotEdit(t *testing.T) {
	for _, role := range []sharedctx.Role{sharedctx.RoleArchitect, sharedctx.RoleStakeholder, sharedctx.Role("")} {
		ownership := ownershipContext{isOwner: false, actorRole: role}
		assert.False(t, canActorEditView(ownership), "role %q should not be able to edit", role)
	}
}

func testView(t *testing.T, ownerID, ownerEmail string) *aggregates.ArchitectureView {
	t.Helper()
	name, err := valueobjects.NewViewName("Test View")
	require.NoError(t, err)
	owner, err := valueobjects.NewViewOwner(ownerID, ownerEmail)
	require.NoError(t, err)
	view, err := aggregates.NewArchitectureView(name, "desc", false, owner)
	require.NoError(t, err)
	return view
}

func TestNewOwnershipContext_CapturesActorAndOwnership(t *testing.T) {
	view := testView(t, "owner-1", "owner@example.com")

	owningActor := sharedctx.NewActor("owner-1", "owner@example.com", sharedctx.RoleStakeholder)
	ownership := newOwnershipContext(view, owningActor)
	assert.True(t, ownership.isOwner)
	assert.False(t, ownership.hasNoOwner)
	assert.Equal(t, sharedctx.RoleStakeholder, ownership.actorRole)

	adminActor := sharedctx.NewActor("admin-1", "admin@example.com", sharedctx.RoleAdmin)
	adminOwnership := newOwnershipContext(view, adminActor)
	assert.False(t, adminOwnership.isOwner)
	assert.Equal(t, sharedctx.RoleAdmin, adminOwnership.actorRole)
}

func TestApplyVisibilityChange(t *testing.T) {
	tests := []struct {
		name        string
		startPublic bool
		actorID     string
		actorRole   sharedctx.Role
		makePrivate bool
		wantErr     error
		wantPrivate bool
		wantOwnerID string
	}{
		{
			name:        "owner makes view private",
			startPublic: true,
			actorID:     "owner-1",
			actorRole:   sharedctx.RoleStakeholder,
			makePrivate: true,
			wantPrivate: true,
			wantOwnerID: "owner-1",
		},
		{
			name:        "admin makes other users view public",
			startPublic: false,
			actorID:     "admin-1",
			actorRole:   sharedctx.RoleAdmin,
			makePrivate: false,
			wantPrivate: false,
			wantOwnerID: "admin-1",
		},
		{
			name:        "non-owner non-admin cannot make view private",
			startPublic: true,
			actorID:     "other-1",
			actorRole:   sharedctx.RoleArchitect,
			makePrivate: true,
			wantErr:     aggregates.ErrOnlyOwnerCanMakePrivate,
		},
		{
			name:        "non-owner non-admin cannot make view public",
			startPublic: false,
			actorID:     "other-1",
			actorRole:   sharedctx.RoleArchitect,
			makePrivate: false,
			wantErr:     ErrNotAuthorizedToChangeVisibility,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := testView(t, "owner-1", "owner@example.com")
			if tt.startPublic {
				require.NoError(t, view.MakePublic(view.Owner()))
			}

			actor := sharedctx.NewActor(tt.actorID, tt.actorID+"@example.com", tt.actorRole)
			ownership := newOwnershipContext(view, actor)
			canEdit := canActorEditView(ownership)

			err := applyVisibilityChange(view, tt.makePrivate, ownership, canEdit)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPrivate, view.IsPrivate())
			assert.Equal(t, tt.wantOwnerID, view.Owner().UserID())
		})
	}
}

func TestChangeViewVisibilityHandler_Handle_MissingActorReturnsErrActorNotFound(t *testing.T) {
	handler := NewChangeViewVisibilityHandler(nil)

	cmd := &commands.ChangeViewVisibility{ViewID: "view-1", IsPrivate: true}

	_, err := handler.Handle(context.Background(), cmd)

	assert.ErrorIs(t, err, ErrActorNotFound)
}

func TestChangeViewVisibilityHandler_Handle_InvalidCommand(t *testing.T) {
	handler := NewChangeViewVisibilityHandler(nil)

	invalidCmd := &commands.CreateView{Name: "Test"}

	_, err := handler.Handle(context.Background(), invalidCmd)

	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
