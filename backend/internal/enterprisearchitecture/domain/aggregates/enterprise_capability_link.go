package aggregates

import (
	"errors"

	"easi/backend/internal/enterprisearchitecture/domain/events"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var ErrCannotLinkInactiveCapability = errors.New("cannot link to an inactive enterprise capability")

type EnterpriseCapabilityLink struct {
	domain.AggregateRoot
	enterpriseCapabilityID valueobjects.EnterpriseCapabilityID
	domainCapabilityID     valueobjects.DomainCapabilityID
	linkedBy               valueobjects.LinkedBy
	linkedAt               valueobjects.LinkedAt
}

func NewEnterpriseCapabilityLink(
	capability *EnterpriseCapability,
	domainCapabilityID valueobjects.DomainCapabilityID,
	linkedBy valueobjects.LinkedBy,
) (*EnterpriseCapabilityLink, error) {
	if !capability.IsActive() {
		return nil, ErrCannotLinkInactiveCapability
	}

	id := valueobjects.NewEnterpriseCapabilityLinkID()
	aggregate := &EnterpriseCapabilityLink{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewEnterpriseCapabilityLinked(
		aggregate.ID(),
		capability.ID(),
		domainCapabilityID.Value(),
		linkedBy.Value(),
	)

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadEnterpriseCapabilityLinkFromHistory(eventHistory []domain.DomainEvent) (*EnterpriseCapabilityLink, error) {
	aggregate := &EnterpriseCapabilityLink{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (l *EnterpriseCapabilityLink) Unlink() error {
	event := events.NewEnterpriseCapabilityUnlinked(
		l.ID(),
		l.enterpriseCapabilityID.Value(),
		l.domainCapabilityID.Value(),
	)

	l.apply(event)
	l.RaiseEvent(event)

	return nil
}

func (l *EnterpriseCapabilityLink) apply(event domain.DomainEvent) {
	switch evt := event.(type) {
	case events.EnterpriseCapabilityLinked:
		l.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
		l.enterpriseCapabilityID = mustNewEnterpriseCapabilityIDFromString(evt.EnterpriseCapabilityID)
		l.domainCapabilityID = mustNewDomainCapabilityIDFromString(evt.DomainCapabilityID)
		l.linkedBy = valueobjects.MustNewLinkedBy(evt.LinkedBy)
		l.linkedAt = valueobjects.NewLinkedAtFromTime(evt.LinkedAt)
	case events.EnterpriseCapabilityUnlinked:
	}
}

func mustNewEnterpriseCapabilityIDFromString(value string) valueobjects.EnterpriseCapabilityID {
	id, err := valueobjects.NewEnterpriseCapabilityIDFromString(value)
	if err != nil {
		panic("corrupted event store: invalid enterprise capability ID: " + value)
	}
	return id
}

func mustNewDomainCapabilityIDFromString(value string) valueobjects.DomainCapabilityID {
	id, err := valueobjects.NewDomainCapabilityIDFromString(value)
	if err != nil {
		panic("corrupted event store: invalid domain capability ID: " + value)
	}
	return id
}

func (l *EnterpriseCapabilityLink) EnterpriseCapabilityID() valueobjects.EnterpriseCapabilityID {
	return l.enterpriseCapabilityID
}

func (l *EnterpriseCapabilityLink) DomainCapabilityID() valueobjects.DomainCapabilityID {
	return l.domainCapabilityID
}

func (l *EnterpriseCapabilityLink) LinkedBy() valueobjects.LinkedBy {
	return l.linkedBy
}

func (l *EnterpriseCapabilityLink) LinkedAt() valueobjects.LinkedAt {
	return l.linkedAt
}
