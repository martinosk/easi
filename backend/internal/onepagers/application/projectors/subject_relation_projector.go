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
	SubjectsByEdge(ctx context.Context, edgeID string) ([]readmodels.SubjectKey, error)
	SubjectsByRelated(ctx context.Context, target readmodels.RelationTarget) ([]readmodels.SubjectKey, error)
}

type BusinessDomainNameStore interface {
	Save(ctx context.Context, businessDomainID, name string) error
	Delete(ctx context.Context, businessDomainID string) error
	Name(ctx context.Context, businessDomainID string) (string, error)
}

type CompletenessRecomputer interface {
	Recompute(ctx context.Context, subjectType string, subjectIDs []string) error
}

type SubjectRelationProjector struct {
	relations    SubjectRelationStore
	domains      BusinessDomainNameStore
	completeness CompletenessRecomputer
}

func NewSubjectRelationProjector(relations SubjectRelationStore, domains BusinessDomainNameStore, completeness CompletenessRecomputer) *SubjectRelationProjector {
	return &SubjectRelationProjector{relations: relations, domains: domains, completeness: completeness}
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
	touched, err := handler(p, ctx, payload)
	if err != nil {
		return fmt.Errorf("cache relations of %s: %w", eventType, err)
	}
	if err := p.recomputeTouched(ctx, touched); err != nil {
		return fmt.Errorf("recompute completeness after %s: %w", eventType, err)
	}
	return nil
}

func (p *SubjectRelationProjector) recomputeTouched(ctx context.Context, touched []readmodels.SubjectKey) error {
	if p.completeness == nil || len(touched) == 0 {
		return nil
	}
	for _, group := range groupBySubjectType(touched) {
		if err := p.completeness.Recompute(ctx, group.subjectType, group.subjectIDs); err != nil {
			return err
		}
	}
	return nil
}

type subjectTypeGroup struct {
	subjectType string
	subjectIDs  []string
}

func groupBySubjectType(keys []readmodels.SubjectKey) []subjectTypeGroup {
	indexBySubjectType := map[string]int{}
	seen := map[readmodels.SubjectKey]struct{}{}
	groups := make([]subjectTypeGroup, 0, len(keys))
	for _, key := range keys {
		if !markSeen(seen, key) {
			continue
		}
		groups = appendToSubjectTypeGroup(groups, indexBySubjectType, key)
	}
	return groups
}

func markSeen(seen map[readmodels.SubjectKey]struct{}, key readmodels.SubjectKey) bool {
	if _, duplicate := seen[key]; duplicate {
		return false
	}
	seen[key] = struct{}{}
	return true
}

func appendToSubjectTypeGroup(groups []subjectTypeGroup, indexBySubjectType map[string]int, key readmodels.SubjectKey) []subjectTypeGroup {
	i, known := indexBySubjectType[key.SubjectType]
	if !known {
		indexBySubjectType[key.SubjectType] = len(groups)
		return append(groups, subjectTypeGroup{subjectType: key.SubjectType, subjectIDs: []string{key.SubjectID}})
	}
	groups[i].subjectIDs = append(groups[i].subjectIDs, key.SubjectID)
	return groups
}

func (p *SubjectRelationProjector) onRealizationLinked(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	return p.linkRealization(ctx, event.CapabilityID, event.ComponentID, event.ID)
}

func (p *SubjectRelationProjector) onRealizationsInherited(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	touched := make([]readmodels.SubjectKey, 0, len(event.InheritedRealizations)*2)
	for _, inherited := range event.InheritedRealizations {
		keys, err := p.linkRealization(ctx, inherited.CapabilityID, inherited.ComponentID, inherited.SourceRealizationID)
		if err != nil {
			return nil, err
		}
		touched = append(touched, keys...)
	}
	return touched, nil
}

func (p *SubjectRelationProjector) linkRealization(ctx context.Context, capabilityID, componentID, edgeID string) ([]readmodels.SubjectKey, error) {
	capability := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: capabilityID}
	application := readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: componentID}
	if err := p.relations.Save(ctx, capability, readmodels.RelationEntry{
		EntryID: entryRealizingApplications, RelatedType: subjectTypeApplication, RelatedID: componentID, EdgeID: edgeID,
	}); err != nil {
		return nil, err
	}
	if err := p.relations.Save(ctx, application, readmodels.RelationEntry{
		EntryID: entryRealizedCapabilities, RelatedType: subjectTypeCapability, RelatedID: capabilityID, EdgeID: edgeID,
	}); err != nil {
		return nil, err
	}
	return []readmodels.SubjectKey{capability, application}, nil
}

func (p *SubjectRelationProjector) onRealizationsUninherited(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	var touched []readmodels.SubjectKey
	for _, removal := range event.Removals {
		affected, err := p.relations.SubjectsByEdge(ctx, removal.SourceRealizationID)
		if err != nil {
			return nil, err
		}
		if err := p.relations.DeleteEdgeForSubjects(ctx, removal.SourceRealizationID, removal.CapabilityIDs); err != nil {
			return nil, err
		}
		touched = append(touched, affected...)
	}
	return touched, nil
}

func (p *SubjectRelationProjector) onDependencyCreated(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	subject := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.SourceCapabilityID}
	if err := p.relations.Save(ctx, subject,
		readmodels.RelationEntry{EntryID: entryDependsOn, RelatedType: subjectTypeCapability, RelatedID: event.TargetCapabilityID, EdgeID: event.ID},
	); err != nil {
		return nil, err
	}
	return []readmodels.SubjectKey{subject}, nil
}

func (p *SubjectRelationProjector) onEdgeDeleted(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	affected, err := p.relations.SubjectsByEdge(ctx, event.ID)
	if err != nil {
		return nil, err
	}
	if err := p.relations.DeleteByEdge(ctx, event.ID); err != nil {
		return nil, err
	}
	return affected, nil
}

func (p *SubjectRelationProjector) onDomainAssigned(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	name, err := p.domains.Name(ctx, event.BusinessDomainID)
	if err != nil {
		return nil, err
	}
	subject := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.CapabilityID}
	if err := p.relations.Save(ctx, subject,
		readmodels.RelationEntry{EntryID: entryBusinessDomains, RelatedID: event.BusinessDomainID, RelatedName: name, EdgeID: event.ID},
	); err != nil {
		return nil, err
	}
	return []readmodels.SubjectKey{subject}, nil
}

func (p *SubjectRelationProjector) onBusinessDomainNamed(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	if err := p.domains.Save(ctx, event.ID, event.Name); err != nil {
		return nil, err
	}
	if err := p.relations.RenameRelated(ctx, readmodels.RelationTarget{EntryID: entryBusinessDomains, RelatedID: event.ID}, event.Name); err != nil {
		return nil, err
	}
	return nil, nil
}

func (p *SubjectRelationProjector) onBusinessDomainDeleted(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	target := readmodels.RelationTarget{EntryID: entryBusinessDomains, RelatedID: event.ID}
	affected, err := p.relations.SubjectsByRelated(ctx, target)
	if err != nil {
		return nil, err
	}
	if err := p.domains.Delete(ctx, event.ID); err != nil {
		return nil, err
	}
	if err := p.relations.DeleteByRelated(ctx, target); err != nil {
		return nil, err
	}
	return affected, nil
}

func (p *SubjectRelationProjector) onCapabilityCreated(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	if event.ParentID == "" {
		return nil, nil
	}
	return p.attachToParent(ctx, event.ID, event.ParentID)
}

func (p *SubjectRelationProjector) onParentChanged(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	capability := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.CapabilityID}
	if err := p.relations.Replace(ctx, capability, entryParentCapability, parentEntries(event.NewParentID)); err != nil {
		return nil, err
	}

	target := readmodels.RelationTarget{EntryID: entryChildCapabilities, RelatedID: event.CapabilityID}
	formerParents, err := p.relations.SubjectsByRelated(ctx, target)
	if err != nil {
		return nil, err
	}
	if err := p.relations.DeleteByRelated(ctx, target); err != nil {
		return nil, err
	}

	touched := append([]readmodels.SubjectKey{capability}, formerParents...)
	if event.NewParentID == "" {
		return touched, nil
	}
	if err := p.saveChild(ctx, event.NewParentID, event.CapabilityID); err != nil {
		return nil, err
	}
	return append(touched, readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: event.NewParentID}), nil
}

func (p *SubjectRelationProjector) attachToParent(ctx context.Context, capabilityID, parentID string) ([]readmodels.SubjectKey, error) {
	capability := readmodels.SubjectKey{SubjectType: subjectTypeCapability, SubjectID: capabilityID}
	if err := p.relations.Save(ctx, capability,
		readmodels.RelationEntry{EntryID: entryParentCapability, RelatedType: subjectTypeCapability, RelatedID: parentID},
	); err != nil {
		return nil, err
	}
	if err := p.saveChild(ctx, parentID, capabilityID); err != nil {
		return nil, err
	}
	return []readmodels.SubjectKey{capability, {SubjectType: subjectTypeCapability, SubjectID: parentID}}, nil
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

func (p *SubjectRelationProjector) onComponentRelationCreated(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	subject := readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: event.SourceComponentID}
	if err := p.relations.Save(ctx, subject,
		readmodels.RelationEntry{EntryID: entryComponentRelations, RelatedType: subjectTypeApplication, RelatedID: event.TargetComponentID, EdgeID: event.ID},
	); err != nil {
		return nil, err
	}
	return []readmodels.SubjectKey{subject}, nil
}

func (p *SubjectRelationProjector) onOriginLinkSet(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	return p.applyOriginLink(ctx, event.OriginType, event.ComponentID, event.EntityID)
}

func (p *SubjectRelationProjector) onOriginLinkReplaced(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	return p.applyOriginLink(ctx, event.OriginType, event.ComponentID, event.NewEntityID)
}

func (p *SubjectRelationProjector) onOriginLinkGone(ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
	return p.applyOriginLink(ctx, event.OriginType, event.ComponentID, "")
}

func (p *SubjectRelationProjector) applyOriginLink(ctx context.Context, originType, componentID, entityID string) ([]readmodels.SubjectKey, error) {
	entries, known := originEntriesByType[originType]
	if !known {
		return nil, fmt.Errorf("unknown origin type %q", originType)
	}

	application := readmodels.SubjectKey{SubjectType: subjectTypeApplication, SubjectID: componentID}
	if err := p.relations.Replace(ctx, application, entries.forward, entries.forwardEntries(entityID)); err != nil {
		return nil, err
	}

	mirrorTarget := readmodels.RelationTarget{EntryID: entries.mirror, RelatedID: componentID}
	formerRelated, err := p.relations.SubjectsByRelated(ctx, mirrorTarget)
	if err != nil {
		return nil, err
	}
	if err := p.relations.DeleteByRelated(ctx, mirrorTarget); err != nil {
		return nil, err
	}

	touched := append([]readmodels.SubjectKey{application}, formerRelated...)
	if entityID == "" {
		return touched, nil
	}
	newRelated := readmodels.SubjectKey{SubjectType: entries.relatedType, SubjectID: entityID}
	if err := p.relations.Save(ctx, newRelated,
		readmodels.RelationEntry{EntryID: entries.mirror, RelatedType: subjectTypeApplication, RelatedID: componentID},
	); err != nil {
		return nil, err
	}
	return append(touched, newRelated), nil
}

func subjectRelationsDeleted(subjectType string) relationHandler {
	return func(p *SubjectRelationProjector, ctx context.Context, event relationEventPayload) ([]readmodels.SubjectKey, error) {
		if err := p.relations.DeleteSubject(ctx, readmodels.SubjectKey{SubjectType: subjectType, SubjectID: event.ID}); err != nil {
			return nil, err
		}
		return nil, nil
	}
}

type relationHandler func(*SubjectRelationProjector, context.Context, relationEventPayload) ([]readmodels.SubjectKey, error)

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
