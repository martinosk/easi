package api

import (
	"context"
	"net/http"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
)

type TimeSuggestionReader interface {
	GetAllSuggestions(ctx context.Context) ([]readmodels.TimeSuggestionDTO, error)
	GetByCapability(ctx context.Context, capabilityID string) ([]readmodels.TimeSuggestionDTO, error)
	GetByComponent(ctx context.Context, componentID string) ([]readmodels.TimeSuggestionDTO, error)
}

type TimeSuggestionsHandlers struct {
	readModel TimeSuggestionReader
	hateoas   *sharedAPI.HATEOASLinks
}

func NewTimeSuggestionsHandlers(readModel TimeSuggestionReader, hateoas *sharedAPI.HATEOASLinks) *TimeSuggestionsHandlers {
	return &TimeSuggestionsHandlers{
		readModel: readModel,
		hateoas:   hateoas,
	}
}

// GetTimeSuggestions godoc
// @Summary Get TIME suggestions
// @Description Retrieves TIME (Tolerate, Invest, Migrate, Eliminate) suggestions based on strategic importance and application fit gaps
// @Tags time-suggestions
// @Produce json
// @Param capabilityId query string false "Filter by capability ID"
// @Param componentId query string false "Filter by component ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]readmodels.TimeSuggestionDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /time-suggestions [get]
func (h *TimeSuggestionsHandlers) GetTimeSuggestions(w http.ResponseWriter, r *http.Request) {
	capabilityID := r.URL.Query().Get("capabilityId")
	componentID := r.URL.Query().Get("componentId")

	suggestions, err := h.fetchSuggestions(r.Context(), capabilityID, componentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	sharedAPI.RespondCollection(w, http.StatusOK, suggestions, h.hateoas.TimeSuggestionsCollectionLinks())
}

func (h *TimeSuggestionsHandlers) fetchSuggestions(ctx context.Context, capabilityID, componentID string) ([]readmodels.TimeSuggestionDTO, error) {
	if capabilityID != "" {
		return h.readModel.GetByCapability(ctx, capabilityID)
	}
	if componentID != "" {
		return h.readModel.GetByComponent(ctx, componentID)
	}
	return h.readModel.GetAllSuggestions(ctx)
}
