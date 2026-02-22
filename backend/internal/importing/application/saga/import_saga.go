package saga

import (
	"context"

	"easi/backend/internal/importing/application/ports"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/valueobjects"
)

type ImportSaga struct {
	components   ports.ComponentGateway
	capabilities ports.CapabilityGateway
	valueStreams ports.ValueStreamGateway
}

func New(
	comp ports.ComponentGateway,
	cap ports.CapabilityGateway,
	vs ports.ValueStreamGateway,
) *ImportSaga {
	return &ImportSaga{
		components:   comp,
		capabilities: cap,
		valueStreams: vs,
	}
}

type sagaState struct {
	sourceToComponentID   map[string]mappedComponentID
	sourceToCapabilityID  map[string]mappedCapabilityID
	sourceToValueStreamID map[string]mappedValueStreamID
	sourceToStageID       map[string]mappedStageID
	createdCapabilityIDs  []mappedCapabilityID
}

func newSagaState() sagaState {
	return sagaState{
		sourceToComponentID:   make(map[string]mappedComponentID),
		sourceToCapabilityID:  make(map[string]mappedCapabilityID),
		sourceToValueStreamID: make(map[string]mappedValueStreamID),
		sourceToStageID:       make(map[string]mappedStageID),
	}
}

func (s *ImportSaga) Execute(ctx context.Context, data aggregates.ParsedData, businessDomainID, capabilityEAOwner string) aggregates.ImportResult {
	state := newSagaState()
	result := aggregates.ImportResult{}

	s.createComponents(ctx, data, &state, &result)
	s.createCapabilities(ctx, data, &state, &result)
	s.assignCapabilityMetadata(ctx, capabilityEAOwner, &state, &result)
	s.createValueStreams(ctx, data, &state, &result)
	s.createRealizations(ctx, data, &state, &result)
	s.createComponentRelations(ctx, data, &state, &result)
	s.assignDomains(domainAssignmentParams{
		ctx: ctx, data: data, businessDomainID: businessDomainID, state: &state, result: &result,
	})
	s.mapCapabilitiesToStages(ctx, data, &state, &result)

	return result
}

func (s *ImportSaga) createComponents(ctx context.Context, data aggregates.ParsedData, state *sagaState, result *aggregates.ImportResult) {
	for _, comp := range data.Components {
		createdID, err := s.components.CreateComponent(ctx, comp.Name, comp.Description)
		if err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(comp.SourceID, comp.Name, err.Error(), "skipped"))
			continue
		}
		state.sourceToComponentID[comp.SourceID] = mappedComponentID(createdID)
		result.ComponentsCreated++
	}
}

func (s *ImportSaga) createCapabilities(ctx context.Context, data aggregates.ParsedData, state *sagaState, result *aggregates.ImportResult) {
	parentMap := buildParentMap(data.Relationships)
	capabilityBySourceID := indexBySourceID(data.Capabilities)
	levels := buildHierarchyLevels(data.Capabilities, parentMap)

	for level, sourceIDs := range levels {
		for _, sourceID := range sourceIDs {
			cap := capabilityBySourceID[sourceID]
			var parentID string
			if parentSourceID, hasParent := parentMap[sourceID]; hasParent {
				parentID = string(state.sourceToCapabilityID[parentSourceID])
			}
			createdID, err := s.capabilities.CreateCapability(ctx, cap.Name, cap.Description, parentID, getLevelString(level))
			if err != nil {
				result.Errors = append(result.Errors, valueobjects.NewImportError(cap.SourceID, cap.Name, err.Error(), "skipped"))
				continue
			}
			state.sourceToCapabilityID[sourceID] = mappedCapabilityID(createdID)
			state.createdCapabilityIDs = append(state.createdCapabilityIDs, mappedCapabilityID(createdID))
			result.CapabilitiesCreated++
		}
	}
}

func (s *ImportSaga) assignCapabilityMetadata(ctx context.Context, eaOwner string, state *sagaState, result *aggregates.ImportResult) {
	if eaOwner == "" || len(state.createdCapabilityIDs) == 0 {
		return
	}
	for _, capID := range state.createdCapabilityIDs {
		if err := s.capabilities.UpdateMetadata(ctx, string(capID), eaOwner, "Active"); err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(string(capID), "", "failed to assign EA Owner: "+err.Error(), "warning"))
		}
	}
}

func (s *ImportSaga) createValueStreams(ctx context.Context, data aggregates.ParsedData, state *sagaState, result *aggregates.ImportResult) {
	for _, vs := range data.ValueStreams {
		vsID, err := s.valueStreams.CreateValueStream(ctx, vs.Name, vs.Description)
		if err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(vs.SourceID, vs.Name, err.Error(), "skipped"))
			continue
		}
		state.sourceToValueStreamID[vs.SourceID] = mappedValueStreamID(vsID)

		stageID, err := s.valueStreams.AddStage(ctx, vsID, "Main Flow", "")
		if err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(vs.SourceID, vs.Name, "failed to create default stage: "+err.Error(), "warning"))
			continue
		}
		state.sourceToStageID[vs.SourceID] = mappedStageID(stageID)
		result.ValueStreamsCreated++
	}
}

func (s *ImportSaga) createRealizations(ctx context.Context, data aggregates.ParsedData, state *sagaState, result *aggregates.ImportResult) {
	for _, rel := range data.Relationships {
		if rel.Type != "Realization" {
			continue
		}
		componentID := state.sourceToComponentID[rel.SourceRef]
		capabilityID := state.sourceToCapabilityID[rel.TargetRef]
		if componentID == "" || capabilityID == "" {
			continue
		}
		notes := buildNotes(rel.Name, rel.Documentation)
		_, err := s.capabilities.LinkSystem(ctx, string(capabilityID), string(componentID), "full", notes)
		if err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(rel.SourceID, rel.Name, err.Error(), "skipped"))
			continue
		}
		result.RealizationsCreated++
	}
}

func (s *ImportSaga) createComponentRelations(ctx context.Context, data aggregates.ParsedData, state *sagaState, result *aggregates.ImportResult) {
	for _, rel := range data.Relationships {
		if rel.Type != "Triggering" && rel.Type != "Serving" {
			continue
		}
		sourceComponentID := state.sourceToComponentID[rel.SourceRef]
		targetComponentID := state.sourceToComponentID[rel.TargetRef]
		if sourceComponentID == "" || targetComponentID == "" {
			continue
		}
		relationType := "Triggers"
		if rel.Type == "Serving" {
			relationType = "Serves"
		}
		notes := buildNotes(rel.Name, rel.Documentation)
		_, err := s.components.CreateRelation(ctx, string(sourceComponentID), string(targetComponentID), relationType, rel.Name, notes)
		if err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(rel.SourceID, rel.Name, err.Error(), "skipped"))
			continue
		}
		result.ComponentRelationsCreated++
	}
}

type domainAssignmentParams struct {
	ctx              context.Context
	data             aggregates.ParsedData
	businessDomainID string
	state            *sagaState
	result           *aggregates.ImportResult
}

func (s *ImportSaga) assignDomains(params domainAssignmentParams) {
	if params.businessDomainID == "" {
		return
	}
	parentMap := buildParentMap(params.data.Relationships)
	for _, capID := range findL1CapabilityIDs(params.data.Capabilities, parentMap, params.state.sourceToCapabilityID) {
		if err := s.capabilities.AssignToDomain(params.ctx, string(capID), params.businessDomainID); err != nil {
			params.result.Errors = append(params.result.Errors, valueobjects.NewImportError(string(capID), "", err.Error(), "skipped"))
			continue
		}
		params.result.DomainAssignments++
	}
}

func findL1CapabilityIDs(capabilities []aggregates.ParsedElement, parentMap map[string]string, sourceToCapabilityID map[string]mappedCapabilityID) []mappedCapabilityID {
	var ids []mappedCapabilityID
	for _, cap := range capabilities {
		if _, hasParent := parentMap[cap.SourceID]; hasParent {
			continue
		}
		if capID := sourceToCapabilityID[cap.SourceID]; capID != "" {
			ids = append(ids, capID)
		}
	}
	return ids
}

func (s *ImportSaga) mapCapabilitiesToStages(ctx context.Context, data aggregates.ParsedData, state *sagaState, result *aggregates.ImportResult) {
	if len(state.sourceToValueStreamID) == 0 {
		return
	}
	for _, rel := range data.Relationships {
		if !isCapabilityStageRelationType(rel.Type) {
			continue
		}
		if !state.hasStageMappingRefs(rel) {
			continue
		}
		if err := s.valueStreams.MapCapabilityToStage(ctx, string(state.sourceToValueStreamID[rel.TargetRef]), string(state.sourceToStageID[rel.TargetRef]), string(state.sourceToCapabilityID[rel.SourceRef])); err != nil {
			result.Errors = append(result.Errors, valueobjects.NewImportError(rel.SourceID, rel.Name, err.Error(), "skipped"))
			continue
		}
		result.CapabilityMappings++
	}
}

func (s *sagaState) hasStageMappingRefs(rel aggregates.ParsedRelationship) bool {
	return s.sourceToCapabilityID[rel.SourceRef] != "" &&
		s.sourceToValueStreamID[rel.TargetRef] != "" &&
		s.sourceToStageID[rel.TargetRef] != ""
}

func buildParentMap(relationships []aggregates.ParsedRelationship) map[string]string {
	parentMap := make(map[string]string)
	for _, rel := range relationships {
		if rel.Type == "Aggregation" || rel.Type == "Composition" {
			parentMap[rel.TargetRef] = rel.SourceRef
		}
	}
	return parentMap
}

func indexBySourceID(elements []aggregates.ParsedElement) map[string]aggregates.ParsedElement {
	m := make(map[string]aggregates.ParsedElement, len(elements))
	for _, e := range elements {
		m[e.SourceID] = e
	}
	return m
}

func buildHierarchyLevels(capabilities []aggregates.ParsedElement, parentMap map[string]string) [][]string {
	roots := findRoots(capabilities, parentMap)
	if len(roots) == 0 {
		return nil
	}

	processed := toSet(roots)
	levels := [][]string{roots}

	for {
		children := findChildrenOfProcessed(capabilities, parentMap, processed)
		if len(children) == 0 {
			break
		}
		levels = append(levels, children)
		addToSet(processed, children)
	}

	return levels
}

func findRoots(capabilities []aggregates.ParsedElement, parentMap map[string]string) []string {
	var roots []string
	for _, cap := range capabilities {
		if _, hasParent := parentMap[cap.SourceID]; !hasParent {
			roots = append(roots, cap.SourceID)
		}
	}
	return roots
}

func findChildrenOfProcessed(capabilities []aggregates.ParsedElement, parentMap map[string]string, processed map[string]bool) []string {
	var children []string
	for _, cap := range capabilities {
		if processed[cap.SourceID] {
			continue
		}
		if parentID, hasParent := parentMap[cap.SourceID]; hasParent && processed[parentID] {
			children = append(children, cap.SourceID)
		}
	}
	return children
}

func toSet(ids []string) map[string]bool {
	s := make(map[string]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

func addToSet(s map[string]bool, ids []string) {
	for _, id := range ids {
		s[id] = true
	}
}

func getLevelString(level int) string {
	switch level {
	case 0:
		return "L1"
	case 1:
		return "L2"
	case 2:
		return "L3"
	default:
		return "L4"
	}
}

func buildNotes(name, documentation string) string {
	if documentation == "" {
		return name
	}
	if name == "" {
		return documentation
	}
	return name + " - " + documentation
}

func isCapabilityStageRelationType(relType string) bool {
	return relType == "Association" || relType == "Serving" || relType == "Triggering" || relType == "Realization"
}
