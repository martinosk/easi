package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func newCompressedRouter(handler http.HandlerFunc) *chi.Mux {
	r := chi.NewRouter()
	r.Use(compressionMiddleware())
	r.Get("/data", handler)
	return r
}

func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

func TestCompressionMiddlewareGzipsJSONWhenClientAcceptsGzip(t *testing.T) {
	body := `{"data":"` + strings.Repeat("x", 4096) + `"}`
	r := newCompressedRouter(jsonHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want %q", got, "gzip")
	}

	gz, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("open gzip reader: %v", err)
	}
	decompressed, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("decompress body: %v", err)
	}
	if string(decompressed) != body {
		t.Fatalf("decompressed body does not match original (len %d vs %d)", len(decompressed), len(body))
	}
}

func TestCompressionMiddlewareSkipsClientsWithoutAcceptEncoding(t *testing.T) {
	body := `{"data":"plain"}`
	r := newCompressedRouter(jsonHandler(body))

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q, want empty", got)
	}
	if rec.Body.String() != body {
		t.Fatalf("body = %q, want %q", rec.Body.String(), body)
	}
}
