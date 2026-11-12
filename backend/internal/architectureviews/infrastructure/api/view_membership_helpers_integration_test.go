//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	sharedcontext "easi/backend/internal/shared/context"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type viewKey string

type elementKey string

type userKey string

type viewTitle string

type eventName string

type idPrefix string

type pathSegment string

type elementKind struct {
	name        string
	elementType repositories.ElementType
	segment     pathSegment
	idParam     string
	addBody     func(id elementKey, pos position) any
	add         func(h *viewTestHarness) http.HandlerFunc
	move        func(h *viewTestHarness) http.HandlerFunc
}

var applicationKind = elementKind{
	name: "application", elementType: repositories.ElementTypeComponent, segment: "components", idParam: "componentId",
	addBody: func(id elementKey, pos position) any {
		return AddComponentRequest{ComponentID: string(id), X: pos.x, Y: pos.y}
	},
	add:  func(h *viewTestHarness) http.HandlerFunc { return h.componentHandlers.AddComponentToView },
	move: func(h *viewTestHarness) http.HandlerFunc { return h.componentHandlers.UpdateComponentPosition },
}

var capabilityKind = elementKind{
	name: "capability", elementType: repositories.ElementTypeCapability, segment: "capabilities", idParam: "capabilityId",
	addBody: func(id elementKey, pos position) any {
		return AddCapabilityRequest{CapabilityID: string(id), X: pos.x, Y: pos.y}
	},
	add:  func(h *viewTestHarness) http.HandlerFunc { return h.elementHandlers.AddCapabilityToView },
	move: func(h *viewTestHarness) http.HandlerFunc { return h.elementHandlers.UpdateCapabilityPosition },
}

var originEntityKind = elementKind{
	name: "origin entity", elementType: repositories.ElementTypeOriginEntity, segment: "origin-entities", idParam: "originEntityId",
	addBody: func(id elementKey, pos position) any {
		return AddOriginEntityRequest{OriginEntityID: string(id), X: pos.x, Y: pos.y}
	},
	add:  func(h *viewTestHarness) http.HandlerFunc { return h.elementHandlers.AddOriginEntityToView },
	move: func(h *viewTestHarness) http.HandlerFunc { return h.elementHandlers.UpdateOriginEntityPosition },
}

var allElementKinds = []elementKind{applicationKind, capabilityKind, originEntityKind}

type viewClient struct {
	t      *testing.T
	tc     *viewTestContext
	h      *viewTestHarness
	userID userKey
}

func newViewClient(t *testing.T, tc *viewTestContext, h *viewTestHarness, userID userKey) *viewClient {
	return &viewClient{t: t, tc: tc, h: h, userID: userID}
}

func newElement(prefix idPrefix) elementKey {
	return elementKey(string(prefix) + "-" + uuid.NewString())
}

func newUser(prefix idPrefix) userKey {
	return userKey(string(prefix) + "-" + uuid.NewString())
}

func actorContext(userID userKey) context.Context {
	ctx := sharedcontext.WithTenant(context.Background(), sharedvo.DefaultTenantID())
	return sharedcontext.WithActor(ctx, sharedcontext.Actor{
		ID:          string(userID),
		Email:       string(userID) + "@example.com",
		Permissions: map[string]bool{"views:write": true},
	})
}

type apiCall struct {
	handler http.HandlerFunc
	method  string
	url     string
	body    any
	params  map[string]string
}

func (c *viewClient) send(call apiCall) *httptest.ResponseRecorder {
	c.t.Helper()
	var raw []byte
	if call.body != nil {
		var err error
		raw, err = json.Marshal(call.body)
		require.NoError(c.t, err)
	}
	w, req := c.tc.makeRequest(c.t, requestSpec{call.method, call.url, raw, call.params})
	call.handler(w, req.WithContext(sharedcontext.WithActor(req.Context(), sharedcontext.Actor{
		ID:          string(c.userID),
		Email:       string(c.userID) + "@example.com",
		Permissions: map[string]bool{"views:write": true},
	})))
	return w
}

func viewURL(viewID viewKey, segments ...pathSegment) string {
	url := "/api/v1/views/" + string(viewID)
	for _, segment := range segments {
		url += "/" + string(segment)
	}
	return url
}

func (c *viewClient) createView(name viewTitle) viewKey {
	c.t.Helper()
	w := c.send(apiCall{c.h.viewHandlers.CreateView, http.MethodPost, "/api/v1/views", CreateViewRequest{Name: string(name)}, nil})
	require.Equal(c.t, http.StatusCreated, w.Code, w.Body.String())
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(c.t, json.NewDecoder(w.Body).Decode(&created))
	c.tc.trackID(created.ID)
	return viewKey(created.ID)
}

func (c *viewClient) createPublicView(name viewTitle) viewKey {
	c.t.Helper()
	viewID := c.createView(name)
	w := c.send(apiCall{c.h.viewHandlers.ChangeVisibility, http.MethodPatch, viewURL(viewID, "visibility"),
		ChangeVisibilityRequest{IsPrivate: false}, map[string]string{"id": string(viewID)}})
	require.Equal(c.t, http.StatusNoContent, w.Code, w.Body.String())
	return viewID
}

func (c *viewClient) place(kind elementKind, viewID viewKey, elementID elementKey, pos position) {
	c.t.Helper()
	w := c.send(apiCall{kind.add(c.h), http.MethodPost, viewURL(viewID, kind.segment), kind.addBody(elementID, pos), map[string]string{"id": string(viewID)}})
	require.Equal(c.t, http.StatusCreated, w.Code, w.Body.String())
}

func (c *viewClient) move(kind elementKind, viewID viewKey, elementID elementKey, pos position) int {
	c.t.Helper()
	body := ElementPositionRequest{X: pos.x, Y: pos.y}
	params := map[string]string{"id": string(viewID), kind.idParam: string(elementID)}
	return c.send(apiCall{kind.move(c.h), http.MethodPatch, viewURL(viewID, kind.segment, pathSegment(elementID), "position"), body, params}).Code
}

func (c *viewClient) saveLayout(viewID viewKey, items ...PositionUpdateItem) int {
	c.t.Helper()
	body := UpdateMultiplePositionsRequest{Positions: items}
	return c.send(apiCall{c.h.componentHandlers.UpdateMultiplePositions, http.MethodPatch, viewURL(viewID, "layout"), body, map[string]string{"id": string(viewID)}}).Code
}

func (c *viewClient) removeApplication(viewID viewKey, componentID elementKey) int {
	c.t.Helper()
	params := map[string]string{"id": string(viewID), "componentId": string(componentID)}
	return c.send(apiCall{c.h.componentHandlers.RemoveComponentFromView, http.MethodDelete, viewURL(viewID, "components", pathSegment(componentID)), nil, params}).Code
}

func (c *viewClient) deleteView(viewID viewKey) {
	c.t.Helper()
	w := c.send(apiCall{c.h.viewHandlers.DeleteView, http.MethodDelete, viewURL(viewID), nil, map[string]string{"id": string(viewID)}})
	require.Equal(c.t, http.StatusNoContent, w.Code, w.Body.String())
}

func (c *viewClient) view(viewID viewKey) readmodels.ArchitectureViewDTO {
	c.t.Helper()
	return c.tc.getViewViaAPI(c.t, c.h, string(viewID))
}

func deletionOf(eventType eventName, entityID elementKey) domain.DomainEvent {
	data := fmt.Sprintf(`{"id":%q}`, entityID)
	return domain.NewGenericDomainEvent(string(entityID), string(eventType), []byte(data), time.Now().UTC())
}

func (h *viewTestHarness) publishInTx(ctx context.Context, deletions ...domain.DomainEvent) error {
	return h.tenantDB.RunInTx(ctx, func(txCtx context.Context) error {
		return h.eventBus.Publish(txCtx, deletions)
	})
}

type storedPlacement struct {
	elementID   string
	elementType string
	x, y        float64
	color       sql.NullString
}

func (h *viewTestHarness) storedPlacements(t *testing.T, viewID viewKey) []storedPlacement {
	t.Helper()
	var placements []storedPlacement
	err := h.tenantDB.WithReadOnlyTx(tenantContext(), func(tx *sql.Tx) error {
		rows, err := tx.Query(`SELECT element_id, element_type, x, y, custom_color FROM architectureviews.view_element_positions
			WHERE tenant_id = $1 AND view_id = $2 ORDER BY element_type, element_id`, testTenantID(), string(viewID))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p storedPlacement
			if err := rows.Scan(&p.elementID, &p.elementType, &p.x, &p.y, &p.color); err != nil {
				return err
			}
			placements = append(placements, p)
		}
		return rows.Err()
	})
	require.NoError(t, err)
	return placements
}

func (h *viewTestHarness) placedElementIDs(t *testing.T, viewID viewKey) []elementKey {
	t.Helper()
	ids := []elementKey{}
	for _, p := range h.storedPlacements(t, viewID) {
		ids = append(ids, elementKey(p.elementID))
	}
	return ids
}

func (h *viewTestHarness) insertPlacement(t *testing.T, viewID viewKey, elementID elementKey, elementType repositories.ElementType) {
	t.Helper()
	_, err := h.tenantDB.ExecContext(tenantContext(), `INSERT INTO architectureviews.view_element_positions
		(view_id, tenant_id, element_id, element_type, x, y, created_at, updated_at) VALUES ($1, $2, $3, $4, 10, 20, NOW(), NOW())`,
		string(viewID), testTenantID(), string(elementID), string(elementType))
	require.NoError(t, err)
}

func (h *viewTestHarness) viewEventTypes(t *testing.T, viewID viewKey) []string {
	t.Helper()
	var eventTypes []string
	err := h.tenantDB.WithReadOnlyTx(tenantContext(), func(tx *sql.Tx) error {
		rows, err := tx.Query(`SELECT event_type FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2 ORDER BY version`,
			testTenantID(), string(viewID))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var eventType string
			if err := rows.Scan(&eventType); err != nil {
				return err
			}
			eventTypes = append(eventTypes, eventType)
		}
		return rows.Err()
	})
	require.NoError(t, err)
	return eventTypes
}

func (h *viewTestHarness) aggregateListsApplication(t *testing.T, viewID viewKey, componentID elementKey) bool {
	t.Helper()
	view, err := h.viewRepo.GetByID(tenantContext(), string(viewID))
	require.NoError(t, err)
	return view.HasComponent(string(componentID))
}

func componentIDsOf(view readmodels.ArchitectureViewDTO) []elementKey {
	ids := []elementKey{}
	for _, c := range view.Components {
		ids = append(ids, elementKey(c.ComponentID))
	}
	return ids
}

func capabilityIDsOf(view readmodels.ArchitectureViewDTO) []elementKey {
	ids := []elementKey{}
	for _, c := range view.Capabilities {
		ids = append(ids, elementKey(c.CapabilityID))
	}
	return ids
}

func originEntityIDsOf(view readmodels.ArchitectureViewDTO) []elementKey {
	ids := []elementKey{}
	for _, o := range view.OriginEntities {
		ids = append(ids, elementKey(o.OriginEntityID))
	}
	return ids
}
