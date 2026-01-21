package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type VendorHandlers struct {
	commandBus       cqrs.CommandBus
	readModel        *readmodels.VendorReadModel
	paginationHelper *sharedAPI.PaginationHelper
	hateoas          *sharedAPI.HATEOASLinks
}

func NewVendorHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.VendorReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *VendorHandlers {
	return &VendorHandlers{
		commandBus:       commandBus,
		readModel:        readModel,
		paginationHelper: sharedAPI.NewPaginationHelper("/api/v1/vendors"),
		hateoas:          hateoas,
	}
}

type CreateVendorRequest struct {
	Name                  string `json:"name"`
	ImplementationPartner string `json:"implementationPartner,omitempty"`
	Notes                 string `json:"notes,omitempty"`
}

type UpdateVendorRequest struct {
	Name                  string `json:"name"`
	ImplementationPartner string `json:"implementationPartner,omitempty"`
	Notes                 string `json:"notes,omitempty"`
}

// CreateVendor godoc
// @Summary Create a new vendor
// @Description Creates a new vendor (external software provider)
// @Tags vendors
// @Accept json
// @Produce json
// @Param vendor body CreateVendorRequest true "Vendor data"
// @Success 201 {object} readmodels.VendorDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /vendors [post]
func (h *VendorHandlers) CreateVendor(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateVendorRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateVendor{
		Name:                  req.Name,
		ImplementationPartner: req.ImplementationPartner,
		Notes:                 req.Notes,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/vendors"), sharedAPI.ResourceID(result.CreatedID))
	vendor, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.HandleErrorWithDefault(w, err, "Failed to retrieve created vendor")
		return
	}

	if vendor == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "Vendor created, processing",
		})
		return
	}

	h.enrichWithLinks(r, vendor)
	sharedAPI.RespondCreated(w, location, vendor)
}

// GetAllVendors godoc
// @Summary Get all vendors
// @Description Retrieves all vendors with cursor-based pagination
// @Tags vendors
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination"
// @Success 200 {object} sharedAPI.PaginatedResponse{data=[]readmodels.VendorDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /vendors [get]
func (h *VendorHandlers) GetAllVendors(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterID, afterName, err := h.paginationHelper.ProcessNameCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	vendors, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterID, afterName)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve vendors")
		return
	}

	for i := range vendors {
		h.enrichWithLinks(r, &vendors[i])
	}

	pageables := ConvertVendorsToNamePageable(vendors)
	nextCursor := h.paginationHelper.GenerateNextNameCursor(pageables, hasMore)
	selfLink := h.paginationHelper.BuildSelfLink(params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       vendors,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/vendors",
	})
}

// GetVendorByID godoc
// @Summary Get a vendor by ID
// @Description Retrieves a specific vendor by its ID
// @Tags vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Success 200 {object} readmodels.VendorDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /vendors/{id} [get]
func (h *VendorHandlers) GetVendorByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	vendor, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve vendor")
		return
	}

	if vendor == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Vendor not found")
		return
	}

	h.enrichWithLinks(r, vendor)
	sharedAPI.RespondJSON(w, http.StatusOK, vendor)
}

// UpdateVendor godoc
// @Summary Update a vendor
// @Description Updates an existing vendor
// @Tags vendors
// @Accept json
// @Produce json
// @Param id path string true "Vendor ID"
// @Param vendor body UpdateVendorRequest true "Updated vendor data"
// @Success 200 {object} readmodels.VendorDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /vendors/{id} [put]
func (h *VendorHandlers) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateVendorRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateVendor{
		ID:                    id,
		Name:                  req.Name,
		ImplementationPartner: req.ImplementationPartner,
		Notes:                 req.Notes,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	vendor, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.HandleErrorWithDefault(w, err, "Failed to retrieve updated vendor")
		return
	}

	if vendor == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Vendor not found")
		return
	}

	h.enrichWithLinks(r, vendor)
	sharedAPI.RespondJSON(w, http.StatusOK, vendor)
}

// DeleteVendor godoc
// @Summary Delete a vendor
// @Description Permanently deletes a vendor
// @Tags vendors
// @Produce json
// @Param id path string true "Vendor ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /vendors/{id} [delete]
func (h *VendorHandlers) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	cmd := &commands.DeleteVendor{
		ID: id,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

func (h *VendorHandlers) enrichWithLinks(r *http.Request, vendor *readmodels.VendorDTO) {
	actor, _ := sharedctx.GetActor(r.Context())
	vendor.Links = h.hateoas.VendorLinksForActor(vendor.ID, actor)
}
