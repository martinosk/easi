package agenthttp_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/infrastructure/agenthttp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func newClientForServer(server *httptest.Server, token string) *agenthttp.Client {
	return agenthttp.NewClient(server.URL+"/api/v1", token)
}

func jsonErrorHandler(statusCode int, errorText, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   errorText,
			"message": message,
		})
	}
}

type bodyExpectation struct {
	t              *testing.T
	method         string
	path           string
	bodyKey        string
	bodyVal        string
	responseStatus int
}

func (e bodyExpectation) handler() http.HandlerFunc {
	e.t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(e.t, e.method, r.Method)
		assert.Equal(e.t, e.path, r.URL.Path)
		assert.Equal(e.t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		require.NoError(e.t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(e.t, e.bodyVal, body[e.bodyKey])

		w.WriteHeader(e.responseStatus)
		json.NewEncoder(w).Encode(body)
	}
}

func TestClient_Get_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/components", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"name": "test"})
	})
	defer server.Close()

	client := newClientForServer(server, "test-token")
	resp, err := client.Get(context.Background(), "/components")

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.True(t, resp.IsSuccess())

	var body map[string]string
	require.NoError(t, json.Unmarshal(resp.Body, &body))
	assert.Equal(t, "test", body["name"])
}

func TestClient_Get_NotFound(t *testing.T) {
	server := newTestServer(jsonErrorHandler(http.StatusNotFound, "Not Found", "Component not found"))
	defer server.Close()

	client := newClientForServer(server, "test-token")
	resp, err := client.Get(context.Background(), "/components/123")

	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.False(t, resp.IsSuccess())
	assert.Equal(t, "Component not found", resp.ErrorMessage())
}

func TestClient_PostAndPut_Success(t *testing.T) {
	cases := []struct {
		name           string
		expect         bodyExpectation
		invoke         func(*agenthttp.Client) (*agenthttp.Response, error)
		wantStatusCode int
	}{
		{
			name: "POST creates resource",
			expect: bodyExpectation{
				method: http.MethodPost, path: "/api/v1/components",
				bodyKey: "name", bodyVal: "new-component", responseStatus: http.StatusCreated,
			},
			invoke: func(c *agenthttp.Client) (*agenthttp.Response, error) {
				return c.Post(context.Background(), "/components", map[string]string{"name": "new-component"})
			},
			wantStatusCode: 201,
		},
		{
			name: "PUT updates resource",
			expect: bodyExpectation{
				method: http.MethodPut, path: "/api/v1/components/abc-123",
				bodyKey: "name", bodyVal: "updated-name", responseStatus: http.StatusOK,
			},
			invoke: func(c *agenthttp.Client) (*agenthttp.Response, error) {
				return c.Put(context.Background(), "/components/abc-123", map[string]string{"name": "updated-name"})
			},
			wantStatusCode: 200,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect.t = t
			server := newTestServer(tc.expect.handler())
			defer server.Close()

			client := newClientForServer(server, "test-token")
			resp, err := tc.invoke(client)

			require.NoError(t, err)
			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)
			assert.True(t, resp.IsSuccess())
		})
	}
}

func TestClient_Delete_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/components/abc-123", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	client := newClientForServer(server, "test-token")
	resp, err := client.Delete(context.Background(), "/components/abc-123")

	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	assert.True(t, resp.IsSuccess())
}

func TestClient_AuthorizationHeader(t *testing.T) {
	methods := []struct {
		name   string
		invoke func(client *agenthttp.Client) (*agenthttp.Response, error)
	}{
		{
			name: "GET",
			invoke: func(c *agenthttp.Client) (*agenthttp.Response, error) {
				return c.Get(context.Background(), "/test")
			},
		},
		{
			name: "POST",
			invoke: func(c *agenthttp.Client) (*agenthttp.Response, error) {
				return c.Post(context.Background(), "/test", map[string]string{"k": "v"})
			},
		},
		{
			name: "PUT",
			invoke: func(c *agenthttp.Client) (*agenthttp.Response, error) {
				return c.Put(context.Background(), "/test", map[string]string{"k": "v"})
			},
		},
		{
			name: "DELETE",
			invoke: func(c *agenthttp.Client) (*agenthttp.Response, error) {
				return c.Delete(context.Background(), "/test")
			},
		},
	}

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "AgentToken my-secret-token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusOK)
			})
			defer server.Close()

			client := newClientForServer(server, "my-secret-token")
			_, err := m.invoke(client)
			require.NoError(t, err)
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newClientForServer(server, "test-token")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Get(ctx, "/components")
	require.Error(t, err)
}

func TestClient_ErrorMessage_ParsesJSON(t *testing.T) {
	server := newTestServer(jsonErrorHandler(http.StatusConflict, "Conflict", "Component with this name already exists"))
	defer server.Close()

	client := newClientForServer(server, "test-token")
	resp, err := client.Get(context.Background(), "/components")

	require.NoError(t, err)
	assert.Equal(t, 409, resp.StatusCode)
	assert.Equal(t, "Component with this name already exists", resp.ErrorMessage())
}

func TestClient_ErrorMessage_Fallback(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "unexpected server error")
	})
	defer server.Close()

	client := newClientForServer(server, "test-token")
	resp, err := client.Get(context.Background(), "/components")

	require.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "Internal Server Error", resp.ErrorMessage())
}
