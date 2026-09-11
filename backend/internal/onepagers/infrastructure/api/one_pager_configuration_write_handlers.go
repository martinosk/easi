package api

import (
	"net/http"

	"easi/backend/internal/onepagers/application/commands"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

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

func (r ChangeRequirementRequest) expectedVersion() int { return r.Version }
func (r VersionRequest) expectedVersion() int           { return r.Version }
func (r ReorderFieldsRequest) expectedVersion() int     { return r.Version }

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
