package aggregates

import (
	"errors"
	"fmt"

	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidSourceCardinality    = errors.New("source capability count does not match the direction type")
	ErrInvalidPlacementCardinality = errors.New("placement count does not match the direction type")
	ErrDuplicateSourceCapabilities = errors.New("source capabilities must be unique")
	ErrNarrativeRequiredToPropose  = errors.New("narrative is required before advancing a direction to proposed")
	ErrInvalidStatusTransition     = errors.New("status transition not allowed from current status")
	ErrDirectionAgreedImmutable    = errors.New("agreed directions are immutable; reject and replace to change")
	ErrCorruptedEvent              = errors.New("corrupted event store: cannot rehydrate direction")
)

type Direction struct {
	domain.AggregateRoot
	enterpriseCapabilityID valueobjects.EnterpriseCapabilityRef
	directionType          valueobjects.DirectionType
	sourceCapabilityIDs    []valueobjects.PhysicalCapabilityRef
	placements             []valueobjects.Placement
	horizon                valueobjects.Horizon
	status                 valueobjects.DirectionStatus
	narrative              valueobjects.Narrative
}

type DraftParams struct {
	EnterpriseCapabilityID valueobjects.EnterpriseCapabilityRef
	Type                   valueobjects.DirectionType
	SourceCapabilityIDs    []valueobjects.PhysicalCapabilityRef
	Placements             []valueobjects.Placement
	Horizon                valueobjects.Horizon
	Narrative              valueobjects.Narrative
}

func DraftDirection(params DraftParams) (*Direction, error) {
	if err := validateSourceCardinality(params.Type, params.SourceCapabilityIDs); err != nil {
		return nil, err
	}
	if err := validatePlacementCardinality(params.Type, params.Placements); err != nil {
		return nil, err
	}

	id := valueobjects.NewDirectionID()
	aggregate := &Direction{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
	}

	event := events.NewDirectionDraftedFromFields(events.DirectionDraftedFields{
		ID:                     id.Value(),
		EnterpriseCapabilityID: params.EnterpriseCapabilityID.Value(),
		Type:                   params.Type.Value(),
		SourceCapabilityIDs:    refsToStrings(params.SourceCapabilityIDs),
		Placements:             placementsToData(params.Placements),
		Horizon:                params.Horizon.Value(),
		Narrative:              params.Narrative.Value(),
	})
	aggregate.raise(event)
	return aggregate, nil
}

func LoadDirectionFromHistory(eventHistory []domain.DomainEvent) (*Direction, error) {
	aggregate := &Direction{
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

func (d *Direction) Propose() error {
	if err := d.requireTransition(valueobjects.DirectionStatusProposed); err != nil {
		return err
	}
	if d.narrative.IsEmpty() {
		return ErrNarrativeRequiredToPropose
	}
	d.raise(events.NewDirectionProposed(d.ID()))
	return nil
}

func (d *Direction) Agree() error {
	if err := d.requireTransition(valueobjects.DirectionStatusAgreed); err != nil {
		return err
	}
	d.raise(events.NewDirectionAgreed(d.ID()))
	return nil
}

func (d *Direction) Reject() error {
	if !d.status.CanReject() {
		return ErrInvalidStatusTransition
	}
	d.raise(events.NewDirectionRejected(d.ID()))
	return nil
}

func (d *Direction) UpdateNarrative(narrative valueobjects.Narrative) error {
	if err := d.requireEditable(); err != nil {
		return err
	}
	d.raise(events.NewDirectionNarrativeUpdated(d.ID(), narrative.Value()))
	return nil
}

func (d *Direction) ChangeHorizon(horizon valueobjects.Horizon) error {
	if err := d.requireEditable(); err != nil {
		return err
	}
	d.raise(events.NewDirectionHorizonChanged(d.ID(), horizon.Value()))
	return nil
}

func (d *Direction) ChangeSourceCapabilities(refs []valueobjects.PhysicalCapabilityRef) error {
	if err := d.requireEditable(); err != nil {
		return err
	}
	if err := validateSourceCardinality(d.directionType, refs); err != nil {
		return err
	}
	d.raise(events.NewDirectionSourceCapabilitiesChanged(d.ID(), refsToStrings(refs)))
	return nil
}

func (d *Direction) ChangePlacements(placements []valueobjects.Placement) error {
	if err := d.requireEditable(); err != nil {
		return err
	}
	if err := validatePlacementCardinality(d.directionType, placements); err != nil {
		return err
	}
	d.raise(events.NewDirectionPlacementsChanged(d.ID(), placementsToData(placements)))
	return nil
}

func (d *Direction) EnterpriseCapabilityID() valueobjects.EnterpriseCapabilityRef {
	return d.enterpriseCapabilityID
}
func (d *Direction) Type() valueobjects.DirectionType     { return d.directionType }
func (d *Direction) Status() valueobjects.DirectionStatus { return d.status }
func (d *Direction) Horizon() valueobjects.Horizon        { return d.horizon }
func (d *Direction) Narrative() valueobjects.Narrative    { return d.narrative }
func (d *Direction) Placements() []valueobjects.Placement {
	out := make([]valueobjects.Placement, len(d.placements))
	copy(out, d.placements)
	return out
}
func (d *Direction) SourceCapabilityIDs() []valueobjects.PhysicalCapabilityRef {
	out := make([]valueobjects.PhysicalCapabilityRef, len(d.sourceCapabilityIDs))
	copy(out, d.sourceCapabilityIDs)
	return out
}

func (d *Direction) requireTransition(target string) error {
	targetStatus, _ := valueobjects.NewDirectionStatus(target)
	if !d.status.CanAdvanceTo(targetStatus) {
		return ErrInvalidStatusTransition
	}
	return nil
}

func (d *Direction) requireEditable() error {
	if d.status.IsAgreed() || d.status.IsRejected() {
		return ErrDirectionAgreedImmutable
	}
	return nil
}

func (d *Direction) raise(event domain.DomainEvent) {
	// In-process events constructed from valid value objects; an apply error
	// here is a programmer error, not a corrupted store.
	if err := d.apply(event); err != nil {
		panic(fmt.Sprintf("architecturedirection: in-process apply failed: %v", err))
	}
	d.RaiseEvent(event)
}

func (d *Direction) apply(event domain.DomainEvent) error {
	if drafted, ok := event.(events.DirectionDrafted); ok {
		return d.applyDrafted(drafted)
	}
	if ok, err := d.applyStatusTransition(event); ok {
		return err
	}
	return d.applyFieldUpdate(event)
}

func (d *Direction) applyStatusTransition(event domain.DomainEvent) (bool, error) {
	var target string
	switch event.(type) {
	case events.DirectionProposed:
		target = valueobjects.DirectionStatusProposed
	case events.DirectionAgreed:
		target = valueobjects.DirectionStatusAgreed
	case events.DirectionRejected:
		target = valueobjects.DirectionStatusRejected
	default:
		return false, nil
	}
	status, err := valueobjects.NewDirectionStatus(target)
	if err != nil {
		return true, fmt.Errorf("%w: invalid status %q: %v", ErrCorruptedEvent, target, err)
	}
	d.status = status
	return true, nil
}

func (d *Direction) applyFieldUpdate(event domain.DomainEvent) error {
	switch evt := event.(type) {
	case events.DirectionNarrativeUpdated:
		return d.applyNarrativeUpdated(evt)
	case events.DirectionHorizonChanged:
		return d.applyHorizonChanged(evt)
	case events.DirectionSourceCapabilitiesChanged:
		return d.applySourceCapabilitiesChanged(evt)
	case events.DirectionPlacementsChanged:
		return d.applyPlacementsChanged(evt)
	}
	return nil
}

func (d *Direction) applyNarrativeUpdated(evt events.DirectionNarrativeUpdated) error {
	n, err := valueobjects.NewNarrative(evt.Narrative)
	if err != nil {
		return fmt.Errorf("%w: narrative: %v", ErrCorruptedEvent, err)
	}
	d.narrative = n
	return nil
}

func (d *Direction) applyHorizonChanged(evt events.DirectionHorizonChanged) error {
	h, err := valueobjects.NewHorizon(evt.Horizon)
	if err != nil {
		return fmt.Errorf("%w: horizon %q: %v", ErrCorruptedEvent, evt.Horizon, err)
	}
	d.horizon = h
	return nil
}

func (d *Direction) applySourceCapabilitiesChanged(evt events.DirectionSourceCapabilitiesChanged) error {
	refs, err := decodePhysicalRefs(evt.SourceCapabilityIDs)
	if err != nil {
		return err
	}
	d.sourceCapabilityIDs = refs
	return nil
}

func (d *Direction) applyPlacementsChanged(evt events.DirectionPlacementsChanged) error {
	placements, err := decodePlacements(evt.Placements)
	if err != nil {
		return err
	}
	d.placements = placements
	return nil
}

func (d *Direction) applyDrafted(evt events.DirectionDrafted) error {
	ecRef, err := valueobjects.NewEnterpriseCapabilityRef(evt.EnterpriseCapabilityID)
	if err != nil {
		return fmt.Errorf("%w: enterprise capability ref %q: %v", ErrCorruptedEvent, evt.EnterpriseCapabilityID, err)
	}
	dt, err := valueobjects.NewDirectionType(evt.Type)
	if err != nil {
		return fmt.Errorf("%w: direction type %q: %v", ErrCorruptedEvent, evt.Type, err)
	}
	sourceRefs, err := decodePhysicalRefs(evt.SourceCapabilityIDs)
	if err != nil {
		return err
	}
	placements, err := decodePlacements(evt.Placements)
	if err != nil {
		return err
	}
	horizon, err := valueobjects.NewHorizon(evt.Horizon)
	if err != nil {
		return fmt.Errorf("%w: horizon %q: %v", ErrCorruptedEvent, evt.Horizon, err)
	}
	narrative, err := valueobjects.NewNarrative(evt.Narrative)
	if err != nil {
		return fmt.Errorf("%w: narrative: %v", ErrCorruptedEvent, err)
	}
	status, err := valueobjects.NewDirectionStatus(valueobjects.DirectionStatusDraft)
	if err != nil {
		return fmt.Errorf("%w: status: %v", ErrCorruptedEvent, err)
	}
	d.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	d.enterpriseCapabilityID = ecRef
	d.directionType = dt
	d.sourceCapabilityIDs = sourceRefs
	d.placements = placements
	d.horizon = horizon
	d.status = status
	d.narrative = narrative
	return nil
}

func validateSourceCardinality(t valueobjects.DirectionType, refs []valueobjects.PhysicalCapabilityRef) error {
	if hasDuplicateRefs(refs) {
		return ErrDuplicateSourceCapabilities
	}
	if t.RequiresExactlyOneSource() {
		if len(refs) != t.ExactSourceCount() {
			return ErrInvalidSourceCardinality
		}
		return nil
	}
	if len(refs) < t.MinSourceCount() {
		return ErrInvalidSourceCardinality
	}
	return nil
}

func validatePlacementCardinality(t valueobjects.DirectionType, placements []valueobjects.Placement) error {
	if !t.IsValidPlacementCount(len(placements)) {
		return ErrInvalidPlacementCardinality
	}
	return nil
}

func hasDuplicateRefs(refs []valueobjects.PhysicalCapabilityRef) bool {
	seen := make(map[string]struct{}, len(refs))
	for _, r := range refs {
		if _, ok := seen[r.Value()]; ok {
			return true
		}
		seen[r.Value()] = struct{}{}
	}
	return false
}

func refsToStrings(refs []valueobjects.PhysicalCapabilityRef) []string {
	out := make([]string, len(refs))
	for i, r := range refs {
		out[i] = r.Value()
	}
	return out
}

func placementsToData(placements []valueobjects.Placement) []events.PlacementData {
	out := make([]events.PlacementData, len(placements))
	for i, p := range placements {
		out[i] = events.PlacementData{
			TargetBusinessDomainID: p.TargetBusinessDomainID(),
			ResultingName:          p.ResultingName(),
		}
	}
	return out
}

func decodePhysicalRefs(values []string) ([]valueobjects.PhysicalCapabilityRef, error) {
	out := make([]valueobjects.PhysicalCapabilityRef, len(values))
	for i, v := range values {
		r, err := valueobjects.NewPhysicalCapabilityRef(v)
		if err != nil {
			return nil, fmt.Errorf("%w: physical capability ref %q: %v", ErrCorruptedEvent, v, err)
		}
		out[i] = r
	}
	return out, nil
}

func decodePlacements(values []events.PlacementData) ([]valueobjects.Placement, error) {
	out := make([]valueobjects.Placement, len(values))
	for i, v := range values {
		p, err := valueobjects.NewPlacement(v.TargetBusinessDomainID, v.ResultingName)
		if err != nil {
			return nil, fmt.Errorf("%w: placement %d: %v", ErrCorruptedEvent, i, err)
		}
		out[i] = p
	}
	return out, nil
}
