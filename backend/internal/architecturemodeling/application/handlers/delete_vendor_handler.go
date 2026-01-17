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

	relations, err := h.relationReadModel.GetByVendorID(ctx, command.ID)
	if err != nil {
		log.Printf("Error querying relationships for vendor %s: %v", command.ID, err)
		return cqrs.EmptyResult(), err
	}

	for _, relation := range relations {
		deleteCmd := &commands.DeletePurchasedFromRelationship{ID: relation.ID}
		if _, err := h.commandBus.Dispatch(ctx, deleteCmd); err != nil {
			log.Printf("Error cascading delete for relationship %s: %v", relation.ID, err)
			continue
		}
		log.Printf("Cascaded delete for purchased from relationship %s", relation.ID)
	}

	return cqrs.EmptyResult(), nil
}
