//go:build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authPL "easi/backend/internal/auth/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	"easi/backend/internal/testing/testdb"
)

type permissionMiddleware struct{}

func (permissionMiddleware) RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := sharedctx.GetActor(r.Context())
			if !ok || !actor.HasPermission(permission.String()) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type stewardshipStack struct {
	t         *testing.T
	admin     *sql.DB
	bus       *events.InMemoryEventBus
	router    chi.Router
	domainIDs []string
	userIDs   []string
}

func tenantContext() context.Context {
	return sharedctx.WithTenant(context.Background(), sharedvo.DefaultTenantID())
}

func newStewardshipStack(t *testing.T) *stewardshipStack {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	tenantDB := database.NewTenantAwareDB(testdb.Open(t))
	store := eventstore.NewPostgresEventStore(tenantDB)
	bus := events.NewInMemoryEventBus()
	store.SetEventBus(bus)

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := sharedctx.Role(r.Header.Get("X-Test-Role"))
			ctx := sharedctx.WithActor(sharedctx.WithTenant(r.Context(), sharedvo.DefaultTenantID()),
				sharedctx.NewActor("actor-"+string(role), string(role)+"@example.com", role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	require.NoError(t, SetupRoutes(RoutesDeps{
		Router:         router,
		CommandBus:     cqrs.NewInMemoryCommandBus(),
		EventStore:     store,
		EventBus:       bus,
		DB:             tenantDB,
		HATEOAS:        sharedAPI.NewHATEOASLinks("/api/v1"),
		AuthMiddleware: permissionMiddleware{},
	}))

	s := &stewardshipStack{t: t, admin: testdb.OpenAdmin(t), bus: bus, router: router}
	t.Cleanup(s.cleanup)
	return s
}

func (s *stewardshipStack) cleanup() {
	tenant := sharedvo.DefaultTenantID().Value()
	for _, domainID := range s.domainIDs {
		_, _ = s.admin.Exec(`DELETE FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id IN (
			SELECT aggregate_id FROM infrastructure.events WHERE tenant_id = $1 AND event_type = 'StewardAssigned' AND event_data::jsonb->>'domainId' = $2)`, tenant, domainID)
		_, _ = s.admin.Exec(`DELETE FROM stewardship.stewardships WHERE tenant_id = $1 AND domain_id = $2`, tenant, domainID)
		_, _ = s.admin.Exec(`DELETE FROM stewardship.domain_cache WHERE tenant_id = $1 AND domain_id = $2`, tenant, domainID)
	}
	for _, userID := range s.userIDs {
		_, _ = s.admin.Exec(`DELETE FROM stewardship.user_cache WHERE tenant_id = $1 AND user_id = $2`, tenant, userID)
	}
}

func (s *stewardshipStack) publish(aggregateID, eventType string, data map[string]any) {
	s.t.Helper()
	payload, err := json.Marshal(data)
	require.NoError(s.t, err)
	event := domain.NewGenericDomainEvent(aggregateID, eventType, payload, time.Now().UTC())
	require.NoError(s.t, s.bus.Publish(tenantContext(), []domain.DomainEvent{event}))
}

func (s *stewardshipStack) domain(name, architectID string) string {
	id := uuid.NewString()
	s.domainIDs = append(s.domainIDs, id)
	s.publish(id, capPL.BusinessDomainCreated, map[string]any{"id": id, "name": name, "domainArchitectId": architectID})
	return id
}

func (s *stewardshipStack) user(name string) string {
	id := uuid.NewString()
	s.userIDs = append(s.userIDs, id)
	s.publish(id, authPL.UserCreated, map[string]any{"id": id, "name": name, "email": id + "@example.com", "status": "active"})
	return id
}

func (s *stewardshipStack) do(role sharedctx.Role, method, path string, body any) *httptest.ResponseRecorder {
	s.t.Helper()
	var reader bytes.Buffer
	if body != nil {
		require.NoError(s.t, json.NewEncoder(&reader).Encode(body))
	}
	req := httptest.NewRequest(method, path, &reader)
	req.Header.Set("X-Test-Role", string(role))
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func (s *stewardshipStack) assign(domainID, concern, stewardID string) *httptest.ResponseRecorder {
	return s.do(sharedctx.RoleArchitect, http.MethodPut, "/stewardships/"+domainID+"/"+concern, AssignStewardRequest{StewardID: stewardID})
}

func (s *stewardshipStack) release(domainID, concern string) *httptest.ResponseRecorder {
	return s.do(sharedctx.RoleArchitect, http.MethodDelete, "/stewardships/"+domainID+"/"+concern, nil)
}

type listedStewardships struct {
	DomainID   string          `json:"domainId"`
	DomainName string          `json:"domainName"`
	Fallback   *listedPerson   `json:"fallback"`
	Data       []listedConcern `json:"data"`
	Links      sharedAPI.Links `json:"_links"`
}

type listedPerson struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}

type listedConcern struct {
	Concern     string          `json:"concern"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Steward     *listedPerson   `json:"steward"`
	AssignedBy  *string         `json:"assignedBy"`
	AssignedAt  *time.Time      `json:"assignedAt"`
	Links       sharedAPI.Links `json:"_links"`
}

func (s *stewardshipStack) list(role sharedctx.Role, domainID string) listedStewardships {
	s.t.Helper()
	rec := s.do(role, http.MethodGet, "/stewardships?domainId="+domainID, nil)
	require.Equal(s.t, http.StatusOK, rec.Code, rec.Body.String())
	var body listedStewardships
	require.NoError(s.t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

func (l listedStewardships) concern(name string) listedConcern {
	for _, c := range l.Data {
		if c.Concern == name {
			return c
		}
	}
	return listedConcern{}
}

func (s *stewardshipStack) eventCount(domainID string) int {
	s.t.Helper()
	var count int
	require.NoError(s.t, s.admin.QueryRow(`SELECT count(*) FROM infrastructure.events
		WHERE tenant_id = $1 AND event_type IN ('StewardAssigned', 'StewardReleased') AND event_data::jsonb->>'domainId' = $2`,
		sharedvo.DefaultTenantID().Value(), domainID).Scan(&count))
	return count
}

func stewardName(c listedConcern) string {
	if c.Steward == nil || c.Steward.Name == nil {
		return ""
	}
	return *c.Steward.Name
}

func TestStewardships_EveryConcernIsListedAssignedOrNot(t *testing.T) {
	s := newStewardshipStack(t)
	alice := s.user("Alice Smith")
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", alice)
	require.Equal(t, http.StatusOK, s.assign(ce, "ownership", mette).Code)

	body := s.list(sharedctx.RoleStakeholder, ce)

	assert.Equal(t, "Customer Engagement", body.DomainName)
	var concerns []string
	for _, c := range body.Data {
		concerns = append(concerns, c.Concern)
		assert.NotEmpty(t, c.Label)
		assert.NotEmpty(t, c.Description)
	}
	assert.Equal(t, []string{"ownership", "assessment", "documentation", "planning", "structure"}, concerns)
	assert.Equal(t, "Mette Gram", stewardName(body.concern("ownership")))
	for _, c := range body.Data[1:] {
		assert.Nil(t, c.Steward, c.Concern)
	}
	require.NotNil(t, body.Fallback)
	assert.Equal(t, alice, body.Fallback.ID)
	assert.Equal(t, "Alice Smith", *body.Fallback.Name)
	assert.Equal(t, "/api/v1/stewardships?domainId="+ce, body.Links["self"].Href)
	assert.Equal(t, "GET", body.Links["self"].Method)
	assert.Equal(t, "/api/v1/business-domains/"+ce, body.Links["x-domain"].Href)
}

func TestStewardships_NoFallbackWithoutADomainArchitect(t *testing.T) {
	s := newStewardshipStack(t)
	logistics := s.domain("Logistics", "")

	body := s.list(sharedctx.RoleStakeholder, logistics)

	assert.Nil(t, body.Fallback)
	for _, c := range body.Data {
		assert.Nil(t, c.Steward)
	}
}

func TestStewardships_FallbackWhoseNameIsUnknown(t *testing.T) {
	s := newStewardshipStack(t)
	ghost := uuid.NewString()
	logistics := s.domain("Logistics", ghost)

	body := s.list(sharedctx.RoleStakeholder, logistics)

	require.NotNil(t, body.Fallback)
	assert.Equal(t, ghost, body.Fallback.ID)
	assert.Nil(t, body.Fallback.Name)
}

func TestStewardships_AssigningIsAttributed(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")

	rec := s.assign(ce, "assessment", mette)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var item listedConcern
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &item))
	assert.Equal(t, "Mette Gram", stewardName(item))
	assert.Equal(t, "architect@example.com", *item.AssignedBy)
	require.NotNil(t, item.AssignedAt)
	assert.WithinDuration(t, time.Now(), *item.AssignedAt, time.Minute)
	assert.Equal(t, "/api/v1/stewardships/"+ce+"/assessment", item.Links["self"].Href)
	assert.Equal(t, "PUT", item.Links["x-assign"].Method)
	assert.Equal(t, "DELETE", item.Links["x-release"].Method)
	assert.Equal(t, 1, s.eventCount(ce))
}

func TestStewardships_RejectedAssignmentsRecordNothing(t *testing.T) {
	s := newStewardshipStack(t)
	jonas := s.user("Jonas Holm")
	s.publish(jonas, authPL.UserDisabled, map[string]any{"id": jonas})
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")

	assert.Equal(t, http.StatusBadRequest, s.assign(ce, "assessment", jonas).Code, "disabled user")
	assert.Equal(t, http.StatusBadRequest, s.assign(ce, "assessment", uuid.NewString()).Code, "unknown user")
	assert.Equal(t, http.StatusBadRequest, s.assign(ce, "budget", mette).Code, "unknown concern")
	assert.Equal(t, 0, s.eventCount(ce))
	assert.Nil(t, s.list(sharedctx.RoleArchitect, ce).concern("assessment").Steward)
}

func TestStewardships_ReenabledUserIsAssignableAgain(t *testing.T) {
	s := newStewardshipStack(t)
	jonas := s.user("Jonas Holm")
	ce := s.domain("Customer Engagement", "")
	s.publish(jonas, authPL.UserDisabled, map[string]any{"id": jonas})
	s.publish(jonas, authPL.UserEnabled, map[string]any{"id": jonas})

	assert.Equal(t, http.StatusOK, s.assign(ce, "assessment", jonas).Code)
}

func TestStewardships_ReassigningReplacesTheSteward(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	jonas := s.user("Jonas Holm")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)

	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", jonas).Code)

	assert.Equal(t, "Jonas Holm", stewardName(s.list(sharedctx.RoleArchitect, ce).concern("assessment")))
	assert.Equal(t, 2, s.eventCount(ce))
}

func TestStewardships_AssigningTheCurrentStewardAgainChangesNothing(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)
	before := s.list(sharedctx.RoleArchitect, ce).concern("assessment")

	rec := s.do(sharedctx.RoleAdmin, http.MethodPut, "/stewardships/"+ce+"/assessment", AssignStewardRequest{StewardID: mette})

	require.Equal(t, http.StatusOK, rec.Code)
	after := s.list(sharedctx.RoleArchitect, ce).concern("assessment")
	assert.Equal(t, "Mette Gram", stewardName(after))
	assert.Equal(t, "architect@example.com", *after.AssignedBy)
	assert.True(t, before.AssignedAt.Equal(*after.AssignedAt))
	assert.Equal(t, 1, s.eventCount(ce))
}

func TestStewardships_OneUserMayStewardManyConcerns(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)

	require.Equal(t, http.StatusOK, s.assign(ce, "documentation", mette).Code)

	body := s.list(sharedctx.RoleArchitect, ce)
	assert.Equal(t, "Mette Gram", stewardName(body.concern("assessment")))
	assert.Equal(t, "Mette Gram", stewardName(body.concern("documentation")))
}

func TestStewardships_ReleasingEndsTheStewardship(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)

	require.Equal(t, http.StatusNoContent, s.release(ce, "assessment").Code)

	released := s.list(sharedctx.RoleArchitect, ce).concern("assessment")
	assert.Nil(t, released.Steward)
	_, hasRelease := released.Links["x-release"]
	assert.False(t, hasRelease)
	assert.Equal(t, 2, s.eventCount(ce))
}

func TestStewardships_ReleasingAnUnassignedConcernChangesNothing(t *testing.T) {
	s := newStewardshipStack(t)
	ce := s.domain("Customer Engagement", "")

	assert.Equal(t, http.StatusNoContent, s.release(ce, "planning").Code)
	assert.Equal(t, 0, s.eventCount(ce))
}

func TestStewardships_ReadersSeeWritersChange(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)

	body := s.list(sharedctx.RoleStakeholder, ce)

	assert.Equal(t, "Mette Gram", stewardName(body.concern("assessment")))
	_, hasCandidates := body.Links["x-candidates"]
	assert.False(t, hasCandidates)
	for _, c := range body.Data {
		assert.Equal(t, "GET", c.Links["self"].Method)
		_, hasAssign := c.Links["x-assign"]
		_, hasRelease := c.Links["x-release"]
		assert.False(t, hasAssign, c.Concern)
		assert.False(t, hasRelease, c.Concern)
	}
	assert.Equal(t, http.StatusForbidden, s.do(sharedctx.RoleStakeholder, http.MethodPut, "/stewardships/"+ce+"/assessment", AssignStewardRequest{StewardID: mette}).Code)
	assert.Equal(t, http.StatusForbidden, s.do(sharedctx.RoleStakeholder, http.MethodDelete, "/stewardships/"+ce+"/assessment", nil).Code)

	writerLinks := s.list(sharedctx.RoleArchitect, ce).Links
	assert.Equal(t, "/api/v1/users?status=active", writerLinks["x-candidates"].Href)
}

func TestStewardships_UnknownDomainIsNotFound(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	unknown := uuid.NewString()

	assert.Equal(t, http.StatusNotFound, s.do(sharedctx.RoleArchitect, http.MethodGet, "/stewardships?domainId="+unknown, nil).Code)
	assert.Equal(t, http.StatusNotFound, s.do(sharedctx.RoleArchitect, http.MethodGet, "/stewardships/"+unknown+"/planning", nil).Code)
	assert.Equal(t, http.StatusNotFound, s.assign(unknown, "planning", mette).Code)
	assert.Equal(t, http.StatusNotFound, s.release(unknown, "planning").Code)
}

func TestStewardships_UnknownConcernAndMissingDomainAreInvalid(t *testing.T) {
	s := newStewardshipStack(t)
	ce := s.domain("Customer Engagement", "")

	assert.Equal(t, http.StatusBadRequest, s.do(sharedctx.RoleArchitect, http.MethodGet, "/stewardships/"+ce+"/budget", nil).Code)
	assert.Equal(t, http.StatusBadRequest, s.release(ce, "budget").Code)
	assert.Equal(t, http.StatusBadRequest, s.do(sharedctx.RoleArchitect, http.MethodGet, "/stewardships", nil).Code)
}

func TestStewardships_DomainDeletionReleasesItsStewardships(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	jonas := s.user("Jonas Holm")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)
	require.Equal(t, http.StatusOK, s.assign(ce, "ownership", jonas).Code)

	s.publish(ce, capPL.BusinessDomainDeleted, map[string]any{"id": ce})

	var released int
	require.NoError(t, s.admin.QueryRow(`SELECT count(*) FROM infrastructure.events
		WHERE tenant_id = $1 AND event_type = 'StewardReleased' AND event_data::jsonb->>'domainId' = $2 AND event_data::jsonb->>'releasedBy' = 'system:domain-deleted'`,
		sharedvo.DefaultTenantID().Value(), ce).Scan(&released))
	assert.Equal(t, 2, released)
	var live int
	require.NoError(t, s.admin.QueryRow(`SELECT count(*) FROM stewardship.stewardships WHERE tenant_id = $1 AND domain_id = $2`,
		sharedvo.DefaultTenantID().Value(), ce).Scan(&live))
	assert.Zero(t, live)
	assert.Equal(t, http.StatusNotFound, s.do(sharedctx.RoleArchitect, http.MethodGet, "/stewardships?domainId="+ce, nil).Code)
}

func TestStewardships_ADisabledStewardStaysRecorded(t *testing.T) {
	s := newStewardshipStack(t)
	mette := s.user("Mette Gram")
	ce := s.domain("Customer Engagement", "")
	require.Equal(t, http.StatusOK, s.assign(ce, "assessment", mette).Code)

	s.publish(mette, authPL.UserDisabled, map[string]any{"id": mette})

	assert.Equal(t, "Mette Gram", stewardName(s.list(sharedctx.RoleArchitect, ce).concern("assessment")))
}

func TestStewardships_DomainRenameAndArchitectChangeReachTheList(t *testing.T) {
	s := newStewardshipStack(t)
	alice := s.user("Alice Smith")
	ce := s.domain("Customer Engagement", "")

	s.publish(ce, capPL.BusinessDomainUpdated, map[string]any{"id": ce, "name": "Customer Experience", "domainArchitectId": alice})

	body := s.list(sharedctx.RoleStakeholder, ce)
	assert.Equal(t, "Customer Experience", body.DomainName)
	require.NotNil(t, body.Fallback)
	assert.Equal(t, "Alice Smith", *body.Fallback.Name)
}
