package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	adPL "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	opevents "easi/backend/internal/onepagers/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type SubjectIndexStore interface {
	Upsert(ctx context.Context, record readmodels.SubjectIndexRecord) error
	Delete(ctx context.Context, subject readmodels.SubjectKey) error
	ApplySubjectChange(ctx context.Context, change readmodels.SubjectChange) error
	ApplyCompleteness(ctx context.Context, subjectType string, required int, filledBySubject map[string]int) error
	SubjectIDs(ctx context.Context, subjectType string) ([]string, error)
	MergeAttributes(ctx context.Context, subject readmodels.SubjectKey, attributes readmodels.SubjectAttributes) error
	ApplyExpertChange(ctx context.Context, subject readmodels.SubjectKey, expert readmodels.SubjectExpert, added bool) error
}

type CompletenessCounter interface {
	CountsForSubjects(ctx context.Context, subjectType string, subjectIDs []string) (int, map[string]int, error)
}

type ConfigurationLookup interface {
	GetByID(ctx context.Context, id string) (*readmodels.ConfigurationRecord, error)
}

type SubjectIndexProjector struct {
	store   SubjectIndexStore
	counter CompletenessCounter
	audit   ports.SubjectAuditReader
	configs ConfigurationLookup
}

func NewSubjectIndexProjector(store SubjectIndexStore, counter CompletenessCounter, audit ports.SubjectAuditReader, configs ConfigurationLookup) *SubjectIndexProjector {
	return &SubjectIndexProjector{store: store, counter: counter, audit: audit, configs: configs}
}

var subjectTypeByCreationEvent = map[string]string{
	capPL.CapabilityCreated:          "capability",
	adPL.EnterpriseCapabilityCreated: "enterprise-capability",
	amPL.ApplicationComponentCreated: "application",
	amPL.AcquiredEntityCreated:       "acquired-entity",
	amPL.VendorCreated:               "vendor",
	amPL.InternalTeamCreated:         "internal-team",
}

var subjectTypeByUpdateEvent = map[string]string{
	capPL.CapabilityUpdated:                "capability",
	capPL.CapabilityMetadataUpdated:        "capability",
	capPL.CapabilityExpertAdded:            "capability",
	capPL.CapabilityExpertRemoved:          "capability",
	adPL.EnterpriseCapabilityUpdated:       "enterprise-capability",
	amPL.ApplicationComponentUpdated:       "application",
	amPL.ApplicationComponentExpertAdded:   "application",
	amPL.ApplicationComponentExpertRemoved: "application",
	amPL.AcquiredEntityUpdated:             "acquired-entity",
	amPL.VendorUpdated:                     "vendor",
	amPL.InternalTeamUpdated:               "internal-team",
}

var factsEventTypes = map[string]struct{}{
	opevents.TypeFieldValueRecorded:    {},
	opevents.TypeFieldValueCleared:     {},
	opevents.TypeOnePagerFactsArchived: {},
}

func SubjectIndexEventTypes() []string {
	types := make([]string, 0)
	for eventType := range subjectTypeByCreationEvent {
		types = append(types, eventType)
	}
	for eventType := range subjectTypeByDeletionEvent {
		types = append(types, eventType)
	}
	for eventType := range subjectTypeByUpdateEvent {
		types = append(types, eventType)
	}
	for eventType := range factsEventTypes {
		types = append(types, eventType)
	}
	types = append(types, attributeOnlyEventTypes()...)
	return append(types, opevents.ConfigurationEventTypes()...)
}

func (p *SubjectIndexProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event data: %w", event.EventType(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), event.OccurredAt(), eventData)
}

type projectedEvent struct {
	subjectType string
	eventType   string
	occurredAt  time.Time
	data        []byte
}

func (e projectedEvent) about(subjectType string) projectedEvent {
	e.subjectType = subjectType
	return e
}

func (p *SubjectIndexProjector) ProjectEvent(ctx context.Context, eventType string, occurredAt time.Time, eventData []byte) error {
	event := projectedEvent{eventType: eventType, occurredAt: occurredAt, data: eventData}
	if handled, err := p.onAttributesOnlyChanged(ctx, event); handled || err != nil {
		return err
	}
	if subjectType, ok := subjectTypeByCreationEvent[eventType]; ok {
		return p.onCreated(ctx, event.about(subjectType))
	}
	if subjectType, ok := subjectTypeByDeletionEvent[eventType]; ok {
		return p.onDeleted(ctx, event.about(subjectType))
	}
	if subjectType, ok := subjectTypeByUpdateEvent[eventType]; ok {
		return p.onSubjectChanged(ctx, event.about(subjectType))
	}
	if _, ok := factsEventTypes[eventType]; ok {
		return p.onFactsChanged(ctx, eventData)
	}
	if isConfigurationEvent(eventType) {
		return p.onConfigurationChanged(ctx, eventData)
	}
	return nil
}

type subjectIdentity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type subjectChangeIdentity struct {
	ID           string `json:"id"`
	CapabilityID string `json:"capabilityId"`
	ComponentID  string `json:"componentId"`
	Name         string `json:"name"`
}

func (e subjectChangeIdentity) subjectID() string {
	if e.ID != "" {
		return e.ID
	}
	if e.CapabilityID != "" {
		return e.CapabilityID
	}
	return e.ComponentID
}

type factsIdentity struct {
	SubjectType string `json:"subjectType"`
	SubjectID   string `json:"subjectId"`
}

func (p *SubjectIndexProjector) onCreated(ctx context.Context, event projectedEvent) error {
	var created subjectIdentity
	if err := json.Unmarshal(event.data, &created); err != nil {
		return fmt.Errorf("unmarshal %s creation event: %w", event.subjectType, err)
	}
	subject := readmodels.SubjectKey{SubjectType: event.subjectType, SubjectID: created.ID}

	attributes, err := publishedAttributes(event.eventType, event.data)
	if err != nil {
		return err
	}

	audit, err := p.audit.Created(ctx, created.ID)
	if err != nil {
		return fmt.Errorf("read creation audit for %s %s: %w", subject.SubjectType, subject.SubjectID, err)
	}

	createdAt := event.occurredAt
	if audit.Found {
		createdAt = audit.CreatedAt
	}
	if err := p.store.Upsert(ctx, readmodels.SubjectIndexRecord{
		SubjectType:    subject.SubjectType,
		SubjectID:      subject.SubjectID,
		Name:           created.Name,
		CreatorActorID: audit.ActorID,
		CreatorEmail:   audit.ActorEmail,
		CreatedAt:      createdAt,
		LastUpdatedAt:  event.occurredAt,
		Attributes:     attributes,
	}); err != nil {
		return err
	}

	return p.recompute(ctx, subject.SubjectType, []string{subject.SubjectID})
}

func (p *SubjectIndexProjector) onDeleted(ctx context.Context, event projectedEvent) error {
	var deleted subjectIdentity
	if err := json.Unmarshal(event.data, &deleted); err != nil {
		return fmt.Errorf("unmarshal %s deletion event: %w", event.subjectType, err)
	}
	return p.store.Delete(ctx, readmodels.SubjectKey{SubjectType: event.subjectType, SubjectID: deleted.ID})
}

func (p *SubjectIndexProjector) onSubjectChanged(ctx context.Context, event projectedEvent) error {
	var changed subjectChangeIdentity
	if err := json.Unmarshal(event.data, &changed); err != nil {
		return fmt.Errorf("unmarshal %s update event: %w", event.subjectType, err)
	}
	if changed.subjectID() == "" {
		return fmt.Errorf("%s update event carries no subject id", event.subjectType)
	}

	subject := readmodels.SubjectKey{SubjectType: event.subjectType, SubjectID: changed.subjectID()}
	if err := p.cacheChangedAttributes(ctx, subject, event); err != nil {
		return err
	}

	counts, err := p.countsFor(ctx, subject)
	if err != nil {
		return err
	}
	return p.store.ApplySubjectChange(ctx, readmodels.SubjectChange{
		Subject:    subject,
		Name:       changed.Name,
		Counts:     counts,
		OccurredAt: event.occurredAt,
	})
}

func (p *SubjectIndexProjector) countsFor(ctx context.Context, subject readmodels.SubjectKey) (readmodels.CompletenessCounts, error) {
	required, filled, err := p.counter.CountsForSubjects(ctx, subject.SubjectType, []string{subject.SubjectID})
	if err != nil {
		return readmodels.CompletenessCounts{}, fmt.Errorf("compute completeness for %s %s: %w", subject.SubjectType, subject.SubjectID, err)
	}
	return readmodels.CompletenessCounts{Required: required, Filled: filled[subject.SubjectID]}, nil
}

func (p *SubjectIndexProjector) cacheChangedAttributes(ctx context.Context, subject readmodels.SubjectKey, event projectedEvent) error {
	change, err := expertChange(event.eventType, event.data)
	if err != nil {
		return err
	}
	if change != nil {
		return p.store.ApplyExpertChange(ctx, subject, change.expert, change.added)
	}

	attributes, err := publishedAttributes(event.eventType, event.data)
	if err != nil {
		return err
	}
	return p.store.MergeAttributes(ctx, subject, attributes)
}

func (p *SubjectIndexProjector) onAttributesOnlyChanged(ctx context.Context, event projectedEvent) (bool, error) {
	subjectType, attributes, err := attributeOnlyChange(event.eventType, event.data)
	if err != nil || subjectType == "" {
		return subjectType != "", err
	}

	var changed subjectChangeIdentity
	if err := json.Unmarshal(event.data, &changed); err != nil {
		return true, fmt.Errorf("unmarshal %s subject identity: %w", event.eventType, err)
	}
	if changed.subjectID() == "" {
		return true, fmt.Errorf("%s carries no subject id", event.eventType)
	}
	return true, p.store.MergeAttributes(ctx, readmodels.SubjectKey{SubjectType: subjectType, SubjectID: changed.subjectID()}, attributes)
}

func (p *SubjectIndexProjector) onFactsChanged(ctx context.Context, eventData []byte) error {
	var event factsIdentity
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal one-pager facts event: %w", err)
	}
	return p.recompute(ctx, event.SubjectType, []string{event.SubjectID})
}

func (p *SubjectIndexProjector) onConfigurationChanged(ctx context.Context, eventData []byte) error {
	var event subjectIdentity
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal one-pager configuration event: %w", err)
	}

	config, err := p.configs.GetByID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("look up one-pager configuration %s: %w", event.ID, err)
	}
	if config == nil {
		return nil
	}

	subjectIDs, err := p.store.SubjectIDs(ctx, config.SubjectType)
	if err != nil {
		return fmt.Errorf("list %s subjects for completeness recompute: %w", config.SubjectType, err)
	}
	return p.recompute(ctx, config.SubjectType, subjectIDs)
}

func (p *SubjectIndexProjector) Recompute(ctx context.Context, subjectType string, subjectIDs []string) error {
	return p.recompute(ctx, subjectType, subjectIDs)
}

func (p *SubjectIndexProjector) recompute(ctx context.Context, subjectType string, subjectIDs []string) error {
	if len(subjectIDs) == 0 {
		return nil
	}
	required, filled, err := p.counter.CountsForSubjects(ctx, subjectType, subjectIDs)
	if err != nil {
		return fmt.Errorf("compute completeness for %s subjects: %w", subjectType, err)
	}
	if err := p.store.ApplyCompleteness(ctx, subjectType, required, filled); err != nil {
		return fmt.Errorf("apply completeness for %s subjects: %w", subjectType, err)
	}
	return nil
}

func isConfigurationEvent(eventType string) bool {
	return slices.Contains(opevents.ConfigurationEventTypes(), eventType)
}
