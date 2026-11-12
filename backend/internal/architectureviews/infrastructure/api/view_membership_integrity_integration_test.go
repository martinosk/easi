//go:build integration
// +build integration

package api

import (
	"context"
	"net/http"
	"testing"

	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type membershipFixture struct {
	h     *viewTestHarness
	owner *viewClient
	other *viewClient
}

func newMembershipFixture(t *testing.T) *membershipFixture {
	tc, cleanup := setupViewTestDB(t)
	t.Cleanup(cleanup)
	h := setupViewHandlers(tc.db)
	f := &membershipFixture{
		h:     h,
		owner: newViewClient(t, tc, h, testActorID),
		other: newViewClient(t, tc, h, newUser("other-user")),
	}
	f.owner.createView("Default landscape")
	return f
}

func (f *membershipFixture) deleteEntities(t *testing.T, eventType eventName, entityIDs ...elementKey) {
	t.Helper()
	ctx := actorContext(newUser("admin-user"))
	for _, id := range entityIDs {
		require.NoError(t, f.h.publishInTx(ctx, deletionOf(eventType, id)))
	}
}

func TestSavingPosition_ElementNotInView_IsRejectedAndStoresNothing(t *testing.T) {
	for _, kind := range allElementKinds {
		t.Run(kind.name, func(t *testing.T) {
			f := newMembershipFixture(t)
			viewID := f.owner.createView("Landscape")

			code := f.owner.move(kind, viewID, newElement("billing"), position{300, 400})

			assert.Equal(t, http.StatusNotFound, code)
			assert.Empty(t, f.h.storedPlacements(t, viewID))
		})
	}
}

func TestSavingPosition_ElementInView_MovesIt(t *testing.T) {
	for _, kind := range allElementKinds {
		t.Run(kind.name, func(t *testing.T) {
			f := newMembershipFixture(t)
			viewID := f.owner.createView("Landscape")
			elementID := newElement("orders")
			f.owner.place(kind, viewID, elementID, position{150.5, 250.5})

			code := f.owner.move(kind, viewID, elementID, position{10, 20})

			assert.Equal(t, http.StatusNoContent, code, "a drag immediately after a drop keeps its position")
			placements := f.h.storedPlacements(t, viewID)
			require.Len(t, placements, 1)
			assert.Equal(t, []float64{10, 20}, []float64{placements[0].x, placements[0].y})
		})
	}
}

func TestAddingApplication_StoresRequestedPosition(t *testing.T) {
	f := newMembershipFixture(t)
	viewID := f.owner.createView("Landscape")
	orders := newElement("orders")

	f.owner.place(applicationKind, viewID, orders, position{150.5, 250.5})

	view := f.owner.view(viewID)
	require.Len(t, view.Components, 1)
	assert.Equal(t, []float64{150.5, 250.5}, []float64{view.Components[0].X, view.Components[0].Y})
	assert.True(t, f.h.aggregateListsApplication(t, viewID, orders))
}

func TestLayoutSave_SkipsApplicationsNotInView(t *testing.T) {
	f := newMembershipFixture(t)
	viewID := f.owner.createView("Landscape")
	orders, billing := newElement("orders"), newElement("billing")
	f.owner.place(applicationKind, viewID, orders, position{0, 0})

	code := f.owner.saveLayout(viewID,
		PositionUpdateItem{ComponentID: string(orders), X: 300, Y: 400},
		PositionUpdateItem{ComponentID: string(billing), X: 500, Y: 600})

	assert.Equal(t, http.StatusNoContent, code)
	view := f.owner.view(viewID)
	require.Equal(t, []elementKey{orders}, componentIDsOf(view))
	assert.Equal(t, []float64{300, 400}, []float64{view.Components[0].X, view.Components[0].Y})
}

func TestLatePositionSave_DoesNotResurrectRemovedApplication(t *testing.T) {
	f := newMembershipFixture(t)
	viewID := f.owner.createView("Landscape")
	billing := newElement("billing")
	f.owner.place(applicationKind, viewID, billing, position{100, 100})
	require.Equal(t, http.StatusNoContent, f.owner.removeApplication(viewID, billing))

	singleCode := f.owner.move(applicationKind, viewID, billing, position{120, 130})
	layoutCode := f.owner.saveLayout(viewID, PositionUpdateItem{ComponentID: string(billing), X: 140, Y: 150})

	assert.Equal(t, http.StatusNotFound, singleCode)
	assert.Equal(t, http.StatusNoContent, layoutCode)
	assert.Empty(t, componentIDsOf(f.owner.view(viewID)))
	assert.Empty(t, f.h.storedPlacements(t, viewID))
}

func TestDeletingApplication_RemovesItFromEveryViewAndRecordsTheRemoval(t *testing.T) {
	f := newMembershipFixture(t)
	publicView := f.owner.createPublicView("Public landscape")
	privateView := f.other.createView("Someone else's private landscape")
	billing := newElement("billing")
	f.owner.place(applicationKind, publicView, billing, position{10, 10})
	f.other.place(applicationKind, privateView, billing, position{20, 20})

	f.deleteEntities(t, archPL.ApplicationComponentDeleted, billing)

	for _, viewID := range []viewKey{publicView, privateView} {
		assert.NotContains(t, componentIDsOf(f.owner.view(viewID)), billing)
		assert.False(t, f.h.aggregateListsApplication(t, viewID, billing))
		history := f.h.viewEventTypes(t, viewID)
		assert.Equal(t, "ComponentRemovedFromView", history[len(history)-1])
	}
}

func TestDeletingApplication_ShownInDeletedView_Succeeds(t *testing.T) {
	f := newMembershipFixture(t)
	deletedView := f.owner.createView("Retired landscape")
	liveView := f.owner.createView("Current landscape")
	billing := newElement("billing")
	f.owner.place(applicationKind, deletedView, billing, position{10, 10})
	f.owner.place(applicationKind, liveView, billing, position{10, 10})
	f.owner.deleteView(deletedView)
	f.h.insertPlacement(t, deletedView, billing, repositories.ElementTypeComponent)

	err := f.h.publishInTx(actorContext(newUser("admin-user")), deletionOf(archPL.ApplicationComponentDeleted, billing))

	require.NoError(t, err)
	assert.Empty(t, componentIDsOf(f.owner.view(liveView)))
}

func TestDeletingView_RemovesAllItsPlacements(t *testing.T) {
	f := newMembershipFixture(t)
	viewID := f.owner.createView("Landscape")
	f.owner.place(applicationKind, viewID, newElement("orders"), position{10, 10})
	f.owner.place(capabilityKind, viewID, newElement("invoicing"), position{20, 20})
	f.owner.place(originEntityKind, viewID, newElement("acme"), position{30, 30})

	f.owner.deleteView(viewID)

	assert.Empty(t, f.h.storedPlacements(t, viewID))
}

func TestDeletingCapability_RemovesItFromEveryView(t *testing.T) {
	f := newMembershipFixture(t)
	publicView := f.owner.createPublicView("Public landscape")
	privateView := f.other.createView("Someone else's private landscape")
	invoicing := newElement("invoicing")
	f.owner.place(capabilityKind, publicView, invoicing, position{10, 10})
	f.other.place(capabilityKind, privateView, invoicing, position{20, 20})

	f.deleteEntities(t, cmPL.CapabilityDeleted, invoicing)

	assert.Empty(t, capabilityIDsOf(f.owner.view(publicView)))
	assert.Empty(t, capabilityIDsOf(f.owner.view(privateView)))
}

func TestDeletingParentCapability_RemovesItsDeletedChildrenFromViews(t *testing.T) {
	f := newMembershipFixture(t)
	viewID := f.owner.createView("Landscape")
	finance, invoicing, payroll := newElement("finance"), newElement("invoicing"), newElement("payroll")
	for _, capabilityID := range []elementKey{finance, invoicing, payroll} {
		f.owner.place(capabilityKind, viewID, capabilityID, position{10, 10})
	}

	err := f.h.publishInTx(actorContext(newUser("admin-user")),
		deletionOf(cmPL.CapabilityDeleted, invoicing),
		deletionOf(cmPL.CapabilityDeleted, finance))

	require.NoError(t, err)
	assert.Equal(t, []elementKey{payroll}, capabilityIDsOf(f.owner.view(viewID)))
}

func TestDeletingOriginEntity_RemovesItFromEveryView(t *testing.T) {
	for _, eventType := range []eventName{archPL.VendorDeleted, archPL.AcquiredEntityDeleted, archPL.InternalTeamDeleted} {
		t.Run(string(eventType), func(t *testing.T) {
			f := newMembershipFixture(t)
			publicView := f.owner.createPublicView("Public landscape")
			privateView := f.other.createView("Someone else's private landscape")
			acme := newElement("acme")
			f.owner.place(originEntityKind, publicView, acme, position{10, 10})
			f.other.place(originEntityKind, privateView, acme, position{20, 20})

			f.deleteEntities(t, eventType, acme)

			assert.Empty(t, originEntityIDsOf(f.owner.view(publicView)))
			assert.Empty(t, originEntityIDsOf(f.owner.view(privateView)))
		})
	}
}

func TestDeletingElement_NeverCascadesToOtherElements(t *testing.T) {
	f := newMembershipFixture(t)
	viewID := f.owner.createView("Landscape")
	billing, orders, invoicing, acme := newElement("billing"), newElement("orders"), newElement("invoicing"), newElement("acme")
	f.owner.place(applicationKind, viewID, billing, position{10, 10})
	f.owner.place(applicationKind, viewID, orders, position{20, 20})
	f.owner.place(capabilityKind, viewID, invoicing, position{30, 30})
	f.owner.place(originEntityKind, viewID, acme, position{40, 40})

	f.deleteEntities(t, archPL.ApplicationComponentDeleted, billing)
	f.deleteEntities(t, cmPL.CapabilityDeleted, invoicing)

	assert.ElementsMatch(t, []elementKey{orders, acme}, f.h.placedElementIDs(t, viewID))
	assert.True(t, f.h.aggregateListsApplication(t, viewID, orders))
}

func TestDeletionCleanupFailure_RollsBackTheDeletionTransaction(t *testing.T) {
	f := newMembershipFixture(t)
	legitView := f.owner.createView("Landscape")
	ghostView := f.owner.createView("Corrupted landscape")
	billing := newElement("billing")
	f.owner.place(applicationKind, legitView, billing, position{10, 10})
	f.h.insertPlacement(t, ghostView, billing, repositories.ElementTypeComponent)

	err := f.h.tenantDB.RunInTx(actorContext(newUser("admin-user")), func(txCtx context.Context) error {
		if _, err := f.h.commandBus.Dispatch(txCtx, &commands.RenameView{ViewID: string(legitView), NewName: "Renamed in the deleting transaction"}); err != nil {
			return err
		}
		return f.h.eventBus.Publish(txCtx, []domain.DomainEvent{deletionOf(archPL.ApplicationComponentDeleted, billing)})
	})

	require.Error(t, err)
	assert.Equal(t, "Landscape", f.owner.view(legitView).Name)
	assert.NotContains(t, f.h.viewEventTypes(t, legitView), "ViewRenamed")
	assert.Equal(t, []elementKey{billing}, componentIDsOf(f.owner.view(legitView)))
	assert.True(t, f.h.aggregateListsApplication(t, legitView, billing))
}
