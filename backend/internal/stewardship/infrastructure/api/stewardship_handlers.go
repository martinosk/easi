package api

import (
	"context"
	"errors"
	"net/http"

	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"
	"easi/backend/internal/stewardship/application/commands"
	"easi/backend/internal/stewardship/application/handlers"
	"easi/backend/internal/stewardship/application/readmodels"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

var errDomainIDRequired = errors.New("domainId query parameter is required")

type StewardshipQueries interface {
	DomainStewardships(ctx context.Context, domainID string) (*readmodels.DomainStewardshipsDTO, error)
}

type StewardshipHandlers struct {
	commandBus cqrs.CommandBus
	queries    StewardshipQueries
	links      *StewardshipLinks
}

func NewStewardshipHandlers(commandBus cqrs.CommandBus, queries StewardshipQueries, links *StewardshipLinks) *StewardshipHandlers {
	return &StewardshipHandlers{commandBus: commandBus, queries: queries, links: links}
}

type DomainStewardshipsResponse struct {
	DomainID   string                             `json:"domainId"`
	DomainName string                             `json:"domainName"`
	Fallback   *readmodels.PersonDTO              `json:"fallback"`
	Data       []readmodels.ConcernStewardshipDTO `json:"data"`
	Links      types.Links                        `json:"_links"`
}

type AssignStewardRequest struct {
	StewardID string `json:"stewardId"`
}

// GetStewardships godoc
// @Summary List the stewardships of a business domain
// @Description Lists every concern in fixed order with its steward (or null when unassigned), attribution and the domain architect as fallback. Callers holding domains:write receive x-assign on every concern and x-release on assigned ones.
// @Tags stewardships
// @Produce json
// @Security CookieAuth
// @Param domainId query string true "Business domain ID"
// @Success 200 {object} DomainStewardshipsResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /stewardships [get]
func (h *StewardshipHandlers) GetStewardships(w http.ResponseWriter, r *http.Request) {
	domainID := r.URL.Query().Get("domainId")
	if domainID == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, errDomainIDRequired, errDomainIDRequired.Error())
		return
	}
	result, ok := h.loadDomainStewardships(w, r, domainID)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	for i := range result.Concerns {
		result.Concerns[i].Links = h.links.ItemLinks(domainID, result.Concerns[i], actor)
	}
	sharedAPI.RespondJSON(w, http.StatusOK, DomainStewardshipsResponse{
		DomainID:   result.DomainID,
		DomainName: result.DomainName,
		Fallback:   result.Fallback,
		Data:       result.Concerns,
		Links:      h.links.CollectionLinks(domainID, actor),
	})
}

// GetStewardship godoc
// @Summary Get the stewardship of one concern in a business domain
// @Description Returns the concern with its steward, or a null steward when unassigned.
// @Tags stewardships
// @Produce json
// @Security CookieAuth
// @Param domainId path string true "Business domain ID"
// @Param concern path string true "Concern (ownership, assessment, documentation, planning, structure)"
// @Success 200 {object} readmodels.ConcernStewardshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /stewardships/{domainId}/{concern} [get]
func (h *StewardshipHandlers) GetStewardship(w http.ResponseWriter, r *http.Request) {
	if _, err := valueobjects.NewConcern(sharedAPI.GetPathParam(r, "concern")); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithConcern(w, r)
}

// AssignSteward godoc
// @Summary Assign the steward of a concern in a business domain
// @Description Makes the given active user the steward of the concern, replacing any current steward. Assigning the current steward again changes nothing.
// @Tags stewardships
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param domainId path string true "Business domain ID"
// @Param concern path string true "Concern (ownership, assessment, documentation, planning, structure)"
// @Param body body AssignStewardRequest true "Steward"
// @Success 200 {object} readmodels.ConcernStewardshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /stewardships/{domainId}/{concern} [put]
func (h *StewardshipHandlers) AssignSteward(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[AssignStewardRequest](w, r)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.AssignSteward{
		DomainID:   sharedAPI.GetPathParam(r, "domainId"),
		Concern:    sharedAPI.GetPathParam(r, "concern"),
		StewardID:  req.StewardID,
		AssignedBy: actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithConcern(w, r)
}

// ReleaseSteward godoc
// @Summary Release the steward of a concern in a business domain
// @Description Ends the stewardship of the concern; it reads as unassigned. Releasing an unassigned concern changes nothing.
// @Tags stewardships
// @Security CookieAuth
// @Param domainId path string true "Business domain ID"
// @Param concern path string true "Concern (ownership, assessment, documentation, planning, structure)"
// @Success 204 "No Content"
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /stewardships/{domainId}/{concern} [delete]
func (h *StewardshipHandlers) ReleaseSteward(w http.ResponseWriter, r *http.Request) {
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.ReleaseSteward{
		DomainID:   sharedAPI.GetPathParam(r, "domainId"),
		Concern:    sharedAPI.GetPathParam(r, "concern"),
		ReleasedBy: actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	sharedAPI.RespondNoContent(w)
}

func (h *StewardshipHandlers) loadDomainStewardships(w http.ResponseWriter, r *http.Request, domainID string) (*readmodels.DomainStewardshipsDTO, bool) {
	result, err := h.queries.DomainStewardships(r.Context(), domainID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return nil, false
	}
	if result == nil {
		sharedAPI.HandleError(w, handlers.ErrDomainNotFound)
		return nil, false
	}
	return result, true
}

func (h *StewardshipHandlers) respondWithConcern(w http.ResponseWriter, r *http.Request) {
	domainID := sharedAPI.GetPathParam(r, "domainId")
	concern := sharedAPI.GetPathParam(r, "concern")
	result, ok := h.loadDomainStewardships(w, r, domainID)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	for _, item := range result.Concerns {
		if item.Concern == concern {
			item.Links = h.links.ItemLinks(domainID, item, actor)
			sharedAPI.RespondJSON(w, http.StatusOK, item)
			return
		}
	}
	sharedAPI.HandleError(w, valueobjects.ErrUnknownConcern)
}
