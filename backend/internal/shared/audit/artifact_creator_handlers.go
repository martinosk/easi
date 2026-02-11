package audit

import (
	"net/http"

	sharedAPI "easi/backend/internal/shared/api"
)

type ArtifactCreatorHandlers struct {
	reader  ArtifactCreatorReader
	hateoas *AuditLinks
}

func NewArtifactCreatorHandlers(reader ArtifactCreatorReader, hateoas ...*AuditLinks) *ArtifactCreatorHandlers {
	h := &ArtifactCreatorHandlers{reader: reader}
	if len(hateoas) > 0 {
		h.hateoas = hateoas[0]
	}
	return h
}

// GetArtifactCreators godoc
// @Summary List artifact creators
// @Description Returns the creator (first event actor) for each tree-relevant aggregate: components, capabilities, vendors, internal teams, and acquired entities. Requires audit:read permission.
// @Tags audit
// @Produce json
// @Success 200 {object} ArtifactCreatorsResponse
// @Failure 401 {object} sharedAPI.ErrorResponse "Authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions - requires audit:read"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security ApiKeyAuth
// @Router /artifact-creators [get]
func (h *ArtifactCreatorHandlers) GetArtifactCreators(w http.ResponseWriter, r *http.Request) {
	creators, err := h.reader.GetArtifactCreators(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve artifact creators")
		return
	}

	if creators == nil {
		creators = []ArtifactCreator{}
	}

	links := map[string]any{
		"self": map[string]string{"href": h.selfLink(), "method": "GET"},
	}

	response := ArtifactCreatorsResponse{
		Data:  creators,
		Links: links,
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

func (h *ArtifactCreatorHandlers) selfLink() string {
	if h.hateoas != nil {
		return h.hateoas.Base() + "/artifact-creators"
	}
	return "/api/v1/artifact-creators"
}
