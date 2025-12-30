package handlers

import (
	"context"
	"log"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type DeleteApplicationComponentHandler struct {
	repository     *repositories.ApplicationComponentRepository
	relationReader *readmodels.ComponentRelationReadModel
	commandBus     cqrs.CommandBus
}

func NewDeleteApplicationComponentHandler(
	repository *repositories.ApplicationComponentRepository,
	relationReader *readmodels.ComponentRelationReadModel,
	commandBus cqrs.CommandBus,
) *DeleteApplicationComponentHandler {
	return &DeleteApplicationComponentHandler{
		repository:     repository,
		relationReader: relationReader,
		commandBus:     commandBus,
	}
}

func (h *DeleteApplicationComponentHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteApplicationComponent)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	component, err := h.repository.GetByID(ctx, componentID.Value())
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if component.IsDeleted() {
		return cqrs.EmptyResult(), nil
	}

	if err := component.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, component); err != nil {
		return cqrs.EmptyResult(), err
	}

	relationsAsSource, err := h.relationReader.GetBySourceID(ctx, componentID.Value())
	if err != nil {
		log.Printf("Error querying relations by source for component %s: %v", componentID.Value(), err)
		return cqrs.EmptyResult(), err
	}

	relationsAsTarget, err := h.relationReader.GetByTargetID(ctx, componentID.Value())
	if err != nil {
		log.Printf("Error querying relations by target for component %s: %v", componentID.Value(), err)
		return cqrs.EmptyResult(), err
	}

	allRelations := make([]readmodels.ComponentRelationDTO, 0, len(relationsAsSource)+len(relationsAsTarget))
	allRelations = append(allRelations, relationsAsSource...)
	allRelations = append(allRelations, relationsAsTarget...)
	log.Printf("Found %d relations to cascade delete for component %s", len(allRelations), componentID.Value())

	for _, relation := range allRelations {
		deleteRelCmd := &commands.DeleteComponentRelation{
			ID: relation.ID,
		}

		if _, err := h.commandBus.Dispatch(ctx, deleteRelCmd); err != nil {
			log.Printf("Error cascading delete for relation %s: %v", relation.ID, err)
			continue
		}

		log.Printf("Cascaded delete for relation %s", relation.ID)
	}

	return cqrs.EmptyResult(), nil
}
