package handlers

import (
	"context"
	"fmt"
	"log"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type DeleteApplicationComponentRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ApplicationComponent, error)
	Save(ctx context.Context, component *aggregates.ApplicationComponent) error
}

type DeleteApplicationComponentRelationReader interface {
	GetBySourceID(ctx context.Context, sourceID string) ([]readmodels.ComponentRelationDTO, error)
	GetByTargetID(ctx context.Context, targetID string) ([]readmodels.ComponentRelationDTO, error)
}

type DeleteApplicationComponentHandler struct {
	repository      DeleteApplicationComponentRepository
	relationReader  DeleteApplicationComponentRelationReader
	componentReader ComponentRecordReader
	commandBus      cqrs.CommandBus
}

func NewDeleteApplicationComponentHandler(
	repository DeleteApplicationComponentRepository,
	relationReader DeleteApplicationComponentRelationReader,
	componentReader ComponentRecordReader,
	commandBus cqrs.CommandBus,
) *DeleteApplicationComponentHandler {
	return &DeleteApplicationComponentHandler{
		repository:      repository,
		relationReader:  relationReader,
		componentReader: componentReader,
		commandBus:      commandBus,
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

	return cqrs.EmptyResult(), h.deleteComponent(ctx, componentID.Value())
}

func (h *DeleteApplicationComponentHandler) deleteComponent(ctx context.Context, componentID string) error {
	component, err := h.repository.GetByID(ctx, componentID)
	if err != nil {
		return err
	}
	if component.IsDeleted() {
		return nil
	}
	if err := h.releaseContainments(ctx, componentID); err != nil {
		return err
	}
	if err := component.Delete(); err != nil {
		return err
	}
	if err := h.repository.Save(ctx, component); err != nil {
		return err
	}
	return h.cascadeRelationDeletes(ctx, componentID)
}

func (h *DeleteApplicationComponentHandler) releaseContainments(ctx context.Context, componentID string) error {
	record, err := h.componentReader.GetByID(ctx, componentID)
	if err != nil {
		return fmt.Errorf("read containments of component %s: %w", componentID, err)
	}
	if record == nil {
		return nil
	}
	for _, part := range record.Parts {
		if _, err := h.commandBus.Dispatch(ctx, releasePartCommand(part)); err != nil {
			return fmt.Errorf("release part %s of component %s: %w", part.ID, componentID, err)
		}
	}
	if record.PartOf == nil {
		return nil
	}
	if _, err := h.commandBus.Dispatch(ctx, &commands.DetachComponent{ComponentID: componentID}); err != nil {
		return fmt.Errorf("detach component %s from parent %s: %w", componentID, record.PartOf.ID, err)
	}
	return nil
}

func releasePartCommand(part readmodels.ContainmentPartDTO) cqrs.Command {
	if part.Kind == valueobjects.ContainmentComposition {
		return &commands.DeleteApplicationComponent{ID: part.ID}
	}
	return &commands.DetachComponent{ComponentID: part.ID}
}

func (h *DeleteApplicationComponentHandler) cascadeRelationDeletes(ctx context.Context, componentID string) error {
	relationsAsSource, err := h.relationReader.GetBySourceID(ctx, componentID)
	if err != nil {
		log.Printf("Error querying relations by source for component %s: %v", componentID, err)
		return err
	}

	relationsAsTarget, err := h.relationReader.GetByTargetID(ctx, componentID)
	if err != nil {
		log.Printf("Error querying relations by target for component %s: %v", componentID, err)
		return err
	}

	allRelations := make([]readmodels.ComponentRelationDTO, 0, len(relationsAsSource)+len(relationsAsTarget))
	allRelations = append(allRelations, relationsAsSource...)
	allRelations = append(allRelations, relationsAsTarget...)
	log.Printf("Found %d relations to cascade delete for component %s", len(allRelations), componentID)

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
	return nil
}
