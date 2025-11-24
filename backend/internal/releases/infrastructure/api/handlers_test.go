package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/releases/domain/aggregates"
	"easi/backend/internal/releases/domain/valueobjects"
	"github.com/go-chi/chi/v5"
)

type mockReleaseRepository struct {
	findLatestFn    func(ctx context.Context) (*aggregates.Release, error)
	findByVersionFn func(ctx context.Context, version valueobjects.Version) (*aggregates.Release, error)
	findAllFn       func(ctx context.Context) ([]*aggregates.Release, error)
}

func (m *mockReleaseRepository) FindLatest(ctx context.Context) (*aggregates.Release, error) {
	if m.findLatestFn != nil {
		return m.findLatestFn(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReleaseRepository) FindByVersion(ctx context.Context, version valueobjects.Version) (*aggregates.Release, error) {
	if m.findByVersionFn != nil {
		return m.findByVersionFn(ctx, version)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReleaseRepository) FindAll(ctx context.Context) ([]*aggregates.Release, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockReleaseRepository) Save(ctx context.Context, release *aggregates.Release) error {
	return errors.New("not implemented - releases are seeded via migrations")
}

func mustCreateVersion(t *testing.T, value string) valueobjects.Version {
	t.Helper()
	version, err := valueobjects.NewVersion(value)
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}
	return version
}

func createTestRelease(t *testing.T, versionStr string, releaseDate time.Time, notes string) *aggregates.Release {
	t.Helper()
	version := mustCreateVersion(t, versionStr)
	return aggregates.NewRelease(version, releaseDate, notes)
}

func TestGetLatest_ReturnsLatestRelease(t *testing.T) {
	releaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	release := createTestRelease(t, "1.0.0", releaseDate, "## Features\n- New feature")

	repo := &mockReleaseRepository{
		findLatestFn: func(ctx context.Context) (*aggregates.Release, error) {
			return release, nil
		},
	}

	handler := NewReleaseHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/latest", nil)
	rec := httptest.NewRecorder()

	handler.GetLatest(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetLatest() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response ReleaseResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Response.Version = %q, want %q", response.Version, "1.0.0")
	}
	if response.Notes != "## Features\n- New feature" {
		t.Errorf("Response.Notes = %q, want %q", response.Notes, "## Features\n- New feature")
	}
	if response.Links["self"] != "/api/v1/releases/1.0.0" {
		t.Errorf("Response._links.self = %q, want %q", response.Links["self"], "/api/v1/releases/1.0.0")
	}
}

func TestGetLatest_Returns404WhenNoReleasesExist(t *testing.T) {
	repo := &mockReleaseRepository{
		findLatestFn: func(ctx context.Context) (*aggregates.Release, error) {
			return nil, errors.New("no releases found")
		},
	}

	handler := NewReleaseHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/latest", nil)
	rec := httptest.NewRecorder()

	handler.GetLatest(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GetLatest() status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetByVersion_ReturnsReleaseForValidVersion(t *testing.T) {
	releaseDate := time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC)
	release := createTestRelease(t, "2.1.0", releaseDate, "Bug fixes")

	repo := &mockReleaseRepository{
		findByVersionFn: func(ctx context.Context, version valueobjects.Version) (*aggregates.Release, error) {
			if version.Value() == "2.1.0" {
				return release, nil
			}
			return nil, errors.New("not found")
		},
	}

	handler := NewReleaseHandler(repo)

	r := chi.NewRouter()
	r.Get("/api/v1/releases/{version}", handler.GetByVersion)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/2.1.0", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetByVersion() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response ReleaseResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Version != "2.1.0" {
		t.Errorf("Response.Version = %q, want %q", response.Version, "2.1.0")
	}
}

func TestGetByVersion_Returns400ForInvalidVersionFormat(t *testing.T) {
	repo := &mockReleaseRepository{}
	handler := NewReleaseHandler(repo)

	r := chi.NewRouter()
	r.Get("/api/v1/releases/{version}", handler.GetByVersion)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/invalid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("GetByVersion() status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByVersion_Returns404WhenReleaseNotFound(t *testing.T) {
	repo := &mockReleaseRepository{
		findByVersionFn: func(ctx context.Context, version valueobjects.Version) (*aggregates.Release, error) {
			return nil, errors.New("release not found")
		},
	}

	handler := NewReleaseHandler(repo)

	r := chi.NewRouter()
	r.Get("/api/v1/releases/{version}", handler.GetByVersion)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/9.9.9", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GetByVersion() status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetAll_ReturnsAllReleasesAsCollection(t *testing.T) {
	releases := []*aggregates.Release{
		createTestRelease(t, "1.0.0", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "Initial release"),
		createTestRelease(t, "1.1.0", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), "Minor update"),
		createTestRelease(t, "2.0.0", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), "Major update"),
	}

	repo := &mockReleaseRepository{
		findAllFn: func(ctx context.Context) ([]*aggregates.Release, error) {
			return releases, nil
		},
	}

	handler := NewReleaseHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases", nil)
	rec := httptest.NewRecorder()

	handler.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetAll() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response struct {
		Data  []ReleaseResponse `json:"data"`
		Links map[string]string `json:"_links"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) != 3 {
		t.Errorf("GetAll() returned %d releases, want 3", len(response.Data))
	}

	if response.Links["self"] != "/api/v1/releases" {
		t.Errorf("Response._links.self = %q, want %q", response.Links["self"], "/api/v1/releases")
	}
}

func TestGetAll_ReturnsEmptyArrayWhenNoReleases(t *testing.T) {
	repo := &mockReleaseRepository{
		findAllFn: func(ctx context.Context) ([]*aggregates.Release, error) {
			return []*aggregates.Release{}, nil
		},
	}

	handler := NewReleaseHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases", nil)
	rec := httptest.NewRecorder()

	handler.GetAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GetAll() status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response struct {
		Data []ReleaseResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Data) != 0 {
		t.Errorf("GetAll() returned %d releases, want 0", len(response.Data))
	}
}

func TestGetAll_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockReleaseRepository{
		findAllFn: func(ctx context.Context) ([]*aggregates.Release, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewReleaseHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases", nil)
	rec := httptest.NewRecorder()

	handler.GetAll(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("GetAll() status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestReleaseResponse_FormatsDateAsRFC3339(t *testing.T) {
	releaseDate := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	release := createTestRelease(t, "1.0.0", releaseDate, "Notes")

	response := toReleaseResponse(release)

	expectedDateStr := "2024-06-15T14:30:00Z"
	if response.ReleaseDate != expectedDateStr {
		t.Errorf("Response.ReleaseDate = %q, want %q", response.ReleaseDate, expectedDateStr)
	}
}

func TestNewReleaseHandler_CreatesHandler(t *testing.T) {
	repo := &mockReleaseRepository{}
	handler := NewReleaseHandler(repo)

	if handler == nil {
		t.Error("NewReleaseHandler() returned nil")
	}
}
