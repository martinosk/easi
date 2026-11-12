package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architectureviews/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type removedPlacement struct {
	elementID   string
	elementType repositories.ElementType
}

type recordingPlacementRemover struct {
	removed []removedPlacement
	err     error
}

func (r *recordingPlacementRemover) DeleteElementFromAllViews(_ context.Context, elementID string, elementType repositories.ElementType) error {
	r.removed = append(r.removed, removedPlacement{elementID: elementID, elementType: elementType})
	return r.err
}

func TestDeletedCapability_RemovesCapabilityPlacementsFromAllViews(t *testing.T) {
	remover := &recordingPlacementRemover{}

	err := NewDeletedCapabilityHandler(remover).Handle(context.Background(), deletionEvent("CapabilityDeleted", "invoicing"))

	require.NoError(t, err)
	assert.Equal(t, []removedPlacement{{elementID: "invoicing", elementType: repositories.ElementTypeCapability}}, remover.removed)
}

func TestDeletedOriginEntity_RemovesOriginEntityPlacementsFromAllViews(t *testing.T) {
	for _, eventType := range []string{"VendorDeleted", "AcquiredEntityDeleted", "InternalTeamDeleted"} {
		t.Run(eventType, func(t *testing.T) {
			remover := &recordingPlacementRemover{}

			err := NewDeletedOriginEntityHandler(remover).Handle(context.Background(), deletionEvent(eventType, "acme"))

			require.NoError(t, err)
			assert.Equal(t, []removedPlacement{{elementID: "acme", elementType: repositories.ElementTypeOriginEntity}}, remover.removed)
		})
	}
}

func TestDeletedElement_EventWithoutIDIsRejected(t *testing.T) {
	remover := &recordingPlacementRemover{}

	err := NewDeletedCapabilityHandler(remover).Handle(context.Background(), deletionEvent("CapabilityDeleted", ""))

	assert.ErrorIs(t, err, ErrDeletedElementWithoutID)
	assert.Empty(t, remover.removed, "an empty id must never reach a tenant-wide delete")
}

func TestDeletedElement_FailingRemovalFailsTheDeletion(t *testing.T) {
	removalErr := errors.New("database unavailable")
	remover := &recordingPlacementRemover{err: removalErr}

	err := NewDeletedOriginEntityHandler(remover).Handle(context.Background(), deletionEvent("VendorDeleted", "acme"))

	assert.ErrorIs(t, err, removalErr)
}
