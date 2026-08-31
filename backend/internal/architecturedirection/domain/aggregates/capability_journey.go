package aggregates

import (
	"errors"
	"fmt"

	"easi/backend/internal/architecturedirection/domain/entities"
	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrJourneyTargetAmongSources            = errors.New("journey target application must not be among the from-applications")
	ErrJourneyMoveRequiresTargetDomain      = errors.New("move journey requires a target business domain")
	ErrJourneyMoveFieldsOnNonMove           = errors.New("move fields are only valid for move journeys")
	ErrJourneyTargetMaturityOnNonMaturity   = errors.New("target maturity is only valid for maturity journeys")
	ErrJourneyMaturityRequiresTarget        = errors.New("maturity journey requires a target maturity")
	ErrJourneyMaturityRefusesApplications   = errors.New("maturity journeys carry no applications")
	ErrJourneyMaturityTargetNotAboveCurrent = errors.New("target maturity must exceed the capability's current maturity")
	ErrInvalidJourneyTransition             = errors.New("journey transition not allowed from its current status")
	ErrJourneyFrozen                        = errors.New("journey is frozen and can no longer be edited")
	ErrJourneyMilestoneNotFound             = entities.ErrMilestoneNotFound
	ErrJourneyMilestoneOrderIncomplete      = entities.ErrMilestoneOrderIncomplete
	ErrJourneyMilestoneOrderDuplicate       = entities.ErrMilestoneOrderDuplicate
	ErrJourneyMilestoneOrderUnchanged       = errors.New("milestone order is unchanged")
	ErrCorruptedCapabilityJourneyEvent      = errors.New("corrupted event store: cannot rehydrate capability journey")
	ErrUnknownCapabilityJourneyEvent        = errors.New("unknown event type for capability journey aggregate")
)

type CapabilityJourney struct {
	domain.AggregateRoot
	capabilityID   valueobjects.PhysicalCapabilityRef
	kind           valueobjects.JourneyKind
	status         valueobjects.JourneyStatus
	fromApps       []valueobjects.ApplicationRef
	toApp          valueobjects.ApplicationRef
	progress       *valueobjects.JourneyProgress
	targetPeriod   *valueobjects.TargetPeriod
	note           sharedvo.Description
	milestones     entities.Milestones
	targetDomain   *valueobjects.BusinessDomainRef
	targetParent   *valueobjects.PhysicalCapabilityRef
	resultingName  string
	targetMaturity *valueobjects.TargetMaturity
}

type CapabilityJourneyFacts struct {
	ID              valueobjects.CapabilityJourneyID
	CapabilityID    valueobjects.PhysicalCapabilityRef
	Kind            valueobjects.JourneyKind
	FromApps        []valueobjects.ApplicationRef
	ToApp           valueobjects.ApplicationRef
	Note            sharedvo.Description
	TargetPeriod    *valueobjects.TargetPeriod
	TargetDomain    *valueobjects.BusinessDomainRef
	TargetParent    *valueobjects.PhysicalCapabilityRef
	ResultingName   string
	TargetMaturity  *valueobjects.TargetMaturity
	CurrentMaturity int
	PlannedBy       string
}

type MilestoneFacts struct {
	MilestoneID  string
	Label        string
	TargetPeriod *valueobjects.TargetPeriod
	Status       valueobjects.MilestoneStatus
	Actor        string
}

type JourneyDetailsFacts struct {
	Note          sharedvo.Description
	TargetPeriod  *valueobjects.TargetPeriod
	ResultingName string
	Actor         string
}

func PlanCapabilityJourney(facts CapabilityJourneyFacts) (*CapabilityJourney, error) {
	if err := facts.Kind.ValidateSourceCount(len(facts.FromApps)); err != nil {
		return nil, err
	}
	if containsApplicationRef(facts.FromApps, facts.ToApp) {
		return nil, ErrJourneyTargetAmongSources
	}
	resultingName, err := validateMoveFields(facts)
	if err != nil {
		return nil, err
	}
	if err := validateMaturityFields(facts); err != nil {
		return nil, err
	}

	aggregate := &CapabilityJourney{
		AggregateRoot: domain.NewAggregateRootWithID(facts.ID.Value()),
	}
	aggregate.raise(events.NewJourneyPlanned(events.JourneyPlannedFields{
		ID:               facts.ID.Value(),
		CapabilityID:     facts.CapabilityID.Value(),
		Kind:             facts.Kind.Value(),
		FromComponentIDs: applicationRefsToStrings(facts.FromApps),
		ToComponentID:    facts.ToApp.Value(),
		Note:             facts.Note.Value(),
		TargetPeriod:     targetPeriodToData(facts.TargetPeriod),
		TargetDomainID:   optionalRefValue(facts.TargetDomain),
		TargetParentID:   optionalRefValue(facts.TargetParent),
		ResultingName:    resultingName,
		TargetMaturity:   targetMaturityToData(facts.TargetMaturity),
		PlannedBy:        facts.PlannedBy,
	}))
	return aggregate, nil
}

func LoadCapabilityJourneyFromHistory(eventHistory []domain.DomainEvent) (*CapabilityJourney, error) {
	aggregate := &CapabilityJourney{
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

func (j *CapabilityJourney) Start(actor string) error {
	if !j.status.CanStart() {
		return ErrInvalidJourneyTransition
	}
	j.raise(events.NewJourneyStarted(events.JourneyStartedFields{ID: j.ID(), StartedBy: actor}))
	return nil
}

func (j *CapabilityJourney) Complete(actor string) error {
	if !j.status.CanComplete() {
		return ErrInvalidJourneyTransition
	}
	j.raise(events.NewJourneyCompleted(events.JourneyCompletedFields{ID: j.ID(), CompletedBy: actor}))
	return nil
}

func (j *CapabilityJourney) Abandon(actor string) error {
	if !j.status.CanAbandon() {
		return ErrInvalidJourneyTransition
	}
	j.raise(events.NewJourneyAbandoned(events.JourneyAbandonedFields{ID: j.ID(), AbandonedBy: actor}))
	return nil
}

func (j *CapabilityJourney) UpdateProgress(progress valueobjects.JourneyProgress, actor string) error {
	if err := j.requireActive(); err != nil {
		return err
	}
	j.raise(events.NewJourneyProgressUpdated(events.JourneyProgressUpdatedFields{
		ID:        j.ID(),
		Progress:  progress.Value(),
		UpdatedBy: actor,
	}))
	return nil
}

func (j *CapabilityJourney) UpdateDetails(details JourneyDetailsFacts) error {
	if err := j.requireActive(); err != nil {
		return err
	}
	resolvedName, err := j.resolveResultingNameUpdate(details.ResultingName)
	if err != nil {
		return err
	}
	j.raise(events.NewJourneyDetailsUpdated(events.JourneyDetailsUpdatedFields{
		ID:            j.ID(),
		Note:          details.Note.Value(),
		TargetPeriod:  targetPeriodToData(details.TargetPeriod),
		ResultingName: resolvedName,
		UpdatedBy:     details.Actor,
	}))
	return nil
}

func (j *CapabilityJourney) ChangeSourceApplications(fromApps []valueobjects.ApplicationRef, actor string) error {
	if err := j.requireActive(); err != nil {
		return err
	}
	if err := j.kind.ValidateSourceCount(len(fromApps)); err != nil {
		return err
	}
	if containsApplicationRef(fromApps, j.toApp) {
		return ErrJourneyTargetAmongSources
	}
	j.raise(events.NewJourneySourceApplicationsChanged(events.JourneySourceApplicationsChangedFields{
		ID:               j.ID(),
		FromComponentIDs: applicationRefsToStrings(fromApps),
		ChangedBy:        actor,
	}))
	return nil
}

func (j *CapabilityJourney) AddMilestone(facts MilestoneFacts) error {
	milestone, err := j.buildMilestone(facts)
	if err != nil {
		return err
	}
	j.raise(events.NewJourneyMilestoneAdded(j.milestoneEventFields(milestone, facts)))
	return nil
}

func (j *CapabilityJourney) UpdateMilestone(facts MilestoneFacts) error {
	milestone, err := j.buildMilestone(facts)
	if err != nil {
		return err
	}
	if !j.milestones.Has(facts.MilestoneID) {
		return ErrJourneyMilestoneNotFound
	}
	j.raise(events.NewJourneyMilestoneUpdated(j.milestoneEventFields(milestone, facts)))
	return nil
}

func (j *CapabilityJourney) RemoveMilestone(milestoneID, actor string) error {
	if err := j.requireActive(); err != nil {
		return err
	}
	if !j.milestones.Has(milestoneID) {
		return ErrJourneyMilestoneNotFound
	}
	j.raise(events.NewJourneyMilestoneRemoved(events.JourneyMilestoneRemovedFields{
		ID:          j.ID(),
		MilestoneID: milestoneID,
		RemovedBy:   actor,
	}))
	return nil
}

func (j *CapabilityJourney) ReorderMilestones(milestoneIDs []string, actor string) error {
	if err := j.requireActive(); err != nil {
		return err
	}
	if err := j.milestones.ValidateSequence(milestoneIDs); err != nil {
		return err
	}
	if j.milestones.InSequence(milestoneIDs) {
		return ErrJourneyMilestoneOrderUnchanged
	}
	j.raise(events.NewJourneyMilestonesReordered(events.JourneyMilestonesReorderedFields{
		ID:           j.ID(),
		MilestoneIDs: append([]string(nil), milestoneIDs...),
		ReorderedBy:  actor,
	}))
	return nil
}

func (j *CapabilityJourney) CapabilityID() valueobjects.PhysicalCapabilityRef { return j.capabilityID }
func (j *CapabilityJourney) Kind() valueobjects.JourneyKind                   { return j.kind }
func (j *CapabilityJourney) Status() valueobjects.JourneyStatus               { return j.status }
func (j *CapabilityJourney) ToApp() valueobjects.ApplicationRef               { return j.toApp }
func (j *CapabilityJourney) Progress() *valueobjects.JourneyProgress          { return j.progress }
func (j *CapabilityJourney) TargetPeriod() *valueobjects.TargetPeriod         { return j.targetPeriod }
func (j *CapabilityJourney) Note() sharedvo.Description                       { return j.note }
func (j *CapabilityJourney) TargetDomain() *valueobjects.BusinessDomainRef    { return j.targetDomain }
func (j *CapabilityJourney) TargetParent() *valueobjects.PhysicalCapabilityRef {
	return j.targetParent
}
func (j *CapabilityJourney) ResultingName() string { return j.resultingName }

func (j *CapabilityJourney) TargetMaturity() *valueobjects.TargetMaturity { return j.targetMaturity }

func (j *CapabilityJourney) FromApps() []valueobjects.ApplicationRef {
	out := make([]valueobjects.ApplicationRef, len(j.fromApps))
	copy(out, j.fromApps)
	return out
}

func (j *CapabilityJourney) Milestones() []entities.Milestone { return j.milestones.List() }

func (j *CapabilityJourney) requireActive() error {
	if j.status.IsTerminal() {
		return ErrJourneyFrozen
	}
	return nil
}

func (j *CapabilityJourney) resolveResultingNameUpdate(resultingName string) (string, error) {
	if !j.kind.IsMove() {
		if resultingName != "" {
			return "", ErrJourneyMoveFieldsOnNonMove
		}
		return "", nil
	}
	name, err := valueobjects.NewResultingCapabilityName(resultingName)
	if err != nil {
		return "", err
	}
	return name.Value(), nil
}

func (j *CapabilityJourney) buildMilestone(facts MilestoneFacts) (entities.Milestone, error) {
	if err := j.requireActive(); err != nil {
		return entities.Milestone{}, err
	}
	return entities.NewMilestone(facts.MilestoneID, facts.Label, facts.TargetPeriod, facts.Status)
}

func (j *CapabilityJourney) milestoneEventFields(m entities.Milestone, facts MilestoneFacts) events.JourneyMilestoneFields {
	return events.JourneyMilestoneFields{
		ID:           j.ID(),
		MilestoneID:  m.ID(),
		Label:        m.Label(),
		TargetPeriod: targetPeriodToData(m.TargetPeriod()),
		Status:       m.Status().Value(),
		Actor:        facts.Actor,
	}
}

func (j *CapabilityJourney) raise(event domain.DomainEvent) {
	if err := j.apply(event); err != nil {
		panic(fmt.Sprintf("architecturedirection: in-process apply failed: %v", err))
	}
	j.RaiseEvent(event)
}

func (j *CapabilityJourney) apply(event domain.DomainEvent) error {
	if planned, ok := event.(events.JourneyPlanned); ok {
		return j.applyPlanned(planned)
	}
	if handled, err := j.applyStatusTransition(event); handled {
		return err
	}
	return j.applyFieldUpdate(event)
}

func (j *CapabilityJourney) applyStatusTransition(event domain.DomainEvent) (bool, error) {
	var target string
	switch event.(type) {
	case events.JourneyStarted:
		target = valueobjects.JourneyStatusInFlight
	case events.JourneyCompleted:
		target = valueobjects.JourneyStatusDone
	case events.JourneyAbandoned:
		target = valueobjects.JourneyStatusAbandoned
	default:
		return false, nil
	}
	status, err := valueobjects.NewJourneyStatus(target)
	if err != nil {
		return true, fmt.Errorf("%w: status %q: %v", ErrCorruptedCapabilityJourneyEvent, target, err)
	}
	j.status = status
	return true, nil
}

func (j *CapabilityJourney) applyFieldUpdate(event domain.DomainEvent) error {
	switch evt := event.(type) {
	case events.JourneyProgressUpdated:
		return j.applyProgressUpdated(evt)
	case events.JourneyDetailsUpdated:
		return j.applyDetailsUpdated(evt)
	case events.JourneySourceApplicationsChanged:
		return j.applySourceApplicationsChanged(evt)
	case events.JourneyMilestoneAdded:
		return j.applyMilestoneRecorded(milestoneSnapshot{id: evt.MilestoneID, label: evt.Label, targetPeriod: evt.TargetPeriod, status: evt.Status})
	case events.JourneyMilestoneUpdated:
		return j.applyMilestoneRecorded(milestoneSnapshot{id: evt.MilestoneID, label: evt.Label, targetPeriod: evt.TargetPeriod, status: evt.Status})
	case events.JourneyMilestoneRemoved:
		return j.applyMilestoneRemoved(evt)
	case events.JourneyMilestonesReordered:
		return j.applyMilestonesReordered(evt)
	default:
		return fmt.Errorf("%w: %T", ErrUnknownCapabilityJourneyEvent, event)
	}
}

func (j *CapabilityJourney) applyPlanned(evt events.JourneyPlanned) error {
	transition, err := decodePlannedTransition(evt)
	if err != nil {
		return err
	}
	destination, err := decodePlannedDestination(evt)
	if err != nil {
		return err
	}
	status, err := valueobjects.NewJourneyStatus(valueobjects.JourneyStatusPlanned)
	if err != nil {
		return fmt.Errorf("%w: status: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}

	j.AggregateRoot = domain.NewAggregateRootWithID(evt.ID)
	j.capabilityID = transition.capabilityID
	j.kind = transition.kind
	j.fromApps = transition.fromApps
	j.toApp = transition.toApp
	j.note = transition.note
	j.targetPeriod = destination.targetPeriod
	j.targetDomain = destination.targetDomain
	j.targetParent = destination.targetParent
	j.resultingName = evt.ResultingName
	j.targetMaturity = destination.targetMaturity
	j.status = status
	j.milestones = entities.NoMilestones()
	j.progress = nil
	return nil
}

func (j *CapabilityJourney) applyProgressUpdated(evt events.JourneyProgressUpdated) error {
	progress, err := valueobjects.NewJourneyProgress(evt.Progress)
	if err != nil {
		return fmt.Errorf("%w: progress %d: %v", ErrCorruptedCapabilityJourneyEvent, evt.Progress, err)
	}
	j.progress = &progress
	return nil
}

func (j *CapabilityJourney) applyDetailsUpdated(evt events.JourneyDetailsUpdated) error {
	note, err := sharedvo.NewDescription(evt.Note)
	if err != nil {
		return fmt.Errorf("%w: note: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	targetPeriod, err := decodeTargetPeriod(evt.TargetPeriod)
	if err != nil {
		return err
	}
	j.note = note
	j.targetPeriod = targetPeriod
	j.resultingName = evt.ResultingName
	return nil
}

func (j *CapabilityJourney) applySourceApplicationsChanged(evt events.JourneySourceApplicationsChanged) error {
	fromApps, err := decodeApplicationRefs(evt.FromComponentIDs)
	if err != nil {
		return err
	}
	j.fromApps = fromApps
	return nil
}

func (j *CapabilityJourney) applyMilestoneRecorded(snapshot milestoneSnapshot) error {
	milestone, err := decodeMilestone(snapshot)
	if err != nil {
		return err
	}
	j.milestones = j.milestones.Record(milestone)
	return nil
}

func (j *CapabilityJourney) applyMilestoneRemoved(evt events.JourneyMilestoneRemoved) error {
	j.milestones = j.milestones.Remove(evt.MilestoneID)
	return nil
}

func (j *CapabilityJourney) applyMilestonesReordered(evt events.JourneyMilestonesReordered) error {
	reordered, err := j.milestones.Reorder(evt.MilestoneIDs)
	if err != nil {
		return fmt.Errorf("%w: milestone order: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	j.milestones = reordered
	return nil
}

func (f CapabilityJourneyFacts) carriesMoveFields() bool {
	if f.TargetDomain != nil || f.TargetParent != nil {
		return true
	}
	return f.ResultingName != ""
}

func validateMoveFields(facts CapabilityJourneyFacts) (string, error) {
	if !facts.Kind.IsMove() {
		if facts.carriesMoveFields() {
			return "", ErrJourneyMoveFieldsOnNonMove
		}
		return "", nil
	}
	if facts.TargetDomain == nil {
		return "", ErrJourneyMoveRequiresTargetDomain
	}
	name, err := valueobjects.NewResultingCapabilityName(facts.ResultingName)
	if err != nil {
		return "", err
	}
	return name.Value(), nil
}

func validateMaturityFields(facts CapabilityJourneyFacts) error {
	if !facts.Kind.IsMaturity() {
		if facts.TargetMaturity != nil {
			return ErrJourneyTargetMaturityOnNonMaturity
		}
		return nil
	}
	if facts.ToApp.Value() != "" {
		return ErrJourneyMaturityRefusesApplications
	}
	if facts.TargetMaturity == nil {
		return ErrJourneyMaturityRequiresTarget
	}
	if facts.TargetMaturity.Value() <= facts.CurrentMaturity {
		return ErrJourneyMaturityTargetNotAboveCurrent
	}
	return nil
}

func containsApplicationRef(refs []valueobjects.ApplicationRef, target valueobjects.ApplicationRef) bool {
	for _, r := range refs {
		if r.Value() == target.Value() {
			return true
		}
	}
	return false
}

func applicationRefsToStrings(refs []valueobjects.ApplicationRef) []string {
	out := make([]string, len(refs))
	for i, r := range refs {
		out[i] = r.Value()
	}
	return out
}

func optionalRefValue[T interface{ Value() string }](ref *T) string {
	if ref == nil {
		return ""
	}
	return (*ref).Value()
}

func targetPeriodToData(tp *valueobjects.TargetPeriod) *events.TargetPeriodData {
	if tp == nil {
		return nil
	}
	return &events.TargetPeriodData{Year: tp.Year(), Quarter: tp.Quarter()}
}

func targetMaturityToData(tm *valueobjects.TargetMaturity) *int {
	if tm == nil {
		return nil
	}
	value := tm.Value()
	return &value
}
