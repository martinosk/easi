package handlers

import (
	"context"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

type CapabilityJourneyRepository interface {
	Save(ctx context.Context, j *aggregates.CapabilityJourney) error
	GetByID(ctx context.Context, id string) (*aggregates.CapabilityJourney, error)
}

type ActiveJourneyLookup interface {
	FindActiveJourneyIDForCapability(ctx context.Context, capabilityID string) (string, bool, error)
}

type PlanJourneyHandler struct {
	repo   CapabilityJourneyRepository
	lookup ActiveJourneyLookup
	refs   JourneyReferenceChecks
}

func NewPlanJourneyHandler(repo CapabilityJourneyRepository, lookup ActiveJourneyLookup, refs JourneyReferenceChecks) *PlanJourneyHandler {
	return &PlanJourneyHandler{repo: repo, lookup: lookup, refs: refs}
}

func (h *PlanJourneyHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.PlanJourney)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}
	facts, err := parsePlanJourneyFacts(command)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.ensureNoActiveJourney(ctx, facts.CapabilityID.Value()); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.verifyReferences(ctx, facts); err != nil {
		return cqrs.EmptyResult(), err
	}
	journey, err := aggregates.PlanCapabilityJourney(facts)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.repo.Save(ctx, journey); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.NewResult(journey.ID()), nil
}

func (h *PlanJourneyHandler) ensureNoActiveJourney(ctx context.Context, capabilityID string) error {
	existingID, exists, err := h.lookup.FindActiveJourneyIDForCapability(ctx, capabilityID)
	if err != nil {
		return err
	}
	if exists {
		return &services.ActiveJourneyError{ExistingJourneyID: existingID}
	}
	return nil
}

func (h *PlanJourneyHandler) verifyReferences(ctx context.Context, facts aggregates.CapabilityJourneyFacts) error {
	if err := requireCapabilityExists(ctx, h.refs.CapabilityExists, facts.CapabilityID.Value()); err != nil {
		return err
	}
	if err := requireComponentExists(ctx, h.refs.ComponentExists, facts.ToApp.Value()); err != nil {
		return err
	}
	if err := verifyComponentsExist(ctx, h.refs.ComponentExists, applicationRefValues(facts.FromApps)); err != nil {
		return err
	}
	return h.verifyMoveReferences(ctx, facts)
}

func (h *PlanJourneyHandler) verifyMoveReferences(ctx context.Context, facts aggregates.CapabilityJourneyFacts) error {
	if !facts.Kind.IsMove() || facts.TargetDomain == nil {
		return nil
	}
	if err := requireDomainExists(ctx, h.refs.DomainExists, facts.TargetDomain.Value()); err != nil {
		return err
	}
	if facts.TargetParent == nil {
		return nil
	}
	return requireCapabilityEffectivelyInDomain(ctx, h.refs.CapabilityEffectivelyInDomain, facts.TargetParent.Value(), facts.TargetDomain.Value())
}

type planJourneyCoreFacts struct {
	capability valueobjects.PhysicalCapabilityRef
	kind       valueobjects.JourneyKind
	fromApps   []valueobjects.ApplicationRef
	toApp      valueobjects.ApplicationRef
	note       sharedvo.Description
}

func parsePlanJourneyCoreFacts(cmd *commands.PlanJourney) (planJourneyCoreFacts, error) {
	capability, err := valueobjects.NewPhysicalCapabilityRef(cmd.CapabilityID)
	if err != nil {
		return planJourneyCoreFacts{}, err
	}
	kind, err := valueobjects.NewJourneyKind(cmd.Kind)
	if err != nil {
		return planJourneyCoreFacts{}, err
	}
	fromApps, err := parseApplicationRefs(cmd.FromComponentIDs)
	if err != nil {
		return planJourneyCoreFacts{}, err
	}
	toApp, err := valueobjects.NewApplicationRef(cmd.ToComponentID)
	if err != nil {
		return planJourneyCoreFacts{}, err
	}
	note, err := sharedvo.NewDescription(cmd.Note)
	if err != nil {
		return planJourneyCoreFacts{}, err
	}
	return planJourneyCoreFacts{capability: capability, kind: kind, fromApps: fromApps, toApp: toApp, note: note}, nil
}

type planJourneyTargetFacts struct {
	period *valueobjects.TargetPeriod
	domain *valueobjects.BusinessDomainRef
	parent *valueobjects.PhysicalCapabilityRef
}

func parsePlanJourneyTargetFacts(cmd *commands.PlanJourney) (planJourneyTargetFacts, error) {
	targetPeriod, err := buildTargetPeriod(cmd.TargetYear, cmd.TargetQuarter)
	if err != nil {
		return planJourneyTargetFacts{}, err
	}
	targetDomain, err := parseOptionalBusinessDomainRef(cmd.TargetDomainID)
	if err != nil {
		return planJourneyTargetFacts{}, err
	}
	targetParent, err := parseOptionalPhysicalCapabilityRef(cmd.TargetParentID)
	if err != nil {
		return planJourneyTargetFacts{}, err
	}
	return planJourneyTargetFacts{period: targetPeriod, domain: targetDomain, parent: targetParent}, nil
}

func parsePlanJourneyFacts(cmd *commands.PlanJourney) (aggregates.CapabilityJourneyFacts, error) {
	core, err := parsePlanJourneyCoreFacts(cmd)
	if err != nil {
		return aggregates.CapabilityJourneyFacts{}, err
	}
	target, err := parsePlanJourneyTargetFacts(cmd)
	if err != nil {
		return aggregates.CapabilityJourneyFacts{}, err
	}
	return aggregates.CapabilityJourneyFacts{
		ID:            valueobjects.NewCapabilityJourneyID(),
		CapabilityID:  core.capability,
		Kind:          core.kind,
		FromApps:      core.fromApps,
		ToApp:         core.toApp,
		Note:          core.note,
		TargetPeriod:  target.period,
		TargetDomain:  target.domain,
		TargetParent:  target.parent,
		ResultingName: cmd.ResultingName,
		PlannedBy:     cmd.PlannedBy,
	}, nil
}
