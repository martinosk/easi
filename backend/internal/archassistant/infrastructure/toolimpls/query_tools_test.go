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
		"valuestreams:read": true,
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
		w.Write(testCollectionJSON(data...))
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
	assert.Contains(t, result.Content, "Application")
	assert.Contains(t, result.Content, "Capabilit")
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

func TestRegisterAllTools_AllRegistered(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	available := newAllToolsRegistry(server).AvailableTools(allPerms(), true)

	expectedTools := []string{
		"list_applications", "get_application_details", "list_application_relations",
		"list_capabilities", "get_capability_details",
		"list_business_domains", "get_business_domain_details",
		"list_value_streams", "get_value_stream_details",
		"search_architecture", "get_portfolio_summary",
		"create_application", "update_application", "delete_application",
		"create_application_relation", "delete_application_relation",
		"create_capability", "update_capability", "delete_capability",
		"realize_capability", "unrealize_capability",
		"create_business_domain", "update_business_domain",
		"assign_capability_to_domain", "remove_capability_from_domain",
	}

	names := make([]string, len(available))
	for i, d := range available {
		names[i] = d.Name
	}

	assert.ElementsMatch(t, expectedTools, names)
	assert.Len(t, available, 25)

	for _, d := range available {
		assert.NotEmpty(t, d.Permission, "tool %s should have a permission", d.Name)
		assert.NotEmpty(t, d.Description, "tool %s should have a description", d.Name)
	}
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

	readOnly := allPerms()
	available := registry.AvailableTools(readOnly, false)

	for _, d := range available {
		assert.Equal(t, tools.AccessRead, d.Access,
			fmt.Sprintf("read-only filter should only show read tools, got %s for %s", d.Access, d.Name))
	}
}
