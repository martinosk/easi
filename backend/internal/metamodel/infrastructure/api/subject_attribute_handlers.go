package api

import (
	"context"
	"errors"
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/metamodel/application/commands"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type SchemaReader interface {
	GetBySubjectType(ctx context.Context, subjectType string) (*readmodels.SubjectAttributeSchemaRecord, error)
}

type SubjectAttributeHandlers struct {
	commandBus      cqrs.CommandBus
	reader          SchemaReader
	links           *MetaModelLinks
	sessionProvider authPL.SessionProvider
}

func NewSubjectAttributeHandlers(
	commandBus cqrs.CommandBus,
	reader SchemaReader,
	links *MetaModelLinks,
	sessionProvider authPL.SessionProvider,
) *SubjectAttributeHandlers {
	return &SubjectAttributeHandlers{commandBus: commandBus, reader: reader, links: links, sessionProvider: sessionProvider}
}

// GetSchema godoc
// @Summary Get the attribute schema of a subject type
// @Description Retrieves the tenant's custom attribute schema for the given subject type, lazily creating an empty schema on first read. The schema is the MetaModel-owned vocabulary of custom attributes; how a one-pager displays or requires them is OnePagers configuration.
// @Tags meta-model
// @Produce json
// @Param subjectType path string true "Subject type" Enums(capability, application, acquired-entity, vendor, internal-team)
// @Success 200 {object} SubjectAttributeSchemaDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes [get]
func (h *SubjectAttributeHandlers) GetSchema(w http.ResponseWriter, r *http.Request) {
	subjectType, ok := h.resolveSubjectType(w, r)
	if !ok {
		return
	}
	record, ok := h.ensureSchema(w, r, subjectType)
	if !ok {
		return
	}
	h.respondWithRecord(w, r, record, http.StatusOK)
}

func (h *SubjectAttributeHandlers) resolveSubjectType(w http.ResponseWriter, r *http.Request) (valueobjects.SubjectType, bool) {
	subjectType, err := valueobjects.NewSubjectType(sharedAPI.GetPathParam(r, "subjectType"))
	if err != nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Unknown subject type")
		return valueobjects.SubjectType{}, false
	}
	return subjectType, true
}

func (h *SubjectAttributeHandlers) ensureSchema(w http.ResponseWriter, r *http.Request, subjectType valueobjects.SubjectType) (*readmodels.SubjectAttributeSchemaRecord, bool) {
	record, err := h.reader.GetBySubjectType(r.Context(), subjectType.Value())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve attribute schema")
		return nil, false
	}
	if record != nil {
		return record, true
	}
	return h.createSchema(w, r, subjectType)
}

func (h *SubjectAttributeHandlers) createSchema(w http.ResponseWriter, r *http.Request, subjectType valueobjects.SubjectType) (*readmodels.SubjectAttributeSchemaRecord, bool) {
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
	createCmd := &commands.CreateSubjectAttributeSchema{TenantID: tenantID.Value(), SubjectType: subjectType.Value(), CreatedBy: email}
	if _, err := h.commandBus.Dispatch(r.Context(), createCmd); err != nil && !errors.Is(err, handlers.ErrSchemaAlreadyExists) {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create attribute schema")
		return nil, false
	}
	record, err := h.reader.GetBySubjectType(r.Context(), subjectType.Value())
	if err != nil || record == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created attribute schema")
		return nil, false
	}
	return record, true
}

func (h *SubjectAttributeHandlers) respondWithRecord(w http.ResponseWriter, r *http.Request, record *readmodels.SubjectAttributeSchemaRecord, status int) {
	actor, _ := sharedctx.GetActor(r.Context())
	dto := BuildSubjectAttributeSchemaDTO(record, h.links, actor)
	if status == http.StatusCreated {
		w.Header().Set("Location", h.links.Base()+subjectAttributesPath(record.SubjectType))
	}
	sharedAPI.RespondJSON(w, status, dto)
}

type schemaWriteContext struct {
	subjectType valueobjects.SubjectType
	schemaID    string
	email       string
	attributeID string
	optionID    string
}

func (h *SubjectAttributeHandlers) prepareWrite(w http.ResponseWriter, r *http.Request, expectedVersion int) (schemaWriteContext, bool) {
	email, err := h.sessionProvider.GetCurrentUserEmail(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return schemaWriteContext{}, false
	}
	subjectType, ok := h.resolveSubjectType(w, r)
	if !ok {
		return schemaWriteContext{}, false
	}
	record, ok := h.ensureSchema(w, r, subjectType)
	if !ok {
		return schemaWriteContext{}, false
	}
	if expectedVersion != record.Version {
		sharedAPI.RespondError(w, http.StatusConflict, nil, "Attribute schema was modified by another user. Please refresh and try again.")
		return schemaWriteContext{}, false
	}
	return schemaWriteContext{
		subjectType: subjectType,
		schemaID:    record.ID,
		email:       email,
		attributeID: sharedAPI.GetPathParam(r, "attributeID"),
		optionID:    sharedAPI.GetPathParam(r, "optionID"),
	}, true
}
