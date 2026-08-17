package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newForwardedRequest(remoteAddr string, headers map[string]string) *http.Request {
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req.RemoteAddr = remoteAddr
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	return req
}

func throughClientIPMiddleware[T any](req *http.Request, read func(*http.Request) T) T {
	var result T
	handler := ClientIP()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		result = read(r)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)
	return result
}

func TestClientIP_ResolvesAddressOfTheClosestUntrustedHop(t *testing.T) {
	spoofedForwardingHeaders := map[string]string{
		"X-Forwarded-For": "203.0.113.9",
		"X-Real-IP":       "203.0.113.9",
		"True-Client-IP":  "203.0.113.9",
	}

	tests := []struct {
		name       string
		cidrs      string
		count      string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "no trusted proxies falls back to the connecting peer",
			remoteAddr: "198.51.100.7:54321",
			want:       "198.51.100.7",
		},
		{
			name:       "no trusted proxies ignores forwarding headers",
			remoteAddr: "198.51.100.7:54321",
			headers:    spoofedForwardingHeaders,
			want:       "198.51.100.7",
		},
		{
			name:       "trusted proxy count reads the entry that proxy appended",
			count:      "1",
			remoteAddr: "10.0.0.1:54321",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.9, 198.51.100.7"},
			want:       "198.51.100.7",
		},
		{
			name:       "trusted proxy count ignores a client supplied prefix",
			count:      "1",
			remoteAddr: "10.0.0.1:54321",
			headers:    map[string]string{"X-Forwarded-For": "127.0.0.1, 198.51.100.7"},
			want:       "198.51.100.7",
		},
		{
			name:       "trusted cidrs walk past every trusted hop",
			cidrs:      "10.0.0.0/8",
			remoteAddr: "10.0.0.1:54321",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.7, 10.0.0.9, 10.0.0.1"},
			want:       "198.51.100.7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TRUSTED_PROXY_CIDRS", tc.cidrs)
			t.Setenv("TRUSTED_PROXY_COUNT", tc.count)

			req := newForwardedRequest(tc.remoteAddr, tc.headers)

			assert.Equal(t, tc.want, throughClientIPMiddleware(req, getClientIP))
		})
	}
}

func TestIsLoopback_TrueForLoopbackConnection(t *testing.T) {
	req := newForwardedRequest("127.0.0.1:54321", nil)

	assert.True(t, isLoopback(req))
}

func TestIsLoopback_IgnoresSpoofedForwardingHeaders(t *testing.T) {
	t.Setenv("TRUSTED_PROXY_CIDRS", "")
	t.Setenv("TRUSTED_PROXY_COUNT", "")

	req := newForwardedRequest("203.0.113.9:54321", map[string]string{
		"X-Forwarded-For": "127.0.0.1",
		"X-Real-IP":       "127.0.0.1",
		"True-Client-IP":  "127.0.0.1",
	})

	assert.False(t, throughClientIPMiddleware(req, isLoopback),
		"forwarding headers must not satisfy the agent-token loopback check")
}
