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
		domainAssignments, domainErrors := o.assignToDomain(ctx, session.BusinessDomainID(), parsedData.Capabilities, parsedData.Relationships, sourceToCapabilityID)
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

func (o *ImportOrchestrator) createRealizations(ctx context.Context, relationships []aggregates.ParsedRelationship, sourceToComponentID, sourceToCapabilityID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	created := 0

	for _, rel := range relationships {
		if rel.Type != "Realization" {
			continue
		}

		componentID, hasComponent := sourceToComponentID[rel.SourceRef]
		capabilityID, hasCapability := sourceToCapabilityID[rel.TargetRef]

		if !hasComponent {
			errors = append(errors, valueobjects.NewImportError(
				rel.SourceID,
				rel.Name,
				"Source component not found",
				"skipped",
			))
			continue
		}

		if !hasCapability {
			errors = append(errors, valueobjects.NewImportError(
				rel.SourceID,
				rel.Name,
				"Target capability not found",
				"skipped",
			))
			continue
		}

		notes := rel.Name
		if rel.Documentation != "" {
			if notes != "" {
				notes += " - "
			}
			notes += rel.Documentation
		}

		cmd := &capabilityCommands.LinkSystemToCapability{
			CapabilityID:     capabilityID,
			ComponentID:      componentID,
			RealizationLevel: "full",
			Notes:            notes,
		}

		if err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(
				rel.SourceID,
				rel.Name,
				err.Error(),
				"skipped",
			))
			continue
		}

		created++
	}

	return created, errors
}

func (o *ImportOrchestrator) assignToDomain(ctx context.Context, domainID string, capabilities []aggregates.ParsedElement, relationships []aggregates.ParsedRelationship, sourceToCapabilityID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	assigned := 0

	l1CapabilityIDs := findL1Capabilities(capabilities, relationships, sourceToCapabilityID)

	for _, capID := range l1CapabilityIDs {
		cmd := &capabilityCommands.AssignCapabilityToDomain{
			CapabilityID:     capID,
			BusinessDomainID: domainID,
		}

		if err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(
				capID,
				"",
				err.Error(),
				"skipped",
			))
			continue
		}

		assigned++
	}

	return assigned, errors
}

func (o *ImportOrchestrator) createComponentRelations(ctx context.Context, relationships []aggregates.ParsedRelationship, sourceToComponentID map[string]string) (int, []valueobjects.ImportError) {
	var errors []valueobjects.ImportError
	created := 0

	for _, rel := range relationships {
		if rel.Type != "Triggering" && rel.Type != "Serving" {
			continue
		}

		sourceComponentID, hasSource := sourceToComponentID[rel.SourceRef]
		targetComponentID, hasTarget := sourceToComponentID[rel.TargetRef]

		if !hasSource {
			errors = append(errors, valueobjects.NewImportError(
				rel.SourceID,
				rel.Name,
				"Source component not found",
				"skipped",
			))
			continue
		}

		if !hasTarget {
			errors = append(errors, valueobjects.NewImportError(
				rel.SourceID,
				rel.Name,
				"Target component not found",
				"skipped",
			))
			continue
		}

		relationType := "Triggers"
		if rel.Type == "Serving" {
			relationType = "Serves"
		}

		description := rel.Name
		if rel.Documentation != "" {
			if description != "" {
				description += " - "
			}
			description += rel.Documentation
		}

		cmd := &architectureCommands.CreateComponentRelation{
			SourceComponentID: sourceComponentID,
			TargetComponentID: targetComponentID,
			RelationType:      relationType,
			Name:              rel.Name,
			Description:       description,
		}

		if err := o.commandBus.Dispatch(ctx, cmd); err != nil {
			errors = append(errors, valueobjects.NewImportError(
				rel.SourceID,
				rel.Name,
				err.Error(),
				"skipped",
			))
			continue
		}

		created++
	}

	return created, errors
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

func buildHierarchyLevels(capabilities []aggregates.ParsedElement, parentMap map[string]string) [][]string {
	levels := make([][]string, 0)
	processed := make(map[string]bool)

	var level0 []string
	for _, cap := range capabilities {
		if _, hasParent := parentMap[cap.SourceID]; !hasParent {
			level0 = append(level0, cap.SourceID)
			processed[cap.SourceID] = true
		}
	}
	if len(level0) > 0 {
		levels = append(levels, level0)
	}

	for {
		var nextLevel []string
		for _, cap := range capabilities {
			if processed[cap.SourceID] {
				continue
			}
			parentID, hasParent := parentMap[cap.SourceID]
			if hasParent && processed[parentID] {
				nextLevel = append(nextLevel, cap.SourceID)
			}
		}
		if len(nextLevel) == 0 {
			break
		}
		for _, sourceID := range nextLevel {
			processed[sourceID] = true
		}
		levels = append(levels, nextLevel)
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
