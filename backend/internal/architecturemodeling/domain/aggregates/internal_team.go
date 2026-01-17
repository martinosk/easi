package aggregates

import (
	"time"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type InternalTeam struct {
	domain.AggregateRoot
	name          valueobjects.EntityName
	department    string
	contactPerson string
	notes         valueobjects.Notes
	createdAt     time.Time
	isDeleted     bool
}

func NewInternalTeam(
	name valueobjects.EntityName,
	department string,
	contactPerson string,
	notes valueobjects.Notes,
) (*InternalTeam, error) {
	aggregate := &InternalTeam{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewInternalTeamCreated(
		aggregate.ID(),
		name.Value(),
		department,
		contactPerson,
		notes.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadInternalTeamFromHistory(events []domain.DomainEvent) (*InternalTeam, error) {
	aggregate := &InternalTeam{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(events, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (t *InternalTeam) Update(
	name valueobjects.EntityName,
	department string,
	contactPerson string,
	notes valueobjects.Notes,
) error {
	event := events.NewInternalTeamUpdated(
		t.ID(),
		name.Value(),
		department,
		contactPerson,
		notes.Value(),
	)

	t.apply(event)
	t.RaiseEvent(event)

	return nil
}

func (t *InternalTeam) Delete() error {
	event := events.NewInternalTeamDeleted(
		t.ID(),
		t.name.Value(),
	)

	t.apply(event)
	t.RaiseEvent(event)

	return nil
}

func (t *InternalTeam) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.InternalTeamCreated:
		t.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		t.name = valueobjects.MustNewEntityName(e.Name)
		t.department = e.Department
		t.contactPerson = e.ContactPerson
		t.notes = valueobjects.MustNewNotes(e.Notes)
		t.createdAt = e.CreatedAt
	case events.InternalTeamUpdated:
		t.name = valueobjects.MustNewEntityName(e.Name)
		t.department = e.Department
		t.contactPerson = e.ContactPerson
		t.notes = valueobjects.MustNewNotes(e.Notes)
	case events.InternalTeamDeleted:
		t.isDeleted = true
	}
}

func (t *InternalTeam) Name() valueobjects.EntityName {
	return t.name
}

func (t *InternalTeam) Department() string {
	return t.department
}

func (t *InternalTeam) ContactPerson() string {
	return t.contactPerson
}

func (t *InternalTeam) Notes() valueobjects.Notes {
	return t.notes
}

func (t *InternalTeam) CreatedAt() time.Time {
	return t.createdAt
}

func (t *InternalTeam) IsDeleted() bool {
	return t.isDeleted
}
