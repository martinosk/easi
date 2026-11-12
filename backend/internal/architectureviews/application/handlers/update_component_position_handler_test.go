package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubElementMover struct {
	ref repositories.ElementRef
	pos repositories.Position
	err error
}

func (m *stubElementMover) MoveElement(_ context.Context, ref repositories.ElementRef, pos repositories.Position) error {
	m.ref, m.pos = ref, pos
	return m.err
}

func positionOf(componentID string) commands.UpdateComponentPosition {
	return commands.UpdateComponentPosition{ViewID: "view-1", ComponentID: componentID, X: 120, Y: 80}
}

func TestUpdateComponentPosition_MovesTheApplicationPlacement(t *testing.T) {
	mover := &stubElementMover{}

	_, err := NewUpdateComponentPositionHandler(mover).Handle(context.Background(), positionOf("orders"))

	require.NoError(t, err)
	assert.Equal(t, repositories.ElementRef{ViewID: "view-1", ElementID: "orders", ElementType: repositories.ElementTypeComponent}, mover.ref)
	assert.Equal(t, repositories.Position{X: 120, Y: 80}, mover.pos)
}

func TestUpdateComponentPosition_ApplicationNotInViewIsNotFound(t *testing.T) {
	mover := &stubElementMover{err: repositories.ErrPlacementNotFound}

	_, err := NewUpdateComponentPositionHandler(mover).Handle(context.Background(), positionOf("billing"))

	assert.ErrorIs(t, err, repositories.ErrPlacementNotFound)
}
