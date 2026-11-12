//go:build integration
// +build integration

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	"easi/backend/internal/testing/testdb"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const membershipBackfillMigration = "172_backfill_architectureviews_view_membership.sql"

var repairTouchedTables = []string{
	"infrastructure.events",
	"architectureviews.view_element_positions",
	"architectureviews.architecture_views",
	"architecturemodeling.application_components",
	"architecturemodeling.vendors",
	"architecturemodeling.acquired_entities",
	"architecturemodeling.internal_teams",
	"capabilitymapping.capabilities",
}

type entityID string

type actorName string

type sqlText string

type hexColor string

type repairIDs struct {
	orders, crm, ghost, neverAdded, billing, payments entityID
	invoicing, missingCapability                      entityID
	acme, goneVendor, acquired, goneTeam              entityID
	mainView, privateView, deletedView, unknownView   entityID
}

func newRepairIDs() repairIDs {
	id := func() entityID { return entityID(uuid.NewString()) }
	return repairIDs{
		orders: id(), crm: id(), ghost: id(), neverAdded: id(), billing: id(), payments: id(),
		invoicing: id(), missingCapability: id(),
		acme: id(), goneVendor: id(), acquired: id(), goneTeam: id(),
		mainView: id(), privateView: id(), deletedView: id(), unknownView: id(),
	}
}

type repairLandscape struct {
	t      *testing.T
	admin  *sql.DB
	app    *sql.DB
	tenant string
	ids    repairIDs
}

func newRepairLandscape(t *testing.T) *repairLandscape {
	t.Helper()
	l := &repairLandscape{
		t:      t,
		admin:  testdb.OpenAdmin(t),
		app:    testdb.Open(t),
		tenant: fmt.Sprintf("vm-repair-%d", time.Now().UnixNano()),
		ids:    newRepairIDs(),
	}
	l.cleanupTenantOnExit()
	return l
}

func (l *repairLandscape) cleanupTenantOnExit() {
	tenant := l.tenant
	l.t.Cleanup(func() {
		for _, table := range repairTouchedTables {
			_, _ = l.admin.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", tenant)
		}
	})
}

func (l *repairLandscape) exec(query sqlText, args ...any) {
	l.t.Helper()
	_, err := l.admin.Exec(string(query), args...)
	require.NoError(l.t, err)
}

func (l *repairLandscape) seedSuppliers() {
	l.t.Helper()
	s := l.ids
	l.exec(`INSERT INTO architecturemodeling.application_components (id, tenant_id, name, created_at, is_deleted) VALUES
		($2, $1, 'Orders', NOW(), FALSE), ($3, $1, 'CRM', NOW(), FALSE), ($4, $1, 'Ghost', NOW(), FALSE),
		($5, $1, 'Never added', NOW(), FALSE), ($6, $1, 'Billing', NOW(), TRUE), ($7, $1, 'Payments', NOW(), TRUE)`,
		l.tenant, s.orders, s.crm, s.ghost, s.neverAdded, s.billing, s.payments)
	l.exec(`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, created_at) VALUES ($2, $1, 'Invoicing', 'L1', NOW())`,
		l.tenant, s.invoicing)
	l.exec(`INSERT INTO architecturemodeling.vendors (id, tenant_id, name, created_at, is_deleted) VALUES
		($2, $1, 'Acme', NOW(), FALSE), ($3, $1, 'Gone vendor', NOW(), TRUE)`, l.tenant, s.acme, s.goneVendor)
	l.exec(`INSERT INTO architecturemodeling.acquired_entities (id, tenant_id, name, created_at, is_deleted) VALUES ($2, $1, 'Acquired', NOW(), NULL)`,
		l.tenant, s.acquired)
}

type seededEvent struct {
	eventType string
	data      map[string]any
}

func viewCreated(viewID entityID, owner actorName) seededEvent {
	return seededEvent{"ViewCreated", map[string]any{
		"id": viewID, "name": "Landscape " + string(viewID[:8]), "description": "", "isPrivate": true,
		"ownerUserId": owner, "ownerEmail": string(owner) + "@example.com", "createdAt": time.Now().UTC(),
	}}
}

type point struct {
	x, y float64
}

func added(viewID, componentID entityID, at point) seededEvent {
	return seededEvent{"ComponentAddedToView", map[string]any{
		"viewId": viewID, "componentId": componentID, "x": at.x, "y": at.y, "addedAt": time.Now().UTC(),
	}}
}

func removed(viewID, componentID entityID) seededEvent {
	return seededEvent{"ComponentRemovedFromView", map[string]any{
		"viewId": viewID, "componentId": componentID, "removedAt": time.Now().UTC(),
	}}
}

func viewDeleted(viewID entityID) seededEvent {
	return seededEvent{"ViewDeleted", map[string]any{"viewId": viewID, "deletedAt": time.Now().UTC()}}
}

type seededView struct {
	id        entityID
	isDeleted bool
}

func (l *repairLandscape) seedView(view seededView, history ...seededEvent) {
	l.t.Helper()
	viewID, isDeleted := view.id, view.isDeleted
	l.exec(`INSERT INTO architectureviews.architecture_views (id, tenant_id, name, is_deleted, is_private, created_at) VALUES ($2, $1, 'Landscape', $3, TRUE, NOW())`,
		l.tenant, viewID, isDeleted)
	for i, event := range history {
		data, err := json.Marshal(event.data)
		require.NoError(l.t, err)
		l.exec(`INSERT INTO infrastructure.events (tenant_id, aggregate_id, event_type, event_data, version, occurred_at, actor_id, actor_email)
			VALUES ($1, $2, $3, $4, $5, NOW(), 'seed-user', 'seed@example.com')`, l.tenant, viewID, event.eventType, data, i+1)
	}
}

type placement struct {
	viewID, elementID entityID
	elementType       ElementType
	x, y              float64
	color             sql.NullString
}

func (l *repairLandscape) seedPlacements(placements ...placement) {
	l.t.Helper()
	for _, p := range placements {
		l.exec(`INSERT INTO architectureviews.view_element_positions (view_id, tenant_id, element_id, element_type, x, y, custom_color, created_at, updated_at)
			VALUES ($2, $1, $3, $4, $5, $6, $7, NOW(), NOW())`, l.tenant, p.viewID, p.elementID, string(p.elementType), p.x, p.y, p.color)
	}
}

func colored(hex hexColor) sql.NullString {
	return sql.NullString{String: string(hex), Valid: true}
}

func (l *repairLandscape) seedViewHistories() {
	l.t.Helper()
	s := l.ids
	l.seedView(seededView{s.mainView, false},
		viewCreated(s.mainView, "architect"),
		added(s.mainView, s.orders, point{5, 6}),
		added(s.mainView, s.billing, point{0, 0}),
		added(s.mainView, s.ghost, point{0, 0}),
		removed(s.mainView, s.ghost),
		added(s.mainView, s.crm, point{70, 80}),
		added(s.mainView, s.payments, point{0, 0}))
	l.seedView(seededView{s.privateView, false},
		viewCreated(s.privateView, "someone-else"),
		added(s.privateView, s.billing, point{0, 0}),
		added(s.privateView, s.payments, point{0, 0}),
		removed(s.privateView, s.payments))
	l.seedView(seededView{s.deletedView, true},
		viewCreated(s.deletedView, "architect"),
		added(s.deletedView, s.orders, point{0, 0}),
		added(s.deletedView, s.billing, point{0, 0}),
		viewDeleted(s.deletedView))
}

func (l *repairLandscape) seedViewPlacements() {
	l.t.Helper()
	s := l.ids
	l.seedPlacements(
		placement{s.mainView, s.orders, ElementTypeComponent, 111, 222, colored("#ff0000")},
		placement{s.mainView, s.billing, ElementTypeComponent, 10, 10, sql.NullString{}},
		placement{s.mainView, s.ghost, ElementTypeComponent, 20, 20, sql.NullString{}},
		placement{s.mainView, s.neverAdded, ElementTypeComponent, 30, 30, sql.NullString{}},
		placement{s.mainView, s.invoicing, ElementTypeCapability, 1, 2, colored("#00ff00")},
		placement{s.mainView, s.missingCapability, ElementTypeCapability, 3, 4, sql.NullString{}},
		placement{s.mainView, s.acme, ElementTypeOriginEntity, 5, 6, sql.NullString{}},
		placement{s.mainView, s.acquired, ElementTypeOriginEntity, 7, 8, sql.NullString{}},
		placement{s.mainView, s.goneVendor, ElementTypeOriginEntity, 9, 10, sql.NullString{}},
		placement{s.mainView, s.goneTeam, ElementTypeOriginEntity, 11, 12, sql.NullString{}},
		placement{s.privateView, s.billing, ElementTypeComponent, 13, 14, sql.NullString{}},
		placement{s.deletedView, s.orders, ElementTypeComponent, 15, 16, sql.NullString{}},
		placement{s.deletedView, s.invoicing, ElementTypeCapability, 17, 18, sql.NullString{}},
		placement{s.unknownView, s.invoicing, ElementTypeCapability, 19, 20, sql.NullString{}},
	)
}

func seededRepairLandscape(t *testing.T) *repairLandscape {
	t.Helper()
	l := newRepairLandscape(t)
	l.seedSuppliers()
	l.seedViewHistories()
	l.seedViewPlacements()
	return l
}

func (l *repairLandscape) runRepair() {
	l.t.Helper()
	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", membershipBackfillMigration))
	require.NoError(l.t, err)

	tx, err := l.app.Begin()
	require.NoError(l.t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(fmt.Sprintf("SET LOCAL app.current_tenant = '%s'", l.tenant))
	require.NoError(l.t, err)
	_, err = tx.Exec(string(migration))
	require.NoError(l.t, err)
	require.NoError(l.t, tx.Commit())
}

func (l *repairLandscape) placements(viewID entityID, elementType ElementType) map[entityID]placement {
	l.t.Helper()
	rows, err := l.admin.Query(`SELECT element_id, x, y, custom_color FROM architectureviews.view_element_positions
		WHERE tenant_id = $1 AND view_id = $2 AND element_type = $3`, l.tenant, viewID, string(elementType))
	require.NoError(l.t, err)
	defer func() { _ = rows.Close() }()
	result := map[entityID]placement{}
	for rows.Next() {
		p := placement{viewID: viewID, elementType: elementType}
		require.NoError(l.t, rows.Scan(&p.elementID, &p.x, &p.y, &p.color))
		result[p.elementID] = p
	}
	require.NoError(l.t, rows.Err())
	return result
}

func (l *repairLandscape) placedIDs(viewID entityID, elementType ElementType) []entityID {
	l.t.Helper()
	return sortedKeys(l.placements(viewID, elementType))
}

func (l *repairLandscape) viewPlacementCount(viewID entityID) int {
	l.t.Helper()
	var count int
	require.NoError(l.t, l.admin.QueryRow(`SELECT COUNT(*) FROM architectureviews.view_element_positions WHERE tenant_id = $1 AND view_id = $2`,
		l.tenant, viewID).Scan(&count))
	return count
}

type storedViewEvent struct {
	version     int
	eventType   string
	componentID entityID
	actorID     string
	actorEmail  string
	data        map[string]any
}

func (l *repairLandscape) history(viewID entityID) []storedViewEvent {
	l.t.Helper()
	rows, err := l.admin.Query(`SELECT version, event_type, event_data, actor_id, actor_email FROM infrastructure.events
		WHERE tenant_id = $1 AND aggregate_id = $2 ORDER BY version`, l.tenant, viewID)
	require.NoError(l.t, err)
	defer func() { _ = rows.Close() }()
	var events []storedViewEvent
	for rows.Next() {
		var e storedViewEvent
		var raw []byte
		require.NoError(l.t, rows.Scan(&e.version, &e.eventType, &raw, &e.actorID, &e.actorEmail))
		require.NoError(l.t, json.Unmarshal(raw, &e.data))
		componentID, _ := e.data["componentId"].(string)
		e.componentID = entityID(componentID)
		events = append(events, e)
	}
	require.NoError(l.t, rows.Err())
	return events
}

func (l *repairLandscape) loadView(viewID entityID) *aggregates.ArchitectureView {
	l.t.Helper()
	ctx := sharedctx.WithTenant(context.Background(), sharedvo.MustNewTenantID(l.tenant))
	repo := NewArchitectureViewRepository(eventstore.NewPostgresEventStore(database.NewTenantAwareDB(l.app)))
	view, err := repo.GetByID(ctx, string(viewID))
	require.NoError(l.t, err)
	return view
}

type tenantSnapshot struct {
	eventCount int
	placements []string
}

func (l *repairLandscape) snapshot() tenantSnapshot {
	l.t.Helper()
	var snap tenantSnapshot
	require.NoError(l.t, l.admin.QueryRow(`SELECT COUNT(*) FROM infrastructure.events WHERE tenant_id = $1`, l.tenant).Scan(&snap.eventCount))
	snap.placements = l.placementRows()
	return snap
}

const placementRowsSQL = `SELECT concat_ws('|', view_id, element_type, element_id, x, y, custom_color, updated_at)
	FROM architectureviews.view_element_positions WHERE tenant_id = $1 ORDER BY 1`

func (l *repairLandscape) placementRows() []string {
	l.t.Helper()
	rows, err := l.admin.Query(placementRowsSQL, l.tenant)
	require.NoError(l.t, err)
	defer func() { _ = rows.Close() }()
	var result []string
	for rows.Next() {
		var row string
		require.NoError(l.t, rows.Scan(&row))
		result = append(result, row)
	}
	require.NoError(l.t, rows.Err())
	return result
}

func sortedKeys[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedIDs[T ~string](ids ...T) []T {
	sorted := append([]T{}, ids...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted
}

func asEntityIDs(raw []string) []entityID {
	ids := make([]entityID, 0, len(raw))
	for _, id := range raw {
		ids = append(ids, entityID(id))
	}
	return ids
}
