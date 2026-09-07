package aggregates

import (
	"errors"
	"fmt"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrOwnershipNotUnknown   = errors.New("ownership can only be nominated or assigned while it is unknown")
	ErrNoNominationToConfirm = errors.New("no nominated owner to confirm")
	ErrNoOwnershipToClear    = errors.New("ownership is already unknown")
)

func (a *ApplicationComponent) NominateOwner(owner valueobjects.OwnerReference) error {
	return a.establishOwner(func() domain.DomainEvent {
		return events.NewApplicationOwnerNominated(a.ID(), owner.Kind(), owner.ID())
	})
}

func (a *ApplicationComponent) ConfirmOwnership() error {
	if !a.ownershipState.IsNominated() || a.owner == nil {
		return ErrNoNominationToConfirm
	}
	return a.transitionOwnership(events.NewApplicationOwnershipConfirmed(
		a.ID(),
		a.owner.Kind(),
		a.owner.ID(),
		a.owner.ResolvedOwnershipState().String(),
	))
}

func (a *ApplicationComponent) AssignOwner(owner valueobjects.OwnerReference) error {
	return a.establishOwner(func() domain.DomainEvent {
		return events.NewApplicationOwnerAssigned(
			a.ID(),
			owner.Kind(),
			owner.ID(),
			owner.ResolvedOwnershipState().String(),
		)
	})
}

func (a *ApplicationComponent) establishOwner(ownerEvent func() domain.DomainEvent) error {
	if !a.ownershipState.IsUnknown() {
		return ErrOwnershipNotUnknown
	}
	return a.transitionOwnership(ownerEvent())
}

func (a *ApplicationComponent) ClearOwnership() error {
	if a.ownershipState.IsUnknown() {
		return ErrNoOwnershipToClear
	}
	return a.transitionOwnership(events.NewApplicationOwnershipCleared(a.ID()))
}

func (a *ApplicationComponent) OwnershipState() valueobjects.OwnershipState {
	return a.ownershipState
}

func (a *ApplicationComponent) Owner() (valueobjects.OwnerReference, bool) {
	if a.owner == nil {
		return valueobjects.OwnerReference{}, false
	}
	return *a.owner, true
}

func (a *ApplicationComponent) transitionOwnership(event domain.DomainEvent) error {
	if err := a.apply(event); err != nil {
		return err
	}
	a.RaiseEvent(event)
	return nil
}

func (a *ApplicationComponent) applyOwnershipEvent(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.ApplicationOwnerNominated:
		return a.applyOwnerNominated(e)
	case events.ApplicationOwnershipConfirmed:
		return a.applyOwnershipResolved(e.OwnerKind, e.OwnerID, e.OwnershipState)
	case events.ApplicationOwnerAssigned:
		return a.applyOwnershipResolved(e.OwnerKind, e.OwnerID, e.OwnershipState)
	case events.ApplicationOwnershipCleared:
		a.ownershipState = valueobjects.UnknownOwnershipState()
		a.owner = nil
	}
	return nil
}

func (a *ApplicationComponent) applyOwnerNominated(e events.ApplicationOwnerNominated) error {
	owner, err := valueobjects.NewOwnerReference(e.OwnerKind, e.OwnerID)
	if err != nil {
		return fmt.Errorf("%w: owner reference: %v", domain.ErrCorruptedEvent, err)
	}
	a.ownershipState = valueobjects.NominatedOwnershipState()
	a.owner = &owner
	return nil
}

func (a *ApplicationComponent) applyOwnershipResolved(ownerKind, ownerID, ownershipState string) error {
	owner, err := valueobjects.NewOwnerReference(ownerKind, ownerID)
	if err != nil {
		return fmt.Errorf("%w: owner reference: %v", domain.ErrCorruptedEvent, err)
	}
	state, err := valueobjects.NewOwnershipState(ownershipState)
	if err != nil {
		return fmt.Errorf("%w: ownership state: %v", domain.ErrCorruptedEvent, err)
	}
	a.ownershipState = state
	a.owner = &owner
	return nil
}
