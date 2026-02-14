package projectors

import (
	"context"
	"encoding/json"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
		log.Printf("Failed to unmarshal ApplicationFitScoreSet event: %v", err)
		return err
	}

	componentName, err := p.fetchComponentName(ctx, event.ComponentID)
	if err != nil {
		return err
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

	return p.fitScoreReadModel.Insert(ctx, dto)
}

func (p *ApplicationFitScoreProjector) fetchComponentName(ctx context.Context, componentID string) (string, error) {
	dto, err := p.componentGateway.GetByID(ctx, componentID)
	if err != nil {
		log.Printf("Failed to get component %s: %v", componentID, err)
		return "", err
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
		log.Printf("Failed to unmarshal ApplicationFitScoreUpdated event: %v", err)
		return err
	}

	score, _ := valueobjects.NewFitScore(event.Score)

	dto := readmodels.ApplicationFitScoreDTO{
		ID:         event.ID,
		Score:      event.Score,
		ScoreLabel: score.Label(),
		Rationale:  event.Rationale,
	}

	return p.fitScoreReadModel.Update(ctx, dto)
}

func (p *ApplicationFitScoreProjector) handleApplicationFitScoreRemoved(ctx context.Context, eventData []byte) error {
	var event events.ApplicationFitScoreRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal ApplicationFitScoreRemoved event: %v", err)
		return err
	}

	return p.fitScoreReadModel.Delete(ctx, event.ID)
}
