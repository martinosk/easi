package api

import (
	"net/http"
	"time"

	"easi/backend/internal/releases/domain"
	"easi/backend/internal/releases/domain/aggregates"
	"easi/backend/internal/releases/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
)

type ReleaseHandler struct {
	repo domain.ReleaseRepository
}

func NewReleaseHandler(repo domain.ReleaseRepository) *ReleaseHandler {
	return &ReleaseHandler{repo: repo}
}

type ReleaseResponse struct {
	Version     string            `json:"version"`
	ReleaseDate string            `json:"releaseDate"`
	Notes       string            `json:"notes"`
	Links       map[string]string `json:"_links,omitempty"`
}

func toReleaseResponse(release *aggregates.Release) ReleaseResponse {
	return ReleaseResponse{
		Version:     release.Version().Value(),
		ReleaseDate: release.ReleaseDate().Format(time.RFC3339),
		Notes:       release.Notes(),
		Links: map[string]string{
			"self": "/api/v1/releases/" + release.Version().Value(),
		},
	}
}

// GetLatest godoc
// @Summary Get latest release
// @Description Returns the most recent release notes
// @Tags releases
// @Produce json
// @Success 200 {object} ReleaseResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Router /releases/latest [get]
func (h *ReleaseHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	release, err := h.repo.FindLatest(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusNotFound, err, "No releases found")
		return
	}
	sharedAPI.RespondJSON(w, http.StatusOK, toReleaseResponse(release))
}

// GetByVersion godoc
// @Summary Get release by version
// @Description Returns release notes for a specific version
// @Tags releases
// @Produce json
// @Param version path string true "Version number (e.g., v1.2.0)"
// @Success 200 {object} ReleaseResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Router /releases/{version} [get]
func (h *ReleaseHandler) GetByVersion(w http.ResponseWriter, r *http.Request) {
	versionStr := chi.URLParam(r, "version")
	version, err := valueobjects.NewVersion(versionStr)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid version format")
		return
	}

	release, err := h.repo.FindByVersion(r.Context(), version)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusNotFound, err, "Release not found")
		return
	}
	sharedAPI.RespondJSON(w, http.StatusOK, toReleaseResponse(release))
}

// GetAll godoc
// @Summary Get all releases
// @Description Returns all release notes ordered by version descending
// @Tags releases
// @Produce json
// @Success 200 {object} object{data=[]ReleaseResponse,_links=map[string]string}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /releases [get]
func (h *ReleaseHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	releases, err := h.repo.FindAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch releases")
		return
	}

	responses := make([]ReleaseResponse, 0, len(releases))
	for _, release := range releases {
		responses = append(responses, toReleaseResponse(release))
	}

	sharedAPI.RespondCollection(w, http.StatusOK, responses, sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/releases", "GET"),
	})
}
