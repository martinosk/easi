package toolimpls_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
	"easi/backend/internal/archassistant/infrastructure/toolimpls"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGenericExecutor(server *httptest.Server, spec toolimpls.AgentToolSpec) *toolimpls.GenericAPIToolExecutor {
	client := agenthttp.NewClient(server.URL+"/api/v1", "test-token")
	return toolimpls.NewGenericExecutor(spec, client)
}

func newCapturingServer(capturedMethod, capturedPath *string, status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*capturedMethod = r.Method
		*capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != "" {
			w.Write([]byte(body))
		}
	}))
}

func newForbiddenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called when validation fails")
	}))
}

func executeSpec(executor *toolimpls.GenericAPIToolExecutor, args map[string]interface{}) tools.ToolResult {
	return executor.Execute(context.Background(), args)
}

func TestGenericExecutor_Validation(t *testing.T) {
	tests := []struct {
		name       string
		spec       toolimpls.AgentToolSpec
		args       map[string]interface{}
		wantSubstr string
	}{
		{
			name: "rejects missing required string param",
			spec: toolimpls.AgentToolSpec{
				Method: "POST", Path: "/components",
				BodyParams: []toolimpls.ParamSpec{{Name: "name", Type: "string", Required: true}},
			},
			args:       map[string]interface{}{},
			wantSubstr: "name is required",
		},
		{
			name: "rejects invalid UUID",
			spec: toolimpls.AgentToolSpec{
				Method: "GET", Path: "/components/{id}",
				PathParams: []toolimpls.ParamSpec{{Name: "id", Type: "uuid", Required: true}},
			},
			args:       map[string]interface{}{"id": "not-a-uuid"},
			wantSubstr: "id must be a valid UUID",
		},
		{
			name: "rejects string too long",
			spec: toolimpls.AgentToolSpec{
				Method: "POST", Path: "/components",
				BodyParams: []toolimpls.ParamSpec{{Name: "name", Type: "string", Required: true}},
			},
			args:       map[string]interface{}{"name": strings.Repeat("a", 201)},
			wantSubstr: "name must be at most 200 characters",
		},
		{
			name: "rejects nil args with required param",
			spec: toolimpls.AgentToolSpec{
				Method: "GET", Path: "/components/{id}",
				PathParams: []toolimpls.ParamSpec{{Name: "id", Type: "uuid", Required: true}},
			},
			args:       nil,
			wantSubstr: "id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newForbiddenServer(t)
			defer server.Close()
			result := executeSpec(newGenericExecutor(server, tt.spec), tt.args)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, tt.wantSubstr)
		})
	}
}

func TestGenericExecutor_PathParamSubstitution(t *testing.T) {
	tests := []struct {
		name       string
		spec       toolimpls.AgentToolSpec
		args       map[string]interface{}
		wantPath   string
	}{
		{
			name: "single path param",
			spec: toolimpls.AgentToolSpec{
				Method: "GET", Path: "/capabilities/{capabilityId}/realizations",
				PathParams: []toolimpls.ParamSpec{{Name: "capabilityId", Type: "uuid", Required: true}},
			},
			args:     map[string]interface{}{"capabilityId": validUUID},
			wantPath: "/api/v1/capabilities/" + validUUID + "/realizations",
		},
		{
			name: "multiple path params",
			spec: toolimpls.AgentToolSpec{
				Method: "DELETE", Path: "/components/{componentId}/relations/{relationId}",
				PathParams: []toolimpls.ParamSpec{
					{Name: "componentId", Type: "uuid", Required: true},
					{Name: "relationId", Type: "uuid", Required: true},
				},
			},
			args:     map[string]interface{}{"componentId": validUUID, "relationId": validUUID2},
			wantPath: "/api/v1/components/" + validUUID + "/relations/" + validUUID2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, path string
			server := newCapturingServer(&method, &path, http.StatusOK, `{}`)
			defer server.Close()
			result := executeSpec(newGenericExecutor(server, tt.spec), tt.args)
			assert.False(t, result.IsError)
			assert.Equal(t, tt.wantPath, path)
		})
	}
}

func newQueryCapturingServer(capturedQuery *string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[]}`))
	}))
}

func newBodyCapturingServer(t *testing.T, capturedBody *map[string]interface{}, status int, responseBody string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*capturedBody = readJSONBody(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(responseBody))
	}))
}

var listSpec = toolimpls.AgentToolSpec{
	Method: "GET", Path: "/components",
	QueryParams: []toolimpls.ParamSpec{
		{Name: "name", Type: "string"},
		{Name: "limit", Type: "integer"},
	},
}

var createSpec = toolimpls.AgentToolSpec{
	Method: "POST", Path: "/components",
	BodyParams: []toolimpls.ParamSpec{
		{Name: "name", Type: "string", Required: true},
		{Name: "description", Type: "string"},
	},
}

func TestGenericExecutor_GETQueryParams(t *testing.T) {
	t.Run("builds query string from provided params", func(t *testing.T) {
		var query string
		server := newQueryCapturingServer(&query)
		defer server.Close()
		result := executeSpec(newGenericExecutor(server, listSpec), map[string]interface{}{"name": "Payment", "limit": float64(10)})
		assert.False(t, result.IsError)
		assert.Contains(t, query, "name=Payment")
		assert.Contains(t, query, "limit=10")
	})

	t.Run("omits absent query params", func(t *testing.T) {
		var query string
		server := newQueryCapturingServer(&query)
		defer server.Close()
		result := executeSpec(newGenericExecutor(server, listSpec), map[string]interface{}{})
		assert.False(t, result.IsError)
		assert.Empty(t, query)
	})
}

func TestGenericExecutor_GETReturnsRawJSON(t *testing.T) {
	expectedJSON := `{"data":[{"id":"abc","name":"Payment Gateway"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(expectedJSON))
	}))
	defer server.Close()

	result := executeSpec(newGenericExecutor(server, toolimpls.AgentToolSpec{Method: "GET", Path: "/components"}), nil)
	assert.False(t, result.IsError)
	assert.Equal(t, expectedJSON, result.Content)
}

func TestGenericExecutor_POSTBody(t *testing.T) {
	t.Run("sends JSON body from provided params", func(t *testing.T) {
		var body map[string]interface{}
		server := newBodyCapturingServer(t, &body, http.StatusCreated, `{"id":"new-id","name":"Payment Gateway"}`)
		defer server.Close()
		result := executeSpec(newGenericExecutor(server, createSpec), map[string]interface{}{
			"name": "Payment Gateway", "description": "Handles payments",
		})
		assert.False(t, result.IsError)
		assert.Equal(t, "Payment Gateway", body["name"])
		assert.Equal(t, "Handles payments", body["description"])
		assert.Contains(t, result.Content, "new-id")
	})

	t.Run("omits absent optional body params", func(t *testing.T) {
		var body map[string]interface{}
		server := newBodyCapturingServer(t, &body, http.StatusCreated, `{"id":"new-id"}`)
		defer server.Close()
		result := executeSpec(newGenericExecutor(server, createSpec), map[string]interface{}{"name": "Test"})
		assert.False(t, result.IsError)
		require.NotNil(t, body)
		assert.Equal(t, "Test", body["name"])
		_, hasDesc := body["description"]
		assert.False(t, hasDesc)
	})
}

func TestGenericExecutor_PUTAndDELETE(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		status     int
	}{
		{"PUT dispatches correctly", "PUT", http.StatusOK},
		{"DELETE dispatches correctly", "DELETE", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMethod, capturedPath string
			server := newCapturingServer(&capturedMethod, &capturedPath, tt.status, `{}`)
			defer server.Close()

			spec := toolimpls.AgentToolSpec{
				Method: tt.method, Path: "/components/{id}",
				PathParams: []toolimpls.ParamSpec{{Name: "id", Type: "uuid", Required: true}},
				BodyParams: []toolimpls.ParamSpec{{Name: "name", Type: "string"}},
			}

			result := executeSpec(newGenericExecutor(server, spec), map[string]interface{}{"id": validUUID, "name": "X"})
			assert.False(t, result.IsError)
			assert.Equal(t, tt.method, capturedMethod)
			assert.Equal(t, "/api/v1/components/"+validUUID, capturedPath)
		})
	}
}

func TestGenericExecutor_ErrorHandling(t *testing.T) {
	t.Run("API error returns ToolResult error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Component not found"}`))
		}))
		defer server.Close()

		spec := toolimpls.AgentToolSpec{
			Method: "GET", Path: "/components/{id}",
			PathParams: []toolimpls.ParamSpec{{Name: "id", Type: "uuid", Required: true}},
		}
		result := executeSpec(newGenericExecutor(server, spec), map[string]interface{}{"id": validUUID})
		assert.True(t, result.IsError)
		assert.Contains(t, result.Content, "Component not found")
	})

	t.Run("unreachable API", func(t *testing.T) {
		client := agenthttp.NewClient("http://localhost:1/api/v1", "test-token")
		executor := toolimpls.NewGenericExecutor(toolimpls.AgentToolSpec{Method: "GET", Path: "/components"}, client)
		result := executor.Execute(context.Background(), nil)
		assert.True(t, result.IsError)
		assert.Contains(t, result.Content, "Failed to reach API")
	})
}

func TestGenericExecutor_ToolSpecCatalog_AllRegistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := agenthttp.NewClient(server.URL+"/api/v1", "test-token")

	registry := tools.NewRegistry()
	toolimpls.RegisterSpecTools(registry, client)

	available := registry.AvailableTools(allPerms(), true)
	names := make([]string, len(available))
	for i, d := range available {
		names[i] = d.Name
	}

	expectedTools := []string{
		"list_applications", "get_application_details",
		"list_capabilities", "get_capability_details",
		"list_business_domains", "get_business_domain_details",
		"list_value_streams", "get_value_stream_details",
		"create_application", "update_application", "delete_application",
		"create_application_relation", "delete_application_relation",
		"create_capability", "update_capability", "delete_capability",
		"realize_capability", "unrealize_capability",
		"create_business_domain", "update_business_domain",
		"assign_capability_to_domain", "remove_capability_from_domain",
	}

	assert.ElementsMatch(t, expectedTools, names)
}

func TestGenericExecutor_IntegrationRoundTrip(t *testing.T) {
	type testCase struct {
		name, toolName, expectMethod, expectPath, responseBody, wantContains string
		args                                                                 map[string]interface{}
		responseStatus                                                       int
	}

	cases := []testCase{
		{"list via GET", "list_applications", "GET", "/api/v1/components", `{"data":[{"id":"abc","name":"Payment Gateway"}]}`, "Payment Gateway", map[string]interface{}{"name": "Payment", "limit": float64(10)}, http.StatusOK},
		{"get by ID", "get_application_details", "GET", "/api/v1/components/" + validUUID, `{"id":"` + validUUID + `"}`, validUUID, map[string]interface{}{"id": validUUID}, http.StatusOK},
		{"create via POST", "create_application", "POST", "/api/v1/components", `{"id":"new-id"}`, "new-id", map[string]interface{}{"name": "New App"}, http.StatusCreated},
		{"update via PUT", "update_application", "PUT", "/api/v1/components/" + validUUID, `{"name":"Updated"}`, "Updated", map[string]interface{}{"id": validUUID, "name": "Updated"}, http.StatusOK},
		{"delete via DELETE", "delete_application", "DELETE", "/api/v1/components/" + validUUID, "", "", map[string]interface{}{"id": validUUID}, http.StatusNoContent},
		{"POST with path params", "realize_capability", "POST", "/api/v1/capabilities/" + validUUID + "/realizations", `{"id":"real-123"}`, "real-123", map[string]interface{}{"capabilityId": validUUID, "applicationId": validUUID2}, http.StatusCreated},
		{"DELETE with path params", "delete_application_relation", "DELETE", "/api/v1/components/" + validUUID + "/relations/" + validUUID2, "", "", map[string]interface{}{"componentId": validUUID, "relationId": validUUID2}, http.StatusNoContent},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var method, path string
			server := newCapturingServer(&method, &path, tc.responseStatus, tc.responseBody)
			defer server.Close()

			registry := tools.NewRegistry()
			toolimpls.RegisterSpecTools(registry, agenthttp.NewClient(server.URL+"/api/v1", "test-token"))

			result, err := registry.Execute(context.Background(), allPerms(), tc.toolName, tc.args)
			require.NoError(t, err)
			assert.False(t, result.IsError, "unexpected error: %s", result.Content)
			assert.Equal(t, tc.expectMethod, method)
			assert.Equal(t, tc.expectPath, path)
			if tc.wantContains != "" {
				assert.Contains(t, result.Content, tc.wantContains)
			}
		})
	}
}
