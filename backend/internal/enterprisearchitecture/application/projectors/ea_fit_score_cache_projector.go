package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type FitScoreCacheWriter interface {
	Upsert(ctx context.Context, entry readmodels.FitScoreEntry) error
	Delete(ctx context.Context, componentID, pillarID string) error
}

type EAFitScoreCacheProjector struct {
	readModel FitScoreCacheWriter
}

func NewEAFitScoreCacheProjector(readModel FitScoreCacheWriter) *EAFitScoreCacheProjector {
	return &EAFitScoreCacheProjector{readModel: readModel}
}

func (p *EAFitScoreCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EAFitScoreCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		cmPL.ApplicationFitScoreSet:     p.handleApplicationFitScoreSet,
		cmPL.ApplicationFitScoreRemoved: p.handleApplicationFitScoreRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type applicationFitScoreSetEvent struct {
	ComponentID string `json:"componentId"`
	PillarID    string `json:"pillarId"`
	Score       int    `json:"score"`
	Rationale   string `json:"rationale"`
}

func (p *EAFitScoreCacheProjector) handleApplicationFitScoreSet(ctx context.Context, eventData []byte) error {
	var event applicationFitScoreSetEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationFitScoreSet event data: %w", err)
		log.Printf("failed to unmarshal ApplicationFitScoreSet event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Upsert(ctx, readmodels.FitScoreEntry{
		ComponentID: event.ComponentID,
		PillarID:    event.PillarID,
		Score:       event.Score,
		Rationale:   event.Rationale,
	}); err != nil {
		return fmt.Errorf("project ApplicationFitScoreSet cache upsert for component %s pillar %s: %w", event.ComponentID, event.PillarID, err)
	}
	return nil
}

type applicationFitScoreRemovedEvent struct {
	ComponentID string `json:"componentId"`
	PillarID    string `json:"pillarId"`
}

func (p *EAFitScoreCacheProjector) handleApplicationFitScoreRemoved(ctx context.Context, eventData []byte) error {
	var event applicationFitScoreRemovedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal ApplicationFitScoreRemoved event data: %w", err)
		log.Printf("failed to unmarshal ApplicationFitScoreRemoved event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.readModel.Delete(ctx, event.ComponentID, event.PillarID); err != nil {
		return fmt.Errorf("project ApplicationFitScoreRemoved cache delete for component %s pillar %s: %w", event.ComponentID, event.PillarID, err)
	}
	return nil
}
