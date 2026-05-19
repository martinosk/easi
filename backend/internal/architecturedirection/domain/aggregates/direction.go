package aggregates

import (
	"errors"

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
	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})
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
	d.apply(event)
	d.RaiseEvent(event)
}

func (d *Direction) apply(event domain.DomainEvent) {
	if drafted, ok := event.(events.DirectionDrafted); ok {
		d.applyDrafted(drafted)
		return
	}
	if d.applyStatusTransition(event) {
		return
	}
	d.applyFieldUpdate(event)
}

func (d *Direction) applyStatusTransition(event domain.DomainEvent) bool {
	switch event.(type) {
	case events.DirectionProposed:
		d.status = mustStatus(valueobjects.DirectionStatusProposed)
	case events.DirectionAgreed:
		d.status = mustStatus(valueobjects.DirectionStatusAgreed)
	case events.DirectionRejected:
		d.status = mustStatus(valueobjects.DirectionStatusRejected)
	default:
		return false
	}
	return true
}

func (d *Direction) applyFieldUpdate(event domain.DomainEvent) {
	switch evt := event.(type) {
	case events.DirectionNarrativeUpdated:
		d.narrative = mustNarrative(evt.Narrative)
	case events.DirectionHorizonChanged:
		d.horizon = mustHorizon(evt.Horizon)
	case events.DirectionSourceCapabilitiesChanged:
		d.sourceCapabilityIDs = mustPhysicalRefs(evt.SourceCapabilityIDs)
	case events.DirectionPlacementsChanged:
		d.placements = mustPlacements(evt.Placements)
	}
}

func (d *Direction) applyDrafted(evt events.DirectionDrafted) {
	d.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	d.enterpriseCapabilityID = mustECRef(evt.EnterpriseCapabilityID)
	d.directionType = mustDirectionType(evt.Type)
	d.sourceCapabilityIDs = mustPhysicalRefs(evt.SourceCapabilityIDs)
	d.placements = mustPlacements(evt.Placements)
	d.horizon = mustHorizon(evt.Horizon)
	d.status = mustStatus(valueobjects.DirectionStatusDraft)
	d.narrative = mustNarrative(evt.Narrative)
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

func mustStatus(value string) valueobjects.DirectionStatus {
	s, err := valueobjects.NewDirectionStatus(value)
	if err != nil {
		panic("corrupted event store: invalid direction status: " + value)
	}
	return s
}

func mustHorizon(value string) valueobjects.Horizon {
	h, err := valueobjects.NewHorizon(value)
	if err != nil {
		panic("corrupted event store: invalid horizon: " + value)
	}
	return h
}

func mustNarrative(value string) valueobjects.Narrative {
	n, err := valueobjects.NewNarrative(value)
	if err != nil {
		panic("corrupted event store: invalid narrative")
	}
	return n
}

func mustDirectionType(value string) valueobjects.DirectionType {
	t, err := valueobjects.NewDirectionType(value)
	if err != nil {
		panic("corrupted event store: invalid direction type: " + value)
	}
	return t
}

func mustECRef(value string) valueobjects.EnterpriseCapabilityRef {
	r, err := valueobjects.NewEnterpriseCapabilityRef(value)
	if err != nil {
		panic("corrupted event store: invalid enterprise capability ref: " + value)
	}
	return r
}

func mustPhysicalRefs(values []string) []valueobjects.PhysicalCapabilityRef {
	out := make([]valueobjects.PhysicalCapabilityRef, len(values))
	for i, v := range values {
		r, err := valueobjects.NewPhysicalCapabilityRef(v)
		if err != nil {
			panic("corrupted event store: invalid physical capability ref: " + v)
		}
		out[i] = r
	}
	return out
}

func mustPlacements(values []events.PlacementData) []valueobjects.Placement {
	out := make([]valueobjects.Placement, len(values))
	for i, v := range values {
		p, err := valueobjects.NewPlacement(v.TargetBusinessDomainID, v.ResultingName)
		if err != nil {
			panic("corrupted event store: invalid placement")
		}
		out[i] = p
	}
	return out
}
