package orchestrator

import (
	"context"

	architectureCommands "easi/backend/internal/architecturemodeling/application/commands"
	capabilityCommands "easi/backend/internal/capabilitymapping/application/commands"
	valueStreamCommands "easi/backend/internal/valuestreams/application/commands"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"
)

type ImportSessionRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.ImportSession, error)
	Save(ctx context.Context, session *aggregates.ImportSession) error
}

type ImportOrchestrator struct {
	commandBus cqrs.CommandBus
	repository ImportSessionRepository
}

func NewImportOrchestrator(commandBus cqrs.CommandBus, repository ImportSessionRepository) *ImportOrchestrator {
	return &ImportOrchestrator{
		commandBus: commandBus,
		repository: repository,
	}
}

type importExecutionContext struct {
	ctx                      context.Context
	session                  *aggregates.ImportSession
	result                   *aggregates.ImportResult
	sourceToComponentID      map[string]string
	sourceToCapabilityID     map[string]string
	sourceToValueStreamID    map[string]string
	sourceToStageID          map[string]string
	createdCapabilityIDs     []string
}

func (o *ImportOrchestrator) Execute(ctx context.Context, session *aggregates.ImportSession) (aggregates.ImportResult, error) {
	result := aggregates.ImportResult{}
	parsedData := session.ParsedData()

	execCtx := &importExecutionContext{
		ctx:                      ctx,
		session:                  session,
		result:                   &result,
		sourceToComponentID:      make(map[string]string),
		sourceToCapabilityID:     make(map[string]string),
		sourceToValueStreamID:    make(map[string]string),
		sourceToStageID:          make(map[string]string),
		createdCapabilityIDs:     make([]string, 0),
	}

	o.executeComponentPhase(execCtx, parsedData)
	o.executeCapabilityPhase(execCtx, parsedData)
	o.executeCapabilityMetadataPhase(execCtx)
	o.executeValueStreamPhase(execCtx, parsedData)
	o.executeRealizationPhase(execCtx, parsedData)
	o.executeComponentRelationPhase(execCtx, parsedData)
	o.executeCapabilityStageMapping(execCtx, parsedData)
	o.executeDomainAssignmentPhase(execCtx, parsedData)
	o.completeSession(execCtx)

	return result, nil
}

func (o *ImportOrchestrator) executeComponentPhase(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	created, errors := o.createComponents(execCtx.ctx, parsedData.Components, execCtx.sourceToComponentID)
	execCtx.result.ComponentsCreated = created
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseCreatingComponents, len(parsedData.Components), created)
}

func (o *ImportOrchestrator) executeCapabilityPhase(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	created, createdIDs, errors := o.createCapabilities(execCtx.ctx, parsedData.Capabilities, parsedData.Relationships, execCtx.sourceToCapabilityID)
	execCtx.result.CapabilitiesCreated = created
	execCtx.createdCapabilityIDs = createdIDs
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseCreatingCapabilities, len(parsedData.Capabilities), created)
}

func (o *ImportOrchestrator) executeCapabilityMetadataPhase(execCtx *importExecutionContext) {
	eaOwner := execCtx.session.CapabilityEAOwner()
	if eaOwner == "" || len(execCtx.createdCapabilityIDs) == 0 {
		return
	}

	assigned, errors := o.assignCapabilityMetadata(execCtx.ctx, execCtx.createdCapabilityIDs, eaOwner)
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseAssigningCapabilityMetadata, len(execCtx.createdCapabilityIDs), assigned)
}

func (o *ImportOrchestrator) executeRealizationPhase(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	created, errors := o.createRealizations(execCtx.ctx, parsedData.Relationships, execCtx.sourceToComponentID, execCtx.sourceToCapabilityID)
	execCtx.result.RealizationsCreated = created
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseCreatingRealizations, countRealizations(parsedData.Relationships), created)
}

func (o *ImportOrchestrator) executeComponentRelationPhase(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	created, errors := o.createComponentRelations(execCtx.ctx, parsedData.Relationships, execCtx.sourceToComponentID)
	execCtx.result.ComponentRelationsCreated = created
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseCreatingComponentRelations, countComponentRelations(parsedData.Relationships), created)
}

func (o *ImportOrchestrator) executeDomainAssignmentPhase(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	if execCtx.session.BusinessDomainID() == "" {
		return
	}
	assignCtx := domainAssignmentContext{
		domainID:             execCtx.session.BusinessDomainID(),
		capabilities:         parsedData.Capabilities,
		relationships:        parsedData.Relationships,
		sourceToCapabilityID: execCtx.sourceToCapabilityID,
	}
	assigned, errors := o.assignToDomain(execCtx.ctx, assignCtx)
	execCtx.result.DomainAssignments = assigned
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
}

func (o *ImportOrchestrator) executeValueStreamPhase(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	created, errors := o.createValueStreams(execCtx.ctx, parsedData.ValueStreams, execCtx.sourceToValueStreamID, execCtx.sourceToStageID)
	execCtx.result.ValueStreamsCreated = created
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseCreatingValueStreams, len(parsedData.ValueStreams), created)
}

func (o *ImportOrchestrator) executeCapabilityStageMapping(execCtx *importExecutionContext, parsedData aggregates.ParsedData) {
	if len(execCtx.sourceToValueStreamID) == 0 {
		return
	}

	mappingCtx := stageMappingContext{
		capabilityID:     execCtx.sourceToCapabilityID,
		valueStreamID:    execCtx.sourceToValueStreamID,
		stageID:          execCtx.sourceToStageID,
		relationships:    parsedData.Relationships,
	}

	mappings, errors := o.mapCapabilitiesToStages(execCtx.ctx, mappingCtx)
	execCtx.result.CapabilityMappings = mappings
	execCtx.result.Errors = append(execCtx.result.Errors, errors...)
	o.saveProgress(execCtx, valueobjects.PhaseMappingCapabilitiesToStages, countCapabilityToStageRelationships(parsedData.Relationships), mappings)
}

type stageMappingContext struct {
	capabilityID  map[string]string
	valueStreamID map[string]string
	stageID       map[string]string
	relationships []aggregates.ParsedRelationship
}

func (o *ImportOrchestrator) completeSession(execCtx *importExecutionContext) {
	if err := execCtx.session.Complete(*execCtx.result); err != nil {
		execCtx.result.Errors = append(execCtx.result.Errors, valueobjects.NewImportError("", "", "failed to mark session complete: "+err.Error(), "warning"))
	}
}

func (o *ImportOrchestrator) saveProgress(execCtx *importExecutionContext, phase string, total, completed int) {
	progress, _ := valueobjects.NewImportProgress(phase, total, completed)
	if err := execCtx.session.UpdateProgress(progress); err != nil {
		execCtx.result.Errors = append(execCtx.result.Errors, valueobjects.NewImportError("", "", "failed to update progress: "+err.Error(), "warning"))
	}
	if err := o.repository.Save(execCtx.ctx, execCtx.session); err != nil {
		execCtx.result.Errors = append(execCtx.result.Errors, valueobjects.NewImportError("", "", "failed to save progress: "+err.Error(), "warning"))
	}
}

func (o *ImportOrchestrator) createComponents(ctx context.Context, components []aggregates.ParsedElement, sourceToID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	created := 0

	for _, comp := range components {
		cmd := &architectureCommands.CreateApplicationComponent{
			Name:        comp.Name,
			Description: comp.Description,
		}

		result, err := o.commandBus.Dispatch(ctx, cmd)
		if err != nil {
			errors = append(errors, valueobjects.NewImportError(
				comp.SourceID,
				comp.Name,
				err.Error(),
				"skipped",
			))
			continue
		}

		sourceToID[comp.SourceID] = result.CreatedID
		created++
	}

	return created, errors
}

func (o *ImportOrchestrator) createCapabilities(ctx context.Context, capabilities []aggregates.ParsedElement, relationships []aggregates.ParsedRelationship, sourceToID map[string]string) (int, []string, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	var createdIDs []string
	created := 0

	parentMap := buildParentMap(relationships)

	capabilityBySourceID := make(map[string]aggregates.ParsedElement)
	for _, cap := range capabilities {
		capabilityBySourceID[cap.SourceID] = cap
	}

	levels := buildHierarchyLevels(capabilities, parentMap)

	for level, sourceIDs := range levels {
		for _, sourceID := range sourceIDs {
			cap := capabilityBySourceID[sourceID]

			var parentID string
			if parentSourceID, hasParent := parentMap[sourceID]; hasParent {
				parentID = sourceToID[parentSourceID]
			}

			levelStr := getLevelString(level)

			cmd := &capabilityCommands.CreateCapability{
				Name:        cap.Name,
				Description: cap.Description,
				ParentID:    parentID,
				Level:       levelStr,
			}

			result, err := o.commandBus.Dispatch(ctx, cmd)
			if err != nil {
				errors = append(errors, valueobjects.NewImportError(
					cap.SourceID,
					cap.Name,
					err.Error(),
					"skipped",
				))
				continue
			}

			sourceToID[sourceID] = result.CreatedID
			createdIDs = append(createdIDs, result.CreatedID)
			created++
		}
	}

	return created, createdIDs, errors
}

func (o *ImportOrchestrator) assignCapabilityMetadata(ctx context.Context, capabilityIDs []string, eaOwner string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	assigned := 0

	for _, capID := range capabilityIDs {
		cmd := &capabilityCommands.UpdateCapabilityMetadata{
			ID:      capID,
			EAOwner: eaOwner,
			Status:  "Active",
		}

		if _, err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(capID, "", "failed to assign EA Owner: "+err.Error(), "warning"))
			continue
		}
		assigned++
	}

	return assigned, errors
}

type relationshipContext struct {
	rel      aggregates.ParsedRelationship
	sourceID string
	targetID string
	notes    string
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

func (o *ImportOrchestrator) processRelationships(
	ctx context.Context,
	relationships []aggregates.ParsedRelationship,
	typeFilter func(string) bool,
	lookupRefs func(rel aggregates.ParsedRelationship) (sourceID, targetID string, sourceErr, targetErr string),
	createCommand func(relCtx relationshipContext) cqrs.Command,
) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	created := 0

	for _, rel := range relationships {
		if !typeFilter(rel.Type) {
			continue
		}

		sourceID, targetID, sourceErr, targetErr := lookupRefs(rel)

		if sourceErr != "" {
			errors = append(errors, valueobjects.NewImportError(rel.SourceID, rel.Name, sourceErr, "skipped"))
			continue
		}
		if targetErr != "" {
			errors = append(errors, valueobjects.NewImportError(rel.SourceID, rel.Name, targetErr, "skipped"))
			continue
		}

		relCtx := relationshipContext{
			rel:      rel,
			sourceID: sourceID,
			targetID: targetID,
			notes:    buildNotes(rel.Name, rel.Documentation),
		}

		if _, err := o.commandBus.Dispatch(ctx, createCommand(relCtx)); err != nil {
			errors = append(errors, valueobjects.NewImportError(rel.SourceID, rel.Name, err.Error(), "skipped"))
			continue
		}
		created++
	}

	return created, errors
}

func (o *ImportOrchestrator) createRealizations(ctx context.Context, relationships []aggregates.ParsedRelationship, sourceToComponentID, sourceToCapabilityID map[string]string) (int, []valueobjects.ImportError) {
	realizationRelationships := filterRealizationRelationships(relationships, sourceToComponentID, sourceToCapabilityID)

	return o.processRelationships(
		ctx,
		realizationRelationships,
		func(t string) bool { return t == "Realization" },
		func(rel aggregates.ParsedRelationship) (string, string, string, string) {
			componentID, hasComponent := sourceToComponentID[rel.SourceRef]
			capabilityID, hasCapability := sourceToCapabilityID[rel.TargetRef]
			var sourceErr, targetErr string
			if !hasComponent {
				sourceErr = "Source component not found"
			}
			if !hasCapability {
				targetErr = "Target capability not found"
			}
			return componentID, capabilityID, sourceErr, targetErr
		},
		func(relCtx relationshipContext) cqrs.Command {
			return &capabilityCommands.LinkSystemToCapability{
				CapabilityID:     relCtx.targetID,
				ComponentID:      relCtx.sourceID,
				RealizationLevel: "full",
				Notes:            relCtx.notes,
			}
		},
	)
}

type domainAssignmentContext struct {
	domainID             string
	capabilities         []aggregates.ParsedElement
	relationships        []aggregates.ParsedRelationship
	sourceToCapabilityID map[string]string
}

func (o *ImportOrchestrator) assignToDomain(ctx context.Context, assignCtx domainAssignmentContext) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	assigned := 0

	l1CapabilityIDs := findL1Capabilities(assignCtx.capabilities, assignCtx.relationships, assignCtx.sourceToCapabilityID)

	for _, capID := range l1CapabilityIDs {
		cmd := &capabilityCommands.AssignCapabilityToDomain{
			CapabilityID:     capID,
			BusinessDomainID: assignCtx.domainID,
		}

		if _, err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(capID, "", err.Error(), "skipped"))
			continue
		}

		assigned++
	}

	return assigned, errors
}

func (o *ImportOrchestrator) createComponentRelations(ctx context.Context, relationships []aggregates.ParsedRelationship, sourceToComponentID map[string]string) (int, []valueobjects.ImportError) {
	componentRelationships := filterComponentRelationships(relationships, sourceToComponentID)

	return o.processRelationships(
		ctx,
		componentRelationships,
		func(t string) bool { return t == "Triggering" || t == "Serving" },
		func(rel aggregates.ParsedRelationship) (string, string, string, string) {
			sourceComponentID, hasSource := sourceToComponentID[rel.SourceRef]
			targetComponentID, hasTarget := sourceToComponentID[rel.TargetRef]
			var sourceErr, targetErr string
			if !hasSource {
				sourceErr = "Source component not found"
			}
			if !hasTarget {
				targetErr = "Target component not found"
			}
			return sourceComponentID, targetComponentID, sourceErr, targetErr
		},
		func(relCtx relationshipContext) cqrs.Command {
			relationType := "Triggers"
			if relCtx.rel.Type == "Serving" {
				relationType = "Serves"
			}
			return &architectureCommands.CreateComponentRelation{
				SourceComponentID: relCtx.sourceID,
				TargetComponentID: relCtx.targetID,
				RelationType:      relationType,
				Name:              relCtx.rel.Name,
				Description:       relCtx.notes,
			}
		},
	)
}

func filterComponentRelationships(relationships []aggregates.ParsedRelationship, sourceToComponentID map[string]string) []aggregates.ParsedRelationship {
	filtered := make([]aggregates.ParsedRelationship, 0)
	for _, rel := range relationships {
		if isComponentRelationType(rel.Type) && isComponentRelationship(rel, sourceToComponentID) {
			filtered = append(filtered, rel)
		}
	}
	return filtered
}

func isComponentRelationType(relType string) bool {
	return relType == "Triggering" || relType == "Serving"
}

func isComponentRelationship(rel aggregates.ParsedRelationship, sourceToComponentID map[string]string) bool {
	return sourceToComponentID[rel.SourceRef] != "" && sourceToComponentID[rel.TargetRef] != ""
}

func filterRealizationRelationships(relationships []aggregates.ParsedRelationship, sourceToComponentID, sourceToCapabilityID map[string]string) []aggregates.ParsedRelationship {
	filtered := make([]aggregates.ParsedRelationship, 0)
	for _, rel := range relationships {
		if rel.Type == "Realization" && isComponentToCapabilityRelationship(rel, sourceToComponentID, sourceToCapabilityID) {
			filtered = append(filtered, rel)
		}
	}
	return filtered
}

func isComponentToCapabilityRelationship(rel aggregates.ParsedRelationship, sourceToComponentID, sourceToCapabilityID map[string]string) bool {
	return sourceToComponentID[rel.SourceRef] != "" && sourceToCapabilityID[rel.TargetRef] != ""
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

func findRootCapabilities(capabilities []aggregates.ParsedElement, parentMap map[string]string) []string {
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
		parentID, hasParent := parentMap[cap.SourceID]
		if hasParent && processed[parentID] {
			children = append(children, cap.SourceID)
		}
	}
	return children
}

func buildHierarchyLevels(capabilities []aggregates.ParsedElement, parentMap map[string]string) [][]string {
	roots := findRootCapabilities(capabilities, parentMap)
	if len(roots) == 0 {
		return nil
	}

	processed := make(map[string]bool)
	for _, id := range roots {
		processed[id] = true
	}

	levels := [][]string{roots}
	for {
		children := findChildrenOfProcessed(capabilities, parentMap, processed)
		if len(children) == 0 {
			break
		}
		levels = append(levels, children)
		for _, id := range children {
			processed[id] = true
		}
	}

	return levels
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

func countRealizations(relationships []aggregates.ParsedRelationship) int {
	count := 0
	for _, rel := range relationships {
		if rel.Type == "Realization" {
			count++
		}
	}
	return count
}

func countComponentRelations(relationships []aggregates.ParsedRelationship) int {
	count := 0
	for _, rel := range relationships {
		if rel.Type == "Triggering" || rel.Type == "Serving" {
			count++
		}
	}
	return count
}

func findL1Capabilities(capabilities []aggregates.ParsedElement, relationships []aggregates.ParsedRelationship, sourceToCapabilityID map[string]string) []string {
	parentMap := buildParentMap(relationships)
	var l1IDs []string

	for _, cap := range capabilities {
		if _, hasParent := parentMap[cap.SourceID]; !hasParent {
			if capID, ok := sourceToCapabilityID[cap.SourceID]; ok {
				l1IDs = append(l1IDs, capID)
			}
		}
	}

	return l1IDs
}

func (o *ImportOrchestrator) createValueStreams(ctx context.Context, valueStreams []aggregates.ParsedElement, sourceToValueStreamID map[string]string, sourceToStageID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	created := 0

	for _, vs := range valueStreams {
		cmd := &valueStreamCommands.CreateValueStream{
			Name:        vs.Name,
			Description: vs.Description,
		}

		result, err := o.commandBus.Dispatch(ctx, cmd)
		if err != nil {
			errors = append(errors, valueobjects.NewImportError(
				vs.SourceID,
				vs.Name,
				err.Error(),
				"skipped",
			))
			continue
		}

		valueStreamID := result.CreatedID
		sourceToValueStreamID[vs.SourceID] = valueStreamID

		// Create default "Main Flow" stage
		stageCmd := &valueStreamCommands.AddStage{
			ValueStreamID: valueStreamID,
			Name:          "Main Flow",
			Description:   "",
		}

		stageResult, err := o.commandBus.Dispatch(ctx, stageCmd)
		if err != nil {
			errors = append(errors, valueobjects.NewImportError(
				vs.SourceID,
				vs.Name,
				"failed to create default stage: "+err.Error(),
				"warning",
			))
			continue
		}

		sourceToStageID[vs.SourceID] = stageResult.CreatedID
		created++
	}

	return created, errors
}

func (o *ImportOrchestrator) mapCapabilitiesToStages(ctx context.Context, mappingCtx stageMappingContext) (int, []valueobjects.ImportError) {
	filtered := filterCapabilityStageRelationships(mappingCtx)
	var errors []valueobjects.ImportError
	mapped := 0

	for _, rel := range filtered {
		if o.addCapabilityToStage(ctx, rel, mappingCtx, &errors) {
			mapped++
		}
	}

	return mapped, errors
}

func filterCapabilityStageRelationships(mappingCtx stageMappingContext) []aggregates.ParsedRelationship {
	filtered := make([]aggregates.ParsedRelationship, 0)
	for _, rel := range mappingCtx.relationships {
		if isValidCapabilityStageRelationship(rel, mappingCtx) {
			filtered = append(filtered, rel)
		}
	}
	return filtered
}

func isValidCapabilityStageRelationship(rel aggregates.ParsedRelationship, mappingCtx stageMappingContext) bool {
	return isCapabilityStageRelationType(rel.Type) &&
		mappingCtx.capabilityID[rel.SourceRef] != "" && mappingCtx.valueStreamID[rel.TargetRef] != ""
}

func isCapabilityStageRelationType(relType string) bool {
	return relType == "Association" || relType == "Serving" || relType == "Triggering" || relType == "Realization"
}

func (o *ImportOrchestrator) addCapabilityToStage(ctx context.Context, rel aggregates.ParsedRelationship, mappingCtx stageMappingContext, errors *[]valueobjects.ImportError) bool {
	stageID := mappingCtx.stageID[rel.TargetRef]
	if stageID == "" {
		*errors = append(*errors, valueobjects.NewImportError(rel.SourceID, rel.Name, "ValueStream stage not found", "skipped"))
		return false
	}

	cmd := &valueStreamCommands.AddStageCapability{
		ValueStreamID: mappingCtx.valueStreamID[rel.TargetRef],
		StageID:       stageID,
		CapabilityID:  mappingCtx.capabilityID[rel.SourceRef],
	}

	if _, err := o.commandBus.Dispatch(ctx, cmd); err != nil {
		*errors = append(*errors, valueobjects.NewImportError(rel.SourceID, rel.Name, err.Error(), "skipped"))
		return false
	}
	return true
}

func countCapabilityToStageRelationships(relationships []aggregates.ParsedRelationship) int {
	count := 0
	for _, rel := range relationships {
		switch rel.Type {
		case "Association", "Serving", "Triggering", "Realization":
			count++
		}
	}
	return count
}
