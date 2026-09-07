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
	FindActiveJourneyIDForCapability(ctx context.Context, capabilityID string, kinds []string) (string, bool, error)
}

type CapabilityMaturityLookup func(ctx context.Context, capabilityID string) (int, error)

type PlanJourneyHandler struct {
	repo     CapabilityJourneyRepository
	lookup   ActiveJourneyLookup
	refs     JourneyReferenceChecks
	maturity CapabilityMaturityLookup
}

func NewPlanJourneyHandler(repo CapabilityJourneyRepository, lookup ActiveJourneyLookup, refs JourneyReferenceChecks, maturity CapabilityMaturityLookup) *PlanJourneyHandler {
	return &PlanJourneyHandler{repo: repo, lookup: lookup, refs: refs, maturity: maturity}
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
	if err := h.ensureNoActiveJourneyOnTrack(ctx, facts); err != nil {
		return cqrs.EmptyResult(), err
	}
	if err := h.verifyReferences(ctx, facts); err != nil {
		return cqrs.EmptyResult(), err
	}
	if facts.CurrentMaturity, err = h.currentMaturity(ctx, facts); err != nil {
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

func (h *PlanJourneyHandler) ensureNoActiveJourneyOnTrack(ctx context.Context, facts aggregates.CapabilityJourneyFacts) error {
	existingID, exists, err := h.lookup.FindActiveJourneyIDForCapability(ctx, facts.CapabilityID.Value(), facts.Kind.TrackKinds())
	if err != nil {
		return err
	}
	if exists {
		return &services.ActiveJourneyError{ExistingJourneyID: existingID}
	}
	return nil
}

func (h *PlanJourneyHandler) currentMaturity(ctx context.Context, facts aggregates.CapabilityJourneyFacts) (int, error) {
	if !facts.Kind.IsMaturity() || h.maturity == nil {
		return 0, nil
	}
	return h.maturity(ctx, facts.CapabilityID.Value())
}

func (h *PlanJourneyHandler) verifyReferences(ctx context.Context, facts aggregates.CapabilityJourneyFacts) error {
	if err := requireReferenceExists(ctx, h.refs.CapabilityExists, facts.CapabilityID); err != nil {
		return err
	}
	if facts.ToApp.Value() != "" {
		if err := requireReferenceExists(ctx, h.refs.ComponentExists, facts.ToApp); err != nil {
			return err
		}
	}
	if err := verifyComponentsExist(ctx, h.refs.ComponentExists, facts.FromApps); err != nil {
		return err
	}
	return h.verifyMoveReferences(ctx, facts)
}

func (h *PlanJourneyHandler) verifyMoveReferences(ctx context.Context, facts aggregates.CapabilityJourneyFacts) error {
	if !facts.Kind.IsMove() || facts.TargetDomain == nil {
		return nil
	}
	if err := requireReferenceExists(ctx, h.refs.DomainExists, facts.TargetDomain); err != nil {
		return err
	}
	if facts.TargetParent == nil {
		return nil
	}
	if err := requireReferenceExists(ctx, h.refs.CapabilityExists, facts.TargetParent); err != nil {
		return err
	}
	return requireCapabilityEffectivelyInDomain(ctx, h.refs.CapabilityEffectivelyInDomain, *facts.TargetParent, *facts.TargetDomain)
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
	toApp, err := parseToApplicationRef(kind, cmd.ToComponentID)
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
	period   *valueobjects.TargetPeriod
	domain   *valueobjects.BusinessDomainRef
	parent   *valueobjects.PhysicalCapabilityRef
	maturity *valueobjects.TargetMaturity
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
	targetMaturity, err := parseOptionalTargetMaturity(cmd.TargetMaturity)
	if err != nil {
		return planJourneyTargetFacts{}, err
	}
	return planJourneyTargetFacts{period: targetPeriod, domain: targetDomain, parent: targetParent, maturity: targetMaturity}, nil
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
		ID:             valueobjects.NewCapabilityJourneyID(),
		CapabilityID:   core.capability,
		Kind:           core.kind,
		FromApps:       core.fromApps,
		ToApp:          core.toApp,
		Note:           core.note,
		TargetPeriod:   target.period,
		TargetDomain:   target.domain,
		TargetParent:   target.parent,
		ResultingName:  cmd.ResultingName,
		TargetMaturity: target.maturity,
		PlannedBy:      cmd.PlannedBy,
	}, nil
}
