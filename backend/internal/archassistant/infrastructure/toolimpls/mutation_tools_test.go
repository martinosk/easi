package toolimpls_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"
const validUUID2 = "660e8400-e29b-41d4-a716-446655440000"

func newTestClient(server *httptest.Server) *agenthttp.Client {
	return agenthttp.NewClient(server.URL+"/api/v1", "test-token")
}

type mockPermissions struct {
	permissions map[string]bool
}

func (m *mockPermissions) HasPermission(perm string) bool {
	return m.permissions[perm]
}

type specToolTestCase struct {
	name           string
	toolName       string
	args           map[string]interface{}
	expectMethod   string
	expectPath     string
	expectBody     map[string]interface{}
	responseStatus int
	responseBody   string
	wantError      bool
	wantContains   []string
}

func runSpecToolTests(t *testing.T, cases []specToolTestCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			capture := newRequestCapture(t, tc)
			result := executeTool(t, newAllToolsRegistry(capture.server), tc.toolName, tc.args)
			assertSpecResult(t, tc, result, capture)
		})
	}
}

type requestCapture struct {
	server       *httptest.Server
	method, path string
	body         map[string]interface{}
}

func newRequestCapture(t *testing.T, tc specToolTestCase) *requestCapture {
	t.Helper()
	c := &requestCapture{}
	c.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.method = r.Method
		c.path = r.URL.Path
		if tc.expectBody != nil {
			c.body = readJSONBody(t, r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(tc.responseStatus)
		if tc.responseBody != "" {
			w.Write([]byte(tc.responseBody))
		}
	}))
	t.Cleanup(c.server.Close)
	return c
}

func assertSpecResult(t *testing.T, tc specToolTestCase, result tools.ToolResult, c *requestCapture) {
	t.Helper()
	assert.Equal(t, tc.wantError, result.IsError, "IsError mismatch: %s", result.Content)
	if tc.expectMethod != "" {
		assert.Equal(t, tc.expectMethod, c.method)
	}
	if tc.expectPath != "" {
		assert.Equal(t, tc.expectPath, c.path)
	}
	assertBodyFields(t, tc.expectBody, c.body)
	for _, s := range tc.wantContains {
		assert.Contains(t, result.Content, s)
	}
}

func assertBodyFields(t *testing.T, expected, actual map[string]interface{}) {
	t.Helper()
	for k, v := range expected {
		assert.Equal(t, v, actual[k], "body field %q", k)
	}
}

func TestMutationTools_CreateEntitySuccess(t *testing.T) {
	runSpecToolTests(t, []specToolTestCase{
		{
			name:           "create application",
			toolName:       "create_application",
			args:           map[string]interface{}{"name": "Payment Gateway", "description": "Handles payments"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/components",
			expectBody:     map[string]interface{}{"name": "Payment Gateway", "description": "Handles payments"},
			responseStatus: http.StatusCreated,
			responseBody:   `{"id":"` + validUUID + `","name":"Payment Gateway"}`,
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
			responseBody:   `{"id":"` + validUUID + `","name":"Payment Processing"}`,
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
			responseBody:   `{"id":"` + validUUID2 + `","name":"Invoice Processing"}`,
			wantContains:   []string{"Invoice Processing"},
		},
		{
			name:           "create business domain",
			toolName:       "create_business_domain",
			args:           map[string]interface{}{"name": "Finance"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/business-domains",
			expectBody:     map[string]interface{}{"name": "Finance"},
			responseStatus: http.StatusCreated,
			responseBody:   `{"id":"` + validUUID + `","name":"Finance"}`,
			wantContains:   []string{"Finance", validUUID},
		},
	})
}

func TestMutationTools_CreateLinkSuccess(t *testing.T) {
	runSpecToolTests(t, []specToolTestCase{
		{
			name:           "create application relation",
			toolName:       "create_application_relation",
			args:           map[string]interface{}{"sourceComponentId": validUUID, "targetComponentId": validUUID2, "relationType": "depends_on"},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/relations",
			expectBody:     map[string]interface{}{"sourceComponentId": validUUID, "targetComponentId": validUUID2, "relationType": "depends_on"},
			responseStatus: http.StatusCreated,
			responseBody:   `{"id":"rel-uuid-123"}`,
			wantContains:   []string{"rel-uuid-123"},
		},
		{
			name:           "realize capability",
			toolName:       "realize_capability",
			args:           map[string]interface{}{"id": validUUID, "componentId": validUUID2},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/capabilities/" + validUUID + "/systems",
			expectBody:     map[string]interface{}{"componentId": validUUID2},
			responseStatus: http.StatusCreated,
			responseBody:   `{"id":"real-uuid-123"}`,
			wantContains:   []string{"real-uuid-123"},
		},
		{
			name:           "assign capability to domain",
			toolName:       "assign_capability_to_domain",
			args:           map[string]interface{}{"domainId": validUUID, "capabilityId": validUUID2},
			expectMethod:   http.MethodPost,
			expectPath:     "/api/v1/business-domains/" + validUUID + "/capabilities",
			expectBody:     map[string]interface{}{"capabilityId": validUUID2},
			responseStatus: http.StatusCreated,
			responseBody:   `{"businessDomainId":"` + validUUID + `","capabilityId":"` + validUUID2 + `"}`,
			wantContains:   []string{validUUID2, validUUID},
		},
	})
}

func TestMutationTools_UpdateSuccess(t *testing.T) {
	runSpecToolTests(t, []specToolTestCase{
		{
			name:           "update application",
			toolName:       "update_application",
			args:           map[string]interface{}{"id": validUUID, "name": "Payment Service"},
			expectMethod:   http.MethodPut,
			expectPath:     "/api/v1/components/" + validUUID,
			expectBody:     map[string]interface{}{"name": "Payment Service"},
			responseStatus: http.StatusOK,
			responseBody:   `{"id":"` + validUUID + `","name":"Payment Service"}`,
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
			responseBody:   `{"id":"` + validUUID + `","name":"Updated Capability"}`,
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
			responseBody:   `{"id":"` + validUUID + `","name":"Updated Domain"}`,
			wantContains:   []string{"Updated Domain"},
		},
	})
}

func TestMutationTools_DeleteSuccess(t *testing.T) {
	runSpecToolTests(t, []specToolTestCase{
		{
			name:           "delete application",
			toolName:       "delete_application",
			args:           map[string]interface{}{"id": validUUID},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/components/" + validUUID,
			responseStatus: http.StatusNoContent,
		},
		{
			name:           "delete capability",
			toolName:       "delete_capability",
			args:           map[string]interface{}{"id": validUUID},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/capabilities/" + validUUID,
			responseStatus: http.StatusNoContent,
		},
		{
			name:           "delete application relation",
			toolName:       "delete_application_relation",
			args:           map[string]interface{}{"id": validUUID},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/relations/" + validUUID,
			responseStatus: http.StatusNoContent,
		},
		{
			name:           "unrealize capability",
			toolName:       "unrealize_capability",
			args:           map[string]interface{}{"id": validUUID},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/capability-realizations/" + validUUID,
			responseStatus: http.StatusNoContent,
		},
		{
			name:           "remove capability from domain",
			toolName:       "remove_capability_from_domain",
			args:           map[string]interface{}{"domainId": validUUID, "capabilityId": validUUID2},
			expectMethod:   http.MethodDelete,
			expectPath:     "/api/v1/business-domains/" + validUUID + "/capabilities/" + validUUID2,
			responseStatus: http.StatusNoContent,
		},
	})
}

func TestMutationTools_ValidationErrors(t *testing.T) {
	runSpecToolTests(t, []specToolTestCase{
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

			result := executeTool(t, newAllToolsRegistry(server), tc.toolName, tc.args)
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

	registry := newAllToolsRegistry(server)
	cases := []struct {
		name     string
		toolName string
		args     map[string]interface{}
	}{
		{"update_application", "update_application", map[string]interface{}{"id": "not-a-uuid", "name": "Test"}},
		{"delete_application", "delete_application", map[string]interface{}{"id": "not-a-uuid"}},
		{"delete_relation invalid id", "delete_application_relation", map[string]interface{}{"id": "bad"}},
		{"realize invalid id", "realize_capability", map[string]interface{}{"id": "bad", "componentId": validUUID}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := executeTool(t, registry, tc.toolName, tc.args)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "valid UUID")
		})
	}
}

func TestMutationTools_APIUnreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	result := executeTool(t, newAllToolsRegistry(server), "create_application", map[string]interface{}{"name": "Test"})

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "Failed to reach API")
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
