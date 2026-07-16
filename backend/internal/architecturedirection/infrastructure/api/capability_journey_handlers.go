package api

import (
	"context"
	"errors"
	"net/http"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

var errCapabilityJourneyMissingAfterMutation = errors.New("capability journey not found after mutation")

type CapabilityJourneyQueries interface {
	GetActiveByCapabilityID(ctx context.Context, capabilityID string) (*readmodels.CapabilityJourneyDTO, error)
	GetHistoryByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.CapabilityJourneyDTO, error)
	GetCurrentByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]readmodels.CapabilityJourneyDTO, error)
	GetAllCurrent(ctx context.Context) ([]readmodels.CapabilityJourneyDTO, error)
	GetByID(ctx context.Context, journeyID string) (*readmodels.CapabilityJourneyDTO, error)
}

type CapabilityJourneyHandlers struct {
	commandBus cqrs.CommandBus
	queries    CapabilityJourneyQueries
	hateoas    *CapabilityJourneyLinks
}

func NewCapabilityJourneyHandlers(commandBus cqrs.CommandBus, queries CapabilityJourneyQueries, hateoas *CapabilityJourneyLinks) *CapabilityJourneyHandlers {
	return &CapabilityJourneyHandlers{commandBus: commandBus, queries: queries, hateoas: hateoas}
}

type CapabilityJourneyResponse struct {
	Journey *readmodels.CapabilityJourneyDTO `json:"journey"`
	Links   sharedAPI.Links                  `json:"_links,omitempty"`
}

type TargetPeriodRequest struct {
	Year    int `json:"year"`
	Quarter int `json:"quarter"`
}

type CaptureJourneyRequest struct {
	Kind             string               `json:"kind"`
	FromComponentIDs []string             `json:"fromComponentIds,omitempty"`
	ToComponentID    string               `json:"toComponentId"`
	Note             string               `json:"note,omitempty"`
	TargetPeriod     *TargetPeriodRequest `json:"targetPeriod,omitempty"`
	TargetDomainID   string               `json:"targetDomainId,omitempty"`
	TargetParentID   string               `json:"targetParentId,omitempty"`
	ResultingName    string               `json:"resultingName,omitempty"`
}

type UpdateJourneyProgressRequest struct {
	Progress int `json:"progress"`
}

type UpdateJourneyDetailsRequest struct {
	Note          string               `json:"note"`
	TargetPeriod  *TargetPeriodRequest `json:"targetPeriod"`
	ResultingName string               `json:"resultingName,omitempty"`
}

func (req UpdateJourneyDetailsRequest) targetPeriod() *TargetPeriodRequest { return req.TargetPeriod }

type ChangeJourneySourceApplicationsRequest struct {
	ComponentIDs []string `json:"componentIds"`
}

type AddJourneyMilestoneRequest struct {
	Label        string               `json:"label"`
	TargetPeriod *TargetPeriodRequest `json:"targetPeriod,omitempty"`
	Status       string               `json:"status,omitempty"`
}

func (req AddJourneyMilestoneRequest) targetPeriod() *TargetPeriodRequest { return req.TargetPeriod }

type UpdateJourneyMilestoneRequest struct {
	Label        string               `json:"label"`
	TargetPeriod *TargetPeriodRequest `json:"targetPeriod,omitempty"`
	Status       string               `json:"status"`
}

func (req UpdateJourneyMilestoneRequest) targetPeriod() *TargetPeriodRequest { return req.TargetPeriod }

// GetJourneyForCapability godoc
// @Summary Get the active journey for a capability
// @Description Returns the capability's active (planned or in-flight) journey, or null if none.
// @Tags capability-journeys
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Success 200 {object} CapabilityJourneyResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/journey [get]
func (h *CapabilityJourneyHandlers) GetJourneyForCapability(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	journey, ok := fetchOrFail(w, r, func(ctx context.Context) (*readmodels.CapabilityJourneyDTO, error) {
		return h.queries.GetActiveByCapabilityID(ctx, capabilityID)
	})
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	h.decorateJourney(journey, actor)
	envelope := CapabilityJourneyResponse{Journey: journey, Links: h.hateoas.ForCapability(capabilityID, journey, actor)}
	sharedAPI.RespondJSON(w, http.StatusOK, envelope)
}

// CaptureJourney godoc
// @Summary Capture a journey on a capability
// @Description Creates a new journey in status "planned"; rejected if an active journey already exists on the capability.
// @Tags capability-journeys
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param body body CaptureJourneyRequest true "Journey data"
// @Success 201 {object} readmodels.CapabilityJourneyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/journey [post]
func (h *CapabilityJourneyHandlers) CaptureJourney(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[CaptureJourneyRequest](w, r)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	year, quarter := requestTargetPeriodParts(req.TargetPeriod)
	cmd := &commands.PlanJourney{
		CapabilityID:     capabilityID,
		Kind:             req.Kind,
		FromComponentIDs: req.FromComponentIDs,
		ToComponentID:    req.ToComponentID,
		Note:             req.Note,
		TargetYear:       year,
		TargetQuarter:    quarter,
		TargetDomainID:   req.TargetDomainID,
		TargetParentID:   req.TargetParentID,
		ResultingName:    req.ResultingName,
		PlannedBy:        actor.Email,
	}
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		h.respondCaptureError(w, err)
		return
	}
	location := sharedAPI.BuildSubResourceLink(capabilitiesPath, sharedAPI.ResourceID(capabilityID), journeySubPath)
	h.respondWithJourneyByID(w, r, respondJourneyParams{journeyID: result.CreatedID, statusCode: http.StatusCreated, location: location})
}

func (h *CapabilityJourneyHandlers) respondCaptureError(w http.ResponseWriter, err error) {
	var conflict *services.ActiveJourneyError
	if errors.As(err, &conflict) {
		sharedAPI.RespondErrorWithLinks(w, sharedAPI.ErrorWithLinksParams{
			StatusCode: http.StatusConflict,
			Message:    conflict.Error(),
			Details:    map[string]string{"existingJourneyId": conflict.ExistingJourneyID},
			Links:      sharedAPI.Links{"x-existing-journey": h.hateoas.Get(journeyItemResourcePath(conflict.ExistingJourneyID))},
		})
		return
	}
	sharedAPI.HandleError(w, err)
}

// GetJourneyHistory godoc
// @Summary Get the full journey history for a capability
// @Description Returns every journey ever captured on the capability, newest first.
// @Tags capability-journeys
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Success 200 {object} sharedAPI.CollectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/journey/history [get]
func (h *CapabilityJourneyHandlers) GetJourneyHistory(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	h.respondJourneyCollection(w, r,
		func(ctx context.Context) ([]readmodels.CapabilityJourneyDTO, error) {
			return h.queries.GetHistoryByCapabilityID(ctx, capabilityID)
		},
		func(sharedctx.Actor) sharedAPI.Links {
			return sharedAPI.Links{"self": h.hateoas.Get(journeyHistoryResourcePath(capabilityID))}
		},
	)
}

// GetCapabilityJourneys godoc
// @Summary Bulk-fetch current journeys
// @Description Returns, for each capability, its active journey plus the most recent terminal journey. Without capabilityIds the whole collection is returned.
// @Tags capability-journeys
// @Produce json
// @Security CookieAuth
// @Param capabilityIds query string false "Comma-separated capability IDs; omit to return the whole collection"
// @Success 200 {object} sharedAPI.CollectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys [get]
func (h *CapabilityJourneyHandlers) GetCapabilityJourneys(w http.ResponseWriter, r *http.Request) {
	h.respondJourneyCollection(w, r,
		func(ctx context.Context) ([]readmodels.CapabilityJourneyDTO, error) {
			if capabilityIDs, filtered := capabilityIDFilter(r); filtered {
				return h.queries.GetCurrentByCapabilityIDs(ctx, capabilityIDs)
			}
			return h.queries.GetAllCurrent(ctx)
		},
		func(actor sharedctx.Actor) sharedAPI.Links {
			return h.hateoas.BulkLinks(string(capabilityJourneysPath), actor)
		},
	)
}

func (h *CapabilityJourneyHandlers) respondJourneyCollection(
	w http.ResponseWriter,
	r *http.Request,
	fetch func(ctx context.Context) ([]readmodels.CapabilityJourneyDTO, error),
	buildLinks func(actor sharedctx.Actor) sharedAPI.Links,
) {
	journeys, ok := fetchOrFail(w, r, fetch)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	h.decorateJourneys(journeys, actor)
	sharedAPI.RespondCollection(w, http.StatusOK, journeys, buildLinks(actor))
}

func fetchOrFail[T any](w http.ResponseWriter, r *http.Request, fetch func(ctx context.Context) (T, error)) (T, bool) {
	result, err := fetch(r.Context())
	if err != nil {
		sharedAPI.HandleError(w, err)
		var zero T
		return zero, false
	}
	return result, true
}

// StartJourney godoc
// @Summary Start a planned journey
// @Tags capability-journeys
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/start [post]
func (h *CapabilityJourneyHandlers) StartJourney(w http.ResponseWriter, r *http.Request) {
	h.dispatchTransition(w, r, func(journeyID string, actor sharedctx.Actor) cqrs.Command {
		return &commands.StartJourney{JourneyID: journeyID, Actor: actor.Email}
	})
}

// CompleteJourney godoc
// @Summary Complete an in-flight journey
// @Description Freezes the journey in status "done". Plan-only: no other context is mutated.
// @Tags capability-journeys
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/complete [post]
func (h *CapabilityJourneyHandlers) CompleteJourney(w http.ResponseWriter, r *http.Request) {
	h.dispatchTransition(w, r, func(journeyID string, actor sharedctx.Actor) cqrs.Command {
		return &commands.CompleteJourney{JourneyID: journeyID, Actor: actor.Email}
	})
}

// AbandonJourney godoc
// @Summary Abandon a planned or in-flight journey
// @Tags capability-journeys
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/abandon [post]
func (h *CapabilityJourneyHandlers) AbandonJourney(w http.ResponseWriter, r *http.Request) {
	h.dispatchTransition(w, r, func(journeyID string, actor sharedctx.Actor) cqrs.Command {
		return &commands.AbandonJourney{JourneyID: journeyID, Actor: actor.Email}
	})
}

func (h *CapabilityJourneyHandlers) dispatchTransition(w http.ResponseWriter, r *http.Request, buildCmd func(journeyID string, actor sharedctx.Actor) cqrs.Command) {
	journeyID := sharedAPI.GetPathParam(r, "journeyId")
	actor, _ := sharedctx.GetActor(r.Context())
	h.dispatchAndRespond(w, r, journeyMutation{journeyID: journeyID, cmd: buildCmd(journeyID, actor), statusCode: http.StatusOK})
}

type journeyMutation struct {
	journeyID  string
	cmd        cqrs.Command
	statusCode int
}

func (h *CapabilityJourneyHandlers) dispatchAndRespond(w http.ResponseWriter, r *http.Request, m journeyMutation) {
	if _, err := h.commandBus.Dispatch(r.Context(), m.cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithJourneyByID(w, r, respondJourneyParams{journeyID: m.journeyID, statusCode: m.statusCode})
}

func decodeJourneyMutation[T any](h *CapabilityJourneyHandlers, w http.ResponseWriter, r *http.Request, statusCode int, buildCmd func(journeyID string, req T, actor sharedctx.Actor) cqrs.Command) {
	journeyID := sharedAPI.GetPathParam(r, "journeyId")
	req, ok := sharedAPI.DecodeRequestOrFail[T](w, r)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	h.dispatchAndRespond(w, r, journeyMutation{journeyID: journeyID, cmd: buildCmd(journeyID, req, actor), statusCode: statusCode})
}

type hasTargetPeriod interface {
	targetPeriod() *TargetPeriodRequest
}

func decodeJourneyMutationWithPeriod[T hasTargetPeriod](
	h *CapabilityJourneyHandlers,
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	buildCmd func(journeyID string, req T, year, quarter *int, actor sharedctx.Actor) cqrs.Command,
) {
	decodeJourneyMutation(h, w, r, statusCode, func(journeyID string, req T, actor sharedctx.Actor) cqrs.Command {
		year, quarter := requestTargetPeriodParts(req.targetPeriod())
		return buildCmd(journeyID, req, year, quarter, actor)
	})
}

// PutJourneyProgress godoc
// @Summary Update a journey's progress
// @Tags capability-journeys
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Param body body UpdateJourneyProgressRequest true "Progress"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/progress [put]
func (h *CapabilityJourneyHandlers) PutJourneyProgress(w http.ResponseWriter, r *http.Request) {
	decodeJourneyMutation(h, w, r, http.StatusOK, func(journeyID string, req UpdateJourneyProgressRequest, actor sharedctx.Actor) cqrs.Command {
		return &commands.UpdateJourneyProgress{JourneyID: journeyID, Progress: req.Progress, Actor: actor.Email}
	})
}

// PutJourneyDetails godoc
// @Summary Replace a journey's note, target period, and resulting name
// @Tags capability-journeys
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Param body body UpdateJourneyDetailsRequest true "Journey details"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/details [put]
func (h *CapabilityJourneyHandlers) PutJourneyDetails(w http.ResponseWriter, r *http.Request) {
	decodeJourneyMutationWithPeriod(h, w, r, http.StatusOK, func(journeyID string, req UpdateJourneyDetailsRequest, year, quarter *int, actor sharedctx.Actor) cqrs.Command {
		return &commands.UpdateJourneyDetails{
			JourneyID:     journeyID,
			Note:          req.Note,
			TargetYear:    year,
			TargetQuarter: quarter,
			ResultingName: req.ResultingName,
			Actor:         actor.Email,
		}
	})
}

// PutJourneySourceApplications godoc
// @Summary Replace a journey's source applications
// @Description Re-validates kind cardinality, that the target is not among the sources, and that every source references an existing application component.
// @Tags capability-journeys
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Param body body ChangeJourneySourceApplicationsRequest true "Source component IDs"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/source-applications [put]
func (h *CapabilityJourneyHandlers) PutJourneySourceApplications(w http.ResponseWriter, r *http.Request) {
	decodeJourneyMutation(h, w, r, http.StatusOK, func(journeyID string, req ChangeJourneySourceApplicationsRequest, actor sharedctx.Actor) cqrs.Command {
		return &commands.ChangeJourneySourceApplications{JourneyID: journeyID, FromComponentIDs: req.ComponentIDs, Actor: actor.Email}
	})
}

// PostJourneyMilestone godoc
// @Summary Add a milestone to an active journey
// @Tags capability-journeys
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Param body body AddJourneyMilestoneRequest true "Milestone data"
// @Success 201 {object} readmodels.CapabilityJourneyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/milestones [post]
func (h *CapabilityJourneyHandlers) PostJourneyMilestone(w http.ResponseWriter, r *http.Request) {
	decodeJourneyMutationWithPeriod(h, w, r, http.StatusCreated, func(journeyID string, req AddJourneyMilestoneRequest, year, quarter *int, actor sharedctx.Actor) cqrs.Command {
		return &commands.AddJourneyMilestone{
			JourneyID:     journeyID,
			Label:         req.Label,
			TargetYear:    year,
			TargetQuarter: quarter,
			Status:        defaultMilestoneStatus(req.Status),
			Actor:         actor.Email,
		}
	})
}

func defaultMilestoneStatus(status string) string {
	if status == "" {
		return valueobjects.MilestoneStatusPlanned
	}
	return status
}

// PutJourneyMilestone godoc
// @Summary Update a journey milestone
// @Tags capability-journeys
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Param milestoneId path string true "Milestone ID"
// @Param body body UpdateJourneyMilestoneRequest true "Milestone data"
// @Success 200 {object} readmodels.CapabilityJourneyDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/milestones/{milestoneId} [put]
func (h *CapabilityJourneyHandlers) PutJourneyMilestone(w http.ResponseWriter, r *http.Request) {
	decodeJourneyMutationWithPeriod(h, w, r, http.StatusOK, func(journeyID string, req UpdateJourneyMilestoneRequest, year, quarter *int, actor sharedctx.Actor) cqrs.Command {
		return &commands.UpdateJourneyMilestone{
			JourneyID:     journeyID,
			MilestoneID:   sharedAPI.GetPathParam(r, "milestoneId"),
			Label:         req.Label,
			TargetYear:    year,
			TargetQuarter: quarter,
			Status:        req.Status,
			Actor:         actor.Email,
		}
	})
}

// DeleteJourneyMilestone godoc
// @Summary Remove a journey milestone
// @Tags capability-journeys
// @Security CookieAuth
// @Param journeyId path string true "Journey ID"
// @Param milestoneId path string true "Milestone ID"
// @Success 204 "No Content"
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capability-journeys/{journeyId}/milestones/{milestoneId} [delete]
func (h *CapabilityJourneyHandlers) DeleteJourneyMilestone(w http.ResponseWriter, r *http.Request) {
	journeyID := sharedAPI.GetPathParam(r, "journeyId")
	milestoneID := sharedAPI.GetPathParam(r, "milestoneId")
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.RemoveJourneyMilestone{JourneyID: journeyID, MilestoneID: milestoneID, Actor: actor.Email}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	sharedAPI.RespondNoContent(w)
}

type respondJourneyParams struct {
	journeyID  string
	statusCode int
	location   string
}

func (h *CapabilityJourneyHandlers) respondWithJourneyByID(w http.ResponseWriter, r *http.Request, p respondJourneyParams) {
	journey, err := h.queries.GetByID(r.Context(), p.journeyID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if journey == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, errCapabilityJourneyMissingAfterMutation, "failed to load journey after mutation")
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	h.decorateJourney(journey, actor)
	if p.statusCode == http.StatusCreated && p.location != "" {
		sharedAPI.RespondCreated(w, p.location, journey)
		return
	}
	sharedAPI.RespondJSON(w, p.statusCode, journey)
}

func (h *CapabilityJourneyHandlers) decorateJourneys(journeys []readmodels.CapabilityJourneyDTO, actor sharedctx.Actor) {
	for i := range journeys {
		h.decorateJourney(&journeys[i], actor)
	}
}

func (h *CapabilityJourneyHandlers) decorateJourney(journey *readmodels.CapabilityJourneyDTO, actor sharedctx.Actor) {
	if journey == nil {
		return
	}
	journey.Links = h.hateoas.ItemLinks(journey, actor)
	for i := range journey.Milestones {
		journey.Milestones[i].Links = h.hateoas.MilestoneLinks(journey, journey.Milestones[i].ID, actor)
	}
}

func requestTargetPeriodParts(period *TargetPeriodRequest) (*int, *int) {
	if period == nil {
		return nil, nil
	}
	year, quarter := period.Year, period.Quarter
	return &year, &quarter
}
