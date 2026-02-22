//go:build !integration

package internal_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"easi/backend/internal/archassistant/infrastructure/toolimpls"
	pl "easi/backend/internal/archassistant/publishedlanguage"
	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	authPL "easi/backend/internal/auth/publishedlanguage"
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

var excludedRoutes = map[string]string{
	"POST /capabilities/*/experts":                             "expert management — operational, not architecture exploration",
	"DELETE /capabilities/*/experts":                           "expert management — operational, not architecture exploration",
	"POST /capabilities/*/tags":                                "tag management — operational, not architecture exploration",
	"PATCH /capabilities/*/parent":                             "hierarchy reparenting — complex operation, not suitable for agent",
	"GET /capabilities/*/dependencies/incoming":                "per-capability view — use list_capability_dependencies instead",
	"GET /capabilities/*/dependencies/outgoing":                "per-capability view — use list_capability_dependencies instead",
	"GET /capability-realizations/by-component/*":              "realizations by component — available in get_application_details",
	"PUT /capability-realizations/*":                           "update realization level — fine-grained, use realize_capability",
	"DELETE /business-domains/*":                               "delete domain — high-impact, reserved for UI",
	"GET /business-domains/*/capabilities":                     "capabilities in domain — available in get_business_domain_details",
	"GET /business-domains/*/capability-realizations":          "domain realizations — composite view, reserved for UI",
	"GET /business-domains/*/capabilities/*/importance":        "per-domain-capability importance — use get_strategy_importance",
	"PUT /business-domains/*/capabilities/*/importance/*":      "update importance — fine-grained, use set_strategy_importance",
	"DELETE /business-domains/*/capabilities/*/importance/*":   "remove importance — fine-grained, reserved for UI",
	"GET /components/expert-roles":                             "reference data for UI dropdowns",
	"POST /components/*/experts":                               "expert management — operational, not architecture exploration",
	"DELETE /components/*/experts":                             "expert management — operational, not architecture exploration",
	"GET /components/*/fit-comparisons":                        "fit comparison view — composite UI visualization",
	"DELETE /components/*/fit-scores/*":                        "remove fit score — fine-grained, reserved for UI",
	"GET /components/*/origin/acquired-via":                    "specific origin type — use get_component_origin for all origins",
	"GET /components/*/origin/built-by":                        "specific origin type — use get_component_origin for all origins",
	"GET /components/*/origin/purchased-from":                  "specific origin type — use get_component_origin for all origins",
	"GET /origin-relationships":                                "all origin relationships — use get_component_origin per component",
	"GET /relations":                                           "all relations — covered by composite list_application_relations tool",
	"GET /relations/*":                                         "single relation — covered by composite list_application_relations tool",
	"GET /relations/from/*":                                    "outgoing relations — covered by composite list_application_relations tool",
	"GET /relations/to/*":                                      "incoming relations — covered by composite list_application_relations tool",
	"PUT /relations/*":                                         "update relation — fine-grained, use create/delete instead",
	"DELETE /acquired-entities/*":                              "origin entity delete — high-impact cascading operation, reserved for UI",
	"DELETE /vendors/*":                                        "origin entity delete — high-impact cascading operation, reserved for UI",
	"DELETE /internal-teams/*":                                 "origin entity delete — high-impact cascading operation, reserved for UI",
	"GET /enterprise-capabilities/*/links":                     "linked capabilities — available in get_enterprise_capability_details",
	"PUT /enterprise-capabilities/*/target-maturity":           "set target maturity — fine-grained, reserved for UI",
	"PUT /enterprise-capabilities/*/strategic-importance/*":    "update importance — fine-grained, use set_enterprise_strategic_importance",
	"DELETE /enterprise-capabilities/*/strategic-importance/*": "remove importance — fine-grained, reserved for UI",
	"GET /domain-capabilities/enterprise-link-status":          "batch link status — UI helper for enterprise capability mapping screen",
	"GET /domain-capabilities/*/enterprise-capability":         "reverse lookup — UI helper for capability mapping screen",
	"GET /domain-capabilities/*/enterprise-link-status":        "single link status — UI helper for capability mapping screen",
	"GET /meta-model/configurations/*":                         "metamodel config by ID — use get_maturity_scale instead",
	"GET /meta-model/strategy-pillars/*":                       "single pillar detail — use get_strategy_pillars for all",
	"PUT /meta-model/maturity-scale":                           "metamodel write — blocked by permission ceiling",
	"POST /meta-model/maturity-scale/reset":                    "metamodel write — blocked by permission ceiling",
	"PATCH /meta-model/strategy-pillars":                       "metamodel write — blocked by permission ceiling",
	"POST /meta-model/strategy-pillars":                        "metamodel write — blocked by permission ceiling",
	"PUT /meta-model/strategy-pillars/*":                       "metamodel write — blocked by permission ceiling",
	"DELETE /meta-model/strategy-pillars/*":                    "metamodel write — blocked by permission ceiling",
	"PUT /meta-model/strategy-pillars/*/fit-configuration":     "metamodel write — blocked by permission ceiling",
	"DELETE /value-streams/*":                                  "value stream delete — high-impact, reserved for UI",
	"DELETE /value-streams/*/stages/*":                         "stage delete — reserved for UI",
	"DELETE /value-streams/*/stages/*/capabilities/*":          "stage-capability unmapping — reserved for UI",
}

func TestToolCatalog_AllRoutesAccountedFor(t *testing.T) {
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
		if toolRoutes[key] {
			return nil
		}
		if _, excluded := excludedRoutes[key]; excluded {
			return nil
		}
		uncovered = append(uncovered, fmt.Sprintf("  %s %s", method, route))
		return nil
	})

	if len(uncovered) > 0 {
		t.Errorf("%d routes have no tool spec and are not explicitly excluded — add a tool or exclude with reason:", len(uncovered))
		for _, r := range uncovered {
			t.Errorf("%s", r)
		}
	}
}

func TestToolCatalog_NoStaleExclusions(t *testing.T) {
	router := buildToolCatalogTestRouter(t)
	registeredRoutes := collectRegisteredRoutes(router)

	for excluded := range excludedRoutes {
		if !registeredRoutes[excluded] {
			t.Errorf("STALE exclusion: %q is excluded but no such route exists — remove from excludedRoutes", excluded)
		}
	}
}
