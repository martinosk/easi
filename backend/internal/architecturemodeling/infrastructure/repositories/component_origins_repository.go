package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
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
		"ComponentOriginsCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			aggregateID, err := repository.GetRequiredString(data, "aggregateId")
			if err != nil {
				return nil, err
			}
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}
			return events.ComponentOriginsCreated{
				BaseEvent:   domain.NewBaseEvent(aggregateID),
				ComponentID: componentID,
				CreatedAt:   createdAt,
			}, nil
		},
		"AcquiredViaRelationshipSet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			entityID, err := repository.GetRequiredString(data, "entityId")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetRequiredString(data, "notes")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}
			return events.AcquiredViaRelationshipSet{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				EntityID:    entityID,
				Notes:       notes,
				LinkedAt:    linkedAt,
			}, nil
		},
		"AcquiredViaRelationshipReplaced": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			oldEntityID, err := repository.GetRequiredString(data, "oldEntityId")
			if err != nil {
				return nil, err
			}
			newEntityID, err := repository.GetRequiredString(data, "newEntityId")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetRequiredString(data, "notes")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}
			return events.AcquiredViaRelationshipReplaced{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				OldEntityID: oldEntityID,
				NewEntityID: newEntityID,
				Notes:       notes,
				LinkedAt:    linkedAt,
			}, nil
		},
		"AcquiredViaNotesUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			entityID, err := repository.GetRequiredString(data, "entityId")
			if err != nil {
				return nil, err
			}
			oldNotes, err := repository.GetRequiredString(data, "oldNotes")
			if err != nil {
				return nil, err
			}
			newNotes, err := repository.GetRequiredString(data, "newNotes")
			if err != nil {
				return nil, err
			}
			return events.AcquiredViaNotesUpdated{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				EntityID:    entityID,
				OldNotes:    oldNotes,
				NewNotes:    newNotes,
			}, nil
		},
		"AcquiredViaRelationshipCleared": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			entityID, err := repository.GetRequiredString(data, "entityId")
			if err != nil {
				return nil, err
			}
			return events.AcquiredViaRelationshipCleared{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				EntityID:    entityID,
			}, nil
		},
		"PurchasedFromRelationshipSet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			vendorID, err := repository.GetRequiredString(data, "vendorId")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetRequiredString(data, "notes")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}
			return events.PurchasedFromRelationshipSet{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				VendorID:    vendorID,
				Notes:       notes,
				LinkedAt:    linkedAt,
			}, nil
		},
		"PurchasedFromRelationshipReplaced": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			oldVendorID, err := repository.GetRequiredString(data, "oldVendorId")
			if err != nil {
				return nil, err
			}
			newVendorID, err := repository.GetRequiredString(data, "newVendorId")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetRequiredString(data, "notes")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}
			return events.PurchasedFromRelationshipReplaced{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				OldVendorID: oldVendorID,
				NewVendorID: newVendorID,
				Notes:       notes,
				LinkedAt:    linkedAt,
			}, nil
		},
		"PurchasedFromNotesUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			vendorID, err := repository.GetRequiredString(data, "vendorId")
			if err != nil {
				return nil, err
			}
			oldNotes, err := repository.GetRequiredString(data, "oldNotes")
			if err != nil {
				return nil, err
			}
			newNotes, err := repository.GetRequiredString(data, "newNotes")
			if err != nil {
				return nil, err
			}
			return events.PurchasedFromNotesUpdated{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				VendorID:    vendorID,
				OldNotes:    oldNotes,
				NewNotes:    newNotes,
			}, nil
		},
		"PurchasedFromRelationshipCleared": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			vendorID, err := repository.GetRequiredString(data, "vendorId")
			if err != nil {
				return nil, err
			}
			return events.PurchasedFromRelationshipCleared{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				VendorID:    vendorID,
			}, nil
		},
		"BuiltByRelationshipSet": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			teamID, err := repository.GetRequiredString(data, "teamId")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetRequiredString(data, "notes")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}
			return events.BuiltByRelationshipSet{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				TeamID:      teamID,
				Notes:       notes,
				LinkedAt:    linkedAt,
			}, nil
		},
		"BuiltByRelationshipReplaced": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			oldTeamID, err := repository.GetRequiredString(data, "oldTeamId")
			if err != nil {
				return nil, err
			}
			newTeamID, err := repository.GetRequiredString(data, "newTeamId")
			if err != nil {
				return nil, err
			}
			notes, err := repository.GetRequiredString(data, "notes")
			if err != nil {
				return nil, err
			}
			linkedAt, err := repository.GetRequiredTime(data, "linkedAt")
			if err != nil {
				return nil, err
			}
			return events.BuiltByRelationshipReplaced{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				OldTeamID:   oldTeamID,
				NewTeamID:   newTeamID,
				Notes:       notes,
				LinkedAt:    linkedAt,
			}, nil
		},
		"BuiltByNotesUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			teamID, err := repository.GetRequiredString(data, "teamId")
			if err != nil {
				return nil, err
			}
			oldNotes, err := repository.GetRequiredString(data, "oldNotes")
			if err != nil {
				return nil, err
			}
			newNotes, err := repository.GetRequiredString(data, "newNotes")
			if err != nil {
				return nil, err
			}
			return events.BuiltByNotesUpdated{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				TeamID:      teamID,
				OldNotes:    oldNotes,
				NewNotes:    newNotes,
			}, nil
		},
		"BuiltByRelationshipCleared": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			teamID, err := repository.GetRequiredString(data, "teamId")
			if err != nil {
				return nil, err
			}
			return events.BuiltByRelationshipCleared{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				TeamID:      teamID,
			}, nil
		},
		"ComponentOriginsDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			componentID, err := repository.GetRequiredString(data, "componentId")
			if err != nil {
				return nil, err
			}
			deletedAt, err := repository.GetRequiredTime(data, "deletedAt")
			if err != nil {
				return nil, err
			}
			return events.ComponentOriginsDeleted{
				BaseEvent:   domain.NewBaseEvent(componentID),
				ComponentID: componentID,
				DeletedAt:   deletedAt,
			}, nil
		},
	},
)
