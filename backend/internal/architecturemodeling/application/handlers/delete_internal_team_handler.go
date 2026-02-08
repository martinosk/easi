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

type DeleteInternalTeamHandler struct {
	repository        *repositories.InternalTeamRepository
	relationReadModel *readmodels.BuiltByRelationshipReadModel
	commandBus        cqrs.CommandBus
}

func NewDeleteInternalTeamHandler(
	repository *repositories.InternalTeamRepository,
	relationReadModel *readmodels.BuiltByRelationshipReadModel,
	commandBus cqrs.CommandBus,
) *DeleteInternalTeamHandler {
	return &DeleteInternalTeamHandler{
		repository:        repository,
		relationReadModel: relationReadModel,
		commandBus:        commandBus,
	}
}

func (h *DeleteInternalTeamHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteInternalTeam)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	team, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if team.IsDeleted() {
		return cqrs.EmptyResult(), nil
	}

	if err := team.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, team); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.EmptyResult(), h.cascadeClear(ctx, command.ID)
}

func (h *DeleteInternalTeamHandler) cascadeClear(ctx context.Context, teamID string) error {
	relations, err := h.relationReadModel.GetByTeamID(ctx, teamID)
	if err != nil {
		log.Printf("Error querying relationships for internal team %s: %v", teamID, err)
		return err
	}

	for _, relation := range relations {
		clearCmd := &commands.ClearOriginLink{ComponentID: relation.ComponentID, OriginType: valueobjects.OriginTypeBuiltBy}
		if _, err := h.commandBus.Dispatch(ctx, clearCmd); err != nil {
			log.Printf("Error cascading clear for relationship on component %s: %v", relation.ComponentID, err)
			continue
		}
		log.Printf("Cascaded clear for built by relationship on component %s", relation.ComponentID)
	}
	return nil
}
