package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrComponentOriginsNotFound = errors.New("component origins not found")

type ComponentOriginsRepository struct {
	*repository.EventSourcedRepository[*aggregates.ComponentOrigins]
}

func NewComponentOriginsRepository(eventStore eventstore.EventStore) *ComponentOriginsRepository {
	return &ComponentOriginsRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			componentOriginsEventDeserializers,
			aggregates.LoadComponentOriginsFromHistory,
			ErrComponentOriginsNotFound,
		),
	}
}

var componentOriginsEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"ComponentOriginsCreated":           repository.JSONDeserializer[events.ComponentOriginsCreated],
		"AcquiredViaRelationshipSet":        repository.JSONDeserializer[events.AcquiredViaRelationshipSet],
		"AcquiredViaRelationshipReplaced":   repository.JSONDeserializer[events.AcquiredViaRelationshipReplaced],
		"AcquiredViaNotesUpdated":           repository.JSONDeserializer[events.AcquiredViaNotesUpdated],
		"AcquiredViaRelationshipCleared":    repository.JSONDeserializer[events.AcquiredViaRelationshipCleared],
		"PurchasedFromRelationshipSet":      repository.JSONDeserializer[events.PurchasedFromRelationshipSet],
		"PurchasedFromRelationshipReplaced": repository.JSONDeserializer[events.PurchasedFromRelationshipReplaced],
		"PurchasedFromNotesUpdated":         repository.JSONDeserializer[events.PurchasedFromNotesUpdated],
		"PurchasedFromRelationshipCleared":  repository.JSONDeserializer[events.PurchasedFromRelationshipCleared],
		"BuiltByRelationshipSet":            repository.JSONDeserializer[events.BuiltByRelationshipSet],
		"BuiltByRelationshipReplaced":       repository.JSONDeserializer[events.BuiltByRelationshipReplaced],
		"BuiltByNotesUpdated":               repository.JSONDeserializer[events.BuiltByNotesUpdated],
		"BuiltByRelationshipCleared":        repository.JSONDeserializer[events.BuiltByRelationshipCleared],
		"ComponentOriginsDeleted":           repository.JSONDeserializer[events.ComponentOriginsDeleted],
	},
)
