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
	ErrJourneyTargetAmongSources       = errors.New("journey target application must not be among the from-applications")
	ErrJourneyMoveRequiresTargetDomain = errors.New("move journey requires a target business domain")
	ErrJourneyMoveFieldsOnNonMove      = errors.New("move fields are only valid for move journeys")
	ErrInvalidJourneyTransition        = errors.New("journey transition not allowed from its current status")
	ErrJourneyFrozen                   = errors.New("journey is frozen and can no longer be edited")
	ErrJourneyMilestoneNotFound        = errors.New("milestone not found on journey")
	ErrCorruptedCapabilityJourneyEvent = errors.New("corrupted event store: cannot rehydrate capability journey")
	ErrUnknownCapabilityJourneyEvent   = errors.New("unknown event type for capability journey aggregate")
)

type CapabilityJourney struct {
	domain.AggregateRoot
	capabilityID  valueobjects.PhysicalCapabilityRef
	kind          valueobjects.JourneyKind
	status        valueobjects.JourneyStatus
	fromApps      []valueobjects.ApplicationRef
	toApp         valueobjects.ApplicationRef
	progress      *valueobjects.JourneyProgress
	targetPeriod  *valueobjects.TargetPeriod
	note          sharedvo.Description
	milestones    []entities.Milestone
	targetDomain  *valueobjects.BusinessDomainRef
	targetParent  *valueobjects.PhysicalCapabilityRef
	resultingName string
}

type CapabilityJourneyFacts struct {
	ID            valueobjects.CapabilityJourneyID
	CapabilityID  valueobjects.PhysicalCapabilityRef
	Kind          valueobjects.JourneyKind
	FromApps      []valueobjects.ApplicationRef
	ToApp         valueobjects.ApplicationRef
	Note          sharedvo.Description
	TargetPeriod  *valueobjects.TargetPeriod
	TargetDomain  *valueobjects.BusinessDomainRef
	TargetParent  *valueobjects.PhysicalCapabilityRef
	ResultingName string
	PlannedBy     string
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
	if !j.hasMilestone(facts.MilestoneID) {
		return ErrJourneyMilestoneNotFound
	}
	j.raise(events.NewJourneyMilestoneUpdated(j.milestoneEventFields(milestone, facts)))
	return nil
}

func (j *CapabilityJourney) RemoveMilestone(milestoneID, actor string) error {
	if err := j.requireActive(); err != nil {
		return err
	}
	if !j.hasMilestone(milestoneID) {
		return ErrJourneyMilestoneNotFound
	}
	j.raise(events.NewJourneyMilestoneRemoved(events.JourneyMilestoneRemovedFields{
		ID:          j.ID(),
		MilestoneID: milestoneID,
		RemovedBy:   actor,
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

func (j *CapabilityJourney) FromApps() []valueobjects.ApplicationRef {
	out := make([]valueobjects.ApplicationRef, len(j.fromApps))
	copy(out, j.fromApps)
	return out
}

func (j *CapabilityJourney) Milestones() []entities.Milestone {
	out := make([]entities.Milestone, len(j.milestones))
	copy(out, j.milestones)
	return out
}

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

func (j *CapabilityJourney) hasMilestone(id string) bool {
	for _, m := range j.milestones {
		if m.ID() == id {
			return true
		}
	}
	return false
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
		return j.applyMilestoneUpsert(milestoneSnapshot{id: evt.MilestoneID, label: evt.Label, targetPeriod: evt.TargetPeriod, status: evt.Status})
	case events.JourneyMilestoneUpdated:
		return j.applyMilestoneUpsert(milestoneSnapshot{id: evt.MilestoneID, label: evt.Label, targetPeriod: evt.TargetPeriod, status: evt.Status})
	case events.JourneyMilestoneRemoved:
		return j.applyMilestoneRemoved(evt)
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
	j.status = status
	j.milestones = []entities.Milestone{}
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

type milestoneSnapshot struct {
	id           string
	label        string
	targetPeriod *events.TargetPeriodData
	status       string
}

func (j *CapabilityJourney) applyMilestoneUpsert(snapshot milestoneSnapshot) error {
	milestone, err := decodeMilestone(snapshot)
	if err != nil {
		return err
	}
	for i, existing := range j.milestones {
		if existing.ID() == milestone.ID() {
			j.milestones[i] = milestone
			return nil
		}
	}
	j.milestones = append(j.milestones, milestone)
	return nil
}

func (j *CapabilityJourney) applyMilestoneRemoved(evt events.JourneyMilestoneRemoved) error {
	updated := make([]entities.Milestone, 0, len(j.milestones))
	for _, m := range j.milestones {
		if m.ID() != evt.MilestoneID {
			updated = append(updated, m)
		}
	}
	j.milestones = updated
	return nil
}

type plannedTransition struct {
	capabilityID valueobjects.PhysicalCapabilityRef
	kind         valueobjects.JourneyKind
	fromApps     []valueobjects.ApplicationRef
	toApp        valueobjects.ApplicationRef
	note         sharedvo.Description
}

func decodePlannedTransition(evt events.JourneyPlanned) (plannedTransition, error) {
	capabilityID, err := valueobjects.NewPhysicalCapabilityRef(evt.CapabilityID)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: capability ref %q: %v", ErrCorruptedCapabilityJourneyEvent, evt.CapabilityID, err)
	}
	kind, err := valueobjects.NewJourneyKind(evt.Kind)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: kind %q: %v", ErrCorruptedCapabilityJourneyEvent, evt.Kind, err)
	}
	fromApps, err := decodeApplicationRefs(evt.FromComponentIDs)
	if err != nil {
		return plannedTransition{}, err
	}
	toApp, err := valueobjects.NewApplicationRef(evt.ToComponentID)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: to component ref %q: %v", ErrCorruptedCapabilityJourneyEvent, evt.ToComponentID, err)
	}
	note, err := sharedvo.NewDescription(evt.Note)
	if err != nil {
		return plannedTransition{}, fmt.Errorf("%w: note: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	return plannedTransition{
		capabilityID: capabilityID,
		kind:         kind,
		fromApps:     fromApps,
		toApp:        toApp,
		note:         note,
	}, nil
}

type plannedDestination struct {
	targetPeriod *valueobjects.TargetPeriod
	targetDomain *valueobjects.BusinessDomainRef
	targetParent *valueobjects.PhysicalCapabilityRef
}

func decodePlannedDestination(evt events.JourneyPlanned) (plannedDestination, error) {
	targetPeriod, err := decodeTargetPeriod(evt.TargetPeriod)
	if err != nil {
		return plannedDestination{}, err
	}
	targetDomain, err := decodeOptionalRef(evt.TargetDomainID, "target domain ref", valueobjects.NewBusinessDomainRef)
	if err != nil {
		return plannedDestination{}, err
	}
	targetParent, err := decodeOptionalRef(evt.TargetParentID, "target parent ref", valueobjects.NewPhysicalCapabilityRef)
	if err != nil {
		return plannedDestination{}, err
	}
	return plannedDestination{
		targetPeriod: targetPeriod,
		targetDomain: targetDomain,
		targetParent: targetParent,
	}, nil
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

func decodeMilestone(snapshot milestoneSnapshot) (entities.Milestone, error) {
	targetPeriod, err := decodeTargetPeriod(snapshot.targetPeriod)
	if err != nil {
		return entities.Milestone{}, err
	}
	milestoneStatus, err := valueobjects.NewMilestoneStatus(snapshot.status)
	if err != nil {
		return entities.Milestone{}, fmt.Errorf("%w: milestone status %q: %v", ErrCorruptedCapabilityJourneyEvent, snapshot.status, err)
	}
	milestone, err := entities.NewMilestone(snapshot.id, snapshot.label, targetPeriod, milestoneStatus)
	if err != nil {
		return entities.Milestone{}, fmt.Errorf("%w: milestone: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	return milestone, nil
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

func decodeApplicationRefs(values []string) ([]valueobjects.ApplicationRef, error) {
	out := make([]valueobjects.ApplicationRef, len(values))
	for i, v := range values {
		ref, err := valueobjects.NewApplicationRef(v)
		if err != nil {
			return nil, fmt.Errorf("%w: component ref %q: %v", ErrCorruptedCapabilityJourneyEvent, v, err)
		}
		out[i] = ref
	}
	return out, nil
}

func optionalRefValue[T interface{ Value() string }](ref *T) string {
	if ref == nil {
		return ""
	}
	return (*ref).Value()
}

func decodeOptionalRef[T any](value, refName string, construct func(string) (T, error)) (*T, error) {
	if value == "" {
		return nil, nil
	}
	ref, err := construct(value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s %q: %v", ErrCorruptedCapabilityJourneyEvent, refName, value, err)
	}
	return &ref, nil
}

func targetPeriodToData(tp *valueobjects.TargetPeriod) *events.TargetPeriodData {
	if tp == nil {
		return nil
	}
	return &events.TargetPeriodData{Year: tp.Year(), Quarter: tp.Quarter()}
}

func decodeTargetPeriod(data *events.TargetPeriodData) (*valueobjects.TargetPeriod, error) {
	if data == nil {
		return nil, nil
	}
	tp, err := valueobjects.NewTargetPeriod(data.Year, data.Quarter)
	if err != nil {
		return nil, fmt.Errorf("%w: target period: %v", ErrCorruptedCapabilityJourneyEvent, err)
	}
	return &tp, nil
}
