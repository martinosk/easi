package api

import (
	"context"
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type FactsReader interface {
	GetForSubject(ctx context.Context, subject readmodels.SubjectKey) ([]readmodels.FactRecord, error)
}

type OnePagerFactsHandlers struct {
	deps OnePagerFactsHandlersDeps
}

type OnePagerFactsHandlersDeps struct {
	CommandBus      cqrs.CommandBus
	Facts           FactsReader
	Configs         ConfigurationReader
	Links           *OnePagerLinks
	SessionProvider authPL.SessionProvider
}

func NewOnePagerFactsHandlers(deps OnePagerFactsHandlersDeps) *OnePagerFactsHandlers {
	return &OnePagerFactsHandlers{deps: deps}
}

type RecordFieldValueRequest struct {
	Value ValueEnvelopeDTO `json:"value"`
}

// GetFacts godoc
// @Summary Get the one-pager facts of a subject
// @Description Retrieves all recorded field values for the subject as {type, version, value} envelopes. Values recorded against retired selection options are flagged.
// @Tags one-pagers
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Param subjectID path string true "Subject ID"
// @Success 200 {object} OnePagerFactsDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/{subjectType}/{subjectID}/facts [get]
func (h *OnePagerFactsHandlers) GetFacts(subjectType valueobjects.SubjectType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subjectID, ok := resolveSubjectID(w, r)
		if !ok {
			return
		}
		h.respondWithFacts(w, r, subjectType, subjectID)
	}
}

// RecordValue godoc
// @Summary Record a field value on a subject's one-pager
// @Description Records a typed value for one custom field as an idempotent replace. The value envelope must match the field's type; selection options must be active on the field definition. Recording a value equal to the current one appends no event.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Param subjectID path string true "Subject ID"
// @Param fieldID path string true "Field ID"
// @Param value body RecordFieldValueRequest true "Value envelope"
// @Success 200 {object} OnePagerFactsDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/{subjectType}/{subjectID}/facts/{fieldID} [put]
func (h *OnePagerFactsHandlers) RecordValue(subjectType valueobjects.SubjectType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := sharedAPI.DecodeRequestOrFail[RecordFieldValueRequest](w, r)
		if !ok {
			return
		}
		h.handleFactsWrite(w, r, subjectType, "Failed to record field value", func(base commands.FactsSubjectField) cqrs.Command {
			return &commands.RecordFieldValue{FactsSubjectField: base, Value: req.Value.toDomain()}
		})
	}
}

// ClearValue godoc
// @Summary Clear a field value on a subject's one-pager
// @Description Clears the recorded value of one custom field. Clearing a field that has no value is a no-op.
// @Tags one-pagers
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Param subjectID path string true "Subject ID"
// @Param fieldID path string true "Field ID"
// @Success 200 {object} OnePagerFactsDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/{subjectType}/{subjectID}/facts/{fieldID} [delete]
func (h *OnePagerFactsHandlers) ClearValue(subjectType valueobjects.SubjectType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.handleFactsWrite(w, r, subjectType, "Failed to clear field value", func(base commands.FactsSubjectField) cqrs.Command {
			return &commands.ClearFieldValue{FactsSubjectField: base}
		})
	}
}

func (h *OnePagerFactsHandlers) handleFactsWrite(
	w http.ResponseWriter,
	r *http.Request,
	subjectType valueobjects.SubjectType,
	failureMessage string,
	command func(base commands.FactsSubjectField) cqrs.Command,
) {
	base, ok := h.prepareFactsWrite(w, r, subjectType)
	if !ok {
		return
	}
	if _, err := h.deps.CommandBus.Dispatch(r.Context(), command(base)); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, failureMessage+": "+err.Error())
		return
	}
	h.respondWithFacts(w, r, subjectType, base.SubjectID)
}

func (h *OnePagerFactsHandlers) prepareFactsWrite(
	w http.ResponseWriter,
	r *http.Request,
	subjectType valueobjects.SubjectType,
) (commands.FactsSubjectField, bool) {
	email, err := h.deps.SessionProvider.GetCurrentUserEmail(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return commands.FactsSubjectField{}, false
	}
	subjectID, ok := resolveSubjectID(w, r)
	if !ok {
		return commands.FactsSubjectField{}, false
	}
	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get tenant")
		return commands.FactsSubjectField{}, false
	}
	return commands.FactsSubjectField{
		TenantID:    tenantID.Value(),
		SubjectType: subjectType.Value(),
		SubjectID:   subjectID,
		FieldID:     sharedAPI.GetPathParam(r, "fieldID"),
		ModifiedBy:  email,
	}, true
}

func resolveSubjectID(w http.ResponseWriter, r *http.Request) (string, bool) {
	subjectID := sharedAPI.GetPathParam(r, "subjectID")
	if subjectID == "" {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Unknown subject")
		return "", false
	}
	return subjectID, true
}

func (h *OnePagerFactsHandlers) respondWithFacts(
	w http.ResponseWriter,
	r *http.Request,
	subjectType valueobjects.SubjectType,
	subjectID string,
) {
	subject := readmodels.SubjectKey{SubjectType: subjectType.Value(), SubjectID: subjectID}
	records, err := h.deps.Facts.GetForSubject(r.Context(), subject)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve one-pager facts")
		return
	}
	config, err := h.deps.Configs.GetBySubjectType(r.Context(), subjectType.Value())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	dto := BuildFactsDTO(factsDTOParams{
		subjectType: subjectType.Value(),
		subjectID:   subjectID,
		records:     records,
		config:      config,
		links:       h.deps.Links,
		ctx:         factsLinkContext{subjectType: subjectType.Value(), subjectID: subjectID, actor: actor},
	})
	sharedAPI.RespondJSON(w, http.StatusOK, dto)
}
