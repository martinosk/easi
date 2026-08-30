package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	metaPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type MaturityScaleStore interface {
	Save(ctx context.Context, sections []readmodels.MaturityScaleSection) error
}

type MaturityScaleProjector struct {
	cache MaturityScaleStore
}

func NewMaturityScaleProjector(cache MaturityScaleStore) *MaturityScaleProjector {
	return &MaturityScaleProjector{cache: cache}
}

func MaturityScaleEventTypes() []string {
	return []string{
		metaPL.MetaModelConfigurationCreated,
		metaPL.MaturityScaleConfigUpdated,
		metaPL.MaturityScaleConfigReset,
	}
}

func (p *MaturityScaleProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

type maturityScalePayload struct {
	Sections    []readmodels.MaturityScaleSection `json:"sections"`
	NewSections []readmodels.MaturityScaleSection `json:"newSections"`
}

func (p *MaturityScaleProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if !slices.Contains(MaturityScaleEventTypes(), eventType) {
		return nil
	}

	var payload maturityScalePayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal maturity scale sections of %s: %w", eventType, err)
	}

	sections := payload.Sections
	if eventType == metaPL.MaturityScaleConfigUpdated {
		sections = payload.NewSections
	}
	if err := p.cache.Save(ctx, sections); err != nil {
		return fmt.Errorf("cache maturity scale from %s: %w", eventType, err)
	}
	return nil
}
