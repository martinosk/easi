package repositories

import (
	"context"
	"errors"
	"time"

	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/events"
	"easi/backend/internal/infrastructure/eventstore"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrImportSessionNotFound = errors.New("import session not found")
)

type ImportSessionRepository struct {
	eventStore eventstore.EventStore
}

func NewImportSessionRepository(eventStore eventstore.EventStore) *ImportSessionRepository {
	return &ImportSessionRepository{
		eventStore: eventStore,
	}
}

func (r *ImportSessionRepository) Save(ctx context.Context, session *aggregates.ImportSession) error {
	uncommittedEvents := session.GetUncommittedChanges()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	err := r.eventStore.SaveEvents(ctx, session.ID(), uncommittedEvents, session.Version()-len(uncommittedEvents))
	if err != nil {
		return err
	}

	session.MarkChangesAsCommitted()
	return nil
}

func (r *ImportSessionRepository) GetByID(ctx context.Context, id string) (*aggregates.ImportSession, error) {
	storedEvents, err := r.eventStore.GetEvents(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(storedEvents) == 0 {
		return nil, ErrImportSessionNotFound
	}

	domainEvents := r.deserializeEvents(storedEvents)

	return aggregates.LoadImportSessionFromHistory(domainEvents)
}

func (r *ImportSessionRepository) deserializeEvents(storedEvents []domain.DomainEvent) []domain.DomainEvent {
	var domainEvents []domain.DomainEvent

	for _, event := range storedEvents {
		eventData := event.EventData()

		switch event.EventType() {
		case "ImportSessionCreated":
			id, _ := eventData["id"].(string)
			sourceFormat, _ := eventData["sourceFormat"].(string)
			businessDomainID, _ := eventData["businessDomainId"].(string)
			preview, _ := eventData["preview"].(map[string]interface{})
			parsedData, _ := eventData["parsedData"].(map[string]interface{})
			createdAtStr, _ := eventData["createdAt"].(string)
			createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)

			concreteEvent := events.NewImportSessionCreated(id, sourceFormat, businessDomainID, preview, parsedData)
			concreteEvent.CreatedAt = createdAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ImportStarted":
			id, _ := eventData["id"].(string)
			totalItems := getInt(eventData, "totalItems")
			startedAtStr, _ := eventData["startedAt"].(string)
			startedAt, _ := time.Parse(time.RFC3339Nano, startedAtStr)

			concreteEvent := events.NewImportStarted(id, totalItems)
			concreteEvent.StartedAt = startedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ImportProgressUpdated":
			id, _ := eventData["id"].(string)
			phase, _ := eventData["phase"].(string)
			totalItems := getInt(eventData, "totalItems")
			completedItems := getInt(eventData, "completedItems")
			updatedAtStr, _ := eventData["updatedAt"].(string)
			updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

			concreteEvent := events.NewImportProgressUpdated(id, phase, totalItems, completedItems)
			concreteEvent.UpdatedAt = updatedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ImportCompleted":
			id, _ := eventData["id"].(string)
			capabilitiesCreated := getInt(eventData, "capabilitiesCreated")
			componentsCreated := getInt(eventData, "componentsCreated")
			realizationsCreated := getInt(eventData, "realizationsCreated")
			domainAssignments := getInt(eventData, "domainAssignments")
			errorsData, _ := eventData["errors"].([]interface{})
			var errorsSlice []map[string]interface{}
			for _, e := range errorsData {
				if m, ok := e.(map[string]interface{}); ok {
					errorsSlice = append(errorsSlice, m)
				}
			}
			completedAtStr, _ := eventData["completedAt"].(string)
			completedAt, _ := time.Parse(time.RFC3339Nano, completedAtStr)

			concreteEvent := events.NewImportCompleted(id, capabilitiesCreated, componentsCreated, realizationsCreated, domainAssignments, errorsSlice)
			concreteEvent.CompletedAt = completedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ImportFailed":
			id, _ := eventData["id"].(string)
			reason, _ := eventData["reason"].(string)
			failedAtStr, _ := eventData["failedAt"].(string)
			failedAt, _ := time.Parse(time.RFC3339Nano, failedAtStr)

			concreteEvent := events.NewImportFailed(id, reason)
			concreteEvent.FailedAt = failedAt
			domainEvents = append(domainEvents, concreteEvent)

		case "ImportSessionCancelled":
			id, _ := eventData["id"].(string)
			cancelledAtStr, _ := eventData["cancelledAt"].(string)
			cancelledAt, _ := time.Parse(time.RFC3339Nano, cancelledAtStr)

			concreteEvent := events.NewImportSessionCancelled(id)
			concreteEvent.CancelledAt = cancelledAt
			domainEvents = append(domainEvents, concreteEvent)

		default:
			continue
		}
	}

	return domainEvents
}

func getInt(data map[string]interface{}, key string) int {
	if v, ok := data[key].(int); ok {
		return v
	}
	if v, ok := data[key].(float64); ok {
		return int(v)
	}
	return 0
}
