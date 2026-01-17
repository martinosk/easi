package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrPurchasedFromRelationshipNotFound = errors.New("purchased from relationship not found")

type PurchasedFromRelationshipRepository struct {
	*repository.EventSourcedRepository[*aggregates.PurchasedFromRelationship]
}

func NewPurchasedFromRelationshipRepository(eventStore eventstore.EventStore) *PurchasedFromRelationshipRepository {
	return &PurchasedFromRelationshipRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			purchasedFromRelationshipEventDeserializers,
			aggregates.LoadPurchasedFromRelationshipFromHistory,
			ErrPurchasedFromRelationshipNotFound,
		),
	}
}

var purchasedFromRelationshipEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"PurchasedFromRelationshipCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			vendorID, err := repository.GetRequiredString(data, "vendorId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			notes, _ := repository.GetOptionalString(data, "notes", "")
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewPurchasedFromRelationshipCreated(id, vendorID, componentID, notes)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"PurchasedFromRelationshipDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			vendorID, err := repository.GetRequiredString(data, "vendorId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}

			return events.NewPurchasedFromRelationshipDeleted(id, vendorID, componentID), nil
		},
	},
)
