package aggregates

import (
	"errors"
	"fmt"
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/stewardship/domain/events"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

var (
	ErrDomainRequired            = errors.New("stewardship requires a business domain")
	ErrActorRequired             = errors.New("stewardship changes must be attributed to an actor")
	ErrStewardshipReleased       = errors.New("stewardship has been released")
	ErrCorruptedStewardshipEvent = errors.New("corrupted event store: cannot rehydrate stewardship")
	ErrUnknownStewardshipEvent   = errors.New("unknown event type for stewardship aggregate")
)

type Stewardship struct {
	domain.AggregateRoot
	domainID   string
	concern    valueobjects.Concern
	steward    valueobjects.StewardRef
	assignedBy string
	assignedAt time.Time
	released   bool
}

type Assignment struct {
	DomainID   string
	Concern    valueobjects.Concern
	Steward    valueobjects.StewardRef
	AssignedBy string
}

func NewStewardship(a Assignment) (*Stewardship, error) {
	if a.DomainID == "" {
		return nil, ErrDomainRequired
	}
	if a.AssignedBy == "" {
		return nil, ErrActorRequired
	}
	s := &Stewardship{AggregateRoot: domain.NewAggregateRoot()}
	s.raise(events.NewStewardAssigned(events.StewardAssignedFields{
		ID:         s.ID(),
		DomainID:   a.DomainID,
		Concern:    a.Concern.Value(),
		StewardID:  a.Steward.Value(),
		AssignedBy: a.AssignedBy,
	}))
	return s, nil
}

func LoadStewardshipFromHistory(history []domain.DomainEvent) (*Stewardship, error) {
	s := &Stewardship{AggregateRoot: domain.NewAggregateRoot()}
	var applyErr error
	s.LoadFromHistory(history, func(event domain.DomainEvent) {
		if applyErr == nil {
			applyErr = s.apply(event)
		}
	})
	if applyErr != nil {
		return nil, applyErr
	}
	return s, nil
}

func (s *Stewardship) Assign(steward valueobjects.StewardRef, assignedBy string) error {
	if assignedBy == "" {
		return ErrActorRequired
	}
	if s.released {
		return ErrStewardshipReleased
	}
	if s.steward.Equals(steward) {
		return nil
	}
	s.raise(events.NewStewardAssigned(events.StewardAssignedFields{
		ID:         s.ID(),
		DomainID:   s.domainID,
		Concern:    s.concern.Value(),
		StewardID:  steward.Value(),
		AssignedBy: assignedBy,
	}))
	return nil
}

func (s *Stewardship) Release(releasedBy string) error {
	if releasedBy == "" {
		return ErrActorRequired
	}
	if s.released {
		return nil
	}
	s.raise(events.NewStewardReleased(s.ID(), s.domainID, s.concern.Value(), releasedBy))
	return nil
}

func (s *Stewardship) raise(event domain.DomainEvent) {
	if err := s.apply(event); err != nil {
		panic(fmt.Sprintf("stewardship: in-process apply failed: %v", err))
	}
	s.RaiseEvent(event)
}

func (s *Stewardship) apply(event domain.DomainEvent) error {
	switch evt := event.(type) {
	case events.StewardAssigned:
		return s.applyAssigned(evt)
	case events.StewardReleased:
		s.released = true
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrUnknownStewardshipEvent, event)
	}
}

func (s *Stewardship) applyAssigned(evt events.StewardAssigned) error {
	concern, err := valueobjects.NewConcern(evt.Concern)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCorruptedStewardshipEvent, err)
	}
	if s.ID() != evt.ID {
		s.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	}
	s.domainID = evt.DomainID
	s.concern = concern
	s.steward = valueobjects.RecordedStewardRef(evt.StewardID)
	s.assignedBy = evt.AssignedBy
	s.assignedAt = evt.AssignedAt
	return nil
}

func (s *Stewardship) DomainID() string                 { return s.domainID }
func (s *Stewardship) Concern() valueobjects.Concern    { return s.concern }
func (s *Stewardship) Steward() valueobjects.StewardRef { return s.steward }
func (s *Stewardship) AssignedBy() string               { return s.assignedBy }
func (s *Stewardship) AssignedAt() time.Time            { return s.assignedAt }
func (s *Stewardship) IsReleased() bool                 { return s.released }
