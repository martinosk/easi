package repositories

import (
	"errors"

	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/shared/infrastructure/repository"
)

var ErrOnePagerConfigurationNotFound = errors.New("one-pager configuration not found")

type OnePagerConfigurationRepository struct {
	*repository.EventSourcedRepository[*aggregates.OnePagerConfiguration]
}

func NewOnePagerConfigurationRepository(eventStore eventstore.EventStore) *OnePagerConfigurationRepository {
	return &OnePagerConfigurationRepository{
		EventSourcedRepository: repository.NewEventSourcedRepository(
			eventStore,
			onePagerEventDeserializers,
			aggregates.LoadOnePagerConfigurationFromHistory,
			ErrOnePagerConfigurationNotFound,
		),
	}
}

var onePagerEventDeserializers = repository.NewEventDeserializers(
	map[string]repository.EventDeserializerFunc{
		"OnePagerConfigurationCreated":   repository.JSONDeserializer[events.OnePagerConfigurationCreated],
		"CustomFieldDefined":             repository.JSONDeserializer[events.CustomFieldDefined],
		"CustomFieldRenamed":             repository.JSONDeserializer[events.CustomFieldRenamed],
		"CustomFieldRequirementChanged":  repository.JSONDeserializer[events.CustomFieldRequirementChanged],
		"CustomFieldRetired":             repository.JSONDeserializer[events.CustomFieldRetired],
		"CustomFieldReactivated":         repository.JSONDeserializer[events.CustomFieldReactivated],
		"BuiltInFieldIncluded":           repository.JSONDeserializer[events.BuiltInFieldIncluded],
		"BuiltInFieldExcluded":           repository.JSONDeserializer[events.BuiltInFieldExcluded],
		"BuiltInFieldRequirementChanged": repository.JSONDeserializer[events.BuiltInFieldRequirementChanged],
		"OnePagerFieldsReordered":        repository.JSONDeserializer[events.OnePagerFieldsReordered],
		"SelectionOptionAdded":           repository.JSONDeserializer[events.SelectionOptionAdded],
		"SelectionOptionRetired":         repository.JSONDeserializer[events.SelectionOptionRetired],
		"NumberFieldBoundsChanged":       repository.JSONDeserializer[events.NumberFieldBoundsChanged],
	},
)
