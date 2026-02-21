//go:build !integration

package internal_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	pl "easi/backend/internal/archassistant/publishedlanguage"
	"easi/backend/internal/archassistant/infrastructure/toolimpls"
	authPL "easi/backend/internal/auth/publishedlanguage"
	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	capabilityAPI "easi/backend/internal/capabilitymapping/infrastructure/api"
	enterpriseArchAPI "easi/backend/internal/enterprisearchitecture/infrastructure/api"
	metamodelAPI "easi/backend/internal/metamodel/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"
	valuestreamsAPI "easi/backend/internal/valuestreams/infrastructure/api"

	"github.com/go-chi/chi/v5"
)

type noopEventStore struct{}

func (n *noopEventStore) SaveEvents(context.Context, string, []domain.DomainEvent, int) error {
	return nil
}
func (n *noopEventStore) GetEvents(context.Context, string) ([]domain.DomainEvent, error) {
	return nil, nil
}

type noopAuthMiddleware struct{}

func (n *noopAuthMiddleware) RequirePermission(_ authPL.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}

type noopSessionProvider struct{}

func (n *noopSessionProvider) GetCurrentUserID(context.Context) (string, error) {
	return "", nil
}

func buildToolCatalogTestRouter(t *testing.T) chi.Router {
	t.Helper()
	r := chi.NewRouter()
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	es := &noopEventStore{}
	auth := &noopAuthMiddleware{}
	hateoas := sharedAPI.NewHATEOASLinks("")

	if err := architectureAPI.SetupArchitectureModelingRoutes(architectureAPI.RouteConfig{
		Router: r, CommandBus: commandBus, EventStore: es, EventBus: eventBus,
		HATEOAS: hateoas, AuthMiddleware: auth,
	}); err != nil {
		t.Fatalf("architecture modeling routes: %v", err)
	}

	if err := capabilityAPI.SetupCapabilityMappingRoutes(&capabilityAPI.RouteConfig{
		Router: r, CommandBus: commandBus, EventStore: es, EventBus: eventBus,
		HATEOAS: hateoas, AuthMiddleware: auth,
	}); err != nil {
		t.Fatalf("capability mapping routes: %v", err)
	}

	if err := valuestreamsAPI.SetupValueStreamsRoutes(&valuestreamsAPI.RouteConfig{
		Router: r, CommandBus: commandBus, EventStore: es, EventBus: eventBus,
		HATEOAS: hateoas, AuthMiddleware: auth,
	}); err != nil {
		t.Fatalf("value streams routes: %v", err)
	}

	if err := enterpriseArchAPI.SetupEnterpriseArchitectureRoutes(enterpriseArchAPI.EnterpriseArchRoutesDeps{
		Router: r, CommandBus: commandBus, EventStore: es, EventBus: eventBus,
		AuthMiddleware: auth,
	}); err != nil {
		t.Fatalf("enterprise architecture routes: %v", err)
	}

	if err := metamodelAPI.SetupMetaModelRoutes(metamodelAPI.MetaModelRoutesDeps{
		Router: r, CommandBus: commandBus, EventStore: es, EventBus: eventBus,
		Hateoas: hateoas, AuthMiddleware: auth,
	}); err != nil {
		t.Fatalf("metamodel routes: %v", err)
	}

	return r
}

var chiParamPattern = regexp.MustCompile(`\{[^}]+\}`)

func normalizeChiPath(path string) string {
	return chiParamPattern.ReplaceAllString(path, "*")
}

func collectRegisteredRoutes(r chi.Router) map[string]bool {
	routes := make(map[string]bool)
	chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.TrimSuffix(route, "/")
		key := method + " " + normalizeChiPath(route)
		routes[key] = true
		return nil
	})
	return routes
}

func logSimilarRoutes(t *testing.T, path string, registeredRoutes map[string]bool) {
	t.Helper()
	prefix := strings.SplitN(normalizeChiPath(path), "/", 3)
	for route := range registeredRoutes {
		if len(prefix) >= 2 && strings.Contains(route, prefix[1]) {
			t.Logf("    %s", route)
		}
	}
}

func TestToolCatalog_AllPathsMatchRegisteredRoutes(t *testing.T) {
	router := buildToolCatalogTestRouter(t)
	registeredRoutes := collectRegisteredRoutes(router)

	specs := toolimpls.AllContextToolSpecs()
	if len(specs) == 0 {
		t.Fatal("no tool specs found — catalog may be broken")
	}

	for _, spec := range specs {
		key := spec.Method + " " + normalizeChiPath(spec.Path)
		if registeredRoutes[key] {
			continue
		}
		t.Errorf("DRIFT: tool %q declares %s %s but no matching route is registered", spec.Name, spec.Method, spec.Path)
		t.Logf("  registered routes with similar prefix:")
		logSimilarRoutes(t, spec.Path, registeredRoutes)
	}
}

func TestToolCatalog_AllPermissionsAreValid(t *testing.T) {
	for _, spec := range toolimpls.AllContextToolSpecs() {
		_, err := authPL.PermissionFromString(spec.Permission)
		if err != nil {
			t.Errorf("tool %q references invalid permission %q", spec.Name, spec.Permission)
		}
	}
}

func TestToolCatalog_NoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, spec := range toolimpls.AllContextToolSpecs() {
		if seen[spec.Name] {
			t.Errorf("duplicate tool name %q", spec.Name)
		}
		seen[spec.Name] = true
	}
}

func TestToolCatalog_AllAccessClassesAreValid(t *testing.T) {
	validAccess := map[pl.AccessClass]bool{
		pl.AccessRead: true, pl.AccessWrite: true,
		pl.AccessCreate: true, pl.AccessUpdate: true, pl.AccessDelete: true,
	}

	for _, spec := range toolimpls.AllContextToolSpecs() {
		if !validAccess[spec.Access] {
			t.Errorf("tool %q has invalid access class %q", spec.Name, spec.Access)
		}
	}
}

func TestToolCatalog_MethodMatchesAccessClass(t *testing.T) {
	expectedAccess := map[string][]pl.AccessClass{
		"GET":    {pl.AccessRead},
		"POST":   {pl.AccessCreate},
		"PUT":    {pl.AccessUpdate},
		"DELETE": {pl.AccessDelete},
		"PATCH":  {pl.AccessUpdate},
	}

	for _, spec := range toolimpls.AllContextToolSpecs() {
		allowed := expectedAccess[spec.Method]
		found := false
		for _, a := range allowed {
			if spec.Access == a {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %q: method %s should have access %v, got %q",
				spec.Name, spec.Method, allowed, spec.Access)
		}
	}
}

func TestToolCatalog_InfoUnexposedRoutes(t *testing.T) {
	router := buildToolCatalogTestRouter(t)

	toolRoutes := make(map[string]bool)
	for _, spec := range toolimpls.AllContextToolSpecs() {
		key := spec.Method + " " + normalizeChiPath(spec.Path)
		toolRoutes[key] = true
	}

	var uncovered []string
	chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.TrimSuffix(route, "/")
		key := method + " " + normalizeChiPath(route)
		if !toolRoutes[key] {
			uncovered = append(uncovered, fmt.Sprintf("  %s %s", method, route))
		}
		return nil
	})

	if len(uncovered) > 0 {
		t.Logf("INFO: %d routes exist without tool specs (excluded by design or not yet covered):", len(uncovered))
		for _, r := range uncovered {
			t.Logf("%s", r)
		}
	}
}
