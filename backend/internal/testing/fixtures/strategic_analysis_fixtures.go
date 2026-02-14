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
)

type StrategicAnalysisFixtures struct {
	tc                      *TestContext
	realizationReadModel    *readmodels.RealizationReadModel
	importanceReadModel     *readmodels.StrategyImportanceReadModel
	fitScoreReadModel       *readmodels.ApplicationFitScoreReadModel
	effectiveImportanceRM   *readmodels.EffectiveCapabilityImportanceReadModel
	capabilityReadModel     *readmodels.CapabilityReadModel
	domainAssignmentRM      *readmodels.DomainCapabilityAssignmentReadModel
	componentCacheRM        *readmodels.ComponentCacheReadModel
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

	pillarsGateway := metamodel.NewDirectStrategyPillarsGateway(tc.TenantDB)

	realizationRepo := repositories.NewRealizationRepository(tc.EventStore)
	capabilityRepo := repositories.NewCapabilityRepository(tc.EventStore)
	importanceRepo := repositories.NewStrategyImportanceRepository(tc.EventStore)
	fitScoreRepo := repositories.NewApplicationFitScoreRepository(tc.EventStore)

	realizationProjector := projectors.NewRealizationProjector(realizationRM, componentCacheRM)
	tc.EventBus.Subscribe("SystemLinkedToCapability", realizationProjector)
	tc.EventBus.Subscribe("SystemRealizationUpdated", realizationProjector)
	tc.EventBus.Subscribe("SystemRealizationDeleted", realizationProjector)
	tc.EventBus.Subscribe("CapabilityRealizationsInherited", realizationProjector)
	tc.EventBus.Subscribe("CapabilityRealizationsUninherited", realizationProjector)

	importanceProjector := projectors.NewStrategyImportanceProjector(importanceRM, domainRM, capabilityRM, pillarsGateway)
	tc.EventBus.Subscribe("StrategyImportanceSet", importanceProjector)
	tc.EventBus.Subscribe("StrategyImportanceUpdated", importanceProjector)
	tc.EventBus.Subscribe("StrategyImportanceRemoved", importanceProjector)

	fitScoreProjector := projectors.NewApplicationFitScoreProjector(fitScoreRM, componentCacheRM, pillarsGateway)
	tc.EventBus.Subscribe("ApplicationFitScoreSet", fitScoreProjector)
	tc.EventBus.Subscribe("ApplicationFitScoreUpdated", fitScoreProjector)
	tc.EventBus.Subscribe("ApplicationFitScoreRemoved", fitScoreProjector)

	capabilityLookupAdapter := adapters.NewCapabilityLookupAdapter(capabilityRM)
	ratingLookupAdapter := adapters.NewRatingLookupAdapter(importanceRM)
	hierarchyService := services.NewCapabilityHierarchyService(capabilityLookupAdapter)
	ratingResolver := services.NewHierarchicalRatingResolver(hierarchyService, ratingLookupAdapter, capabilityLookupAdapter)

	recomputer := projectors.NewEffectiveImportanceRecomputer(effectiveImportanceRM, ratingResolver, hierarchyService, nil)

	importanceChangeProjector := projectors.NewImportanceChangeEffectiveProjector(recomputer, importanceRM)
	tc.EventBus.Subscribe("StrategyImportanceSet", importanceChangeProjector)
	tc.EventBus.Subscribe("StrategyImportanceUpdated", importanceChangeProjector)
	tc.EventBus.Subscribe("StrategyImportanceRemoved", importanceChangeProjector)

	hierarchyChangeProjector := projectors.NewHierarchyChangeEffectiveProjector(recomputer, effectiveImportanceRM)
	tc.EventBus.Subscribe("CapabilityParentChanged", hierarchyChangeProjector)
	tc.EventBus.Subscribe("CapabilityDeleted", hierarchyChangeProjector)

	ancestryChecker := projectors.NewDomainAncestryChecker(hierarchyService, domainAssignmentRM)
	domainAssignmentEffectiveProjector := projectors.NewDomainAssignmentEffectiveProjector(recomputer, ancestryChecker, pillarsGateway)
	tc.EventBus.Subscribe("CapabilityAssignedToDomain", domainAssignmentEffectiveProjector)
	tc.EventBus.Subscribe("CapabilityUnassignedFromDomain", domainAssignmentEffectiveProjector)

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
		tc:                      tc,
		realizationReadModel:    realizationRM,
		importanceReadModel:     importanceRM,
		fitScoreReadModel:       fitScoreRM,
		effectiveImportanceRM:   effectiveImportanceRM,
		capabilityReadModel:     capabilityRM,
		domainAssignmentRM:      domainAssignmentRM,
		componentCacheRM:        componentCacheRM,
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
