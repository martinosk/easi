package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	opevents "easi/backend/internal/onepagers/domain/events"
	domain "easi/backend/internal/shared/eventsourcing"
)

type SubjectIndexStore interface {
	Upsert(ctx context.Context, record readmodels.SubjectIndexRecord) error
	Delete(ctx context.Context, subject readmodels.SubjectKey) error
	ApplySubjectChange(ctx context.Context, change readmodels.SubjectChange) error
	ApplyCompleteness(ctx context.Context, subject readmodels.SubjectKey, counts readmodels.CompletenessCounts) error
	SubjectIDs(ctx context.Context, subjectType string) ([]string, error)
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
	eaPL.EnterpriseCapabilityCreated: "enterprise-capability",
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
	eaPL.EnterpriseCapabilityUpdated:       "enterprise-capability",
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
	return append(types, opevents.ConfigurationEventTypes()...)
}

func (p *SubjectIndexProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event data: %w", event.EventType(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), event.OccurredAt(), eventData)
}

func (p *SubjectIndexProjector) ProjectEvent(ctx context.Context, eventType string, occurredAt time.Time, eventData []byte) error {
	if subjectType, ok := subjectTypeByCreationEvent[eventType]; ok {
		return p.onCreated(ctx, subjectType, occurredAt, eventData)
	}
	if subjectType, ok := subjectTypeByDeletionEvent[eventType]; ok {
		return p.onDeleted(ctx, subjectType, eventData)
	}
	if subjectType, ok := subjectTypeByUpdateEvent[eventType]; ok {
		return p.onSubjectChanged(ctx, subjectType, occurredAt, eventData)
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

type factsIdentity struct {
	SubjectType string `json:"subjectType"`
	SubjectID   string `json:"subjectId"`
}

func (p *SubjectIndexProjector) onCreated(ctx context.Context, subjectType string, occurredAt time.Time, eventData []byte) error {
	var event subjectIdentity
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s creation event: %w", subjectType, err)
	}

	audit, err := p.audit.Created(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("read creation audit for %s %s: %w", subjectType, event.ID, err)
	}

	required, filled, err := p.counter.CountsForSubjects(ctx, subjectType, []string{event.ID})
	if err != nil {
		return fmt.Errorf("compute completeness for created %s %s: %w", subjectType, event.ID, err)
	}

	createdAt := occurredAt
	if audit.Found {
		createdAt = audit.CreatedAt
	}
	return p.store.Upsert(ctx, readmodels.SubjectIndexRecord{
		SubjectType:    subjectType,
		SubjectID:      event.ID,
		Name:           event.Name,
		CreatorActorID: audit.ActorID,
		CreatorEmail:   audit.ActorEmail,
		CreatedAt:      createdAt,
		LastUpdatedAt:  occurredAt,
		RequiredCount:  required,
		FilledCount:    filled[event.ID],
	})
}

func (p *SubjectIndexProjector) onDeleted(ctx context.Context, subjectType string, eventData []byte) error {
	var event subjectIdentity
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s deletion event: %w", subjectType, err)
	}
	return p.store.Delete(ctx, readmodels.SubjectKey{SubjectType: subjectType, SubjectID: event.ID})
}

func (p *SubjectIndexProjector) onSubjectChanged(ctx context.Context, subjectType string, occurredAt time.Time, eventData []byte) error {
	var event subjectIdentity
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s update event: %w", subjectType, err)
	}

	required, filled, err := p.counter.CountsForSubjects(ctx, subjectType, []string{event.ID})
	if err != nil {
		return fmt.Errorf("compute completeness for updated %s %s: %w", subjectType, event.ID, err)
	}
	return p.store.ApplySubjectChange(ctx, readmodels.SubjectChange{
		Subject:    readmodels.SubjectKey{SubjectType: subjectType, SubjectID: event.ID},
		Name:       event.Name,
		Counts:     readmodels.CompletenessCounts{Required: required, Filled: filled[event.ID]},
		OccurredAt: occurredAt,
	})
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

func (p *SubjectIndexProjector) recompute(ctx context.Context, subjectType string, subjectIDs []string) error {
	if len(subjectIDs) == 0 {
		return nil
	}
	required, filled, err := p.counter.CountsForSubjects(ctx, subjectType, subjectIDs)
	if err != nil {
		return fmt.Errorf("compute completeness for %s subjects: %w", subjectType, err)
	}
	for _, subjectID := range subjectIDs {
		subject := readmodels.SubjectKey{SubjectType: subjectType, SubjectID: subjectID}
		if err := p.store.ApplyCompleteness(ctx, subject, readmodels.CompletenessCounts{Required: required, Filled: filled[subjectID]}); err != nil {
			return fmt.Errorf("apply completeness for %s %s: %w", subjectType, subjectID, err)
		}
	}
	return nil
}

func isConfigurationEvent(eventType string) bool {
	for _, configEvent := range opevents.ConfigurationEventTypes() {
		if configEvent == eventType {
			return true
		}
	}
	return false
}
