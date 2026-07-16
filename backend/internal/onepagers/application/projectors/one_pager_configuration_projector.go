package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ConfigurationStore interface {
	Insert(ctx context.Context, record readmodels.ConfigurationRecord) error
	GetByID(ctx context.Context, id string) (*readmodels.ConfigurationRecord, error)
	Update(ctx context.Context, params readmodels.UpdateParams) error
}

type OnePagerConfigurationProjector struct {
	store ConfigurationStore
}

func NewOnePagerConfigurationProjector(store ConfigurationStore) *OnePagerConfigurationProjector {
	return &OnePagerConfigurationProjector{store: store}
}

func (p *OnePagerConfigurationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event data: %w", event.EventType(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *OnePagerConfigurationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	if eventType == events.TypeOnePagerConfigurationCreated {
		return p.handleCreated(ctx, eventData)
	}
	mutation, found := documentMutations[eventType]
	if !found {
		return nil
	}
	return p.applyMutation(ctx, eventType, eventData, mutation)
}

func (p *OnePagerConfigurationProjector) handleCreated(ctx context.Context, eventData []byte) error {
	var event events.OnePagerConfigurationCreated
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal OnePagerConfigurationCreated event: %w", err)
	}

	displayOrder := make([]readmodels.FieldRefRecord, len(event.BuiltIns))
	for i, entryID := range event.BuiltIns {
		displayOrder[i] = readmodels.FieldRefRecord{Kind: "builtIn", ID: entryID}
	}

	return p.store.Insert(ctx, readmodels.ConfigurationRecord{
		ID:          event.ID,
		TenantID:    event.TenantID,
		SubjectType: event.SubjectType,
		Document: readmodels.ConfigurationDocument{
			CustomFields: []readmodels.CustomFieldRecord{},
			DisplayOrder: displayOrder,
		},
		Version:    1,
		CreatedAt:  event.CreatedAt,
		ModifiedAt: event.CreatedAt,
		ModifiedBy: event.CreatedBy,
	})
}

type documentMutation func(doc readmodels.ConfigurationDocument, eventData []byte) (readmodels.ConfigurationDocument, error)

var documentMutations = map[string]documentMutation{
	events.TypeCustomFieldDefined:             mutate(applyCustomFieldDefined),
	events.TypeCustomFieldRenamed:             mutate(applyCustomFieldRenamed),
	events.TypeCustomFieldRequirementChanged:  mutate(applyCustomFieldRequirementChanged),
	events.TypeCustomFieldRetired:             mutate(applyCustomFieldRetired),
	events.TypeCustomFieldReactivated:         mutate(applyCustomFieldReactivated),
	events.TypeBuiltInFieldIncluded:           mutate(applyBuiltInFieldIncluded),
	events.TypeBuiltInFieldExcluded:           mutate(applyBuiltInFieldExcluded),
	events.TypeBuiltInFieldRequirementChanged: mutate(applyBuiltInFieldRequirementChanged),
	events.TypeOnePagerFieldsReordered:        mutate(applyFieldsReordered),
	events.TypeSelectionOptionAdded:           mutate(applySelectionOptionAdded),
	events.TypeSelectionOptionRetired:         mutate(applySelectionOptionRetired),
	events.TypeNumberFieldBoundsChanged:       mutate(applyNumberFieldBoundsChanged),
}

func mutate[E any](apply func(doc readmodels.ConfigurationDocument, event *E) readmodels.ConfigurationDocument) documentMutation {
	return func(doc readmodels.ConfigurationDocument, eventData []byte) (readmodels.ConfigurationDocument, error) {
		var event E
		if err := json.Unmarshal(eventData, &event); err != nil {
			return doc, fmt.Errorf("unmarshal event: %w", err)
		}
		return apply(doc, &event), nil
	}
}

func (p *OnePagerConfigurationProjector) applyMutation(ctx context.Context, eventType string, eventData []byte, mutation documentMutation) error {
	var base events.ConfigurationEventBase
	if err := json.Unmarshal(eventData, &base); err != nil {
		return fmt.Errorf("unmarshal %s event base: %w", eventType, err)
	}

	record, err := p.store.GetByID(ctx, base.ID)
	if err != nil {
		return err
	}
	if record == nil {
		slog.WarnContext(ctx, "one-pager configuration not found for projection", "configID", base.ID, "eventType", eventType)
		return nil
	}

	document, err := mutation(record.Document, eventData)
	if err != nil {
		return fmt.Errorf("apply %s to one-pager configuration %s: %w", eventType, base.ID, err)
	}

	return p.store.Update(ctx, readmodels.UpdateParams{
		ID:         base.ID,
		Document:   document,
		Version:    base.Version,
		ModifiedAt: base.ModifiedAt,
		ModifiedBy: base.ModifiedBy,
	})
}

func applyCustomFieldDefined(doc readmodels.ConfigurationDocument, event *events.CustomFieldDefined) readmodels.ConfigurationDocument {
	options := make([]readmodels.OptionRecord, len(event.Options))
	for i, option := range event.Options {
		options[i] = readmodels.OptionRecord{ID: option.ID, Label: option.Label, Active: option.Active}
	}
	doc.CustomFields = append(doc.CustomFields, readmodels.CustomFieldRecord{
		ID:       event.FieldID,
		Name:     event.Name,
		Type:     event.FieldType,
		Required: event.Required,
		HelpText: event.HelpText,
		Active:   true,
		Options:  options,
	})
	doc.DisplayOrder = append(doc.DisplayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: event.FieldID})
	return doc
}

func applyCustomFieldRenamed(doc readmodels.ConfigurationDocument, event *events.CustomFieldRenamed) readmodels.ConfigurationDocument {
	return updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		field.Name = event.NewName
		field.HelpText = event.NewHelpText
	})
}

func applyCustomFieldRequirementChanged(doc readmodels.ConfigurationDocument, event *events.CustomFieldRequirementChanged) readmodels.ConfigurationDocument {
	return updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		field.Required = event.Required
	})
}

func applyCustomFieldRetired(doc readmodels.ConfigurationDocument, event *events.CustomFieldRetired) readmodels.ConfigurationDocument {
	doc = updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		field.Active = false
	})
	doc.DisplayOrder = removeRef(doc.DisplayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: event.FieldID})
	return doc
}

func applyCustomFieldReactivated(doc readmodels.ConfigurationDocument, event *events.CustomFieldReactivated) readmodels.ConfigurationDocument {
	doc = updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		field.Active = true
	})
	doc.DisplayOrder = append(doc.DisplayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: event.FieldID})
	return doc
}

func applyBuiltInFieldIncluded(doc readmodels.ConfigurationDocument, event *events.BuiltInFieldIncluded) readmodels.ConfigurationDocument {
	doc.DisplayOrder = append(doc.DisplayOrder, readmodels.FieldRefRecord{Kind: "builtIn", ID: event.EntryID})
	return doc
}

func applyBuiltInFieldExcluded(doc readmodels.ConfigurationDocument, event *events.BuiltInFieldExcluded) readmodels.ConfigurationDocument {
	doc.DisplayOrder = removeRef(doc.DisplayOrder, readmodels.FieldRefRecord{Kind: "builtIn", ID: event.EntryID})
	return doc
}

func applyBuiltInFieldRequirementChanged(doc readmodels.ConfigurationDocument, event *events.BuiltInFieldRequirementChanged) readmodels.ConfigurationDocument {
	fields := make([]readmodels.BuiltInFieldRecord, len(doc.BuiltInFields))
	copy(fields, doc.BuiltInFields)
	for i := range fields {
		if fields[i].ID == event.EntryID {
			fields[i].Required = event.Required
			doc.BuiltInFields = fields
			return doc
		}
	}
	fields = append(fields, readmodels.BuiltInFieldRecord{ID: event.EntryID, Required: event.Required})
	doc.BuiltInFields = fields
	return doc
}

func applyFieldsReordered(doc readmodels.ConfigurationDocument, event *events.OnePagerFieldsReordered) readmodels.ConfigurationDocument {
	order := make([]readmodels.FieldRefRecord, len(event.Order))
	for i, ref := range event.Order {
		order[i] = readmodels.FieldRefRecord{Kind: ref.Kind, ID: ref.ID}
	}
	doc.DisplayOrder = order
	return doc
}

func applySelectionOptionAdded(doc readmodels.ConfigurationDocument, event *events.SelectionOptionAdded) readmodels.ConfigurationDocument {
	return updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		field.Options = append(field.Options, readmodels.OptionRecord{ID: event.OptionID, Label: event.Label, Active: true})
	})
}

func applySelectionOptionRetired(doc readmodels.ConfigurationDocument, event *events.SelectionOptionRetired) readmodels.ConfigurationDocument {
	return updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		for i := range field.Options {
			if field.Options[i].ID == event.OptionID {
				field.Options[i].Active = false
				return
			}
		}
	})
}

func applyNumberFieldBoundsChanged(doc readmodels.ConfigurationDocument, event *events.NumberFieldBoundsChanged) readmodels.ConfigurationDocument {
	return updateField(doc, event.FieldID, func(field *readmodels.CustomFieldRecord) {
		field.Min = event.Min
		field.Max = event.Max
	})
}

func updateField(doc readmodels.ConfigurationDocument, fieldID string, modify func(*readmodels.CustomFieldRecord)) readmodels.ConfigurationDocument {
	fields := make([]readmodels.CustomFieldRecord, len(doc.CustomFields))
	copy(fields, doc.CustomFields)
	for i := range fields {
		if fields[i].ID == fieldID {
			modify(&fields[i])
			break
		}
	}
	doc.CustomFields = fields
	return doc
}

func removeRef(order []readmodels.FieldRefRecord, target readmodels.FieldRefRecord) []readmodels.FieldRefRecord {
	result := make([]readmodels.FieldRefRecord, 0, len(order))
	for _, ref := range order {
		if ref != target {
			result = append(result, ref)
		}
	}
	return result
}
