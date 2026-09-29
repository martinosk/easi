package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/stewardship/application/commands"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

func assign(stewardID, concern string) *commands.AssignSteward {
	return &commands.AssignSteward{DomainID: "domain-1", Concern: concern, StewardID: stewardID, AssignedBy: "alice@example.com"}
}

func TestAssignSteward_CreatesAStewardshipForAnUnassignedConcern(t *testing.T) {
	f := newHandlerFixture()

	result, err := f.assignHandler().Handle(context.Background(), assign("mette", "assessment"))

	require.NoError(t, err)
	s := f.repo.only()
	require.NotNil(t, s)
	assert.Equal(t, s.ID(), result.CreatedID)
	assert.Equal(t, "mette", s.Steward().Value())
	assert.Equal(t, "alice@example.com", s.AssignedBy())
}

func TestAssignSteward_ReplacesTheStewardOfAStewardedConcern(t *testing.T) {
	f := newHandlerFixture()
	_, err := f.assignHandler().Handle(context.Background(), assign("mette", "assessment"))
	require.NoError(t, err)

	_, err = f.assignHandler().Handle(context.Background(), assign("jonas", "assessment"))

	require.NoError(t, err)
	require.Len(t, f.repo.byID, 1)
	assert.Equal(t, "jonas", f.repo.only().Steward().Value())
}

func TestAssignSteward_TheCurrentStewardAgainRecordsNothing(t *testing.T) {
	f := newHandlerFixture()
	_, err := f.assignHandler().Handle(context.Background(), assign("mette", "assessment"))
	require.NoError(t, err)
	savesBefore := f.repo.saves

	_, err = f.assignHandler().Handle(context.Background(), &commands.AssignSteward{DomainID: "domain-1", Concern: "assessment", StewardID: "mette", AssignedBy: "bob@example.com"})

	require.NoError(t, err)
	assert.Equal(t, savesBefore, f.repo.saves)
	assert.Equal(t, "alice@example.com", f.repo.only().AssignedBy())
}

func TestAssignSteward_OneUserMayStewardManyConcerns(t *testing.T) {
	f := newHandlerFixture()
	_, err := f.assignHandler().Handle(context.Background(), assign("mette", "assessment"))
	require.NoError(t, err)

	_, err = f.assignHandler().Handle(context.Background(), assign("mette", "documentation"))

	require.NoError(t, err)
	assert.Len(t, f.repo.byID, 2)
}

func TestAssignSteward_RejectionsRecordNothing(t *testing.T) {
	tests := []struct {
		name    string
		command *commands.AssignSteward
		wantErr error
	}{
		{"unknown concern", assign("mette", "budget"), valueobjects.ErrUnknownConcern},
		{"unknown user", assign("nobody", "assessment"), valueobjects.ErrStewardNotFound},
		{"disabled user", assign("gone", "assessment"), valueobjects.ErrStewardDisabled},
		{"unknown domain", &commands.AssignSteward{DomainID: "domain-x", Concern: "assessment", StewardID: "mette", AssignedBy: "alice@example.com"}, ErrDomainNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newHandlerFixture()

			_, err := f.assignHandler().Handle(context.Background(), tt.command)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Zero(t, f.repo.saves)
		})
	}
}

func TestAssignSteward_RejectsAForeignCommand(t *testing.T) {
	_, err := newHandlerFixture().assignHandler().Handle(context.Background(), &commands.ReleaseSteward{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
