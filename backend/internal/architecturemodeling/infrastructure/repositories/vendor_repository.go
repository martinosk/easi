package repositories

import (
	"errors"

	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
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
		"VendorCreated": repository.JSONDeserializer[events.VendorCreated],
		"VendorUpdated": repository.JSONDeserializer[events.VendorUpdated],
		"VendorDeleted": repository.JSONDeserializer[events.VendorDeleted],
	},
)
