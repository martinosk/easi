package handlers

import (
	"context"
	"log"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteVendorHandler struct {
	repository        *repositories.VendorRepository
	relationReadModel *readmodels.PurchasedFromRelationshipReadModel
	commandBus        cqrs.CommandBus
}

func NewDeleteVendorHandler(
	repository *repositories.VendorRepository,
	relationReadModel *readmodels.PurchasedFromRelationshipReadModel,
	commandBus cqrs.CommandBus,
) *DeleteVendorHandler {
	return &DeleteVendorHandler{
		repository:        repository,
		relationReadModel: relationReadModel,
		commandBus:        commandBus,
	}
}

func (h *DeleteVendorHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteVendor)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	vendor, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if vendor.IsDeleted() {
		return cqrs.EmptyResult(), nil
	}

	if err := vendor.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, vendor); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), h.cascadeClear(ctx, command.ID)
}

func (h *DeleteVendorHandler) cascadeClear(ctx context.Context, vendorID string) error {
	relations, err := h.relationReadModel.GetByVendorID(ctx, vendorID)
	if err != nil {
		log.Printf("Error querying relationships for vendor %s: %v", vendorID, err)
		return err
	}

	for _, relation := range relations {
		clearCmd := &commands.ClearOriginLink{ComponentID: relation.ComponentID, OriginType: "purchased-from"}
		if _, err := h.commandBus.Dispatch(ctx, clearCmd); err != nil {
			log.Printf("Error cascading clear for relationship on component %s: %v", relation.ComponentID, err)
			continue
		}
		log.Printf("Cascaded clear for purchased from relationship on component %s", relation.ComponentID)
	}
	return nil
}
