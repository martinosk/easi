package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/application/readmodels"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ArtifactDeletionProjector struct {
	readModel    *readmodels.EditGrantReadModel
	commandBus   cqrs.CommandBus
	artifactType string
}

func NewArtifactDeletionProjector(readModel *readmodels.EditGrantReadModel, commandBus cqrs.CommandBus, artifactType string) *ArtifactDeletionProjector {
	return &ArtifactDeletionProjector{
		readModel:    readModel,
		commandBus:   commandBus,
		artifactType: artifactType,
	}
}

type artifactDeletedEvent struct {
	ID string `json:"id"`
}

func (p *ArtifactDeletionProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal event data for artifact deletion: %w", err)
	}

	var deleted artifactDeletedEvent
	if err := json.Unmarshal(eventData, &deleted); err != nil {
		return fmt.Errorf("unmarshal artifact deleted event: %w", err)
	}

	artifactID := deleted.ID
	if artifactID == "" {
		artifactID = event.AggregateID()
	}

	grantIDs, err := p.readModel.GetActiveGrantIDsForArtifact(ctx, p.artifactType, artifactID)
	if err != nil {
		return fmt.Errorf("get active grants for deleted %s %s: %w", p.artifactType, artifactID, err)
	}

	for _, grantID := range grantIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cmd := &commands.RevokeEditGrant{
			ID:        grantID,
			RevokedBy: "system:artifact-deleted",
		}
		if _, err := p.commandBus.Dispatch(ctx, cmd); err != nil {
			return fmt.Errorf("revoke edit grant %s for deleted %s %s: %w", grantID, p.artifactType, artifactID, err)
		}
	}

	return nil
}
