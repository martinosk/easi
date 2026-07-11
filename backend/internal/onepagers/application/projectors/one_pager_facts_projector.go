package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type FactsStore interface {
	Upsert(ctx context.Context, record readmodels.FactRecord) error
	Clear(ctx context.Context, params readmodels.ClearFactParams) error
	DeleteForSubject(ctx context.Context, subject readmodels.SubjectKey) error
}

type OnePagerFactsProjector struct {
	store FactsStore
}

func NewOnePagerFactsProjector(store FactsStore) *OnePagerFactsProjector {
	return &OnePagerFactsProjector{store: store}
}

func (p *OnePagerFactsProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event data: %w", event.EventType(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *OnePagerFactsProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case events.TypeFieldValueRecorded:
		return p.projectRecorded(ctx, eventData)
	case events.TypeFieldValueCleared:
		return p.projectCleared(ctx, eventData)
	case events.TypeOnePagerFactsArchived:
		return p.projectArchived(ctx, eventData)
	}
	return nil
}

func (p *OnePagerFactsProjector) projectRecorded(ctx context.Context, eventData []byte) error {
	var event events.FieldValueRecorded
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal FieldValueRecorded event: %w", err)
	}

	value, err := valueobjects.FieldValueFromEnvelope(event.Value)
	if err != nil {
		return fmt.Errorf("decode field value for field %s: %w", event.FieldID, err)
	}

	envelope := event.Value
	return p.store.Upsert(ctx, readmodels.FactRecord{
		FactsID:     event.ID,
		TenantID:    event.TenantID,
		SubjectType: event.SubjectType,
		SubjectID:   event.SubjectID,
		FieldID:     event.FieldID,
		Value:       &envelope,
		ValueType:   envelope.Type,
		DisplayText: valueobjects.DisplayText(value),
		ModifiedAt:  event.ModifiedAt,
		ModifiedBy:  event.ModifiedBy,
	})
}

func (p *OnePagerFactsProjector) projectCleared(ctx context.Context, eventData []byte) error {
	var event events.FieldValueCleared
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal FieldValueCleared event: %w", err)
	}

	return p.store.Clear(ctx, readmodels.ClearFactParams{
		SubjectType: event.SubjectType,
		SubjectID:   event.SubjectID,
		FieldID:     event.FieldID,
		ModifiedAt:  event.ModifiedAt,
		ModifiedBy:  event.ModifiedBy,
	})
}

func (p *OnePagerFactsProjector) projectArchived(ctx context.Context, eventData []byte) error {
	var event events.OnePagerFactsArchived
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal OnePagerFactsArchived event: %w", err)
	}

	return p.store.DeleteForSubject(ctx, readmodels.SubjectKey{SubjectType: event.SubjectType, SubjectID: event.SubjectID})
}
