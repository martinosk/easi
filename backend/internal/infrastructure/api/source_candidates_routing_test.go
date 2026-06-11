package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestChiServesStaticSourceCandidatesAlongsideCapabilitySubrouter(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/capabilities", func(r chi.Router) {
		r.Get("/{id}", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("byid")) })
	})
	r.Get("/capabilities/source-candidates", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("candidates")) })

	cases := map[string]string{
		"/capabilities/source-candidates": "candidates",
		"/capabilities/abc":               "byid",
	}
	for target, want := range cases {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Body.String() != want {
			t.Fatalf("%s routed to %q, want %q", target, rec.Body.String(), want)
		}
	}
}
