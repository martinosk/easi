package api

import (
	"context"
	"errors"
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/handlers"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ConfigurationReader interface {
	GetBySubjectType(ctx context.Context, subjectType string) (*readmodels.ConfigurationRecord, error)
}

type OnePagerConfigurationHandlers struct {
	commandBus      cqrs.CommandBus
	reader          ConfigurationReader
	links           *OnePagerLinks
	sessionProvider authPL.SessionProvider
}

func NewOnePagerConfigurationHandlers(
	commandBus cqrs.CommandBus,
	reader ConfigurationReader,
	links *OnePagerLinks,
	sessionProvider authPL.SessionProvider,
) *OnePagerConfigurationHandlers {
	return &OnePagerConfigurationHandlers{
		commandBus:      commandBus,
		reader:          reader,
		links:           links,
		sessionProvider: sessionProvider,
	}
}

// GetConfiguration godoc
// @Summary Get the one-pager configuration for a subject type
// @Description Retrieves the tenant's one-pager configuration for the given subject type, lazily creating the default configuration (all catalog built-in fields in catalog order, no custom fields) on first read.
// @Tags one-pagers
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, enterprise-capability, application, acquired-entity, vendor, internal-team)
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType} [get]
func (h *OnePagerConfigurationHandlers) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	subjectType, ok := h.resolveSubjectType(w, r)
	if !ok {
		return
	}

	record, ok := h.ensureConfiguration(w, r, subjectType)
	if !ok {
		return
	}

	h.respondWithRecord(w, r, record, http.StatusOK)
}

func (h *OnePagerConfigurationHandlers) resolveSubjectType(w http.ResponseWriter, r *http.Request) (valueobjects.SubjectType, bool) {
	subjectType, err := valueobjects.NewSubjectType(sharedAPI.GetPathParam(r, "subjectType"))
	if err != nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Unknown subject type")
		return valueobjects.SubjectType{}, false
	}
	return subjectType, true
}

func (h *OnePagerConfigurationHandlers) ensureConfiguration(
	w http.ResponseWriter,
	r *http.Request,
	subjectType valueobjects.SubjectType,
) (*readmodels.ConfigurationRecord, bool) {
	record, err := h.reader.GetBySubjectType(r.Context(), subjectType.Value())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve configuration")
		return nil, false
	}
	if record != nil {
		return record, true
	}
	return h.createDefaultConfiguration(w, r, subjectType)
}

func (h *OnePagerConfigurationHandlers) createDefaultConfiguration(
	w http.ResponseWriter,
	r *http.Request,
	subjectType valueobjects.SubjectType,
) (*readmodels.ConfigurationRecord, bool) {
	email, err := h.sessionProvider.GetCurrentUserEmail(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return nil, false
	}
	tenantID, err := sharedctx.GetTenant(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to get tenant")
		return nil, false
	}

	createCmd := &commands.CreateOnePagerConfiguration{
		TenantID:    tenantID.Value(),
		SubjectType: subjectType.Value(),
		CreatedBy:   email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), createCmd); err != nil && !errors.Is(err, handlers.ErrConfigurationAlreadyExists) {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create configuration")
		return nil, false
	}

	record, err := h.reader.GetBySubjectType(r.Context(), subjectType.Value())
	if err != nil || record == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created configuration")
		return nil, false
	}
	return record, true
}

func (h *OnePagerConfigurationHandlers) respondWithRecord(w http.ResponseWriter, r *http.Request, record *readmodels.ConfigurationRecord, status int) {
	actor, _ := sharedctx.GetActor(r.Context())
	dto := BuildConfigurationDTO(record, h.links, actor)
	if status == http.StatusCreated {
		w.Header().Set("Location", h.links.Base()+configurationPath(record.SubjectType))
	}
	sharedAPI.RespondJSON(w, status, dto)
}

type writeContext struct {
	subjectType valueobjects.SubjectType
	configID    string
	email       string
	fieldID     string
	entryID     string
	optionID    string
}

func (h *OnePagerConfigurationHandlers) prepareWrite(w http.ResponseWriter, r *http.Request, expectedVersion int) (writeContext, bool) {
	email, err := h.sessionProvider.GetCurrentUserEmail(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return writeContext{}, false
	}

	subjectType, ok := h.resolveSubjectType(w, r)
	if !ok {
		return writeContext{}, false
	}

	record, ok := h.ensureConfiguration(w, r, subjectType)
	if !ok {
		return writeContext{}, false
	}

	if expectedVersion != record.Version {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "Configuration was modified by another user. Please refresh and try again.")
		return writeContext{}, false
	}

	return writeContext{
		subjectType: subjectType,
		configID:    record.ID,
		email:       email,
		fieldID:     sharedAPI.GetPathParam(r, "fieldID"),
		entryID:     sharedAPI.GetPathParam(r, "entryID"),
		optionID:    sharedAPI.GetPathParam(r, "optionID"),
	}, true
}
