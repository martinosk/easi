package api

import (
	"context"
	"net/http"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type StandardApplicationQueries interface {
	GetCurrentByEnterpriseCapability(ctx context.Context, ecID string) (*readmodels.StandardApplicationDTO, error)
	GetHistoryByAggregateID(ctx context.Context, id string) (*readmodels.StandardApplicationHistoryDTO, error)
	FindAggregateIDForEnterpriseCapability(ctx context.Context, ecID string) (string, bool, error)
}

type StandardApplicationHandlers struct {
	commandBus cqrs.CommandBus
	queries    StandardApplicationQueries
	hateoas    *StandardApplicationLinks
}

func NewStandardApplicationHandlers(commandBus cqrs.CommandBus, queries StandardApplicationQueries, hateoas *StandardApplicationLinks) *StandardApplicationHandlers {
	return &StandardApplicationHandlers{
		commandBus: commandBus,
		queries:    queries,
		hateoas:    hateoas,
	}
}

type SetStandardApplicationRequest struct {
	ApplicationID string `json:"applicationId"`
	Narrative     string `json:"narrative"`
}

type ECStandardApplicationResponse struct {
	Standard *readmodels.StandardApplicationDTO `json:"standard"`
	Links    sharedAPI.Links                    `json:"_links,omitempty"`
}

// GetStandardApplicationForEnterpriseCapability godoc
// @Summary Get the standard application for an enterprise capability
// @Description Returns the current standard application on an enterprise capability, or null if none.
// @Tags standard-applications
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} ECStandardApplicationResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/standard-application [get]
func (h *StandardApplicationHandlers) GetStandardApplicationForEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	standard, err := h.queries.GetCurrentByEnterpriseCapability(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	envelope := ECStandardApplicationResponse{
		Standard: standard,
		Links:    h.hateoas.EnvelopeLinks(ecID, standard != nil, actor),
	}
	if standard != nil {
		standard.Links = h.hateoas.StandardLinks(ecID, actor)
	}
	sharedAPI.RespondJSON(w, http.StatusOK, envelope)
}

// SetStandardApplication godoc
// @Summary Set or change the standard application for an enterprise capability
// @Description Sets the standard application on the EC; if one already exists, replaces it. Narrative is required. The previous standard is preserved in history. Returns 201 on first creation (with Location header) and 200 on replacement.
// @Tags standard-applications
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Param body body SetStandardApplicationRequest true "Standard application data"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.StandardApplicationDTO
// @Success 201 {object} easi_backend_internal_architecturedirection_application_readmodels.StandardApplicationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/standard-application [put]
func (h *StandardApplicationHandlers) SetStandardApplication(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[SetStandardApplicationRequest](w, r)
	if !ok {
		return
	}
	existedBefore, err := h.queries.GetCurrentByEnterpriseCapability(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	cmd := &commands.SetStandardApplication{
		EnterpriseCapabilityID: ecID,
		ApplicationID:          req.ApplicationID,
		Narrative:              req.Narrative,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	statusCode := http.StatusOK
	if existedBefore == nil {
		statusCode = http.StatusCreated
	}
	h.respondWithCurrentStandard(w, r, ecID, statusCode)
}

// GetStandardApplicationHistory godoc
// @Summary Get the history of standard applications for an enterprise capability
// @Description Returns the reverse-chronological history of standard applications set on the EC.
// @Tags standard-applications
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.StandardApplicationHistoryDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/standard-application/history [get]
func (h *StandardApplicationHandlers) GetStandardApplicationHistory(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	aggregateID, exists, err := h.queries.FindAggregateIDForEnterpriseCapability(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	history := emptyHistoryFor(ecID)
	if exists {
		fetched, err := h.queries.GetHistoryByAggregateID(r.Context(), aggregateID)
		if err != nil {
			sharedAPI.HandleError(w, err)
			return
		}
		if fetched != nil {
			history = fetched
		}
	}
	history.Links = h.hateoas.HistoryLinks(ecID)
	sharedAPI.RespondJSON(w, http.StatusOK, history)
}

func emptyHistoryFor(ecID string) *readmodels.StandardApplicationHistoryDTO {
	return &readmodels.StandardApplicationHistoryDTO{
		StandardApplicationID:  ecID,
		EnterpriseCapabilityID: ecID,
		Entries:                []readmodels.StandardApplicationHistoryEntryDTO{},
	}
}

func (h *StandardApplicationHandlers) respondWithCurrentStandard(w http.ResponseWriter, r *http.Request, ecID string, statusCode int) {
	standard, err := h.queries.GetCurrentByEnterpriseCapability(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if standard == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, ErrNoStandardApplication, "failed to load standard application")
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	standard.Links = h.hateoas.StandardLinks(ecID, actor)
	if statusCode == http.StatusCreated {
		location := sharedAPI.BuildSubResourceLink(enterpriseCapabilitiesPath, sharedAPI.ResourceID(ecID), standardApplicationSubPath)
		sharedAPI.RespondCreated(w, location, standard)
		return
	}
	sharedAPI.RespondJSON(w, statusCode, standard)
}
