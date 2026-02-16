package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/architecturemodeling"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationFitScoreProjector struct {
	fitScoreReadModel *readmodels.ApplicationFitScoreReadModel
	componentGateway  architecturemodeling.ComponentGateway
	pillarsGateway    mmPL.StrategyPillarsGateway
}

func NewApplicationFitScoreProjector(
	fitScoreReadModel *readmodels.ApplicationFitScoreReadModel,
	componentGateway architecturemodeling.ComponentGateway,
	pillarsGateway mmPL.StrategyPillarsGateway,
) *ApplicationFitScoreProjector {
	return &ApplicationFitScoreProjector{
		fitScoreReadModel: fitScoreReadModel,
		componentGateway:  componentGateway,
		pillarsGateway:    pillarsGateway,
	}
}

func (p *ApplicationFitScoreProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *ApplicationFitScoreProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"ApplicationFitScoreSet":     p.handleApplicationFitScoreSet,
		"ApplicationFitScoreUpdated": p.handleApplicationFitScoreUpdated,
		"ApplicationFitScoreRemoved": p.handleApplicationFitScoreRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *ApplicationFitScoreProjector) handleApplicationFitScoreSet(ctx context.Context, eventData []byte) error {
	var event events.ApplicationFitScoreSet
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationFitScoreSet event data: %w", err)
		log.Printf("failed to unmarshal ApplicationFitScoreSet event: %v", wrappedErr)
		return wrappedErr
	}

	componentName, err := p.fetchComponentName(ctx, event.ComponentID)
	if err != nil {
		return fmt.Errorf("resolve component name for fit score %s component %s: %w", event.ID, event.ComponentID, err)
	}

	pillarName := p.resolvePillarName(ctx, event.PillarID, event.PillarName)
	score, _ := valueobjects.NewFitScore(event.Score)

	dto := readmodels.ApplicationFitScoreDTO{
		ID:            event.ID,
		ComponentID:   event.ComponentID,
		ComponentName: componentName,
		PillarID:      event.PillarID,
		PillarName:    pillarName,
		Score:         event.Score,
		ScoreLabel:    score.Label(),
		Rationale:     event.Rationale,
		ScoredAt:      event.ScoredAt,
		ScoredBy:      event.ScoredBy,
	}

	if err := p.fitScoreReadModel.Insert(ctx, dto); err != nil {
		return fmt.Errorf("project ApplicationFitScoreSet for fit score %s: %w", event.ID, err)
	}
	return nil
}

func (p *ApplicationFitScoreProjector) fetchComponentName(ctx context.Context, componentID string) (string, error) {
	dto, err := p.componentGateway.GetByID(ctx, componentID)
	if err != nil {
		log.Printf("Failed to get component %s: %v", componentID, err)
		return "", fmt.Errorf("load component %s for fit score projection: %w", componentID, err)
	}
	if dto == nil {
		return "", nil
	}
	return dto.Name, nil
}

func (p *ApplicationFitScoreProjector) resolvePillarName(ctx context.Context, pillarID, eventPillarName string) string {
	if eventPillarName != "" {
		return eventPillarName
	}
	if p.pillarsGateway == nil {
		return ""
	}
	pillar, _ := p.pillarsGateway.GetActivePillar(ctx, pillarID)
	if pillar == nil {
		return ""
	}
	return pillar.Name
}

func (p *ApplicationFitScoreProjector) handleApplicationFitScoreUpdated(ctx context.Context, eventData []byte) error {
	var event events.ApplicationFitScoreUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationFitScoreUpdated event data: %w", err)
		log.Printf("failed to unmarshal ApplicationFitScoreUpdated event: %v", wrappedErr)
		return wrappedErr
	}

	score, _ := valueobjects.NewFitScore(event.Score)

	dto := readmodels.ApplicationFitScoreDTO{
		ID:         event.ID,
		Score:      event.Score,
		ScoreLabel: score.Label(),
		Rationale:  event.Rationale,
	}

	if err := p.fitScoreReadModel.Update(ctx, dto); err != nil {
		return fmt.Errorf("project ApplicationFitScoreUpdated for fit score %s: %w", event.ID, err)
	}
	return nil
}

func (p *ApplicationFitScoreProjector) handleApplicationFitScoreRemoved(ctx context.Context, eventData []byte) error {
	var event events.ApplicationFitScoreRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationFitScoreRemoved event data: %w", err)
		log.Printf("failed to unmarshal ApplicationFitScoreRemoved event: %v", wrappedErr)
		return wrappedErr
	}

	if err := p.fitScoreReadModel.Delete(ctx, event.ID); err != nil {
		return fmt.Errorf("project ApplicationFitScoreRemoved for fit score %s: %w", event.ID, err)
	}
	return nil
}
