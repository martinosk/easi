package api

import (
	"context"
	"net/http"

	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type SubjectIDsSource interface {
	SubjectIDs(ctx context.Context, subjectType string) ([]string, error)
}

type CompletenessIndicatorsSource interface {
	ForSubjects(ctx context.Context, subjectType string, subjectIDs []string) (map[string]bool, bool, error)
}

type OnePagerCompletenessDTO struct {
	SubjectID string `json:"subjectId"`
	Complete  bool   `json:"complete"`
}

type OnePagerCompletenessResponse struct {
	Data  []OnePagerCompletenessDTO `json:"data"`
	Links types.Links               `json:"_links,omitempty"`
}

type OnePagerCompletenessHandlers struct {
	subjects   SubjectIDsSource
	indicators CompletenessIndicatorsSource
	links      *OnePagerLinks
}

func NewOnePagerCompletenessHandlers(subjects SubjectIDsSource, indicators CompletenessIndicatorsSource, links *OnePagerLinks) *OnePagerCompletenessHandlers {
	return &OnePagerCompletenessHandlers{subjects: subjects, indicators: indicators, links: links}
}

// GetCompleteness godoc
// @Summary Get one-pager completeness for every subject of a type
// @Description Returns, for each subject of the given type, whether all required one-pager fields are filled. The collection is empty when the subject type has no required field, in which case no indicator applies (spec 178 rule 10).
// @Tags one-pagers
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Success 200 {object} OnePagerCompletenessResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/{subjectType}/completeness [get]
func (h *OnePagerCompletenessHandlers) GetCompleteness(subjectType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := h.completeness(r.Context(), subjectType)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to evaluate one-pager completeness")
			return
		}
		sharedAPI.RespondJSON(w, http.StatusOK, OnePagerCompletenessResponse{
			Data:  data,
			Links: types.Links{"self": h.links.Get("/one-pagers/" + subjectType + "/completeness")},
		})
	}
}

func (h *OnePagerCompletenessHandlers) completeness(ctx context.Context, subjectType string) ([]OnePagerCompletenessDTO, error) {
	subjectIDs, err := h.subjects.SubjectIDs(ctx, subjectType)
	if err != nil {
		return nil, err
	}
	indicators, applies, err := h.indicators.ForSubjects(ctx, subjectType, subjectIDs)
	if err != nil {
		return nil, err
	}
	data := make([]OnePagerCompletenessDTO, 0, len(subjectIDs))
	if !applies {
		return data, nil
	}
	for _, subjectID := range subjectIDs {
		data = append(data, OnePagerCompletenessDTO{SubjectID: subjectID, Complete: indicators[subjectID]})
	}
	return data, nil
}
