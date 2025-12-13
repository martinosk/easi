package orchestrator

import (
	"context"

	architectureCommands "easi/backend/internal/architecturemodeling/application/commands"
	capabilityCommands "easi/backend/internal/capabilitymapping/application/commands"
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

func (o *ImportOrchestrator) Execute(ctx context.Context, session *aggregates.ImportSession) (aggregates.ImportResult, error) {
	result := aggregates.ImportResult{}
	parsedData := session.ParsedData()

	sourceToComponentID := make(map[string]string)
	sourceToCapabilityID := make(map[string]string)

	componentsCreated, componentErrors := o.createComponents(ctx, parsedData.Components, sourceToComponentID)
	result.ComponentsCreated = componentsCreated
	result.Errors = append(result.Errors, componentErrors...)

	progress, _ := valueobjects.NewImportProgress(valueobjects.PhaseCreatingComponents, len(parsedData.Components), componentsCreated)
	session.UpdateProgress(progress)
	o.repository.Save(ctx, session)

	capabilitiesCreated, capabilityErrors := o.createCapabilities(ctx, parsedData.Capabilities, parsedData.Relationships, sourceToCapabilityID)
	result.CapabilitiesCreated = capabilitiesCreated
	result.Errors = append(result.Errors, capabilityErrors...)

	progress, _ = valueobjects.NewImportProgress(valueobjects.PhaseCreatingCapabilities, len(parsedData.Capabilities), capabilitiesCreated)
	session.UpdateProgress(progress)
	o.repository.Save(ctx, session)

	realizationsCreated, realizationErrors := o.createRealizations(ctx, parsedData.Relationships, sourceToComponentID, sourceToCapabilityID)
	result.RealizationsCreated = realizationsCreated
	result.Errors = append(result.Errors, realizationErrors...)

	progress, _ = valueobjects.NewImportProgress(valueobjects.PhaseCreatingRealizations, countRealizations(parsedData.Relationships), realizationsCreated)
	session.UpdateProgress(progress)
	o.repository.Save(ctx, session)

	componentRelationsCreated, componentRelationErrors := o.createComponentRelations(ctx, parsedData.Relationships, sourceToComponentID)
	result.ComponentRelationsCreated = componentRelationsCreated
	result.Errors = append(result.Errors, componentRelationErrors...)

	progress, _ = valueobjects.NewImportProgress(valueobjects.PhaseCreatingComponentRelations, countComponentRelations(parsedData.Relationships), componentRelationsCreated)
	session.UpdateProgress(progress)
	o.repository.Save(ctx, session)

	if session.BusinessDomainID() != "" {
		assignCtx := domainAssignmentContext{
			domainID:             session.BusinessDomainID(),
			capabilities:         parsedData.Capabilities,
			relationships:        parsedData.Relationships,
			sourceToCapabilityID: sourceToCapabilityID,
		}
		domainAssignments, domainErrors := o.assignToDomain(ctx, assignCtx)
		result.DomainAssignments = domainAssignments
		result.Errors = append(result.Errors, domainErrors...)
	}

	session.Complete(result)

	return result, nil
}

func (o *ImportOrchestrator) createComponents(ctx context.Context, components []aggregates.ParsedElement, sourceToID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	created := 0

	for _, comp := range components {
		cmd := &architectureCommands.CreateApplicationComponent{
			Name:        comp.Name,
			Description: comp.Description,
		}

		if err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(
				comp.SourceID,
				comp.Name,
				err.Error(),
				"skipped",
			))
			continue
		}

		sourceToID[comp.SourceID] = cmd.ID
		created++
	}

	return created, errors
}

func (o *ImportOrchestrator) createCapabilities(ctx context.Context, capabilities []aggregates.ParsedElement, relationships []aggregates.ParsedRelationship, sourceToID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
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

			if err := o.commandBus.Dispatch(ctx, cmd); err != nil {
				errors = append(errors, valueobjects.NewImportError(
					cap.SourceID,
					cap.Name,
					err.Error(),
					"skipped",
				))
				continue
			}

			sourceToID[sourceID] = cmd.ID
			created++
		}
	}

	return created, errors
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

		if err := o.commandBus.Dispatch(ctx, createCommand(relCtx)); err != nil {
			errors = append(errors, valueobjects.NewImportError(rel.SourceID, rel.Name, err.Error(), "skipped"))
			continue
		}
		created++
	}

	return created, errors
}

func (o *ImportOrchestrator) createRealizations(ctx context.Context, relationships []aggregates.ParsedRelationship, sourceToComponentID, sourceToCapabilityID map[string]string) (int, []valueobjects.ImportError) {
	return o.processRelationships(
		ctx,
		relationships,
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

		if err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(capID, "", err.Error(), "skipped"))
			continue
		}

		assigned++
	}

	return assigned, errors
}

func (o *ImportOrchestrator) createComponentRelations(ctx context.Context, relationships []aggregates.ParsedRelationship, sourceToComponentID map[string]string) (int, []valueobjects.ImportError) {
	return o.processRelationships(
		ctx,
		relationships,
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
