package projectors

import (
	"context"
	"encoding/json"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	domain "easi/backend/internal/shared/eventsourcing"
)

type StrategyImportanceProjector struct {
	importanceReadModel *readmodels.StrategyImportanceReadModel
	domainReadModel     *readmodels.BusinessDomainReadModel
	capabilityReadModel *readmodels.CapabilityReadModel
	pillarsGateway      metamodel.StrategyPillarsGateway
}

func NewStrategyImportanceProjector(
	importanceReadModel *readmodels.StrategyImportanceReadModel,
	domainReadModel *readmodels.BusinessDomainReadModel,
	capabilityReadModel *readmodels.CapabilityReadModel,
	pillarsGateway metamodel.StrategyPillarsGateway,
) *StrategyImportanceProjector {
	return &StrategyImportanceProjector{
		importanceReadModel: importanceReadModel,
		domainReadModel:     domainReadModel,
		capabilityReadModel: capabilityReadModel,
		pillarsGateway:      pillarsGateway,
	}
}

func (p *StrategyImportanceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *StrategyImportanceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"StrategyImportanceSet":     p.handleStrategyImportanceSet,
		"StrategyImportanceUpdated": p.handleStrategyImportanceUpdated,
		"StrategyImportanceRemoved": p.handleStrategyImportanceRemoved,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

func (p *StrategyImportanceProjector) handleStrategyImportanceSet(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceSet
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceSet event: %v", err)
		return err
	}

	domainName, err := p.fetchDomainName(ctx, event.BusinessDomainID)
	if err != nil {
		return err
	}

	capabilityName, err := p.fetchCapabilityName(ctx, event.CapabilityID)
	if err != nil {
		return err
	}

	pillarName := p.resolvePillarName(ctx, event.PillarID, event.PillarName)
	importance, _ := valueobjects.NewImportance(event.Importance)

	dto := readmodels.StrategyImportanceDTO{
		ID:                 event.ID,
		BusinessDomainID:   event.BusinessDomainID,
		BusinessDomainName: domainName,
		CapabilityID:       event.CapabilityID,
		CapabilityName:     capabilityName,
		PillarID:           event.PillarID,
		PillarName:         pillarName,
		Importance:         event.Importance,
		ImportanceLabel:    importance.Label(),
		Rationale:          event.Rationale,
		SetAt:              event.SetAt,
	}

	return p.importanceReadModel.Insert(ctx, dto)
}

type namedEntity interface {
	GetName() string
}

type entityNameGetter[T namedEntity] func(ctx context.Context, id string) (T, error)

func fetchEntityName[T namedEntity](ctx context.Context, id, entityType string, getter entityNameGetter[T]) (string, error) {
	entity, err := getter(ctx, id)
	if err != nil {
		log.Printf("Failed to get %s %s: %v", entityType, id, err)
		return "", err
	}
	var zero T
	if any(entity) == any(zero) {
		return "", nil
	}
	return entity.GetName(), nil
}

func (p *StrategyImportanceProjector) fetchDomainName(ctx context.Context, domainID string) (string, error) {
	return fetchEntityName(ctx, domainID, "business domain", func(ctx context.Context, id string) (*domainNameWrapper, error) {
		dto, err := p.domainReadModel.GetByID(ctx, id)
		if dto == nil {
			return nil, err
		}
		return &domainNameWrapper{name: dto.Name}, err
	})
}

func (p *StrategyImportanceProjector) fetchCapabilityName(ctx context.Context, capabilityID string) (string, error) {
	return fetchEntityName(ctx, capabilityID, "capability", func(ctx context.Context, id string) (*capabilityNameWrapper, error) {
		dto, err := p.capabilityReadModel.GetByID(ctx, id)
		if dto == nil {
			return nil, err
		}
		return &capabilityNameWrapper{name: dto.Name}, err
	})
}

type domainNameWrapper struct{ name string }

func (w *domainNameWrapper) GetName() string { return w.name }

type capabilityNameWrapper struct{ name string }

func (w *capabilityNameWrapper) GetName() string { return w.name }

func (p *StrategyImportanceProjector) resolvePillarName(ctx context.Context, pillarID, eventPillarName string) string {
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

func (p *StrategyImportanceProjector) handleStrategyImportanceUpdated(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceUpdated
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceUpdated event: %v", err)
		return err
	}

	importance, _ := valueobjects.NewImportance(event.Importance)

	dto := readmodels.StrategyImportanceDTO{
		ID:              event.ID,
		Importance:      event.Importance,
		ImportanceLabel: importance.Label(),
		Rationale:       event.Rationale,
	}

	return p.importanceReadModel.Update(ctx, dto)
}

func (p *StrategyImportanceProjector) handleStrategyImportanceRemoved(ctx context.Context, eventData []byte) error {
	var event events.StrategyImportanceRemoved
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal StrategyImportanceRemoved event: %v", err)
		return err
	}

	return p.importanceReadModel.Delete(ctx, event.ID)
}
