package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrNoRoleToClear                  = errors.New("component holds no realization role to clear")
	ErrCorruptedRealizationRolesEvent = errors.New("corrupted event store: cannot rehydrate realization roles")
	ErrUnknownRealizationRolesEvent   = errors.New("unknown event type for realization roles aggregate")
)

type componentRoleState struct {
	role          valueobjects.RealizationRole
	realizationID string
	assignedBy    string
	assignedAt    time.Time
}

type RealizationRoles struct {
	domain.AggregateRoot
	capabilityID valueobjects.PhysicalCapabilityRef
	roles        map[string]componentRoleState
}

type RealizationRolesFacts struct {
	CapabilityID  valueobjects.PhysicalCapabilityRef
	ComponentID   valueobjects.ApplicationRef
	RealizationID string
	Role          valueobjects.RealizationRole
	AssignedBy    string
}

func NewRealizationRoles(facts RealizationRolesFacts) (*RealizationRoles, error) {
	id := valueobjects.NewRealizationRolesID()
	aggregate := &RealizationRoles{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
		roles:         map[string]componentRoleState{},
	}
	aggregate.raise(events.NewRealizationRoleAssigned(events.RealizationRoleAssignedFields{
		ID:            id.Value(),
		CapabilityID:  facts.CapabilityID.Value(),
		ComponentID:   facts.ComponentID.Value(),
		RealizationID: facts.RealizationID,
		Role:          facts.Role.Value(),
		AssignedBy:    facts.AssignedBy,
	}))
	return aggregate, nil
}

func LoadRealizationRolesFromHistory(eventHistory []domain.DomainEvent) (*RealizationRoles, error) {
	aggregate := &RealizationRoles{
		AggregateRoot: domain.NewAggregateRoot(),
		roles:         map[string]componentRoleState{},
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

func (r *RealizationRoles) Assign(component valueobjects.ApplicationRef, realizationID string, role valueobjects.RealizationRole, assignedBy string) error {
	displaced := ""
	if role.Value() == valueobjects.RealizationRoleStandard {
		if holder, ok := r.currentStandardHolder(); ok && holder != component.Value() {
			displaced = holder
		}
	}
	r.raise(events.NewRealizationRoleAssigned(events.RealizationRoleAssignedFields{
		ID:                   r.ID(),
		CapabilityID:         r.capabilityID.Value(),
		ComponentID:          component.Value(),
		RealizationID:        realizationID,
		Role:                 role.Value(),
		DisplacedComponentID: displaced,
		AssignedBy:           assignedBy,
	}))
	return nil
}

func (r *RealizationRoles) Clear(component valueobjects.ApplicationRef, clearedBy string) error {
	if _, ok := r.roles[component.Value()]; !ok {
		return ErrNoRoleToClear
	}
	r.raise(events.NewRealizationRoleCleared(events.RealizationRoleClearedFields{
		ID:           r.ID(),
		CapabilityID: r.capabilityID.Value(),
		ComponentID:  component.Value(),
		ClearedBy:    clearedBy,
	}))
	return nil
}

func (r *RealizationRoles) CapabilityID() valueobjects.PhysicalCapabilityRef { return r.capabilityID }

func (r *RealizationRoles) RoleFor(component valueobjects.ApplicationRef) (valueobjects.RealizationRole, bool) {
	entry, ok := r.roles[component.Value()]
	if !ok {
		return valueobjects.RealizationRole{}, false
	}
	return entry.role, true
}

func (r *RealizationRoles) currentStandardHolder() (string, bool) {
	for componentID, entry := range r.roles {
		if entry.role.Value() == valueobjects.RealizationRoleStandard {
			return componentID, true
		}
	}
	return "", false
}

func (r *RealizationRoles) raise(event domain.DomainEvent) {
	if err := r.apply(event); err != nil {
		panic(fmt.Sprintf("architecturedirection: in-process apply failed: %v", err))
	}
	r.RaiseEvent(event)
}

func (r *RealizationRoles) apply(event domain.DomainEvent) error {
	switch evt := event.(type) {
	case events.RealizationRoleAssigned:
		return r.applyAssigned(evt)
	case events.RealizationRoleCleared:
		return r.applyCleared(evt)
	default:
		return fmt.Errorf("%w: %T", ErrUnknownRealizationRolesEvent, event)
	}
}

func (r *RealizationRoles) applyAssigned(evt events.RealizationRoleAssigned) error {
	capabilityID, err := valueobjects.NewPhysicalCapabilityRef(evt.CapabilityID)
	if err != nil {
		return fmt.Errorf("%w: capability ref %q: %v", ErrCorruptedRealizationRolesEvent, evt.CapabilityID, err)
	}
	role, err := valueobjects.NewRealizationRole(evt.Role)
	if err != nil {
		return fmt.Errorf("%w: role %q: %v", ErrCorruptedRealizationRolesEvent, evt.Role, err)
	}
	if r.ID() != evt.ID {
		r.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	}
	if r.roles == nil {
		r.roles = map[string]componentRoleState{}
	}
	r.capabilityID = capabilityID
	if evt.DisplacedComponentID != "" {
		delete(r.roles, evt.DisplacedComponentID)
	}
	r.roles[evt.ComponentID] = componentRoleState{
		role:          role,
		realizationID: evt.RealizationID,
		assignedBy:    evt.AssignedBy,
		assignedAt:    evt.OccurredOn,
	}
	return nil
}

func (r *RealizationRoles) applyCleared(evt events.RealizationRoleCleared) error {
	if r.ID() != evt.ID {
		r.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	}
	if r.roles == nil {
		r.roles = map[string]componentRoleState{}
	}
	delete(r.roles, evt.ComponentID)
	return nil
}
