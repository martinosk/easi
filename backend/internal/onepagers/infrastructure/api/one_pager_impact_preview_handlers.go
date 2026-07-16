package api

import (
	"context"
	"net/http"

	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
)

type impactPreviewReader interface {
	Preview(ctx context.Context, subjectType valueobjects.SubjectType, field queries.PreviewField) (*queries.ImpactPreview, error)
}

type ImpactPreviewHandlers struct {
	query impactPreviewReader
	links *OnePagerLinks
}

func NewImpactPreviewHandlers(query impactPreviewReader, links *OnePagerLinks) *ImpactPreviewHandlers {
	return &ImpactPreviewHandlers{query: query, links: links}
}

// GetImpactPreview godoc
// @Summary Preview the impact of a field requirement change
// @Description Side-effect-free preview of how many subjects would be marked incomplete by making a field required. For an existing custom field, counts the subjects of the type lacking a recorded value; for a built-in field (fieldKind=builtIn), counts the subjects lacking a value for that built-in through the supplier read models; without fieldId, counts the full subject population for a new custom field being defined. Appends no events and changes no configuration or facts.
// @Tags one-pagers
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Param fieldId query string false "Existing custom field ID or built-in catalog entry ID; omit for a new custom field being defined"
// @Param fieldKind query string false "Field kind discriminator: 'custom' (default) or 'builtIn'" Enums(custom, builtIn)
// @Success 200 {object} ImpactPreviewDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/impact-preview [get]
func (h *ImpactPreviewHandlers) GetImpactPreview(w http.ResponseWriter, r *http.Request) {
	subjectType, err := valueobjects.NewSubjectType(sharedAPI.GetPathParam(r, "subjectType"))
	if err != nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Unknown subject type")
		return
	}

	preview, err := h.query.Preview(r.Context(), subjectType, previewFieldFrom(r))
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	sharedAPI.RespondJSON(w, http.StatusOK, BuildImpactPreviewDTO(preview, h.links, previewFieldKind(r)))
}

func previewFieldFrom(r *http.Request) queries.PreviewField {
	return queries.PreviewField{Kind: previewFieldKind(r), ID: r.URL.Query().Get("fieldId")}
}

func previewFieldKind(r *http.Request) string {
	kind := r.URL.Query().Get("fieldKind")
	if kind == string(valueobjects.FieldRefKindBuiltIn) {
		return string(valueobjects.FieldRefKindBuiltIn)
	}
	return string(valueobjects.FieldRefKindCustom)
}
