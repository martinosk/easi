//go:build integration
// +build integration

package repositories

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMembershipRepair_ApplicationPlacementsEqualTheExistingMembersTheHistoryLists(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	assert.Equal(t, sortedIDs(l.ids.orders, l.ids.crm), l.placedIDs(l.ids.mainView, ElementTypeComponent))
	assert.Empty(t, l.placedIDs(l.ids.privateView, ElementTypeComponent))
}

func TestMembershipRepair_PurgesApplicationsWhoseLastMembershipEventIsARemoval(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	assert.NotContains(t, l.placedIDs(l.ids.mainView, ElementTypeComponent), l.ids.ghost)
}

func TestMembershipRepair_PurgesApplicationsThatWereNeverAdded(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	assert.NotContains(t, l.placedIDs(l.ids.mainView, ElementTypeComponent), l.ids.neverAdded)
}

func TestMembershipRepair_RestoresListedMemberMissingFromPlacementsAtItsEventPosition(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	crm, ok := l.placements(l.ids.mainView, ElementTypeComponent)[l.ids.crm]
	require.True(t, ok)
	assert.Equal(t, []float64{70, 80}, []float64{crm.x, crm.y})
}

func TestMembershipRepair_AppendsRemovalForEveryDeletedApplicationTheHistoryStillLists(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	mainHistory := l.history(l.ids.mainView)
	require.Len(t, mainHistory, 9)
	appended := mainHistory[7:]
	assert.ElementsMatch(t, []entityID{l.ids.billing, l.ids.payments}, []entityID{appended[0].componentID, appended[1].componentID})
	for _, event := range appended {
		assertSystemRemoval(t, event, l.ids.mainView)
	}

	privateHistory := l.history(l.ids.privateView)
	require.Len(t, privateHistory, 5, "payments was already removed, so only billing gets a removal")
	assert.Equal(t, l.ids.billing, privateHistory[4].componentID)
	assertSystemRemoval(t, privateHistory[4], l.ids.privateView)
}

func assertSystemRemoval(t *testing.T, event storedViewEvent, viewID entityID) {
	t.Helper()
	assert.Equal(t, "ComponentRemovedFromView", event.eventType)
	assert.Equal(t, string(viewID), event.data["viewId"])
	assert.Equal(t, []string{"componentId", "removedAt", "viewId"}, sortedKeys(event.data))
	_, err := time.Parse(time.RFC3339Nano, fmt.Sprint(event.data["removedAt"]))
	assert.NoError(t, err)
	assert.Equal(t, "system", event.actorID)
	assert.Equal(t, "system@easi.app", event.actorEmail)
}

func TestMembershipRepair_AffectedAggregatesLoadConsistentlyWithTheirPlacements(t *testing.T) {
	l := seededRepairLandscape(t)
	deleted := map[entityID]bool{l.ids.billing: true, l.ids.payments: true}

	l.runRepair()

	for _, viewID := range []entityID{l.ids.mainView, l.ids.privateView} {
		view := l.loadView(viewID)
		history := l.history(viewID)
		assert.Equal(t, len(history), view.Version())
		assert.Equal(t, len(history), history[len(history)-1].version, "versions stay consecutive")
		assert.Equal(t, sortedIDs(asEntityIDs(view.Components())...), l.placedIDs(viewID, ElementTypeComponent))
		for _, member := range view.Components() {
			assert.False(t, deleted[entityID(member)], "deleted application %s is still a member", member)
		}
	}
}

func TestMembershipRepair_PurgesPlacementsOfCapabilitiesAndOriginEntitiesThatNoLongerExist(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	assert.Equal(t, []entityID{l.ids.invoicing}, l.placedIDs(l.ids.mainView, ElementTypeCapability))
	assert.Equal(t, sortedIDs(l.ids.acme, l.ids.acquired), l.placedIDs(l.ids.mainView, ElementTypeOriginEntity))
}

func TestMembershipRepair_KeepsLegitimatePlacementsUnchanged(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	orders := l.placements(l.ids.mainView, ElementTypeComponent)[l.ids.orders]
	assert.Equal(t, []float64{111, 222}, []float64{orders.x, orders.y})
	assert.Equal(t, "#ff0000", orders.color.String)
	invoicing := l.placements(l.ids.mainView, ElementTypeCapability)[l.ids.invoicing]
	assert.Equal(t, []float64{1, 2}, []float64{invoicing.x, invoicing.y})
	assert.Equal(t, "#00ff00", invoicing.color.String)
	acme := l.placements(l.ids.mainView, ElementTypeOriginEntity)[l.ids.acme]
	assert.Equal(t, []float64{5, 6}, []float64{acme.x, acme.y})
}

func TestMembershipRepair_DeletedAndUnknownViewsHoldNoPlacements(t *testing.T) {
	l := seededRepairLandscape(t)

	l.runRepair()

	assert.Zero(t, l.viewPlacementCount(l.ids.deletedView))
	assert.Zero(t, l.viewPlacementCount(l.ids.unknownView))
	assert.Len(t, l.history(l.ids.deletedView), 4, "deleted views get no appended removals")
}

func TestMembershipRepair_SecondRunInsertsNoEventsAndDeletesNoRows(t *testing.T) {
	l := seededRepairLandscape(t)
	l.runRepair()
	afterFirstRun := l.snapshot()

	l.runRepair()

	assert.Equal(t, afterFirstRun, l.snapshot())
}

func TestMembershipRepair_ScopedRunLeavesOtherTenantsUntouched(t *testing.T) {
	l := seededRepairLandscape(t)
	repaired := l.tenant
	controlTenant := fmt.Sprintf("vm-control-%d", time.Now().UnixNano())
	l.tenant = controlTenant
	l.cleanupTenantOnExit()
	l.seedView(seededView{l.ids.mainView, false}, viewCreated(l.ids.mainView, "architect"), added(l.ids.mainView, l.ids.billing, point{0, 0}))
	l.seedPlacements(
		placement{l.ids.mainView, l.ids.billing, ElementTypeComponent, 1, 1, sql.NullString{}},
		placement{l.ids.mainView, l.ids.ghost, ElementTypeComponent, 2, 2, sql.NullString{}},
		placement{l.ids.mainView, l.ids.missingCapability, ElementTypeCapability, 3, 3, sql.NullString{}})
	before := l.snapshot()

	l.tenant = repaired
	l.runRepair()

	l.tenant = controlTenant
	assert.Equal(t, before, l.snapshot())
}
