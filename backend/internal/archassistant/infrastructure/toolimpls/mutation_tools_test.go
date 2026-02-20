package toolimpls_test

import (
	"context"
	"encoding/json"
	"io"
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

const validUUID = "550e8400-e29b-41d4-a716-446655440000"
const validUUID2 = "660e8400-e29b-41d4-a716-446655440000"

func newTestClient(server *httptest.Server) *agenthttp.Client {
	return agenthttp.NewClient(server.URL+"/api/v1", "test-token")
}

type toolTestCase struct {
	name           string
	toolName       string
	args           map[string]interface{}
	expectMethod   string
	expectPath     string
	expectBody     map[string]interface{}
	responseStatus int
	responseBody   map[string]interface{}
	wantError      bool
	wantContains   []string
}

func runToolTests(t *testing.T, cases []toolTestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := buildHandler(t, tc)
			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			result := executeRegisteredTool(t, server, tc.toolName, tc.args)

			assert.Equal(t, tc.wantError, result.IsError, "IsError mismatch: %s", result.Content)
			for _, s := range tc.wantContains {
				assert.Contains(t, result.Content, s)
			}
		})
	}
}

func buildHandler(t *testing.T, tc toolTestCase) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tc.wantError {
			t.Fatal("API should not be called for validation errors")
		}
		assert.Equal(t, tc.expectMethod, r.Method)
		assert.Equal(t, tc.expectPath, r.URL.Path)
		if tc.expectBody != nil {
			body := readJSONBody(t, r)
			for k, v := range tc.expectBody {
				assert.Equal(t, v, body[k], "body field %q", k)
			}
		}
		if tc.responseBody != nil {
			jsonResponse(w, tc.responseStatus, tc.responseBody)
		} else {
			w.WriteHeader(tc.responseStatus)
		}
	})
}

func executeRegisteredTool(t *testing.T, server *httptest.Server, name string, args map[string]interface{}) tools.ToolResult {
	t.Helper()
	client := agenthttp.NewClient(server.URL+"/api/v1", "test-token")
	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, client)
	return executeTool(t, registry, name, args)
}

func TestMutationTools_CreateSuccess(t *testing.T) {
	runToolTests(t, []toolTestCase{
		{
			name:           "create application",
			toolName:       "create_application",
			args:           map[string]interface{}{"name": "Payment Gateway", "description": "Handles payments"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/components",
			expectBody:     map[string]interface{}{"name": "Payment Gateway", "description": "Handles payments"},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"id": validUUID, "name": "Payment Gateway"},
			wantContains:   []string{"Payment Gateway", validUUID},
		},
		{
			name:           "create L1 capability",
			toolName:       "create_capability",
			args:           map[string]interface{}{"name": "Payment Processing", "level": "L1"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/capabilities",
			expectBody:     map[string]interface{}{"name": "Payment Processing", "level": "L1"},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"id": validUUID, "name": "Payment Processing"},
			wantContains:   []string{"Payment Processing", validUUID},
		},
		{
			name:           "create L2 capability with parent",
			toolName:       "create_capability",
			args:           map[string]interface{}{"name": "Invoice Processing", "level": "L2", "parentId": validUUID},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/capabilities",
			expectBody:     map[string]interface{}{"name": "Invoice Processing", "level": "L2", "parentId": validUUID},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"id": validUUID2, "name": "Invoice Processing"},
			wantContains:   []string{"Invoice Processing", "L2"},
		},
		{
			name:           "create business domain",
			toolName:       "create_business_domain",
			args:           map[string]interface{}{"name": "Finance"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/business-domains",
			expectBody:     map[string]interface{}{"name": "Finance"},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"id": validUUID, "name": "Finance"},
			wantContains:   []string{"Finance", validUUID},
		},
		{
			name:           "create application relation",
			toolName:       "create_application_relation",
			args:           map[string]interface{}{"sourceId": validUUID, "targetId": validUUID2, "type": "depends_on"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/components/" + validUUID + "/relations",
			expectBody:     map[string]interface{}{"targetId": validUUID2, "type": "depends_on"},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"id": "rel-uuid-123"},
			wantContains:   []string{"relation"},
		},
		{
			name:           "realize capability",
			toolName:       "realize_capability",
			args:           map[string]interface{}{"capabilityId": validUUID, "applicationId": validUUID2},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/capabilities/" + validUUID + "/realizations",
			expectBody:     map[string]interface{}{"applicationId": validUUID2},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"id": "real-uuid-123"},
			wantContains:   []string{"Linked"},
		},
		{
			name:           "assign capability to domain",
			toolName:       "assign_capability_to_domain",
			args:           map[string]interface{}{"domainId": validUUID, "capabilityId": validUUID2},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/business-domains/" + validUUID + "/capabilities",
			expectBody:     map[string]interface{}{"capabilityId": validUUID2},
			responseStatus: http.StatusCreated,
			responseBody:   map[string]interface{}{"businessDomainId": validUUID, "capabilityId": validUUID2},
			wantContains:   []string{"Assigned", validUUID2, validUUID},
		},
	})
}

func TestMutationTools_UpdateSuccess(t *testing.T) {
	runToolTests(t, []toolTestCase{
		{
			name:           "update application",
			toolName:       "update_application",
			args:           map[string]interface{}{"id": validUUID, "name": "Payment Service"},
			expectMethod:   http.MethodPut,
			expectPath:     "/api/v1/components/" + validUUID,
			expectBody:     map[string]interface{}{"name": "Payment Service"},
			responseStatus: http.StatusOK,
			responseBody:   map[string]interface{}{"id": validUUID, "name": "Payment Service"},
			wantContains:   []string{"Payment Service"},
		},
		{
			name:           "update capability",
			toolName:       "update_capability",
			args:           map[string]interface{}{"id": validUUID, "name": "Updated Capability"},
			expectMethod:   http.MethodPut,
			expectPath:     "/api/v1/capabilities/" + validUUID,
			expectBody:     map[string]interface{}{"name": "Updated Capability"},
			responseStatus: http.StatusOK,
			responseBody:   map[string]interface{}{"id": validUUID, "name": "Updated Capability"},
			wantContains:   []string{"Updated Capability"},
		},
		{
			name:           "update business domain",
			toolName:       "update_business_domain",
			args:           map[string]interface{}{"id": validUUID, "name": "Updated Domain"},
			expectMethod:   http.MethodPut,
			expectPath:     "/api/v1/business-domains/" + validUUID,
			expectBody:     map[string]interface{}{"name": "Updated Domain"},
			responseStatus: http.StatusOK,
			responseBody:   map[string]interface{}{"id": validUUID, "name": "Updated Domain"},
			wantContains:   []string{"Updated Domain"},
		},
	})
}

func TestMutationTools_DeleteSuccess(t *testing.T) {
	runToolTests(t, []toolTestCase{
		{
			name:           "delete application",
			toolName:       "delete_application",
			args:           map[string]interface{}{"id": validUUID},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/components/" + validUUID,
			responseStatus: http.StatusNoContent,
			wantContains:   []string{"Deleted application", validUUID},
		},
		{
			name:           "delete capability",
			toolName:       "delete_capability",
			args:           map[string]interface{}{"id": validUUID},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/capabilities/" + validUUID,
			responseStatus: http.StatusNoContent,
			wantContains:   []string{"Deleted capability", validUUID},
		},
		{
			name:           "delete application relation",
			toolName:       "delete_application_relation",
			args:           map[string]interface{}{"componentId": validUUID, "relationId": validUUID2},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/components/" + validUUID + "/relations/" + validUUID2,
			responseStatus: http.StatusNoContent,
			wantContains:   []string{"Deleted relation"},
		},
		{
			name:           "unrealize capability",
			toolName:       "unrealize_capability",
			args:           map[string]interface{}{"capabilityId": validUUID, "realizationId": validUUID2},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/capabilities/" + validUUID + "/realizations/" + validUUID2,
			responseStatus: http.StatusNoContent,
			wantContains:   []string{"Unlinked"},
		},
		{
			name:           "remove capability from domain",
			toolName:       "remove_capability_from_domain",
			args:           map[string]interface{}{"domainId": validUUID, "capabilityId": validUUID2},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/business-domains/" + validUUID + "/capabilities/" + validUUID2,
			responseStatus: http.StatusNoContent,
			wantContains:   []string{"Removed", validUUID2, validUUID},
		},
	})
}

func TestMutationTools_ValidationErrors(t *testing.T) {
	runToolTests(t, []toolTestCase{
		{
			name:         "create application missing name",
			toolName:     "create_application",
			args:         map[string]interface{}{},
			wantError:    true,
			wantContains: []string{"name is required"},
		},
		{
			name:         "create capability missing level",
			toolName:     "create_capability",
			args:         map[string]interface{}{"name": "Payment Processing"},
			wantError:    true,
			wantContains: []string{"level is required"},
		},
		{
			name:         "create application name too long",
			toolName:     "create_application",
			args:         map[string]interface{}{"name": strings.Repeat("a", 201)},
			wantError:    true,
			wantContains: []string{"200 characters"},
		},
	})
}

func TestMutationTools_APIError(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		message  string
		toolName string
		args     map[string]interface{}
	}{
		{
			name:     "400 Bad Request",
			status:   http.StatusBadRequest,
			message:  "Name cannot be empty",
			toolName: "create_application",
			args:     map[string]interface{}{"name": "Test"},
		},
		{
			name:     "403 Forbidden",
			status:   http.StatusForbidden,
			message:  "Insufficient permissions",
			toolName: "create_application",
			args:     map[string]interface{}{"name": "Test"},
		},
		{
			name:     "500 Internal Server Error",
			status:   http.StatusInternalServerError,
			message:  "Unexpected error occurred",
			toolName: "delete_application",
			args:     map[string]interface{}{"id": validUUID},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonError(w, tc.status, tc.message)
			}))
			t.Cleanup(server.Close)

			result := executeRegisteredTool(t, server, tc.toolName, tc.args)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, tc.message)
		})
	}
}

func TestMutationTools_InvalidID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called with invalid UUID")
	}))
	t.Cleanup(server.Close)

	cases := []struct {
		name     string
		toolName string
		args     map[string]interface{}
	}{
		{"update_application", "update_application", map[string]interface{}{"id": "not-a-uuid", "name": "Test"}},
		{"delete_application", "delete_application", map[string]interface{}{"id": "not-a-uuid"}},
		{"delete_relation invalid componentId", "delete_application_relation", map[string]interface{}{"componentId": "bad", "relationId": validUUID}},
		{"delete_relation invalid relationId", "delete_application_relation", map[string]interface{}{"componentId": validUUID, "relationId": "bad"}},
		{"realize invalid capabilityId", "realize_capability", map[string]interface{}{"capabilityId": "bad", "applicationId": validUUID}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := executeRegisteredTool(t, server, tc.toolName, tc.args)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "valid UUID")
		})
	}
}

func TestMutationTools_APIUnreachable(t *testing.T) {
	client := agenthttp.NewClient("http://localhost:1/api/v1", "test-token")
	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, client)

	result := executeTool(t, registry, "create_application", map[string]interface{}{"name": "Test"})

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "Failed to reach API")
}

func TestMutationTools_AllRegistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := agenthttp.NewClient(server.URL+"/api/v1", "test-token")
	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, client)

	allPerms := &mockPermissions{permissions: map[string]bool{
		"components:write":   true,
		"capabilities:write": true,
		"domains:write":      true,
	}}

	available := registry.AvailableTools(allPerms, true)

	expectedTools := []string{
		"create_application", "update_application", "delete_application",
		"create_capability", "update_capability", "delete_capability",
		"create_business_domain", "update_business_domain",
		"assign_capability_to_domain", "remove_capability_from_domain",
		"create_application_relation", "delete_application_relation",
		"realize_capability", "unrealize_capability",
	}

	names := make([]string, len(available))
	for i, d := range available {
		names[i] = d.Name
	}
	assert.ElementsMatch(t, expectedTools, names)

	for _, d := range available {
		assert.Equal(t, tools.AccessWrite, d.Access, "tool %s should be AccessWrite", d.Name)
	}
}

func jsonResponse(w http.ResponseWriter, status int, body map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]interface{}{"message": message})
}

func readJSONBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &body))
	return body
}

type mockPermissions struct {
	permissions map[string]bool
}

func (m *mockPermissions) HasPermission(perm string) bool {
	return m.permissions[perm]
}

func executeTool(t *testing.T, registry *tools.Registry, name string, args map[string]interface{}) tools.ToolResult {
	t.Helper()
	allPerms := &mockPermissions{permissions: map[string]bool{
		"components:write":   true,
		"capabilities:write": true,
		"domains:write":      true,
	}}
	result, err := registry.Execute(context.Background(), allPerms, name, args)
	require.NoError(t, err)
	return result
}
