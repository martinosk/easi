package api

import (
	"net/http"

	"easi/backend/internal/onepagers/application/commands"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type DefineCustomFieldRequest struct {
	Name      string   `json:"name"`
	FieldType string   `json:"fieldType"`
	Required  bool     `json:"required"`
	HelpText  string   `json:"helpText"`
	Options   []string `json:"options"`
	Version   int      `json:"version"`
}

type RenameCustomFieldRequest struct {
	Name      string `json:"name"`
	HelpText  string `json:"helpText"`
	FieldType string `json:"fieldType,omitempty"`
	Version   int    `json:"version"`
}

type ChangeRequirementRequest struct {
	Required bool `json:"required"`
	Version  int  `json:"version"`
}

type VersionRequest struct {
	Version int `json:"version"`
}

type ReorderFieldsRequest struct {
	Order   []FieldRefDTO `json:"order"`
	Version int           `json:"version"`
}

type AddSelectionOptionRequest struct {
	Label   string `json:"label"`
	Version int    `json:"version"`
}

type SetNumberFieldBoundsRequest struct {
	Min     *float64 `json:"min,omitempty"`
	Max     *float64 `json:"max,omitempty"`
	Version int      `json:"version"`
}

func (r DefineCustomFieldRequest) expectedVersion() int    { return r.Version }
func (r RenameCustomFieldRequest) expectedVersion() int    { return r.Version }
func (r ChangeRequirementRequest) expectedVersion() int    { return r.Version }
func (r VersionRequest) expectedVersion() int              { return r.Version }
func (r ReorderFieldsRequest) expectedVersion() int        { return r.Version }
func (r AddSelectionOptionRequest) expectedVersion() int   { return r.Version }
func (r SetNumberFieldBoundsRequest) expectedVersion() int { return r.Version }

type versionedRequest interface {
	expectedVersion() int
}

type configurationWrite[R versionedRequest] struct {
	successStatus  int
	failureMessage string
	command        func(req R, wc writeContext) cqrs.Command
}

func handleConfigurationWrite[R versionedRequest](
	h *OnePagerConfigurationHandlers,
	w http.ResponseWriter,
	r *http.Request,
	op configurationWrite[R],
) {
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
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated configuration")
		return
	}
	h.respondWithRecord(w, r, record, op.successStatus)
}

var defineCustomFieldWrite = configurationWrite[DefineCustomFieldRequest]{
	successStatus:  http.StatusCreated,
	failureMessage: "Failed to define custom field",
	command: func(req DefineCustomFieldRequest, wc writeContext) cqrs.Command {
		return &commands.DefineCustomField{
			ConfigID:     wc.configID,
			Name:         req.Name,
			FieldType:    req.FieldType,
			Required:     req.Required,
			HelpText:     req.HelpText,
			OptionLabels: req.Options,
			ModifiedBy:   wc.email,
		}
	},
}

// DefineCustomField godoc
// @Summary Define a custom field
// @Description Adds a typed custom field definition to the subject type's one-pager configuration. Selection fields require at least one option.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param field body DefineCustomFieldRequest true "Custom field definition"
// @Success 201 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields [post]
func (h *OnePagerConfigurationHandlers) DefineCustomField(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, defineCustomFieldWrite)
}

var renameCustomFieldWrite = configurationWrite[RenameCustomFieldRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to rename custom field",
	command: func(req RenameCustomFieldRequest, wc writeContext) cqrs.Command {
		return &commands.RenameCustomField{
			ConfigID:      wc.configID,
			FieldID:       wc.fieldID,
			Name:          req.Name,
			HelpText:      req.HelpText,
			RequestedType: req.FieldType,
			ModifiedBy:    wc.email,
		}
	},
}

// RenameCustomField godoc
// @Summary Rename a custom field
// @Description Updates the display name and help text of an active custom field. The field keeps its ID, type, required flag, and display position. Supplying a fieldType different from the field's type is rejected: field types are immutable.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param field body RenameCustomFieldRequest true "New display metadata"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID} [put]
func (h *OnePagerConfigurationHandlers) RenameCustomField(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, renameCustomFieldWrite)
}

var changeRequirementWrite = configurationWrite[ChangeRequirementRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to change field requirement",
	command: func(req ChangeRequirementRequest, wc writeContext) cqrs.Command {
		return &commands.ChangeCustomFieldRequirement{
			ConfigID:   wc.configID,
			FieldID:    wc.fieldID,
			Required:   req.Required,
			ModifiedBy: wc.email,
		}
	},
}

// ChangeCustomFieldRequirement godoc
// @Summary Change the required flag of a custom field
// @Description Marks an active custom field as required or optional. The change only affects the configuration; no recorded data is validated or blocked.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param requirement body ChangeRequirementRequest true "Required flag"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID}/requirement [put]
func (h *OnePagerConfigurationHandlers) ChangeCustomFieldRequirement(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, changeRequirementWrite)
}

var retireCustomFieldWrite = configurationWrite[VersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to retire custom field",
	command: func(_ VersionRequest, wc writeContext) cqrs.Command {
		return &commands.RetireCustomField{
			ConfigID:   wc.configID,
			FieldID:    wc.fieldID,
			ModifiedBy: wc.email,
		}
	},
}

// RetireCustomField godoc
// @Summary Retire a custom field
// @Description Retires an active custom field. The field leaves the display order but remains on the configuration and can be reactivated later.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param version body VersionRequest true "Expected configuration version"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID}/retire [post]
func (h *OnePagerConfigurationHandlers) RetireCustomField(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, retireCustomFieldWrite)
}

var reactivateCustomFieldWrite = configurationWrite[VersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to reactivate custom field",
	command: func(_ VersionRequest, wc writeContext) cqrs.Command {
		return &commands.ReactivateCustomField{
			ConfigID:   wc.configID,
			FieldID:    wc.fieldID,
			ModifiedBy: wc.email,
		}
	},
}

// ReactivateCustomField godoc
// @Summary Reactivate a retired custom field
// @Description Reactivates a retired custom field with its original ID, type, required flag, and options. The field re-enters the display order at the end.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param version body VersionRequest true "Expected configuration version"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID}/reactivate [post]
func (h *OnePagerConfigurationHandlers) ReactivateCustomField(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, reactivateCustomFieldWrite)
}

var includeBuiltInFieldWrite = configurationWrite[VersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to include built-in field",
	command: func(_ VersionRequest, wc writeContext) cqrs.Command {
		return &commands.IncludeBuiltInField{
			ConfigID:   wc.configID,
			EntryID:    wc.entryID,
			ModifiedBy: wc.email,
		}
	},
}

// IncludeBuiltInField godoc
// @Summary Include a built-in field
// @Description Includes a catalog built-in field on the one-pager. The field enters the display order at the end.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param entryID path string true "Catalog entry ID"
// @Param version body VersionRequest true "Expected configuration version"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/built-in-fields/{entryID}/include [post]
func (h *OnePagerConfigurationHandlers) IncludeBuiltInField(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, includeBuiltInFieldWrite)
}

var excludeBuiltInFieldWrite = configurationWrite[VersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to exclude built-in field",
	command: func(_ VersionRequest, wc writeContext) cqrs.Command {
		return &commands.ExcludeBuiltInField{
			ConfigID:   wc.configID,
			EntryID:    wc.entryID,
			ModifiedBy: wc.email,
		}
	},
}

// ExcludeBuiltInField godoc
// @Summary Exclude a built-in field
// @Description Excludes an included built-in field from the one-pager. The field leaves the display order but remains available in the catalog.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param entryID path string true "Catalog entry ID"
// @Param version body VersionRequest true "Expected configuration version"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/built-in-fields/{entryID}/exclude [post]
func (h *OnePagerConfigurationHandlers) ExcludeBuiltInField(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, excludeBuiltInFieldWrite)
}

var changeBuiltInRequirementWrite = configurationWrite[ChangeRequirementRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to change built-in field requirement",
	command: func(req ChangeRequirementRequest, wc writeContext) cqrs.Command {
		return &commands.ChangeBuiltInFieldRequirement{
			ConfigID:   wc.configID,
			EntryID:    wc.entryID,
			Required:   req.Required,
			ModifiedBy: wc.email,
		}
	},
}

// ChangeBuiltInFieldRequirement godoc
// @Summary Change the required flag of a built-in field
// @Description Marks an included built-in field as required or optional. Only included built-in fields can be required; targeting an excluded or unknown built-in is rejected. The change only affects the configuration; no recorded data is validated or blocked.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param entryID path string true "Catalog entry ID"
// @Param requirement body ChangeRequirementRequest true "Required flag"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/built-in-fields/{entryID}/requirement [put]
func (h *OnePagerConfigurationHandlers) ChangeBuiltInFieldRequirement(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, changeBuiltInRequirementWrite)
}

var reorderFieldsWrite = configurationWrite[ReorderFieldsRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to reorder fields",
	command: func(req ReorderFieldsRequest, wc writeContext) cqrs.Command {
		order := make([]commands.FieldRefInput, len(req.Order))
		for i, ref := range req.Order {
			order[i] = commands.FieldRefInput{Kind: ref.Kind, ID: ref.ID}
		}
		return &commands.ReorderOnePagerFields{
			ConfigID:   wc.configID,
			Order:      order,
			ModifiedBy: wc.email,
		}
	},
}

// ReorderFields godoc
// @Summary Reorder the one-pager fields
// @Description Replaces the single interleaved display order over included built-in and active custom fields. The new order must contain every such field exactly once.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param order body ReorderFieldsRequest true "New display order"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/display-order [put]
func (h *OnePagerConfigurationHandlers) ReorderFields(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, reorderFieldsWrite)
}

var addSelectionOptionWrite = configurationWrite[AddSelectionOptionRequest]{
	successStatus:  http.StatusCreated,
	failureMessage: "Failed to add selection option",
	command: func(req AddSelectionOptionRequest, wc writeContext) cqrs.Command {
		return &commands.AddSelectionOption{
			ConfigID:   wc.configID,
			FieldID:    wc.fieldID,
			Label:      req.Label,
			ModifiedBy: wc.email,
		}
	},
}

// AddSelectionOption godoc
// @Summary Add an option to a Selection field
// @Description Adds a new active option to an active Selection custom field.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param option body AddSelectionOptionRequest true "Option label"
// @Success 201 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID}/options [post]
func (h *OnePagerConfigurationHandlers) AddSelectionOption(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, addSelectionOptionWrite)
}

var retireSelectionOptionWrite = configurationWrite[VersionRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to retire selection option",
	command: func(_ VersionRequest, wc writeContext) cqrs.Command {
		return &commands.RetireSelectionOption{
			ConfigID:   wc.configID,
			FieldID:    wc.fieldID,
			OptionID:   wc.optionID,
			ModifiedBy: wc.email,
		}
	},
}

// RetireSelectionOption godoc
// @Summary Retire a Selection field option
// @Description Retires an active option of a Selection custom field. Retired options remain on the definition. The last active option cannot be retired.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param optionID path string true "Option ID"
// @Param version body VersionRequest true "Expected configuration version"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID}/options/{optionID}/retire [post]
func (h *OnePagerConfigurationHandlers) RetireSelectionOption(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, retireSelectionOptionWrite)
}

var setNumberFieldBoundsWrite = configurationWrite[SetNumberFieldBoundsRequest]{
	successStatus:  http.StatusOK,
	failureMessage: "Failed to set number field bounds",
	command: func(req SetNumberFieldBoundsRequest, wc writeContext) cqrs.Command {
		return &commands.SetNumberFieldBounds{
			ConfigID:   wc.configID,
			FieldID:    wc.fieldID,
			Min:        req.Min,
			Max:        req.Max,
			ModifiedBy: wc.email,
		}
	},
}

// SetNumberFieldBounds godoc
// @Summary Set a Number field's bounds
// @Description Sets, tightens, loosens, or clears an active Number custom field's minimum and/or maximum bounds. Bounds only gate new facts writes; already-recorded values are never altered or hidden. Setting bounds on a non-Number field is rejected.
// @Tags one-pagers
// @Accept json
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param fieldID path string true "Field ID"
// @Param bounds body SetNumberFieldBoundsRequest true "New bounds"
// @Success 200 {object} OnePagerConfigurationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /one-pagers/configurations/{subjectType}/custom-fields/{fieldID}/bounds [put]
func (h *OnePagerConfigurationHandlers) SetNumberFieldBounds(w http.ResponseWriter, r *http.Request) {
	handleConfigurationWrite(h, w, r, setNumberFieldBoundsWrite)
}
