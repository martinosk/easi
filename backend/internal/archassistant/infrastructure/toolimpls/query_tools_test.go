package toolimpls_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/toolimpls"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCollectionJSON(data ...map[string]interface{}) []byte {
	b, _ := json.Marshal(map[string]interface{}{"data": data})
	return b
}

func newMockAPI(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}
	return httptest.NewServer(mux)
}

func newAllToolsRegistry(server *httptest.Server) *tools.Registry {
	client := newTestClient(server)
	registry := tools.NewRegistry()
	toolimpls.RegisterAllTools(registry, client)
	return registry
}

func allPerms() *mockPermissions {
	return &mockPermissions{permissions: map[string]bool{
		"components:read": true, "components:write": true,
		"capabilities:read": true, "capabilities:write": true,
		"domains:read": true, "domains:write": true,
		"enterprise-arch:read": true, "enterprise-arch:write": true,
		"valuestreams:read": true, "valuestreams:write": true,
		"metamodel:read": true,
		"assistant:use":  true,
	}}
}

func executeTool(t *testing.T, registry *tools.Registry, name string, args map[string]interface{}) tools.ToolResult {
	t.Helper()
	result, err := registry.Execute(context.Background(), allPerms(), name, args)
	require.NoError(t, err)
	return result
}

func jsonCollectionHandler(data ...map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(testCollectionJSON(data...))
	}
}

func TestListApplicationRelations_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/relations/from/" + validUUID: jsonCollectionHandler(
			map[string]interface{}{"id": "rel-1", "sourceComponentId": validUUID, "targetComponentId": "target-1", "relationType": "uses", "name": "API Call"},
		),
		"/api/v1/relations/to/" + validUUID: jsonCollectionHandler(
			map[string]interface{}{"id": "rel-2", "sourceComponentId": "source-1", "targetComponentId": validUUID, "relationType": "depends_on", "name": "Data Feed"},
		),
	})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "list_application_relations", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "API Call")
	assert.Contains(t, result.Content, "Data Feed")
	assert.Contains(t, result.Content, "uses")
	assert.Contains(t, result.Content, "depends_on")
}

func TestListApplicationRelations_NoRelations(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/relations/from/" + validUUID: jsonCollectionHandler(),
		"/api/v1/relations/to/" + validUUID:   jsonCollectionHandler(),
	})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "list_application_relations", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No relations found")
}

func TestSearchArchitecture_CombinesResults(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components":       jsonCollectionHandler(map[string]interface{}{"id": "app-1", "name": "Payment Gateway"}),
		"/api/v1/capabilities":     jsonCollectionHandler(map[string]interface{}{"id": "cap-1", "name": "Payment Processing", "level": "L1"}),
		"/api/v1/business-domains": jsonCollectionHandler(),
	})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "search_architecture", map[string]interface{}{"query": "Payment"})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Payment Gateway")
	assert.Contains(t, result.Content, "Payment Processing")
	assert.Contains(t, result.Content, "Applications (")
	assert.Contains(t, result.Content, "Capabilities (")
}

func TestSearchArchitecture_NoResults(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components":       jsonCollectionHandler(),
		"/api/v1/capabilities":     jsonCollectionHandler(),
		"/api/v1/business-domains": jsonCollectionHandler(),
	})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "search_architecture", map[string]interface{}{"query": "Nonexistent"})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No results found for 'Nonexistent'")
}

func TestSearchArchitecture_MissingQuery(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "search_architecture", nil)

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "query")
}

func TestGetPortfolioSummary_AggregatesCounts(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components":       jsonCollectionHandler(map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}, map[string]interface{}{"id": "3"}),
		"/api/v1/capabilities":     jsonCollectionHandler(map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}),
		"/api/v1/business-domains": jsonCollectionHandler(map[string]interface{}{"id": "1"}),
		"/api/v1/value-streams":    jsonCollectionHandler(map[string]interface{}{"id": "1"}, map[string]interface{}{"id": "2"}),
		"/api/v1/relations":        jsonCollectionHandler(map[string]interface{}{"id": "1"}),
	})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "get_portfolio_summary", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Applications: 3")
	assert.Contains(t, result.Content, "Capabilities: 2")
	assert.Contains(t, result.Content, "Business Domains: 1")
	assert.Contains(t, result.Content, "Value Streams: 2")
	assert.Contains(t, result.Content, "Relations: 1")
}

// TestGetPortfolioSummary_PartialAPIFailure documents that the tool silently
// reports 0 for any endpoint that fails, rather than returning an error.
// This is a pinning test — if the desired behaviour changes to return an error,
// this test should be updated to reflect the new contract.
func TestGetPortfolioSummary_PartialAPIFailure_ZeroForFailedEndpoints(t *testing.T) {
	// Only register the /components endpoint; all others will 404.
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components": jsonCollectionHandler(map[string]interface{}{"id": "1"}),
	})
	defer server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "get_portfolio_summary", nil)

	// Tool must not return an error — partial failures are silently zeroed.
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Applications: 1")
	assert.Contains(t, result.Content, "Capabilities: 0")
	assert.Contains(t, result.Content, "Business Domains: 0")
	assert.Contains(t, result.Content, "Value Streams: 0")
	assert.Contains(t, result.Content, "Relations: 0")
}

func TestRegisterAllTools_AllExpectedToolsPresent(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	available := newAllToolsRegistry(server).AvailableTools(allPerms(), true)
	names := make([]string, len(available))
	for i, d := range available {
		names[i] = d.Name
	}

	assert.ElementsMatch(t, allExpectedToolNames, names)
}

func TestRegisterAllTools_AllToolsHavePermission(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	for _, d := range newAllToolsRegistry(server).AvailableTools(allPerms(), true) {
		assert.NotEmpty(t, d.Permission, "tool %s should have a permission", d.Name)
	}
}

func TestRegisterAllTools_AllToolsHaveDescription(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	for _, d := range newAllToolsRegistry(server).AvailableTools(allPerms(), true) {
		assert.NotEmpty(t, d.Description, "tool %s should have a description", d.Name)
	}
}

func TestQueryDomainModel(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()
	registry := newAllToolsRegistry(server)

	t.Run("returns content for each topic", func(t *testing.T) {
		topics := []string{
			"capability-hierarchy", "business-domains", "realizations",
			"strategy", "enterprise-capabilities", "time-classification",
			"value-streams", "component-origins", "overview",
		}
		for _, topic := range topics {
			t.Run(topic, func(t *testing.T) {
				result := executeTool(t, registry, "query_domain_model", map[string]interface{}{"topic": topic})
				assert.False(t, result.IsError, "topic %s returned error: %s", topic, result.Content)
				assert.NotEmpty(t, result.Content, "topic %s returned empty content", topic)
			})
		}
	})

	t.Run("returns error for unknown topic", func(t *testing.T) {
		result := executeTool(t, registry, "query_domain_model", map[string]interface{}{"topic": "unknown-topic"})
		assert.True(t, result.IsError)
		assert.Contains(t, result.Content, "Unknown topic")
	})

	t.Run("returns error for missing topic", func(t *testing.T) {
		result := executeTool(t, registry, "query_domain_model", nil)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Content, "topic is required")
	})
}

func TestCompositeTools_InvalidID(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()
	registry := newAllToolsRegistry(server)

	result := executeTool(t, registry, "list_application_relations", map[string]interface{}{"id": "not-uuid"})
	assert.True(t, result.IsError)

	result = executeTool(t, registry, "list_application_relations", nil)
	assert.True(t, result.IsError)
}

func TestCompositeTools_AccessClass(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()
	registry := newAllToolsRegistry(server)

	// allPerms() grants every permission. The second argument (false) is what
	// filters AvailableTools to read-access tools only, regardless of the
	// caller's permission set.
	allPermsCtx := allPerms()
	available := registry.AvailableTools(allPermsCtx, false)

	for _, d := range available {
		assert.Equal(t, tools.AccessRead, d.Access,
			fmt.Sprintf("read-only filter should only show read tools, got %s for %s", d.Access, d.Name))
	}
}
