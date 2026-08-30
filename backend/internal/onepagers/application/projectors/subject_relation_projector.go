package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type SubjectRelationStore interface {
	Save(ctx context.Context, subject readmodels.SubjectKey, entry readmodels.RelationEntry) error
	Replace(ctx context.Context, subject readmodels.SubjectKey, entryID string, entries []readmodels.RelationEntry) error
	DeleteByEdge(ctx context.Context, edgeID string) error
	DeleteEdgeForSubjects(ctx context.Context, edgeID string, subjectIDs []string) error
	DeleteByRelated(ctx context.Context, target readmodels.RelationTarget) error
	DeleteSubject(ctx context.Context, subject readmodels.SubjectKey) error
	RenameRelated(ctx context.Context, target readmodels.RelationTarget, name string) error
}

type BusinessDomainNameStore interface {
	Save(ctx context.Context, businessDomainID, name string) error
	Delete(ctx context.Context, businessDomainID string) error
	Name(ctx context.Context, businessDomainID string) (string, error)
}

type SubjectRelationProjector struct {
	relations SubjectRelationStore
	domains   BusinessDomainNameStore
}

func NewSubjectRelationProjector(relations SubjectRelationStore, domains BusinessDomainNameStore) *SubjectRelationProjector {
	return &SubjectRelationProjector{relations: relations, domains: domains}
}

func SubjectRelationEventTypes() []string {
	types := make([]string, 0, len(relationHandlers))
	for eventType := range relationHandlers {
		types = append(types, eventType)
	}
	return types
}

func (p *SubjectRelationProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *SubjectRelationProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handler, found := relationHandlers[eventType]
	if !found {
		return nil
	}
	var payload relationEventPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal relation payload of %s: %w", eventType, err)
	}
	if err := handler(p, ctx, payload); err != nil {
		return fmt.Errorf("cache relations of %s: %w", eventType, err)
	}
	return nil
}

func (p *SubjectRelationProjector) onRealizationLinked(ctx context.Context, event relationEventPayload) error {
	return p.linkRealization(ctx, event.CapabilityID, event.ComponentID, event.ID)
}

func (p *SubjectRelationProjector) onRealizationsInherited(ctx context.Context, event relationEventPayload) error {
	for _, inherited := range event.InheritedRealizations {
		if err := p.linkRealization(ctx, inherited.CapabilityID, inherited.ComponentID, inherited.SourceRealizationID); err != nil {
			return err
		}
	}
	return nil
}

func (p *SubjectRelationProjector) linkRealization(ctx context.Context, capabilityID, componentID, edgeID string) error {
	capability := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: capabilityID}
	application := readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: componentID}
	if err := p.relations.Save(ctx, capability, readmodels.RelationEntry{
		EntryID: entryRealizingApplications, RelatedType: subjectTypeApplication, RelatedID: componentID, EdgeID: edgeID,
	}); err != nil {
		return err
	}
	return p.relations.Save(ctx, application, readmodels.RelationEntry{
		EntryID: entryRealizedCapabilities, RelatedType: subjectTypeCapability, RelatedID: capabilityID, EdgeID: edgeID,
	})
}

func (p *SubjectRelationProjector) onRealizationsUninherited(ctx context.Context, event relationEventPayload) error {
	for _, removal := range event.Removals {
		if err := p.relations.DeleteEdgeForSubjects(ctx, removal.SourceRealizationID, removal.CapabilityIDs); err != nil {
			return err
		}
	}
	return nil
}

func (p *SubjectRelationProjector) onDependencyCreated(ctx context.Context, event relationEventPayload) error {
	return p.relations.Save(ctx,
		readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.SourceCapabilityID},
		readmodels.RelationEntry{EntryID: entryDependsOn, RelatedType: subjectTypeCapability, RelatedID: event.TargetCapabilityID, EdgeID: event.ID},
	)
}

func (p *SubjectRelationProjector) onEdgeDeleted(ctx context.Context, event relationEventPayload) error {
	return p.relations.DeleteByEdge(ctx, event.ID)
}

func (p *SubjectRelationProjector) onDomainAssigned(ctx context.Context, event relationEventPayload) error {
	name, err := p.domains.Name(ctx, event.BusinessDomainID)
	if err != nil {
		return err
	}
	return p.relations.Save(ctx,
		readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.CapabilityID},
		readmodels.RelationEntry{EntryID: entryBusinessDomains, RelatedID: event.BusinessDomainID, RelatedName: name, EdgeID: event.ID},
	)
}

func (p *SubjectRelationProjector) onBusinessDomainNamed(ctx context.Context, event relationEventPayload) error {
	if err := p.domains.Save(ctx, event.ID, event.Name); err != nil {
		return err
	}
	return p.relations.RenameRelated(ctx, readmodels.RelationTarget{EntryID: entryBusinessDomains, RelatedID: event.ID}, event.Name)
}

func (p *SubjectRelationProjector) onBusinessDomainDeleted(ctx context.Context, event relationEventPayload) error {
	if err := p.domains.Delete(ctx, event.ID); err != nil {
		return err
	}
	return p.relations.DeleteByRelated(ctx, readmodels.RelationTarget{EntryID: entryBusinessDomains, RelatedID: event.ID})
}

func (p *SubjectRelationProjector) onCapabilityCreated(ctx context.Context, event relationEventPayload) error {
	if event.ParentID == "" {
		return nil
	}
	return p.attachToParent(ctx, event.ID, event.ParentID)
}

func (p *SubjectRelationProjector) onParentChanged(ctx context.Context, event relationEventPayload) error {
	capability := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.CapabilityID}
	if err := p.relations.Replace(ctx, capability, entryParentCapability, parentEntries(event.NewParentID)); err != nil {
		return err
	}
	if err := p.relations.DeleteByRelated(ctx, readmodels.RelationTarget{EntryID: entryChildCapabilities, RelatedID: event.CapabilityID}); err != nil {
		return err
	}
	if event.NewParentID == "" {
		return nil
	}
	return p.saveChild(ctx, event.NewParentID, event.CapabilityID)
}

func (p *SubjectRelationProjector) attachToParent(ctx context.Context, capabilityID, parentID string) error {
	if err := p.relations.Save(ctx,
		readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: capabilityID},
		readmodels.RelationEntry{EntryID: entryParentCapability, RelatedType: subjectTypeCapability, RelatedID: parentID},
	); err != nil {
		return err
	}
	return p.saveChild(ctx, parentID, capabilityID)
}

func (p *SubjectRelationProjector) saveChild(ctx context.Context, parentID, capabilityID string) error {
	return p.relations.Save(ctx,
		readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: parentID},
		readmodels.RelationEntry{EntryID: entryChildCapabilities, RelatedType: subjectTypeCapability, RelatedID: capabilityID},
	)
}

func parentEntries(parentID string) []readmodels.RelationEntry {
	if parentID == "" {
		return nil
	}
	return []readmodels.RelationEntry{{EntryID: entryParentCapability, RelatedType: subjectTypeCapability, RelatedID: parentID}}
}

func (p *SubjectRelationProjector) onComponentRelationCreated(ctx context.Context, event relationEventPayload) error {
	return p.relations.Save(ctx,
		readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: event.SourceComponentID},
		readmodels.RelationEntry{EntryID: entryComponentRelations, RelatedType: subjectTypeApplication, RelatedID: event.TargetComponentID, EdgeID: event.ID},
	)
}

func (p *SubjectRelationProjector) onOriginLinkSet(ctx context.Context, event relationEventPayload) error {
	return p.applyOriginLink(ctx, event.OriginType, event.ComponentID, event.EntityID)
}

func (p *SubjectRelationProjector) onOriginLinkReplaced(ctx context.Context, event relationEventPayload) error {
	return p.applyOriginLink(ctx, event.OriginType, event.ComponentID, event.NewEntityID)
}

func (p *SubjectRelationProjector) onOriginLinkGone(ctx context.Context, event relationEventPayload) error {
	return p.applyOriginLink(ctx, event.OriginType, event.ComponentID, "")
}

func (p *SubjectRelationProjector) applyOriginLink(ctx context.Context, originType, componentID, entityID string) error {
	entries, known := originEntriesByType[originType]
	if !known {
		return fmt.Errorf("unknown origin type %q", originType)
	}

	application := readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: componentID}
	if err := p.relations.Replace(ctx, application, entries.forward, entries.forwardEntries(entityID)); err != nil {
		return err
	}
	if err := p.relations.DeleteByRelated(ctx, readmodels.RelationTarget{EntryID: entries.mirror, RelatedID: componentID}); err != nil {
		return err
	}
	if entityID == "" {
		return nil
	}
	return p.relations.Save(ctx,
		readmodels.SubjectKey{SubjectType: entries.relatedType, SubjectID: entityID},
		readmodels.RelationEntry{EntryID: entries.mirror, RelatedType: subjectTypeApplication, RelatedID: componentID},
	)
}

func subjectRelationsDeleted(subjectType string) relationHandler {
	return func(p *SubjectRelationProjector, ctx context.Context, event relationEventPayload) error {
		return p.relations.DeleteSubject(ctx, readmodels.SubjectKey{SubjectType: subjectType, SubjectID: event.ID})
	}
}

type relationHandler func(*SubjectRelationProjector, context.Context, relationEventPayload) error

var relationHandlers = buildRelationHandlers()

func buildRelationHandlers() map[string]relationHandler {
	handlers := map[string]relationHandler{
		capPL.CapabilityCreated:                 (*SubjectRelationProjector).onCapabilityCreated,
		capPL.CapabilityParentChanged:           (*SubjectRelationProjector).onParentChanged,
		capPL.SystemLinkedToCapability:          (*SubjectRelationProjector).onRealizationLinked,
		capPL.SystemRealizationDeleted:          (*SubjectRelationProjector).onEdgeDeleted,
		capPL.CapabilityRealizationsInherited:   (*SubjectRelationProjector).onRealizationsInherited,
		capPL.CapabilityRealizationsUninherited: (*SubjectRelationProjector).onRealizationsUninherited,
		capPL.CapabilityDependencyCreated:       (*SubjectRelationProjector).onDependencyCreated,
		capPL.CapabilityDependencyDeleted:       (*SubjectRelationProjector).onEdgeDeleted,
		capPL.CapabilityAssignedToDomain:        (*SubjectRelationProjector).onDomainAssigned,
		capPL.CapabilityUnassignedFromDomain:    (*SubjectRelationProjector).onEdgeDeleted,
		capPL.BusinessDomainCreated:             (*SubjectRelationProjector).onBusinessDomainNamed,
		capPL.BusinessDomainUpdated:             (*SubjectRelationProjector).onBusinessDomainNamed,
		capPL.BusinessDomainDeleted:             (*SubjectRelationProjector).onBusinessDomainDeleted,
		amPL.ComponentRelationCreated:           (*SubjectRelationProjector).onComponentRelationCreated,
		amPL.ComponentRelationDeleted:           (*SubjectRelationProjector).onEdgeDeleted,
		amPL.OriginLinkSet:                      (*SubjectRelationProjector).onOriginLinkSet,
		amPL.OriginLinkReplaced:                 (*SubjectRelationProjector).onOriginLinkReplaced,
		amPL.OriginLinkCleared:                  (*SubjectRelationProjector).onOriginLinkGone,
		amPL.OriginLinkDeleted:                  (*SubjectRelationProjector).onOriginLinkGone,
	}
	for eventType, subjectType := range subjectTypeByDeletionEvent {
		handlers[eventType] = subjectRelationsDeleted(subjectType)
	}
	return handlers
}
