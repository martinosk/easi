package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type SubjectAttributeSchemaStore interface {
	Insert(ctx context.Context, record readmodels.SubjectAttributeSchemaRecord) error
	GetByID(ctx context.Context, id string) (*readmodels.SubjectAttributeSchemaRecord, error)
	Update(ctx context.Context, params readmodels.UpdateSchemaParams) error
}

type SubjectAttributeSchemaProjector struct {
	store SubjectAttributeSchemaStore
}

func NewSubjectAttributeSchemaProjector(store SubjectAttributeSchemaStore) *SubjectAttributeSchemaProjector {
	return &SubjectAttributeSchemaProjector{store: store}
}

func (p *SubjectAttributeSchemaProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event data: %w", event.EventType(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *SubjectAttributeSchemaProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType == events.TypeSubjectAttributeSchemaCreated {
		return p.handleCreated(ctx, eventData)
	}
	mutation, found := attributeMutations[eventType]
	if !found {
		return nil
	}
	return p.applyMutation(ctx, eventType, eventData, mutation)
}

func (p *SubjectAttributeSchemaProjector) handleCreated(ctx context.Context, eventData []byte) error {
	var event events.SubjectAttributeSchemaCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal SubjectAttributeSchemaCreated event: %w", err)
	}
	return p.store.Insert(ctx, readmodels.SubjectAttributeSchemaRecord{
		ID:          event.ID,
		TenantID:    event.TenantID,
		SubjectType: event.SubjectType,
		Attributes:  []readmodels.SubjectAttributeRecord{},
		Version:     1,
		CreatedAt:   event.CreatedAt,
		ModifiedAt:  event.CreatedAt,
		ModifiedBy:  event.CreatedBy,
	})
}

type attributeMutation func(attributes []readmodels.SubjectAttributeRecord, eventData []byte) ([]readmodels.SubjectAttributeRecord, error)

var attributeMutations = map[string]attributeMutation{
	events.TypeSubjectAttributeDefined:       mutate(applyDefined),
	events.TypeSubjectAttributeRenamed:       mutate(applyRenamed),
	events.TypeSubjectAttributeRetired:       mutate(applyRetired),
	events.TypeSubjectAttributeReactivated:   mutate(applyReactivated),
	events.TypeSubjectAttributeOptionAdded:   mutate(applyOptionAdded),
	events.TypeSubjectAttributeOptionRetired: mutate(applyOptionRetired),
	events.TypeSubjectAttributeBoundsChanged: mutate(applyBoundsChanged),
}

func mutate[E any](apply func(attributes []readmodels.SubjectAttributeRecord, event *E) []readmodels.SubjectAttributeRecord) attributeMutation {
	return func(attributes []readmodels.SubjectAttributeRecord, eventData []byte) ([]readmodels.SubjectAttributeRecord, error) {
		var event E
		if err := json.Unmarshal(eventData, &event); err != nil {
			return attributes, fmt.Errorf("unmarshal event: %w", err)
		}
		return apply(attributes, &event), nil
	}
}

func (p *SubjectAttributeSchemaProjector) applyMutation(ctx context.Context, eventType string, eventData []byte, mutation attributeMutation) error {
	var base events.SchemaEventBase
	if err := json.Unmarshal(eventData, &base); err != nil {
		return fmt.Errorf("unmarshal %s event base: %w", eventType, err)
	}
	record, err := p.store.GetByID(ctx, base.ID)
	if err != nil {
		return err
	}
	if record == nil {
		slog.WarnContext(ctx, "subject attribute schema not found for projection", "schemaID", base.ID, "eventType", eventType)
		return nil
	}
	attributes, err := mutation(record.Attributes, eventData)
	if err != nil {
		return fmt.Errorf("apply %s to subject attribute schema %s: %w", eventType, base.ID, err)
	}
	return p.store.Update(ctx, readmodels.UpdateSchemaParams{
		ID:         base.ID,
		Attributes: attributes,
		Version:    base.Version,
		ModifiedAt: base.ModifiedAt,
		ModifiedBy: base.ModifiedBy,
	})
}

func applyDefined(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeDefined) []readmodels.SubjectAttributeRecord {
	options := make([]readmodels.AttributeOptionRecord, len(event.Options))
	for i, option := range event.Options {
		options[i] = readmodels.AttributeOptionRecord{ID: option.ID, Label: option.Label, Active: option.Active}
	}
	if len(options) == 0 {
		options = nil
	}
	return append(attributes, readmodels.SubjectAttributeRecord{
		ID:       event.AttributeID,
		Name:     event.Name,
		Type:     event.AttributeType,
		HelpText: event.HelpText,
		Active:   true,
		Options:  options,
		Min:      event.Min,
		Max:      event.Max,
	})
}

func applyRenamed(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeRenamed) []readmodels.SubjectAttributeRecord {
	return updateAttribute(attributes, event.AttributeID, func(attribute *readmodels.SubjectAttributeRecord) {
		attribute.Name = event.NewName
		attribute.HelpText = event.NewHelpText
	})
}

func applyRetired(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeRetired) []readmodels.SubjectAttributeRecord {
	return updateAttribute(attributes, event.AttributeID, func(attribute *readmodels.SubjectAttributeRecord) { attribute.Active = false })
}

func applyReactivated(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeReactivated) []readmodels.SubjectAttributeRecord {
	return updateAttribute(attributes, event.AttributeID, func(attribute *readmodels.SubjectAttributeRecord) { attribute.Active = true })
}

func applyOptionAdded(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeOptionAdded) []readmodels.SubjectAttributeRecord {
	return updateAttribute(attributes, event.AttributeID, func(attribute *readmodels.SubjectAttributeRecord) {
		attribute.Options = append(attribute.Options, readmodels.AttributeOptionRecord{ID: event.OptionID, Label: event.Label, Active: true})
	})
}

func applyOptionRetired(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeOptionRetired) []readmodels.SubjectAttributeRecord {
	return updateAttribute(attributes, event.AttributeID, func(attribute *readmodels.SubjectAttributeRecord) {
		options := make([]readmodels.AttributeOptionRecord, len(attribute.Options))
		copy(options, attribute.Options)
		for i := range options {
			if options[i].ID == event.OptionID {
				options[i].Active = false
			}
		}
		attribute.Options = options
	})
}

func applyBoundsChanged(attributes []readmodels.SubjectAttributeRecord, event *events.SubjectAttributeBoundsChanged) []readmodels.SubjectAttributeRecord {
	return updateAttribute(attributes, event.AttributeID, func(attribute *readmodels.SubjectAttributeRecord) {
		attribute.Min = event.Min
		attribute.Max = event.Max
	})
}

func updateAttribute(attributes []readmodels.SubjectAttributeRecord, id string, modify func(*readmodels.SubjectAttributeRecord)) []readmodels.SubjectAttributeRecord {
	updated := make([]readmodels.SubjectAttributeRecord, len(attributes))
	copy(updated, attributes)
	for i := range updated {
		if updated[i].ID == id {
			modify(&updated[i])
			break
		}
	}
	return updated
}
