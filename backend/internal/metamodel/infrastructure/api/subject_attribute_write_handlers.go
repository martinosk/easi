package api

import (
	"net/http"

	"easi/backend/internal/metamodel/application/commands"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type DefineSubjectAttributeRequest struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	HelpText string   `json:"helpText"`
	Options  []string `json:"options"`
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	Version  int      `json:"version"`
}

type RenameSubjectAttributeRequest struct {
	Name     string `json:"name"`
	HelpText string `json:"helpText"`
	Type     string `json:"type,omitempty"`
	Version  int    `json:"version"`
}

type SchemaVersionRequest struct {
	Version int `json:"version"`
}

type AddAttributeOptionRequest struct {
	Label   string `json:"label"`
	Version int    `json:"version"`
}

type SetAttributeBoundsRequest struct {
	Min     *float64 `json:"min,omitempty"`
	Max     *float64 `json:"max,omitempty"`
	Version int      `json:"version"`
}

func (r DefineSubjectAttributeRequest) expectedVersion() int { return r.Version }
func (r RenameSubjectAttributeRequest) expectedVersion() int { return r.Version }
func (r SchemaVersionRequest) expectedVersion() int          { return r.Version }
func (r AddAttributeOptionRequest) expectedVersion() int     { return r.Version }
func (r SetAttributeBoundsRequest) expectedVersion() int     { return r.Version }

type versionedSchemaRequest interface {
	expectedVersion() int
}

type schemaWrite[R versionedSchemaRequest] struct {
	successStatus  int
	failureMessage string
	command        func(req R, wc schemaWriteContext) cqrs.Command
}

func handleSchemaWrite[R versionedSchemaRequest](h *SubjectAttributeHandlers, w http.ResponseWriter, r *http.Request, op schemaWrite[R]) {
	req, ok := sharedAPI.DecodeRequestOrFail[R](w, r)
	if !ok {
		return
	}
	wc, ok := h.prepareWrite(w, r, req.expectedVersion())
	if !ok {
		return
	}
	if _, err := h.commandBus.Dispatch(r.Context(), op.command(req, wc)); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, op.failureMessage+": "+err.Error())
		return
	}
	record, err := h.reader.GetBySubjectType(r.Context(), wc.subjectType.Value())
	if err != nil || record == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated attribute schema")
		return
	}
	h.respondWithRecord(w, r, record, op.successStatus)
}

var defineAttributeWrite = schemaWrite[DefineSubjectAttributeRequest]{
	successStatus:  http.StatusCreated,
	failureMessage: "Failed to define attribute",
	command: func(req DefineSubjectAttributeRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.DefineSubjectAttribute{
			SchemaID:      wc.schemaID,
			Name:          req.Name,
			AttributeType: req.Type,
			HelpText:      req.HelpText,
			OptionLabels:  req.Options,
			Min:           req.Min,
			Max:           req.Max,
			ModifiedBy:    wc.email,
		}
	},
}

// DefineAttribute godoc
// @Summary Define a custom attribute
// @Description Adds a typed custom attribute to the subject type's schema. Selection attributes require at least one option; number attributes may carry bounds. The new attribute is published to consuming contexts, which decide how to present it.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attribute body DefineSubjectAttributeRequest true "Attribute definition"
// @Success 201 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes [post]
func (h *SubjectAttributeHandlers) DefineAttribute(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, defineAttributeWrite)
}

var renameAttributeWrite = schemaWrite[RenameSubjectAttributeRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to rename attribute",
	command: func(req RenameSubjectAttributeRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.RenameSubjectAttribute{
			SchemaID:      wc.schemaID,
			AttributeID:   wc.attributeID,
			Name:          req.Name,
			HelpText:      req.HelpText,
			RequestedType: req.Type,
			ModifiedBy:    wc.email,
		}
	},
}

// RenameAttribute godoc
// @Summary Rename a custom attribute
// @Description Updates the name and help text of an active attribute. The attribute keeps its ID and type; supplying a different type is rejected because attribute types are immutable.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attributeID path string true "Attribute ID"
// @Param attribute body RenameSubjectAttributeRequest true "New name and help text"
// @Success 200 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes/{attributeID} [put]
func (h *SubjectAttributeHandlers) RenameAttribute(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, renameAttributeWrite)
}

var retireAttributeWrite = schemaWrite[SchemaVersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to retire attribute",
	command: func(_ SchemaVersionRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.RetireSubjectAttribute{SchemaID: wc.schemaID, AttributeID: wc.attributeID, ModifiedBy: wc.email}
	},
}

// RetireAttribute godoc
// @Summary Retire a custom attribute
// @Description Retires an active attribute. Consuming contexts stop offering it for editing while previously recorded facts stay readable. A retired attribute can be reactivated.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attributeID path string true "Attribute ID"
// @Param version body SchemaVersionRequest true "Expected schema version"
// @Success 200 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes/{attributeID}/retire [post]
func (h *SubjectAttributeHandlers) RetireAttribute(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, retireAttributeWrite)
}

var reactivateAttributeWrite = schemaWrite[SchemaVersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to reactivate attribute",
	command: func(_ SchemaVersionRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.ReactivateSubjectAttribute{SchemaID: wc.schemaID, AttributeID: wc.attributeID, ModifiedBy: wc.email}
	},
}

// ReactivateAttribute godoc
// @Summary Reactivate a retired custom attribute
// @Description Reactivates a retired attribute with its original ID, type, options and bounds.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attributeID path string true "Attribute ID"
// @Param version body SchemaVersionRequest true "Expected schema version"
// @Success 200 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes/{attributeID}/reactivate [post]
func (h *SubjectAttributeHandlers) ReactivateAttribute(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, reactivateAttributeWrite)
}

var addOptionWrite = schemaWrite[AddAttributeOptionRequest]{
	successStatus:  http.StatusCreated,
	failureMessage: "Failed to add attribute option",
	command: func(req AddAttributeOptionRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.AddSubjectAttributeOption{SchemaID: wc.schemaID, AttributeID: wc.attributeID, Label: req.Label, ModifiedBy: wc.email}
	},
}

// AddOption godoc
// @Summary Add an option to a selection attribute
// @Description Adds a new active option to an active selection attribute.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attributeID path string true "Attribute ID"
// @Param option body AddAttributeOptionRequest true "Option label"
// @Success 201 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes/{attributeID}/options [post]
func (h *SubjectAttributeHandlers) AddOption(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, addOptionWrite)
}

var retireOptionWrite = schemaWrite[SchemaVersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to retire attribute option",
	command: func(_ SchemaVersionRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.RetireSubjectAttributeOption{SchemaID: wc.schemaID, AttributeID: wc.attributeID, OptionID: wc.optionID, ModifiedBy: wc.email}
	},
}

// RetireOption godoc
// @Summary Retire a selection attribute option
// @Description Retires an active option of a selection attribute. Retired options remain on the definition so recorded facts stay readable. The last active option cannot be retired.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attributeID path string true "Attribute ID"
// @Param optionID path string true "Option ID"
// @Param version body SchemaVersionRequest true "Expected schema version"
// @Success 200 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes/{attributeID}/options/{optionID}/retire [post]
func (h *SubjectAttributeHandlers) RetireOption(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, retireOptionWrite)
}

var setBoundsWrite = schemaWrite[SetAttributeBoundsRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to set attribute bounds",
	command: func(req SetAttributeBoundsRequest, wc schemaWriteContext) cqrs.Command {
		return &commands.SetSubjectAttributeBounds{SchemaID: wc.schemaID, AttributeID: wc.attributeID, Min: req.Min, Max: req.Max, ModifiedBy: wc.email}
	},
}

// SetBounds godoc
// @Summary Set a number attribute's bounds
// @Description Sets, tightens, loosens or clears the minimum and maximum bounds of an active number attribute. Bounds gate new facts only; recorded values are never altered. Setting bounds on a non-number attribute is rejected.
// @Tags meta-model
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param attributeID path string true "Attribute ID"
// @Param bounds body SetAttributeBoundsRequest true "New bounds"
// @Success 200 {object} SubjectAttributeSchemaDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /meta-model/subject-types/{subjectType}/attributes/{attributeID}/bounds [put]
func (h *SubjectAttributeHandlers) SetBounds(w http.ResponseWriter, r *http.Request) {
	handleSchemaWrite(h, w, r, setBoundsWrite)
}
