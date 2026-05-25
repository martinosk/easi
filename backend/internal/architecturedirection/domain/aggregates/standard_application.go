package aggregates

import (
	"errors"
	"fmt"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrNarrativeRequiredForStandardApplication = errors.New("narrative is required to set a standard application")
	ErrCorruptedStandardApplicationEvent       = errors.New("corrupted event store: cannot rehydrate standard application")
	ErrUnknownStandardApplicationEvent         = errors.New("unknown event type for standard application aggregate")
)

type StandardApplication struct {
	domain.AggregateRoot
	enterpriseCapabilityID valueobjects.EnterpriseCapabilityRef
	currentApplication     valueobjects.ApplicationRef
	currentNarrative       sharedvo.Description
}

func NewStandardApplication(
	ec valueobjects.EnterpriseCapabilityRef,
	app valueobjects.ApplicationRef,
	narrative sharedvo.Description,
) (*StandardApplication, error) {
	if narrative.IsEmpty() {
		return nil, ErrNarrativeRequiredForStandardApplication
	}
	id := valueobjects.NewStandardApplicationID()
	aggregate := &StandardApplication{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}
	aggregate.raise(events.NewStandardApplicationSet(events.StandardApplicationSetFields{
		ID:                     id.Value(),
		EnterpriseCapabilityID: ec.Value(),
		ApplicationID:          app.Value(),
		Narrative:              narrative.Value(),
	}))
	return aggregate, nil
}

func LoadStandardApplicationFromHistory(eventHistory []domain.DomainEvent) (*StandardApplication, error) {
	aggregate := &StandardApplication{
		AggregateRoot: domain.NewAggregateRoot(),
	}
	var applyErr error
	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = aggregate.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}
	return aggregate, nil
}

func (s *StandardApplication) Change(app valueobjects.ApplicationRef, narrative sharedvo.Description) error {
	if narrative.IsEmpty() {
		return ErrNarrativeRequiredForStandardApplication
	}
	s.raise(events.NewStandardApplicationSet(events.StandardApplicationSetFields{
		ID:                     s.ID(),
		EnterpriseCapabilityID: s.enterpriseCapabilityID.Value(),
		ApplicationID:          app.Value(),
		PreviousApplicationID:  s.currentApplication.Value(),
		Narrative:              narrative.Value(),
	}))
	return nil
}

func (s *StandardApplication) EnterpriseCapabilityID() valueobjects.EnterpriseCapabilityRef {
	return s.enterpriseCapabilityID
}

func (s *StandardApplication) CurrentApplication() valueobjects.ApplicationRef {
	return s.currentApplication
}

func (s *StandardApplication) CurrentNarrative() sharedvo.Description {
	return s.currentNarrative
}

func (s *StandardApplication) raise(event domain.DomainEvent) {
	if err := s.apply(event); err != nil {
		panic(fmt.Sprintf("architecturedirection: in-process apply failed: %v", err))
	}
	s.RaiseEvent(event)
}

func (s *StandardApplication) apply(event domain.DomainEvent) error {
	evt, ok := event.(events.StandardApplicationSet)
	if !ok {
		return fmt.Errorf("%w: %T", ErrUnknownStandardApplicationEvent, event)
	}
	ecRef, err := valueobjects.NewEnterpriseCapabilityRef(evt.EnterpriseCapabilityID)
	if err != nil {
		return fmt.Errorf("%w: enterprise capability ref %q: %v", ErrCorruptedStandardApplicationEvent, evt.EnterpriseCapabilityID, err)
	}
	appRef, err := valueobjects.NewApplicationRef(evt.ApplicationID)
	if err != nil {
		return fmt.Errorf("%w: application ref %q: %v", ErrCorruptedStandardApplicationEvent, evt.ApplicationID, err)
	}
	narrative, err := sharedvo.NewDescription(evt.Narrative)
	if err != nil {
		return fmt.Errorf("%w: narrative: %v", ErrCorruptedStandardApplicationEvent, err)
	}
	if s.ID() != evt.ID {
		s.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	}
	s.enterpriseCapabilityID = ecRef
	s.currentApplication = appRef
	s.currentNarrative = narrative
	return nil
}
