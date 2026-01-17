package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrVendorNotFound = errors.New("vendor not found")

type VendorRepository struct {
	*repository.EventSourcedRepository[*aggregates.Vendor]
}

func NewVendorRepository(eventStore eventstore.EventStore) *VendorRepository {
	return &VendorRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			vendorEventDeserializers,
			aggregates.LoadVendorFromHistory,
			ErrVendorNotFound,
		),
	}
}

var vendorEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"VendorCreated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			implementationPartner, _ := repository.GetOptionalString(data, "implementationPartner", "")
			notes, _ := repository.GetOptionalString(data, "notes", "")
			createdAt, err := repository.GetRequiredTime(data, "createdAt")
			if err != nil {
				return nil, err
			}

			evt := events.NewVendorCreated(id, name, implementationPartner, notes)
			evt.CreatedAt = createdAt
			return evt, nil
		},
		"VendorUpdated": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}
			implementationPartner, _ := repository.GetOptionalString(data, "implementationPartner", "")
			notes, _ := repository.GetOptionalString(data, "notes", "")

			return events.NewVendorUpdated(id, name, implementationPartner, notes), nil
		},
		"VendorDeleted": func(data map[string]interface{}) (domain.DomainEvent, error) {
			id, err := repository.GetRequiredString(data, "id")
			if err != nil {
				return nil, err
			}
			name, err := repository.GetRequiredString(data, "name")
			if err != nil {
				return nil, err
			}

			return events.NewVendorDeleted(id, name), nil
		},
	},
)
