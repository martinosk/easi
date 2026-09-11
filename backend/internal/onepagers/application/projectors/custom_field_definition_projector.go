package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type CustomFieldDefinitionStore interface {
	Save(ctx context.Context, definition readmodels.SubjectDefinition) error
	Get(ctx context.Context, fieldID string) (*readmodels.CustomFieldRecord, error)
}

type CustomFieldDefinitionProjector struct {
	cache CustomFieldDefinitionStore
}

func NewCustomFieldDefinitionProjector(cache CustomFieldDefinitionStore) *CustomFieldDefinitionProjector {
	return &CustomFieldDefinitionProjector{cache: cache}
}

func CustomFieldDefinitionEventTypes() []string {
	return mmPL.SubjectAttributeEventTypes()
}

func (p *CustomFieldDefinitionProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *CustomFieldDefinitionProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType == mmPL.SubjectAttributeDefined {
		return p.handleDefined(ctx, eventData)
	}
	mutation, found := definitionMutations[eventType]
	if !found {
		return nil
	}
	return p.applyMutation(ctx, eventType, eventData, mutation)
}

func (p *CustomFieldDefinitionProjector) handleDefined(ctx context.Context, eventData []byte) error {
	var payload mmPL.SubjectAttributeDefinedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal SubjectAttributeDefined: %w", err)
	}
	options := make([]readmodels.OptionRecord, len(payload.Options))
	for i, option := range payload.Options {
		options[i] = readmodels.OptionRecord{ID: option.ID, Label: option.Label, Active: option.Active}
	}
	if len(options) == 0 {
		options = nil
	}
	return p.cache.Save(ctx, readmodels.SubjectDefinition{
		SubjectType: payload.SubjectType,
		Field: readmodels.CustomFieldRecord{
			ID:       payload.AttributeID,
			Name:     payload.Name,
			Type:     payload.AttributeType,
			HelpText: payload.HelpText,
			Active:   true,
			Options:  options,
			Min:      payload.Min,
			Max:      payload.Max,
		},
	})
}

type definitionMutation func(field readmodels.CustomFieldRecord, eventData []byte) (readmodels.CustomFieldRecord, error)

var definitionMutations = map[string]definitionMutation{
	mmPL.SubjectAttributeRenamed:       mutateDefinition(applyDefinitionRenamed),
	mmPL.SubjectAttributeRetired:       mutateDefinition(applyDefinitionRetired),
	mmPL.SubjectAttributeReactivated:   mutateDefinition(applyDefinitionReactivated),
	mmPL.SubjectAttributeOptionAdded:   mutateDefinition(applyDefinitionOptionAdded),
	mmPL.SubjectAttributeOptionRetired: mutateDefinition(applyDefinitionOptionRetired),
	mmPL.SubjectAttributeBoundsChanged: mutateDefinition(applyDefinitionBoundsChanged),
}

func mutateDefinition[E any](apply func(field readmodels.CustomFieldRecord, event *E) readmodels.CustomFieldRecord) definitionMutation {
	return func(field readmodels.CustomFieldRecord, eventData []byte) (readmodels.CustomFieldRecord, error) {
		var event E
		if err := json.Unmarshal(eventData, &event); err != nil {
			return field, fmt.Errorf("unmarshal event: %w", err)
		}
		return apply(field, &event), nil
	}
}

func (p *CustomFieldDefinitionProjector) applyMutation(ctx context.Context, eventType string, eventData []byte, mutation definitionMutation) error {
	var base mmPL.SubjectAttributeEventBase
	if err := json.Unmarshal(eventData, &base); err != nil {
		return fmt.Errorf("unmarshal %s event base: %w", eventType, err)
	}
	field, err := p.cache.Get(ctx, base.AttributeID)
	if err != nil {
		return err
	}
	if field == nil {
		slog.WarnContext(ctx, "custom field definition not cached", "attributeID", base.AttributeID, "eventType", eventType)
		return nil
	}
	mutated, err := mutation(*field, eventData)
	if err != nil {
		return fmt.Errorf("apply %s to custom field definition %s: %w", eventType, base.AttributeID, err)
	}
	return p.cache.Save(ctx, readmodels.SubjectDefinition{SubjectType: base.SubjectType, Field: mutated})
}

func applyDefinitionRenamed(field readmodels.CustomFieldRecord, event *mmPL.SubjectAttributeRenamedPayload) readmodels.CustomFieldRecord {
	field.Name = event.NewName
	field.HelpText = event.NewHelpText
	return field
}

func applyDefinitionRetired(field readmodels.CustomFieldRecord, _ *mmPL.SubjectAttributeRetiredPayload) readmodels.CustomFieldRecord {
	field.Active = false
	return field
}

func applyDefinitionReactivated(field readmodels.CustomFieldRecord, _ *mmPL.SubjectAttributeReactivatedPayload) readmodels.CustomFieldRecord {
	field.Active = true
	return field
}

func applyDefinitionOptionAdded(field readmodels.CustomFieldRecord, event *mmPL.SubjectAttributeOptionAddedPayload) readmodels.CustomFieldRecord {
	field.Options = append(copyOptionRecords(field.Options), readmodels.OptionRecord{ID: event.OptionID, Label: event.Label, Active: true})
	return field
}

func applyDefinitionOptionRetired(field readmodels.CustomFieldRecord, event *mmPL.SubjectAttributeOptionRetiredPayload) readmodels.CustomFieldRecord {
	options := copyOptionRecords(field.Options)
	for i := range options {
		if options[i].ID == event.OptionID {
			options[i].Active = false
		}
	}
	field.Options = options
	return field
}

func applyDefinitionBoundsChanged(field readmodels.CustomFieldRecord, event *mmPL.SubjectAttributeBoundsChangedPayload) readmodels.CustomFieldRecord {
	field.Min = event.Min
	field.Max = event.Max
	return field
}

func copyOptionRecords(options []readmodels.OptionRecord) []readmodels.OptionRecord {
	copied := make([]readmodels.OptionRecord, len(options))
	copy(copied, options)
	return copied
}
