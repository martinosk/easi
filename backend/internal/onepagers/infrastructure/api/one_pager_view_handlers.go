package api

import (
	"context"
	"net/http"

	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
)

type onePagerReader interface {
	Get(ctx context.Context, subjectType valueobjects.SubjectType, subjectID string) (*queries.OnePager, error)
}

type OnePagerViewHandlers struct {
	query onePagerReader
	links *OnePagerLinks
}

func NewOnePagerViewHandlers(query onePagerReader, links *OnePagerLinks) *OnePagerViewHandlers {
	return &OnePagerViewHandlers{query: query, links: links}
}

// GetOnePager godoc
// @Summary Get the composed one-pager for a subject
// @Description Assembles the tenant's one-pager configuration, the subject's recorded field values, and built-in field data sourced from the owning context into a single field list in the configured interleaved display order, alongside a completeness summary of the active required custom fields.
// @Tags one-pagers
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Param subjectID path string true "Subject ID"
// @Success 200 {object} OnePagerDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/{subjectType}/{subjectID} [get]
func (h *OnePagerViewHandlers) GetOnePager(subjectType valueobjects.SubjectType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subjectID, ok := resolveSubjectID(w, r)
		if !ok {
			return
		}
		onePager, err := h.query.Get(r.Context(), subjectType, subjectID)
		if err != nil {
			sharedAPI.HandleError(w, err)
			return
		}
		sharedAPI.RespondJSON(w, http.StatusOK, BuildOnePagerDTO(onePager, h.links))
	}
}
