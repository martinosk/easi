package repositories

import (
	"errors"

	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrBusinessDomainNotFound = errors.New("business domain not found")

type BusinessDomainRepository struct {
	*repository.EventSourcedRepository[*aggregates.BusinessDomain]
}

func NewBusinessDomainRepository(eventStore eventstore.EventStore) *BusinessDomainRepository {
	return &BusinessDomainRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			businessDomainEventDeserializers,
			aggregates.LoadBusinessDomainFromHistory,
			ErrBusinessDomainNotFound,
		),
	}
}

var businessDomainEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"BusinessDomainCreated": repository.JSONDeserializer[events.BusinessDomainCreated],
		"BusinessDomainUpdated": repository.JSONDeserializer[events.BusinessDomainUpdated],
		"BusinessDomainDeleted": repository.JSONDeserializer[events.BusinessDomainDeleted],
	},
)
