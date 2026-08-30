//go:build integration

package fixtures

import (
	capCommands "easi/backend/internal/capabilitymapping/application/commands"
	capHandlers "easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/shared/events"
)

func subscribePillarCacheEvents(eventBus *events.InMemoryEventBus, projector *projectors.StrategyPillarCacheProjector) {
	for _, eventType := range []string{
		mmPL.MetaModelConfigurationCreated, mmPL.StrategyPillarAdded, mmPL.StrategyPillarUpdated,
		mmPL.StrategyPillarRemoved, mmPL.PillarFitConfigurationUpdated,
	} {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeStrategicRealizationEvents(eventBus *events.InMemoryEventBus, projector *projectors.RealizationProjector) {
	for _, eventType := range []string{
		cmPL.SystemLinkedToCapability, cmPL.SystemRealizationUpdated, cmPL.SystemRealizationDeleted,
		cmPL.CapabilityRealizationsInherited, cmPL.CapabilityRealizationsUninherited,
	} {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeStrategyImportanceEvents(eventBus *events.InMemoryEventBus, projector *projectors.StrategyImportanceProjector) {
	for _, eventType := range []string{cmPL.StrategyImportanceSet, cmPL.StrategyImportanceUpdated, cmPL.StrategyImportanceRemoved} {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeApplicationFitScoreEvents(eventBus *events.InMemoryEventBus, projector *projectors.ApplicationFitScoreProjector) {
	for _, eventType := range []string{cmPL.ApplicationFitScoreSet, cmPL.ApplicationFitScoreUpdated, cmPL.ApplicationFitScoreRemoved} {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeImportanceChangeEffectiveEvents(eventBus *events.InMemoryEventBus, projector *projectors.ImportanceChangeEffectiveProjector) {
	for _, eventType := range []string{cmPL.StrategyImportanceSet, cmPL.StrategyImportanceUpdated, cmPL.StrategyImportanceRemoved} {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeHierarchyChangeEffectiveEvents(eventBus *events.InMemoryEventBus, projector *projectors.HierarchyChangeEffectiveProjector) {
	for _, eventType := range []string{cmPL.CapabilityParentChanged, cmPL.CapabilityDeleted} {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeDomainAssignmentEffectiveEvents(eventBus *events.InMemoryEventBus, projector *projectors.DomainAssignmentEffectiveProjector) {
	for _, eventType := range []string{cmPL.CapabilityAssignedToDomain, cmPL.CapabilityUnassignedFromDomain} {
		eventBus.Subscribe(eventType, projector)
	}
}

type StrategicAnalysisFixtures struct {
	tc                    *TestContext
	realizationReadModel  *readmodels.RealizationReadModel
	importanceReadModel   *readmodels.StrategyImportanceReadModel
	fitScoreReadModel     *readmodels.ApplicationFitScoreReadModel
	effectiveImportanceRM *readmodels.EffectiveCapabilityImportanceReadModel
	capabilityReadModel   *readmodels.CapabilityReadModel
	domainAssignmentRM    *readmodels.DomainCapabilityAssignmentReadModel
	componentCacheRM      *readmodels.ComponentCacheReadModel
	strategyPillarCacheRM *readmodels.StrategyPillarCacheReadModel
}

func NewStrategicAnalysisFixtures(tc *TestContext) *StrategicAnalysisFixtures {
	realizationRM := readmodels.NewRealizationReadModel(tc.TenantDB)
	importanceRM := readmodels.NewStrategyImportanceReadModel(tc.TenantDB)
	fitScoreRM := readmodels.NewApplicationFitScoreReadModel(tc.TenantDB)
	effectiveImportanceRM := readmodels.NewEffectiveCapabilityImportanceReadModel(tc.TenantDB)
	capabilityRM := readmodels.NewCapabilityReadModel(tc.TenantDB)
	domainAssignmentRM := readmodels.NewDomainCapabilityAssignmentReadModel(tc.TenantDB)
	componentCacheRM := readmodels.NewComponentCacheReadModel(tc.TenantDB)
	domainRM := readmodels.NewBusinessDomainReadModel(tc.TenantDB)
	strategyPillarCacheRM := readmodels.NewStrategyPillarCacheReadModel(tc.TenantDB)

	pillarsGateway := metamodel.NewLocalStrategyPillarsGateway(strategyPillarCacheRM)

	realizationRepo := repositories.NewRealizationRepository(tc.EventStore)
	capabilityRepo := repositories.NewCapabilityRepository(tc.EventStore)
	importanceRepo := repositories.NewStrategyImportanceRepository(tc.EventStore)
	fitScoreRepo := repositories.NewApplicationFitScoreRepository(tc.EventStore)

	pillarCacheProjector := projectors.NewStrategyPillarCacheProjector(strategyPillarCacheRM)
	subscribePillarCacheEvents(tc.EventBus, pillarCacheProjector)

	realizationProjector := projectors.NewRealizationProjector(realizationRM, componentCacheRM)
	subscribeStrategicRealizationEvents(tc.EventBus, realizationProjector)

	importanceProjector := projectors.NewStrategyImportanceProjector(importanceRM, domainRM, capabilityRM, pillarsGateway)
	subscribeStrategyImportanceEvents(tc.EventBus, importanceProjector)

	fitScoreProjector := projectors.NewApplicationFitScoreProjector(fitScoreRM, componentCacheRM, pillarsGateway)
	subscribeApplicationFitScoreEvents(tc.EventBus, fitScoreProjector)

	capabilityLookupAdapter := adapters.NewCapabilityLookupAdapter(capabilityRM)
	ratingLookupAdapter := adapters.NewRatingLookupAdapter(importanceRM)
	hierarchyService := services.NewCapabilityHierarchyService(capabilityLookupAdapter)
	ratingResolver := services.NewHierarchicalRatingResolver(hierarchyService, ratingLookupAdapter, capabilityLookupAdapter)

	recomputer := projectors.NewEffectiveImportanceRecomputer(effectiveImportanceRM, ratingResolver, hierarchyService, nil)

	importanceChangeProjector := projectors.NewImportanceChangeEffectiveProjector(recomputer, importanceRM)
	subscribeImportanceChangeEffectiveEvents(tc.EventBus, importanceChangeProjector)

	hierarchyChangeProjector := projectors.NewHierarchyChangeEffectiveProjector(recomputer, effectiveImportanceRM)
	subscribeHierarchyChangeEffectiveEvents(tc.EventBus, hierarchyChangeProjector)

	ancestryChecker := projectors.NewDomainAncestryChecker(hierarchyService, domainAssignmentRM)
	domainAssignmentEffectiveProjector := projectors.NewDomainAssignmentEffectiveProjector(recomputer, ancestryChecker, pillarsGateway)
	subscribeDomainAssignmentEffectiveEvents(tc.EventBus, domainAssignmentEffectiveProjector)

	tc.CommandBus.Register("LinkSystemToCapability", capHandlers.NewLinkSystemToCapabilityHandler(realizationRepo, capabilityRepo, capabilityRM, componentCacheRM))
	tc.CommandBus.Register("UpdateSystemRealization", capHandlers.NewUpdateSystemRealizationHandler(realizationRepo))
	tc.CommandBus.Register("DeleteSystemRealization", capHandlers.NewDeleteSystemRealizationHandler(realizationRepo))

	importanceDeps := capHandlers.StrategyImportanceDeps{
		ImportanceRepo:   importanceRepo,
		DomainReader:     domainRM,
		CapabilityReader: capabilityRM,
		ImportanceReader: importanceRM,
		PillarsGateway:   pillarsGateway,
	}
	tc.CommandBus.Register("SetStrategyImportance", capHandlers.NewSetStrategyImportanceHandler(importanceDeps))
	tc.CommandBus.Register("UpdateStrategyImportance", capHandlers.NewUpdateStrategyImportanceHandler(importanceRepo))
	tc.CommandBus.Register("RemoveStrategyImportance", capHandlers.NewRemoveStrategyImportanceHandler(importanceRepo))

	fitScoreDeps := capHandlers.ApplicationFitScoreDeps{
		FitScoreRepo:   fitScoreRepo,
		FitScoreReader: fitScoreRM,
		PillarsGateway: pillarsGateway,
	}
	tc.CommandBus.Register("SetApplicationFitScore", capHandlers.NewSetApplicationFitScoreHandler(fitScoreDeps))
	tc.CommandBus.Register("UpdateApplicationFitScore", capHandlers.NewUpdateApplicationFitScoreHandler(fitScoreRepo))
	tc.CommandBus.Register("RemoveApplicationFitScore", capHandlers.NewRemoveApplicationFitScoreHandler(fitScoreRepo))

	return &StrategicAnalysisFixtures{
		tc:                    tc,
		realizationReadModel:  realizationRM,
		importanceReadModel:   importanceRM,
		fitScoreReadModel:     fitScoreRM,
		effectiveImportanceRM: effectiveImportanceRM,
		capabilityReadModel:   capabilityRM,
		domainAssignmentRM:    domainAssignmentRM,
		componentCacheRM:      componentCacheRM,
		strategyPillarCacheRM: strategyPillarCacheRM,
	}
}

func (f *StrategicAnalysisFixtures) LinkSystemToCapability(capabilityID, componentID string) string {
	cmd := &capCommands.LinkSystemToCapability{
		CapabilityID:     capabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
	}

	result := f.tc.MustDispatch(cmd)
	f.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (f *StrategicAnalysisFixtures) SetStrategyImportance(domainID, capabilityID, pillarID string, importance int) string {
	cmd := &capCommands.SetStrategyImportance{
		BusinessDomainID: domainID,
		CapabilityID:     capabilityID,
		PillarID:         pillarID,
		Importance:       importance,
	}

	result := f.tc.MustDispatch(cmd)
	f.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (f *StrategicAnalysisFixtures) SetApplicationFitScore(componentID, pillarID string, score int) string {
	cmd := &capCommands.SetApplicationFitScore{
		ComponentID: componentID,
		PillarID:    pillarID,
		Score:       score,
		ScoredBy:    "test-user",
	}

	result := f.tc.MustDispatch(cmd)
	f.tc.TrackID(result.CreatedID)
	return result.CreatedID
}

func (f *StrategicAnalysisFixtures) RealizationReadModel() *readmodels.RealizationReadModel {
	return f.realizationReadModel
}

func (f *StrategicAnalysisFixtures) ImportanceReadModel() *readmodels.StrategyImportanceReadModel {
	return f.importanceReadModel
}

func (f *StrategicAnalysisFixtures) FitScoreReadModel() *readmodels.ApplicationFitScoreReadModel {
	return f.fitScoreReadModel
}

func (f *StrategicAnalysisFixtures) EffectiveImportanceReadModel() *readmodels.EffectiveCapabilityImportanceReadModel {
	return f.effectiveImportanceRM
}

func (f *StrategicAnalysisFixtures) StrategyPillarCacheReadModel() *readmodels.StrategyPillarCacheReadModel {
	return f.strategyPillarCacheRM
}
