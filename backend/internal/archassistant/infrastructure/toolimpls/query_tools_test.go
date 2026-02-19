package toolimpls_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func testResourceJSON(resource map[string]interface{}) []byte {
	b, _ := json.Marshal(resource)
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

func newQueryRegistry(server *httptest.Server) *tools.Registry {
	client := newTestClient(server)
	registry := tools.NewRegistry()
	toolimpls.RegisterQueryTools(registry, client)
	return registry
}

func allQueryPerms() *mockPermissions {
	return &mockPermissions{permissions: map[string]bool{
		"components:read": true, "components:write": true,
		"capabilities:read": true, "capabilities:write": true,
		"domains:read": true, "domains:write": true,
		"valuestreams:read": true,
	}}
}

func executeQueryTool(t *testing.T, registry *tools.Registry, name string, args map[string]interface{}) tools.ToolResult {
	t.Helper()
	result, err := registry.Execute(context.Background(), allQueryPerms(), name, args)
	require.NoError(t, err)
	return result
}

func jsonCollectionHandler(data ...map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(testCollectionJSON(data...))
	}
}

func jsonResourceHandler(resource map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(testResourceJSON(resource))
	}
}

func TestListApplications_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components": jsonCollectionHandler(
			map[string]interface{}{"id": "abc-123", "name": "Payment Gateway", "description": "Handles payments"},
			map[string]interface{}{"id": "def-456", "name": "Order Service", "description": "Order processing"},
			map[string]interface{}{"id": "ghi-789", "name": "Legacy CRM", "description": ""},
		),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_applications", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Found 3 applications")
	assert.Contains(t, result.Content, "Payment Gateway")
	assert.Contains(t, result.Content, "abc-123")
	assert.Contains(t, result.Content, "Order Service")
	assert.Contains(t, result.Content, "Legacy CRM")
}

func TestListApplications_WithFilter(t *testing.T) {
	var capturedQuery string
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components": func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			w.Write(testCollectionJSON(map[string]interface{}{"id": "abc-123", "name": "Payment Gateway"}))
		},
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_applications", map[string]interface{}{
		"name": "Payment", "limit": float64(10),
	})

	assert.False(t, result.IsError)
	assert.Contains(t, capturedQuery, "name=Payment")
	assert.Contains(t, capturedQuery, "limit=10")
}

func TestListApplications_EmptyResult(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components": jsonCollectionHandler(),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_applications", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No applications found")
}

func TestListApplications_LimitClampAndFilterCap(t *testing.T) {
	cases := []struct {
		name          string
		args          map[string]interface{}
		expectInQuery string
	}{
		{
			name:          "limit clamped to 50",
			args:          map[string]interface{}{"limit": float64(999)},
			expectInQuery: "limit=50",
		},
		{
			name:          "long filter is capped",
			args:          map[string]interface{}{"name": strings.Repeat("a", 300)},
			expectInQuery: "limit=20",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedQuery string
			server := newMockAPI(t, map[string]http.HandlerFunc{
				"/api/v1/components": func(w http.ResponseWriter, r *http.Request) {
					capturedQuery = r.URL.RawQuery
					w.Header().Set("Content-Type", "application/json")
					w.Write(testCollectionJSON())
				},
			})
			defer server.Close()

			executeQueryTool(t, newQueryRegistry(server), "list_applications", tc.args)
			assert.Contains(t, capturedQuery, tc.expectInQuery)
		})
	}
}

func TestGetApplicationDetails_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components/" + validUUID: jsonResourceHandler(map[string]interface{}{
			"id": validUUID, "name": "Payment Gateway", "description": "Handles all payment processing",
		}),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "get_application_details", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Payment Gateway")
	assert.Contains(t, result.Content, validUUID)
	assert.Contains(t, result.Content, "Handles all payment processing")
}

func TestGetApplicationDetails_NotFound(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components/" + validUUID: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Component not found"})
		},
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "get_application_details", map[string]interface{}{"id": validUUID})

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestQueryTool_InvalidID(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	registry := newQueryRegistry(server)
	cases := []struct {
		name     string
		toolName string
		args     map[string]interface{}
	}{
		{"get_application_details with invalid ID", "get_application_details", map[string]interface{}{"id": "not-a-uuid"}},
		{"get_application_details with missing ID", "get_application_details", nil},
		{"get_capability_details with invalid ID", "get_capability_details", map[string]interface{}{"id": "invalid"}},
		{"get_business_domain_details with invalid ID", "get_business_domain_details", map[string]interface{}{"id": "invalid"}},
		{"get_value_stream_details with invalid ID", "get_value_stream_details", map[string]interface{}{"id": "invalid"}},
		{"list_application_relations with invalid ID", "list_application_relations", map[string]interface{}{"id": "not-uuid"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := executeQueryTool(t, registry, tc.toolName, tc.args)
			assert.True(t, result.IsError)
		})
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

	result := executeQueryTool(t, newQueryRegistry(server), "list_application_relations", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "API Call")
	assert.Contains(t, result.Content, "Data Feed")
	assert.Contains(t, result.Content, "uses")
	assert.Contains(t, result.Content, "depends_on")
}

func TestListCapabilities_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/capabilities": jsonCollectionHandler(
			map[string]interface{}{"id": "cap-1", "name": "Sales Management", "level": "L1", "description": "Sales operations"},
			map[string]interface{}{"id": "cap-2", "name": "Order Processing", "level": "L2", "parentId": "cap-1"},
		),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_capabilities", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Sales Management")
	assert.Contains(t, result.Content, "Order Processing")
	assert.Contains(t, result.Content, "L1")
	assert.Contains(t, result.Content, "L2")
}

func TestListCapabilities_WithFilter(t *testing.T) {
	var capturedQuery string
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/capabilities": func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			w.Write(testCollectionJSON(map[string]interface{}{"id": "cap-1", "name": "Sales Management", "level": "L1"}))
		},
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_capabilities", map[string]interface{}{
		"name": "Sales", "limit": float64(5),
	})

	assert.False(t, result.IsError)
	assert.Contains(t, capturedQuery, "name=Sales")
	assert.Contains(t, capturedQuery, "limit=5")
}

func TestGetCapabilityDetails_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/capabilities/" + validUUID: jsonResourceHandler(map[string]interface{}{
			"id": validUUID, "name": "Sales Management", "description": "Manages all sales operations", "level": "L1",
		}),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "get_capability_details", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Sales Management")
	assert.Contains(t, result.Content, validUUID)
	assert.Contains(t, result.Content, "L1")
}

func TestListBusinessDomains_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/business-domains": jsonCollectionHandler(
			map[string]interface{}{"id": "dom-1", "name": "Sales", "description": "Sales domain", "capabilityCount": 5},
			map[string]interface{}{"id": "dom-2", "name": "Marketing", "description": "Marketing domain", "capabilityCount": 3},
		),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_business_domains", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Sales")
	assert.Contains(t, result.Content, "Marketing")
	assert.Contains(t, result.Content, "2 business domains")
}

func TestGetBusinessDomainDetails_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/business-domains/" + validUUID: jsonResourceHandler(map[string]interface{}{
			"id": validUUID, "name": "Sales", "description": "Sales business domain", "capabilityCount": 5,
		}),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "get_business_domain_details", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Sales")
	assert.Contains(t, result.Content, validUUID)
}

func TestListValueStreams_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/value-streams": jsonCollectionHandler(
			map[string]interface{}{"id": "vs-1", "name": "Customer Onboarding", "description": "Onboarding flow", "stageCount": 4},
			map[string]interface{}{"id": "vs-2", "name": "Order Fulfillment", "description": "Fulfillment flow", "stageCount": 6},
		),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_value_streams", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Customer Onboarding")
	assert.Contains(t, result.Content, "Order Fulfillment")
	assert.Contains(t, result.Content, "2 value streams")
}

func TestGetValueStreamDetails_Success(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/value-streams/" + validUUID: jsonResourceHandler(map[string]interface{}{
			"id": validUUID, "name": "Customer Onboarding", "description": "Onboarding flow", "stageCount": 4,
			"stages": []map[string]interface{}{
				{"id": "stage-1", "name": "Registration", "position": 1},
				{"id": "stage-2", "name": "Verification", "position": 2},
			},
			"stageCapabilities": []map[string]interface{}{
				{"stageId": "stage-1", "capabilityId": "cap-1", "capabilityName": "Identity Management"},
			},
		}),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "get_value_stream_details", map[string]interface{}{"id": validUUID})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Customer Onboarding")
	assert.Contains(t, result.Content, "Registration")
	assert.Contains(t, result.Content, "Verification")
	assert.Contains(t, result.Content, "Identity Management")
}

func TestSearchArchitecture_CombinesResults(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components":       jsonCollectionHandler(map[string]interface{}{"id": "app-1", "name": "Payment Gateway"}),
		"/api/v1/capabilities":     jsonCollectionHandler(map[string]interface{}{"id": "cap-1", "name": "Payment Processing", "level": "L1"}),
		"/api/v1/business-domains": jsonCollectionHandler(),
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "search_architecture", map[string]interface{}{"query": "Payment"})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Payment Gateway")
	assert.Contains(t, result.Content, "Payment Processing")
	assert.Contains(t, result.Content, "Application")
	assert.Contains(t, result.Content, "Capabilit")
}

func TestSearchArchitecture_MissingQuery(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "search_architecture", nil)

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

	result := executeQueryTool(t, newQueryRegistry(server), "get_portfolio_summary", nil)

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Applications: 3")
	assert.Contains(t, result.Content, "Capabilities: 2")
	assert.Contains(t, result.Content, "Business Domains: 1")
	assert.Contains(t, result.Content, "Value Streams: 2")
	assert.Contains(t, result.Content, "Relations: 1")
}

func TestQueryTool_APIError(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{
		"/api/v1/components": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Database connection failed"})
		},
	})
	defer server.Close()

	result := executeQueryTool(t, newQueryRegistry(server), "list_applications", nil)

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "Database connection failed")
}

func TestRegisterQueryTools_AllRegistered(t *testing.T) {
	server := newMockAPI(t, map[string]http.HandlerFunc{})
	defer server.Close()

	available := newQueryRegistry(server).AvailableTools(allQueryPerms(), false)

	expectedTools := []string{
		"list_applications", "get_application_details", "list_application_relations",
		"list_capabilities", "get_capability_details",
		"list_business_domains", "get_business_domain_details",
		"list_value_streams", "get_value_stream_details",
		"search_architecture", "get_portfolio_summary",
	}

	names := make([]string, len(available))
	for i, d := range available {
		names[i] = d.Name
	}

	assert.ElementsMatch(t, expectedTools, names)
	assert.Len(t, available, 11)

	for _, d := range available {
		assert.Equal(t, tools.AccessRead, d.Access, "tool %s should be AccessRead", d.Name)
	}
}
