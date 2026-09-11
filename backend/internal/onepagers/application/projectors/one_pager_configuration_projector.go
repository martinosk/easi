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
			CustomFields: []readmodels.FieldRequirementRecord{},
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
	events.TypeCustomFieldDefined:             mutate(applyLegacyCustomFieldDefined),
	events.TypeCustomFieldRenamed:             mutate(applyMetadataOnly[events.CustomFieldRenamed]),
	events.TypeCustomFieldRequirementChanged:  mutate(applyCustomFieldRequirementChanged),
	events.TypeCustomFieldRetired:             mutate(applyCustomFieldRetired),
	events.TypeCustomFieldReactivated:         mutate(applyCustomFieldReactivated),
	events.TypeCustomFieldIncluded:            mutate(applyCustomFieldIncluded),
	events.TypeCustomFieldExcluded:            mutate(applyCustomFieldExcluded),
	events.TypeBuiltInFieldIncluded:           mutate(applyBuiltInFieldIncluded),
	events.TypeBuiltInFieldExcluded:           mutate(applyBuiltInFieldExcluded),
	events.TypeBuiltInFieldRequirementChanged: mutate(applyBuiltInFieldRequirementChanged),
	events.TypeOnePagerFieldsReordered:        mutate(applyFieldsReordered),
	events.TypeSelectionOptionAdded:           mutate(applyMetadataOnly[events.SelectionOptionAdded]),
	events.TypeSelectionOptionRetired:         mutate(applyMetadataOnly[events.SelectionOptionRetired]),
	events.TypeNumberFieldBoundsChanged:       mutate(applyMetadataOnly[events.NumberFieldBoundsChanged]),
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

func applyMetadataOnly[E any](doc readmodels.ConfigurationDocument, _ *E) readmodels.ConfigurationDocument {
	return doc
}

func applyLegacyCustomFieldDefined(doc readmodels.ConfigurationDocument, event *events.CustomFieldDefined) readmodels.ConfigurationDocument {
	doc = setRequirement(doc, event.FieldID, event.Required)
	return includeCustomField(doc, event.FieldID)
}

func applyCustomFieldRequirementChanged(doc readmodels.ConfigurationDocument, event *events.CustomFieldRequirementChanged) readmodels.ConfigurationDocument {
	return setRequirement(doc, event.FieldID, event.Required)
}

func applyCustomFieldRetired(doc readmodels.ConfigurationDocument, event *events.CustomFieldRetired) readmodels.ConfigurationDocument {
	return excludeCustomField(doc, event.FieldID)
}

func applyCustomFieldReactivated(doc readmodels.ConfigurationDocument, event *events.CustomFieldReactivated) readmodels.ConfigurationDocument {
	return includeCustomField(doc, event.FieldID)
}

func applyCustomFieldIncluded(doc readmodels.ConfigurationDocument, event *events.CustomFieldIncluded) readmodels.ConfigurationDocument {
	return includeCustomField(doc, event.FieldID)
}

func applyCustomFieldExcluded(doc readmodels.ConfigurationDocument, event *events.CustomFieldExcluded) readmodels.ConfigurationDocument {
	return excludeCustomField(doc, event.FieldID)
}

func includeCustomField(doc readmodels.ConfigurationDocument, fieldID string) readmodels.ConfigurationDocument {
	ref := readmodels.FieldRefRecord{Kind: "custom", ID: fieldID}
	doc.DisplayOrder = append(removeRef(doc.DisplayOrder, ref), ref)
	if _, found := requirementIndex(doc.CustomFields, fieldID); !found {
		doc.CustomFields = append(copyRequirements(doc.CustomFields), readmodels.FieldRequirementRecord{ID: fieldID})
	}
	return doc
}

func excludeCustomField(doc readmodels.ConfigurationDocument, fieldID string) readmodels.ConfigurationDocument {
	doc.DisplayOrder = removeRef(doc.DisplayOrder, readmodels.FieldRefRecord{Kind: "custom", ID: fieldID})
	return doc
}

func setRequirement(doc readmodels.ConfigurationDocument, fieldID string, required bool) readmodels.ConfigurationDocument {
	doc.CustomFields = withRequirement(doc.CustomFields, fieldID, required)
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
	doc.BuiltInFields = withRequirement(doc.BuiltInFields, event.EntryID, event.Required)
	return doc
}

func withRequirement(records []readmodels.FieldRequirementRecord, id string, required bool) []readmodels.FieldRequirementRecord {
	updated := copyRequirements(records)
	if index, found := requirementIndex(updated, id); found {
		updated[index].Required = required
		return updated
	}
	return append(updated, readmodels.FieldRequirementRecord{ID: id, Required: required})
}

func requirementIndex(records []readmodels.FieldRequirementRecord, id string) (int, bool) {
	for i, record := range records {
		if record.ID == id {
			return i, true
		}
	}
	return 0, false
}

func copyRequirements(records []readmodels.FieldRequirementRecord) []readmodels.FieldRequirementRecord {
	copied := make([]readmodels.FieldRequirementRecord, len(records))
	copy(copied, records)
	return copied
}

func applyFieldsReordered(doc readmodels.ConfigurationDocument, event *events.OnePagerFieldsReordered) readmodels.ConfigurationDocument {
	order := make([]readmodels.FieldRefRecord, len(event.Order))
	for i, ref := range event.Order {
		order[i] = readmodels.FieldRefRecord{Kind: ref.Kind, ID: ref.ID}
	}
	doc.DisplayOrder = order
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
