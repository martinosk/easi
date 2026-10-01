//go:build integration

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	directionPL "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
	pl "easi/backend/internal/stewardship/publishedlanguage"
	"easi/backend/internal/testing/testdb"
)

var everyHomePermission = []string{"capabilities:read", "components:read", "architecture-direction:read", "domains:read"}

type homeStack struct {
	t      *testing.T
	ctx    context.Context
	bus    *events.InMemoryEventBus
	router chi.Router
	tenant string
}

type homeCaller struct {
	id          string
	email       string
	permissions []string
}

func newHomeStack(t *testing.T) *homeStack {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	tenantDB := database.NewTenantAwareDB(testdb.Open(t))
	bus := events.NewInMemoryEventBus()
	s := &homeStack{t: t, bus: bus, tenant: "home-api-" + uuid.NewString()[:8]}
	s.ctx = sharedctx.WithTenant(context.Background(), sharedvo.MustNewTenantID(s.tenant))

	s.router = chi.NewRouter()
	s.router.Use(s.actorFromHeaders)
	require.NoError(t, SetupRoutes(RoutesDeps{
		Router:         s.router,
		CommandBus:     cqrs.NewInMemoryCommandBus(),
		EventBus:       bus,
		DB:             tenantDB,
		HATEOAS:        sharedAPI.NewHATEOASLinks("/api/v1"),
		AuthMiddleware: permissionMiddleware{},
	}))
	admin := testdb.OpenAdmin(t)
	t.Cleanup(func() {
		for _, table := range []string{"stewardship.stewardships", "stewardship.domain_cache", "stewardship.user_cache",
			"stewardship.capability_cache", "stewardship.domain_assignment_cache", "stewardship.application_cache",
			"stewardship.realization_cache", "stewardship.time_assessment_cache", "stewardship.edit_grant_cache"} {
			_, _ = admin.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", s.tenant)
		}
	})
	return s
}

func (s *homeStack) actorFromHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := sharedctx.NewActor(r.Header.Get("X-Test-User"), r.Header.Get("X-Test-Email"), sharedctx.RoleStakeholder)
		actor.Permissions = map[string]bool{}
		for _, permission := range strings.Split(r.Header.Get("X-Test-Permissions"), ",") {
			actor.Permissions[permission] = true
		}
		ctx := sharedctx.WithActor(sharedctx.WithTenant(r.Context(), sharedvo.MustNewTenantID(s.tenant)), actor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *homeStack) publish(eventType string, data map[string]any) {
	s.t.Helper()
	payload, err := json.Marshal(data)
	require.NoError(s.t, err)
	event := domain.NewGenericDomainEvent(uuid.NewString(), eventType, payload, time.Now().UTC())
	require.NoError(s.t, s.bus.Publish(s.ctx, []domain.DomainEvent{event}))
}

func (s *homeStack) get(path string, caller homeCaller) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Test-User", caller.id)
	req.Header.Set("X-Test-Email", caller.email)
	req.Header.Set("X-Test-Permissions", strings.Join(caller.permissions, ","))
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func (s *homeStack) home(caller homeCaller) map[string]any {
	s.t.Helper()
	rec := s.get("/home", caller)
	require.Equal(s.t, http.StatusOK, rec.Code, rec.Body.String())
	var body map[string]any
	require.NoError(s.t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

type homeLandscape struct {
	engagement, finance, customerManagement, invoicing, crm string
	owner                                                   homeCaller
}

func newHomeCaller(permissions ...string) homeCaller {
	id := uuid.NewString()
	return homeCaller{id: id, email: id + "@example.com", permissions: permissions}
}

func (s *homeStack) seed() homeLandscape {
	l := homeLandscape{engagement: uuid.NewString(), finance: uuid.NewString(), customerManagement: uuid.NewString(),
		invoicing: uuid.NewString(), crm: uuid.NewString(), owner: newHomeCaller(everyHomePermission...)}
	accounting := uuid.NewString()
	s.publish(capPL.BusinessDomainCreated, map[string]any{"id": l.engagement, "name": "Customer Engagement"})
	s.publish(capPL.BusinessDomainCreated, map[string]any{"id": l.finance, "name": "Finance"})
	s.publish(capPL.CapabilityCreated, map[string]any{"id": l.customerManagement, "name": "Customer Management", "level": "L1"})
	s.publish(capPL.CapabilityCreated, map[string]any{"id": accounting, "name": "Accounting", "level": "L1"})
	s.publish(capPL.CapabilityCreated, map[string]any{"id": l.invoicing, "name": "Invoicing", "level": "L2", "parentId": accounting})
	s.publish(capPL.CapabilityAssignedToDomain, map[string]any{"capabilityId": l.customerManagement, "businessDomainId": l.engagement})
	s.publish(capPL.CapabilityAssignedToDomain, map[string]any{"capabilityId": accounting, "businessDomainId": l.finance})
	s.publish(amPL.ApplicationComponentCreated, map[string]any{"id": l.crm, "name": "CRM"})
	s.publish(capPL.SystemLinkedToCapability, map[string]any{"id": uuid.NewString(), "capabilityId": l.customerManagement, "componentId": l.crm})
	s.publish(capPL.SystemLinkedToCapability, map[string]any{"id": uuid.NewString(), "capabilityId": l.invoicing, "componentId": l.crm})
	s.publish(directionPL.TimeAssessmentRecorded, map[string]any{"capabilityId": l.invoicing, "componentId": l.crm, "grade": "Eliminate", "occurredOn": time.Now()})
	s.publish(capPL.CapabilityMetadataUpdated, map[string]any{"id": l.invoicing, "status": "Active", "eaOwner": l.owner.id})
	s.publish(amPL.ApplicationOwnershipConfirmed, map[string]any{"componentId": l.crm, "ownerKind": "user", "ownerId": l.owner.id, "ownershipState": "owned"})
	return l
}

func (s *homeStack) assignSteward(domainID, concern, stewardID string) string {
	id := uuid.NewString()
	s.publish(pl.StewardAssigned, map[string]any{"id": id, "domainId": domainID, "concern": concern, "stewardId": stewardID, "assignedBy": "admin@example.com", "assignedAt": time.Now()})
	return id
}

func path(body map[string]any, keys ...string) any {
	var current any = body
	for _, key := range keys {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[key]
	}
	return current
}

func items(body map[string]any) []map[string]any {
	raw, _ := path(body, "myWork", "items").([]any)
	result := make([]map[string]any, len(raw))
	for i, item := range raw {
		result[i] = item.(map[string]any)
	}
	return result
}

func TestHome_ComposedFromSupplierEventsWithCardLinks(t *testing.T) {
	s := newHomeStack(t)
	l := s.seed()

	rec := s.get("/home", l.owner)
	body := s.home(l.owner)

	assert.Equal(t, "private, no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "Cookie", rec.Header().Get("Vary"))
	assert.Equal(t, "personal", path(body, "scope", "kind"))
	assert.Equal(t, float64(1), path(body, "portfolio", "capabilities", "total"))
	assert.Equal(t, float64(1), path(body, "portfolio", "applications", "total"))
	assert.Equal(t, []any{"Customer Engagement", "Finance"}, path(body, "portfolio", "domains", "names"))
	assert.Equal(t, float64(2), path(body, "portfolio", "time", "total"))
	assert.Equal(t, "/api/v1/home", path(body, "_links", "self", "href"))
	cards := items(body)
	require.Len(t, cards, 2)
	assert.Equal(t, "Invoicing", cards[0]["name"])
	assert.Equal(t, "eliminate", cards[0]["dominantGrade"])
	assert.Equal(t, "/api/v1/one-pagers/capability/"+l.invoicing, path(cards[0], "_links", "x-one-pager", "href"))
	assert.Equal(t, "/api/v1/one-pagers/application/"+l.crm, path(cards[1], "_links", "x-one-pager", "href"))
}

func TestHome_SectionsFollowReadPermissions(t *testing.T) {
	s := newHomeStack(t)
	l := s.seed()
	s.assignSteward(l.engagement, "assessment", l.owner.id)
	without := func(missing string) homeCaller {
		caller := l.owner
		caller.permissions = nil
		for _, permission := range everyHomePermission {
			if permission != missing {
				caller.permissions = append(caller.permissions, permission)
			}
		}
		return caller
	}

	t.Run("capabilities:read", func(t *testing.T) {
		body := s.home(without("capabilities:read"))
		assert.Nil(t, path(body, "portfolio", "capabilities"))
		assert.NotNil(t, path(body, "portfolio", "applications"))
		assert.Equal(t, float64(1), path(body, "myWork", "total"))
		assert.Equal(t, "application", items(body)[0]["subjectType"])
	})
	t.Run("components:read", func(t *testing.T) {
		body := s.home(without("components:read"))
		assert.Nil(t, path(body, "portfolio", "applications"))
		assert.Equal(t, float64(1), path(body, "myWork", "total"))
		assert.Equal(t, "capability", items(body)[0]["subjectType"])
		assert.NotNil(t, path(items(body)[0], "_links", "x-one-pager"))
	})
	t.Run("architecture-direction:read", func(t *testing.T) {
		body := s.home(without("architecture-direction:read"))
		assert.Nil(t, path(body, "portfolio", "time"))
		for _, card := range items(body) {
			assert.NotContains(t, card, "dominantGrade")
		}
	})
	t.Run("domains:read", func(t *testing.T) {
		body := s.home(without("domains:read"))
		assert.Nil(t, path(body, "portfolio", "domains", "names"))
		assert.Equal(t, float64(2), path(body, "portfolio", "domains", "total"))
		stewardship := path(body, "scope", "stewardships").([]any)[0].(map[string]any)
		assert.NotContains(t, stewardship, "domain")
		assert.Equal(t, "assessment", stewardship["concern"])
	})
	t.Run("nothing readable is still 200", func(t *testing.T) {
		body := s.home(homeCaller{id: l.owner.id, email: l.owner.email})
		assert.Equal(t, "personal", path(body, "scope", "kind"))
		assert.Equal(t, float64(0), path(body, "myWork", "total"))
	})
}

func TestHome_IgnoresQueryParametersAndServesOnlyTheCaller(t *testing.T) {
	s := newHomeStack(t)
	l := s.seed()
	stranger := newHomeCaller(everyHomePermission...)

	body := s.home(stranger)
	rec := s.get("/home?userId="+l.owner.id+"&email="+l.owner.email, stranger)

	assert.Equal(t, "empty", path(body, "scope", "kind"))
	assert.Nil(t, path(body, "myWork"))
	assert.JSONEq(t, mustJSON(t, body), rec.Body.String())
}

func TestHome_FollowsChangesToAnchors(t *testing.T) {
	s := newHomeStack(t)
	l := s.seed()
	mette := newHomeCaller(everyHomePermission...)
	assessment := s.assignSteward(l.engagement, "assessment", mette.id)

	assert.Equal(t, []any{"Customer Engagement"}, path(s.home(mette), "portfolio", "domains", "names"))

	ownership := s.assignSteward(l.finance, "ownership", mette.id)
	assert.Equal(t, []any{"Customer Engagement", "Finance"}, path(s.home(mette), "portfolio", "domains", "names"))

	s.publish(pl.StewardReleased, map[string]any{"id": assessment})
	s.publish(pl.StewardReleased, map[string]any{"id": ownership})
	assert.Equal(t, "empty", path(s.home(mette), "scope", "kind"))
}

func TestHome_EditGrantLifecycleThroughEvents(t *testing.T) {
	s := newHomeStack(t)
	l := s.seed()
	ole := newHomeCaller(everyHomePermission...)
	grantID := uuid.NewString()
	s.publish(adPL.EditGrantActivated, map[string]any{"id": grantID, "artifactType": "component", "artifactId": l.crm,
		"granteeEmail": ole.email, "expiresAt": "2099-10-21T10:00:00Z"})

	cards := items(s.home(ole))
	require.Len(t, cards, 1)
	assert.Equal(t, "edit-grant", cards[0]["relation"])
	assert.Equal(t, "2099-10-21", cards[0]["grantExpiresOn"])

	s.publish(adPL.EditGrantExpired, map[string]any{"id": grantID})
	assert.Equal(t, "empty", path(s.home(ole), "scope", "kind"))
}

func TestHome_TenantFallbackForStewardAssigners(t *testing.T) {
	s := newHomeStack(t)
	s.seed()
	per := newHomeCaller(append([]string{"domains:write"}, everyHomePermission...)...)

	body := s.home(per)

	assert.Equal(t, "tenant", path(body, "scope", "kind"))
	assert.Equal(t, float64(3), path(body, "portfolio", "capabilities", "total"))
	assert.Equal(t, map[string]any{"total": float64(0), "items": []any{}}, path(body, "myWork"))
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return string(data)
}
