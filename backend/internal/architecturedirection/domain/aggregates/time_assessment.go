package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrTimeAssessmentAlreadyRemoved = errors.New("time assessment has already been removed")
	ErrCorruptedTimeAssessmentEvent = errors.New("corrupted event store: cannot rehydrate time assessment")
	ErrUnknownTimeAssessmentEvent   = errors.New("unknown event type for time assessment aggregate")
)

type TimeAssessment struct {
	domain.AggregateRoot
	capabilityID valueobjects.PhysicalCapabilityRef
	componentID  valueobjects.ApplicationRef
	grade        valueobjects.TimeGrade
	rationale    sharedvo.Description
	assessedBy   string
	assessedAt   time.Time
	removed      bool
}

type TimeAssessmentFacts struct {
	CapabilityID  valueobjects.PhysicalCapabilityRef
	ComponentID   valueobjects.ApplicationRef
	RealizationID string
	Grade         valueobjects.TimeGrade
	Rationale     sharedvo.Description
	AssessedBy    string
}

func NewTimeAssessment(facts TimeAssessmentFacts) (*TimeAssessment, error) {
	id := valueobjects.NewTimeAssessmentID()
	aggregate := &TimeAssessment{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}
	aggregate.raise(events.NewTimeAssessmentRecorded(events.TimeAssessmentRecordedFields{
		ID:            id.Value(),
		CapabilityID:  facts.CapabilityID.Value(),
		ComponentID:   facts.ComponentID.Value(),
		RealizationID: facts.RealizationID,
		Grade:         facts.Grade.Value(),
		Rationale:     facts.Rationale.Value(),
		AssessedBy:    facts.AssessedBy,
	}))
	return aggregate, nil
}

func LoadTimeAssessmentFromHistory(eventHistory []domain.DomainEvent) (*TimeAssessment, error) {
	aggregate := &TimeAssessment{
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

func (t *TimeAssessment) Reassess(realizationID string, grade valueobjects.TimeGrade, rationale sharedvo.Description, assessedBy string) error {
	if t.removed {
		return ErrTimeAssessmentAlreadyRemoved
	}
	t.raise(events.NewTimeAssessmentRecorded(events.TimeAssessmentRecordedFields{
		ID:            t.ID(),
		CapabilityID:  t.capabilityID.Value(),
		ComponentID:   t.componentID.Value(),
		RealizationID: realizationID,
		Grade:         grade.Value(),
		PreviousGrade: t.grade.Value(),
		Rationale:     rationale.Value(),
		AssessedBy:    assessedBy,
	}))
	return nil
}

func (t *TimeAssessment) Remove(removedBy string) error {
	if t.removed {
		return ErrTimeAssessmentAlreadyRemoved
	}
	t.raise(events.NewTimeAssessmentRemoved(events.TimeAssessmentRemovedFields{
		ID:           t.ID(),
		CapabilityID: t.capabilityID.Value(),
		ComponentID:  t.componentID.Value(),
		RemovedBy:    removedBy,
	}))
	return nil
}

func (t *TimeAssessment) CapabilityID() valueobjects.PhysicalCapabilityRef { return t.capabilityID }
func (t *TimeAssessment) ComponentID() valueobjects.ApplicationRef         { return t.componentID }
func (t *TimeAssessment) Grade() valueobjects.TimeGrade                    { return t.grade }
func (t *TimeAssessment) Rationale() sharedvo.Description                  { return t.rationale }
func (t *TimeAssessment) AssessedBy() string                               { return t.assessedBy }
func (t *TimeAssessment) AssessedAt() time.Time                            { return t.assessedAt }
func (t *TimeAssessment) IsRemoved() bool                                  { return t.removed }

func (t *TimeAssessment) raise(event domain.DomainEvent) {
	if err := t.apply(event); err != nil {
		panic(fmt.Sprintf("architecturedirection: in-process apply failed: %v", err))
	}
	t.RaiseEvent(event)
}

func (t *TimeAssessment) apply(event domain.DomainEvent) error {
	switch evt := event.(type) {
	case events.TimeAssessmentRecorded:
		return t.applyRecorded(evt)
	case events.TimeAssessmentRemoved:
		return t.applyRemoved(evt)
	default:
		return fmt.Errorf("%w: %T", ErrUnknownTimeAssessmentEvent, event)
	}
}

func (t *TimeAssessment) applyRecorded(evt events.TimeAssessmentRecorded) error {
	capabilityID, err := valueobjects.NewPhysicalCapabilityRef(evt.CapabilityID)
	if err != nil {
		return fmt.Errorf("%w: capability ref %q: %v", ErrCorruptedTimeAssessmentEvent, evt.CapabilityID, err)
	}
	componentID, err := valueobjects.NewApplicationRef(evt.ComponentID)
	if err != nil {
		return fmt.Errorf("%w: component ref %q: %v", ErrCorruptedTimeAssessmentEvent, evt.ComponentID, err)
	}
	grade, err := valueobjects.NewTimeGrade(evt.Grade)
	if err != nil {
		return fmt.Errorf("%w: grade %q: %v", ErrCorruptedTimeAssessmentEvent, evt.Grade, err)
	}
	rationale, err := sharedvo.NewDescription(evt.Rationale)
	if err != nil {
		return fmt.Errorf("%w: rationale: %v", ErrCorruptedTimeAssessmentEvent, err)
	}
	if t.ID() != evt.ID {
		t.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	}
	t.capabilityID = capabilityID
	t.componentID = componentID
	t.grade = grade
	t.rationale = rationale
	t.assessedBy = evt.AssessedBy
	t.assessedAt = evt.OccurredOn
	t.removed = false
	return nil
}

func (t *TimeAssessment) applyRemoved(evt events.TimeAssessmentRemoved) error {
	if t.ID() != evt.ID {
		t.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	}
	t.removed = true
	return nil
}
