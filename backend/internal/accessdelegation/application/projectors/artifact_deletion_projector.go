package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/accessdelegation/application/commands"
	"easi/backend/internal/accessdelegation/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/cqrs"
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
		log.Printf("Failed to marshal event data for artifact deletion: %v", err)
		return err
	}

	var deleted artifactDeletedEvent
	if err := json.Unmarshal(eventData, &deleted); err != nil {
		log.Printf("Failed to unmarshal artifact deleted event: %v", err)
		return err
	}

	artifactID := deleted.ID
	if artifactID == "" {
		artifactID = event.AggregateID()
	}

	grantIDs, err := p.readModel.GetActiveGrantIDsForArtifact(ctx, p.artifactType, artifactID)
	if err != nil {
		log.Printf("Failed to get active grants for deleted %s %s: %v", p.artifactType, artifactID, err)
		return err
	}

	for _, grantID := range grantIDs {
		cmd := &commands.RevokeEditGrant{
			ID:        grantID,
			RevokedBy: "system:artifact-deleted",
		}
		if _, err := p.commandBus.Dispatch(ctx, cmd); err != nil {
			log.Printf("Failed to revoke edit grant %s for deleted %s %s: %v", grantID, p.artifactType, artifactID, err)
		}
	}

	return nil
}
