package aggregates

import (
	"fmt"

	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func (f *OnePagerFacts) applyAndRaise(event domain.DomainEvent) error {
	if err := f.apply(event); err != nil {
		return err
	}
	f.RaiseEvent(event)
	return nil
}

func (f *OnePagerFacts) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.FieldValueRecorded:
		return f.applyFieldValueRecorded(e)
	case events.FieldValueCleared:
		return f.applyFieldValueCleared(e)
	case events.OnePagerFactsArchived:
		return f.applyArchived(e)
	}
	return nil
}

func (f *OnePagerFacts) applyFieldValueRecorded(e events.FieldValueRecorded) error {
	if err := f.applyIdentity(e.FactsEventBase); err != nil {
		return err
	}
	value, err := valueobjects.FieldValueFromEnvelope(e.Value)
	if err != nil {
		return fmt.Errorf("%w: field value for field %q: %v", domain.ErrCorruptedEvent, e.FieldID, err)
	}
	f.values[e.FieldID] = value
	return nil
}

func (f *OnePagerFacts) applyFieldValueCleared(e events.FieldValueCleared) error {
	if err := f.applyIdentity(e.FactsEventBase); err != nil {
		return err
	}
	delete(f.values, e.FieldID)
	return nil
}

func (f *OnePagerFacts) applyArchived(e events.OnePagerFactsArchived) error {
	if err := f.applyIdentity(e.FactsEventBase); err != nil {
		return err
	}
	f.archived = true
	return nil
}

func (f *OnePagerFacts) applyIdentity(base events.FactsEventBase) error {
	if f.ID() == base.ID {
		return nil
	}

	tenantID, err := sharedvo.NewTenantID(base.TenantID)
	if err != nil {
		return fmt.Errorf("%w: tenant ID %q: %v", domain.ErrCorruptedEvent, base.TenantID, err)
	}
	subjectRef, err := valueobjects.NewSubjectRef(base.SubjectType, base.SubjectID)
	if err != nil {
		return fmt.Errorf("%w: subject %s/%s: %v", domain.ErrCorruptedEvent, base.SubjectType, base.SubjectID, err)
	}

	f.AggregateRoot = domain.NewAggregateRootWithID(base.ID)
	f.tenantID = tenantID
	f.subjectRef = subjectRef
	return nil
}
