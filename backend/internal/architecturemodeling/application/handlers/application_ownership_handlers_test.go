package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOwnershipComponentRepository struct {
	loaded  *aggregates.ApplicationComponent
	saved   []*aggregates.ApplicationComponent
	getErr  error
	saveErr error
}

func (m *mockOwnershipComponentRepository) GetByID(_ context.Context, _ string) (*aggregates.ApplicationComponent, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.loaded, nil
}

func (m *mockOwnershipComponentRepository) Save(_ context.Context, component *aggregates.ApplicationComponent) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, component)
	return nil
}

type fakeOwnerExistence struct {
	exists bool
	err    error
}

func (f *fakeOwnerExistence) Exists(_ context.Context, _ string) (bool, error) {
	return f.exists, f.err
}

type ownershipHandlerBuilder func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler

func nominateHandlerWith(users, teams OwnerExistence) ownershipHandlerBuilder {
	return func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler {
		return NewNominateApplicationComponentOwnerHandler(repo, users, teams)
	}
}

func assignHandlerWith(users, teams OwnerExistence) ownershipHandlerBuilder {
	return func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler {
		return NewAssignApplicationComponentOwnerHandler(repo, users, teams)
	}
}

func existingOwners() (OwnerExistence, OwnerExistence) {
	return &fakeOwnerExistence{exists: true}, &fakeOwnerExistence{exists: true}
}

func buildOwnershipComponent(t *testing.T, arrange func(*aggregates.ApplicationComponent) error) *aggregates.ApplicationComponent {
	t.Helper()
	name, err := valueobjects.NewComponentName("Billing Service")
	require.NoError(t, err)
	component, err := aggregates.NewApplicationComponent(name, valueobjects.MustNewDescription(""))
	require.NoError(t, err)
	if arrange != nil {
		require.NoError(t, arrange(component))
	}
	component.MarkChangesAsCommitted()
	return component
}

func withOwner(kind, id string) func(*aggregates.ApplicationComponent) error {
	return func(component *aggregates.ApplicationComponent) error {
		ref, err := valueobjects.NewOwnerReference(kind, id)
		if err != nil {
			return err
		}
		return component.NominateOwner(ref)
	}
}

func TestOwnershipHandlers_Transitions(t *testing.T) {
	users, teams := existingOwners()
	cases := []struct {
		name      string
		arrange   func(*aggregates.ApplicationComponent) error
		build     ownershipHandlerBuilder
		command   func(componentID string) cqrs.Command
		wantState string
	}{
		{
			name:  "nominate user",
			build: nominateHandlerWith(users, teams),
			command: func(id string) cqrs.Command {
				return &commands.NominateApplicationComponentOwner{ComponentID: id, OwnerKind: valueobjects.OwnerKindUser, OwnerID: "user-1"}
			},
			wantState: valueobjects.OwnershipStateNominated,
		},
		{
			name:    "confirm nominated team",
			arrange: withOwner(valueobjects.OwnerKindTeam, "team-1"),
			build: func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler {
				return NewConfirmApplicationComponentOwnershipHandler(repo)
			},
			command: func(id string) cqrs.Command {
				return &commands.ConfirmApplicationComponentOwnership{ComponentID: id}
			},
			wantState: valueobjects.OwnershipStateManaged,
		},
		{
			name:  "assign team",
			build: assignHandlerWith(users, teams),
			command: func(id string) cqrs.Command {
				return &commands.AssignApplicationComponentOwner{ComponentID: id, OwnerKind: valueobjects.OwnerKindTeam, OwnerID: "team-1"}
			},
			wantState: valueobjects.OwnershipStateManaged,
		},
		{
			name: "clear assigned owner",
			arrange: func(component *aggregates.ApplicationComponent) error {
				ref, err := valueobjects.NewOwnerReference(valueobjects.OwnerKindUser, "user-1")
				if err != nil {
					return err
				}
				return component.AssignOwner(ref)
			},
			build: func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler {
				return NewClearApplicationComponentOwnershipHandler(repo)
			},
			command: func(id string) cqrs.Command {
				return &commands.ClearApplicationComponentOwnership{ComponentID: id}
			},
			wantState: valueobjects.OwnershipStateUnknown,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			component := buildOwnershipComponent(t, tc.arrange)
			repo := &mockOwnershipComponentRepository{loaded: component}

			_, err := tc.build(repo).Handle(context.Background(), tc.command(component.ID()))

			require.NoError(t, err)
			require.Len(t, repo.saved, 1)
			assert.Equal(t, tc.wantState, repo.saved[0].OwnershipState().String())
		})
	}
}

func TestOwnershipHandlers_Rejections(t *testing.T) {
	users, teams := existingOwners()
	cases := []struct {
		name    string
		build   ownershipHandlerBuilder
		command cqrs.Command
		wantErr error
	}{
		{
			name:    "nominate unknown user",
			build:   nominateHandlerWith(&fakeOwnerExistence{exists: false}, teams),
			command: &commands.NominateApplicationComponentOwner{ComponentID: "comp-1", OwnerKind: valueobjects.OwnerKindUser, OwnerID: "user-unknown"},
			wantErr: ErrOwnerNotFound,
		},
		{
			name:    "nominate invalid kind",
			build:   nominateHandlerWith(users, teams),
			command: &commands.NominateApplicationComponentOwner{ComponentID: "comp-1", OwnerKind: "department", OwnerID: "dep-1"},
			wantErr: valueobjects.ErrInvalidOwnerKind,
		},
		{
			name:    "assign unknown team",
			build:   assignHandlerWith(users, &fakeOwnerExistence{exists: false}),
			command: &commands.AssignApplicationComponentOwner{ComponentID: "comp-1", OwnerKind: valueobjects.OwnerKindTeam, OwnerID: "team-unknown"},
			wantErr: ErrOwnerNotFound,
		},
		{
			name: "confirm without nomination",
			build: func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler {
				return NewConfirmApplicationComponentOwnershipHandler(repo)
			},
			command: &commands.ConfirmApplicationComponentOwnership{ComponentID: "comp-1"},
			wantErr: aggregates.ErrNoNominationToConfirm,
		},
		{
			name: "clear already unknown",
			build: func(repo *mockOwnershipComponentRepository) cqrs.CommandHandler {
				return NewClearApplicationComponentOwnershipHandler(repo)
			},
			command: &commands.ClearApplicationComponentOwnership{ComponentID: "comp-1"},
			wantErr: aggregates.ErrNoOwnershipToClear,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockOwnershipComponentRepository{loaded: buildOwnershipComponent(t, nil)}

			_, err := tc.build(repo).Handle(context.Background(), tc.command)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, repo.saved)
		})
	}
}

func TestOwnershipHandlers_DirectoryErrorFails(t *testing.T) {
	repo := &mockOwnershipComponentRepository{loaded: buildOwnershipComponent(t, nil)}
	handler := NewNominateApplicationComponentOwnerHandler(repo, &fakeOwnerExistence{err: errors.New("db down")}, &fakeOwnerExistence{exists: true})

	_, err := handler.Handle(context.Background(), &commands.NominateApplicationComponentOwner{
		ComponentID: "comp-1",
		OwnerKind:   valueobjects.OwnerKindUser,
		OwnerID:     "user-1",
	})

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}

func TestOwnershipHandlers_LoadErrorFails(t *testing.T) {
	repo := &mockOwnershipComponentRepository{getErr: errors.New("not found")}
	handler := NewClearApplicationComponentOwnershipHandler(repo)

	_, err := handler.Handle(context.Background(), &commands.ClearApplicationComponentOwnership{ComponentID: "comp-1"})

	assert.Error(t, err)
	assert.Empty(t, repo.saved)
}
