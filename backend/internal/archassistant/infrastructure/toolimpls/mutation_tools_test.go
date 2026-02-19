package toolimpls_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
	"easi/backend/internal/archassistant/infrastructure/toolimpls"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(server *httptest.Server) *agenthttp.Client {
	return agenthttp.NewClient(server.URL+"/api/v1", "test-token")
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

const validUUID = "550e8400-e29b-41d4-a716-446655440000"
const validUUID2 = "660e8400-e29b-41d4-a716-446655440000"

func TestCreateApplication_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/components", r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, "Payment Gateway", body["name"])
		assert.Equal(t, "Handles payments", body["description"])
		jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"id":   validUUID,
			"name": "Payment Gateway",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "create_application", map[string]interface{}{
		"name":        "Payment Gateway",
		"description": "Handles payments",
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Payment Gateway")
	assert.Contains(t, result.Content, validUUID)
}

func TestCreateApplication_MissingName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called when name is missing")
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "create_application", map[string]interface{}{})

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "name is required")
}

func TestUpdateApplication_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/components/"+validUUID, r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, "Payment Service", body["name"])
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"id":   validUUID,
			"name": "Payment Service",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "update_application", map[string]interface{}{
		"id":   validUUID,
		"name": "Payment Service",
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Payment Service")
}

func TestDeleteApplication_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/components/"+validUUID, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "delete_application", map[string]interface{}{
		"id": validUUID,
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Deleted application")
	assert.Contains(t, result.Content, validUUID)
}

func TestCreateCapability_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/capabilities", r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, "Payment Processing", body["name"])
		assert.Equal(t, validUUID2, body["domainId"])
		jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"id":   validUUID,
			"name": "Payment Processing",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "create_capability", map[string]interface{}{
		"name":     "Payment Processing",
		"domainId": validUUID2,
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Payment Processing")
	assert.Contains(t, result.Content, validUUID)
}

func TestCreateBusinessDomain_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/business-domains", r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, "Finance", body["name"])
		jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"id":   validUUID,
			"name": "Finance",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "create_business_domain", map[string]interface{}{
		"name": "Finance",
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Finance")
	assert.Contains(t, result.Content, validUUID)
}

func TestCreateApplicationRelation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/components/"+validUUID+"/relations", r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, validUUID2, body["targetId"])
		assert.Equal(t, "depends_on", body["type"])
		jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"id":       "rel-uuid-123",
			"sourceId": validUUID,
			"targetId": validUUID2,
			"type":     "depends_on",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "create_application_relation", map[string]interface{}{
		"sourceId": validUUID,
		"targetId": validUUID2,
		"type":     "depends_on",
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "relation")
}

func TestDeleteApplicationRelation_Success(t *testing.T) {
	relID := "770e8400-e29b-41d4-a716-446655440000"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/components/"+validUUID+"/relations/"+relID, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "delete_application_relation", map[string]interface{}{
		"componentId": validUUID,
		"relationId":  relID,
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Deleted relation")
}

func TestRealizeCapability_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/capabilities/"+validUUID+"/realizations", r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, validUUID2, body["applicationId"])
		jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"id":            "real-uuid-123",
			"capabilityId":  validUUID,
			"applicationId": validUUID2,
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "realize_capability", map[string]interface{}{
		"capabilityId":  validUUID,
		"applicationId": validUUID2,
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Linked")
}

func TestUnrealizeCapability_Success(t *testing.T) {
	realID := "880e8400-e29b-41d4-a716-446655440000"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/capabilities/"+validUUID+"/realizations/"+realID, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "unrealize_capability", map[string]interface{}{
		"capabilityId":  validUUID,
		"realizationId": realID,
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Unlinked")
}

func TestMutationTool_APIError(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		message    string
		toolName   string
		args       map[string]interface{}
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
			defer server.Close()

			registry := tools.NewRegistry()
			toolimpls.RegisterMutationTools(registry, newTestClient(server))

			result := executeTool(t, registry, tc.toolName, tc.args)

			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, tc.message)
		})
	}
}

func TestMutationTool_InvalidID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called with invalid UUID")
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	cases := []struct {
		name     string
		toolName string
		args     map[string]interface{}
	}{
		{
			name:     "update_application with invalid ID",
			toolName: "update_application",
			args:     map[string]interface{}{"id": "not-a-uuid", "name": "Test"},
		},
		{
			name:     "delete_application with invalid ID",
			toolName: "delete_application",
			args:     map[string]interface{}{"id": "not-a-uuid"},
		},
		{
			name:     "delete_application_relation with invalid componentId",
			toolName: "delete_application_relation",
			args:     map[string]interface{}{"componentId": "bad", "relationId": validUUID},
		},
		{
			name:     "delete_application_relation with invalid relationId",
			toolName: "delete_application_relation",
			args:     map[string]interface{}{"componentId": validUUID, "relationId": "bad"},
		},
		{
			name:     "realize_capability with invalid capabilityId",
			toolName: "realize_capability",
			args:     map[string]interface{}{"capabilityId": "bad", "applicationId": validUUID},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := executeTool(t, registry, tc.toolName, tc.args)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "valid UUID")
		})
	}
}

func TestRegisterMutationTools_AllRegistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	allPerms := &mockPermissions{permissions: map[string]bool{
		"components:write":   true,
		"capabilities:write": true,
		"domains:write":      true,
	}}

	available := registry.AvailableTools(allPerms, true)

	expectedTools := []string{
		"create_application",
		"update_application",
		"delete_application",
		"create_capability",
		"update_capability",
		"delete_capability",
		"create_business_domain",
		"update_business_domain",
		"create_application_relation",
		"delete_application_relation",
		"realize_capability",
		"unrealize_capability",
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

func TestUpdateCapability_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/capabilities/"+validUUID, r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, "Updated Capability", body["name"])
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"id":   validUUID,
			"name": "Updated Capability",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "update_capability", map[string]interface{}{
		"id":   validUUID,
		"name": "Updated Capability",
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Updated Capability")
}

func TestDeleteCapability_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/capabilities/"+validUUID, r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "delete_capability", map[string]interface{}{
		"id": validUUID,
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Deleted capability")
}

func TestUpdateBusinessDomain_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/business-domains/"+validUUID, r.URL.Path)
		body := readJSONBody(t, r)
		assert.Equal(t, "Updated Domain", body["name"])
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"id":   validUUID,
			"name": "Updated Domain",
		})
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	result := executeTool(t, registry, "update_business_domain", map[string]interface{}{
		"id":   validUUID,
		"name": "Updated Domain",
	})

	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Updated Domain")
}

func TestCreateApplication_NameTooLong(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called with name too long")
	}))
	defer server.Close()

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, newTestClient(server))

	longName := make([]byte, 201)
	for i := range longName {
		longName[i] = 'a'
	}

	result := executeTool(t, registry, "create_application", map[string]interface{}{
		"name": string(longName),
	})

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "200 characters")
}

func TestCreateApplication_APIUnreachable(t *testing.T) {
	client := agenthttp.NewClient("http://localhost:1/api/v1", "test-token")

	registry := tools.NewRegistry()
	toolimpls.RegisterMutationTools(registry, client)

	result := executeTool(t, registry, "create_application", map[string]interface{}{
		"name": "Test",
	})

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "Failed to reach API")
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
