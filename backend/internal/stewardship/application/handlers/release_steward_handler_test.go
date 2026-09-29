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

func release(domainID, concern string) *commands.ReleaseSteward {
	return &commands.ReleaseSteward{DomainID: domainID, Concern: concern, ReleasedBy: "bob@example.com"}
}

func TestReleaseSteward_EndsTheStewardship(t *testing.T) {
	f := newHandlerFixture()
	_, err := f.assignHandler().Handle(context.Background(), assign("mette", "assessment"))
	require.NoError(t, err)

	_, err = f.releaseHandler().Handle(context.Background(), release("domain-1", "assessment"))

	require.NoError(t, err)
	assert.True(t, f.repo.only().IsReleased())
	_, live, _ := f.repo.FindLiveStewardshipID(context.Background(), "domain-1", "assessment")
	assert.False(t, live)
}

func TestReleaseSteward_AnUnassignedConcernRecordsNothing(t *testing.T) {
	f := newHandlerFixture()

	_, err := f.releaseHandler().Handle(context.Background(), release("domain-1", "planning"))

	require.NoError(t, err)
	assert.Zero(t, f.repo.saves)
}

func TestReleaseSteward_Rejections(t *testing.T) {
	tests := []struct {
		name    string
		command *commands.ReleaseSteward
		wantErr error
	}{
		{"unknown concern", release("domain-1", "budget"), valueobjects.ErrUnknownConcern},
		{"unknown domain", release("domain-x", "planning"), ErrDomainNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newHandlerFixture().releaseHandler().Handle(context.Background(), tt.command)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestReleaseSteward_RejectsAForeignCommand(t *testing.T) {
	_, err := newHandlerFixture().releaseHandler().Handle(context.Background(), &commands.AssignSteward{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
