package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type EffectiveBDStore interface {
	Upsert(ctx context.Context, dto readmodels.CMEffectiveBusinessDomainDTO) error
	Delete(ctx context.Context, capabilityID string) error
	GetByCapabilityID(ctx context.Context, capabilityID string) (*readmodels.CMEffectiveBusinessDomainDTO, error)
	UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, bdID string, bdName string) error
}

type BusinessDomainNameProvider interface {
	GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error)
}

type CapabilityChildProvider interface {
	GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error)
}

type EffectiveBusinessDomainProjector struct {
	store       EffectiveBDStore
	bdProvider  BusinessDomainNameProvider
	capProvider CapabilityChildProvider
}

func NewEffectiveBusinessDomainProjector(
	store EffectiveBDStore,
	bdProvider BusinessDomainNameProvider,
	capProvider CapabilityChildProvider,
) *EffectiveBusinessDomainProjector {
	return &EffectiveBusinessDomainProjector{
		store:       store,
		bdProvider:  bdProvider,
		capProvider: capProvider,
	}
}

func (p *EffectiveBusinessDomainProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *EffectiveBusinessDomainProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		"CapabilityCreated":              p.handleCapabilityCreated,
		"CapabilityDeleted":              p.handleCapabilityDeleted,
		"CapabilityParentChanged":        p.handleCapabilityParentChanged,
		"CapabilityLevelChanged":         p.handleCapabilityLevelChanged,
		"CapabilityAssignedToDomain":     p.handleCapabilityAssignedToDomain,
		"CapabilityUnassignedFromDomain": p.handleCapabilityUnassignedFromDomain,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type capCreatedEvent struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId"`
	Level    string `json:"level"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityCreated(ctx context.Context, eventData []byte) error {
	var event capCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityCreated event data: %w", err)
		log.Printf("failed to unmarshal CapabilityCreated event: %v", wrappedErr)
		return wrappedErr
	}

	l1CapabilityID := event.ID
	var businessDomainID, businessDomainName string

	if event.Level != "L1" && event.ParentID != "" {
		parentBD, err := p.store.GetByCapabilityID(ctx, event.ParentID)
		if err != nil {
			log.Printf("Failed to get parent effective BD for %s: %v", event.ParentID, err)
			return fmt.Errorf("load parent effective business domain for capability %s parent %s: %w", event.ID, event.ParentID, err)
		} else if parentBD != nil {
			l1CapabilityID = parentBD.L1CapabilityID
			businessDomainID = parentBD.BusinessDomainID
			businessDomainName = parentBD.BusinessDomainName
		}
	}

	if err := p.store.Upsert(ctx, readmodels.CMEffectiveBusinessDomainDTO{
		CapabilityID:       event.ID,
		L1CapabilityID:     l1CapabilityID,
		BusinessDomainID:   businessDomainID,
		BusinessDomainName: businessDomainName,
	}); err != nil {
		return fmt.Errorf("project CapabilityCreated effective business domain upsert for capability %s: %w", event.ID, err)
	}
	return nil
}

type capDeletedEvent struct {
	ID string `json:"id"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event capDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityDeleted event data: %w", err)
		log.Printf("failed to unmarshal CapabilityDeleted event: %v", wrappedErr)
		return wrappedErr
	}
	if err := p.store.Delete(ctx, event.ID); err != nil {
		return fmt.Errorf("project CapabilityDeleted effective business domain delete for capability %s: %w", event.ID, err)
	}
	return nil
}

type capParentChangedEvent struct {
	CapabilityID string `json:"capabilityId"`
	NewParentID  string `json:"newParentId"`
	NewLevel     string `json:"newLevel"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event capParentChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityParentChanged event data: %w", err)
		log.Printf("failed to unmarshal CapabilityParentChanged event: %v", wrappedErr)
		return wrappedErr
	}

	return p.updateSubtreeEffectiveBD(ctx, event.CapabilityID, event.NewParentID, event.NewLevel)
}

func (p *EffectiveBusinessDomainProjector) resolveL1AndBD(ctx context.Context, capabilityID, newParentID, newLevel string) (l1ID, bdID, bdName string) {
	if newLevel == "L1" {
		return capabilityID, "", ""
	}

	if newParentID != "" {
		parentBD, err := p.store.GetByCapabilityID(ctx, newParentID)
		if err == nil && parentBD != nil {
			return parentBD.L1CapabilityID, parentBD.BusinessDomainID, parentBD.BusinessDomainName
		}
	}

	return capabilityID, "", ""
}

func (p *EffectiveBusinessDomainProjector) collectSubtreeIDs(ctx context.Context, rootID string) ([]string, error) {
	result := []string{rootID}
	if p.capProvider == nil {
		return result, nil
	}

	children, err := p.capProvider.GetChildren(ctx, rootID)
	if err != nil {
		log.Printf("Failed to get children for %s: %v", rootID, err)
		return nil, fmt.Errorf("collect subtree children for capability %s: %w", rootID, err)
	}

	for _, child := range children {
		childIDs, childErr := p.collectSubtreeIDs(ctx, child.ID)
		if childErr != nil {
			return nil, childErr
		}
		result = append(result, childIDs...)
	}

	return result, nil
}

type capLevelChangedEvent struct {
	CapabilityID string `json:"capabilityId"`
	NewLevel     string `json:"newLevel"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityLevelChanged(ctx context.Context, eventData []byte) error {
	var event capLevelChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityLevelChanged event data: %w", err)
		log.Printf("failed to unmarshal CapabilityLevelChanged event: %v", wrappedErr)
		return wrappedErr
	}

	existing, err := p.store.GetByCapabilityID(ctx, event.CapabilityID)
	if err != nil || existing == nil {
		if err != nil {
			return fmt.Errorf("load effective business domain for capability %s level change: %w", event.CapabilityID, err)
		}
		return nil
	}

	return p.updateSubtreeEffectiveBD(ctx, event.CapabilityID, "", event.NewLevel)
}

func (p *EffectiveBusinessDomainProjector) updateSubtreeEffectiveBD(ctx context.Context, capabilityID, parentID, newLevel string) error {
	newL1, bdID, bdName := p.resolveL1AndBD(ctx, capabilityID, parentID, newLevel)
	subtreeIDs, err := p.collectSubtreeIDs(ctx, capabilityID)
	if err != nil {
		return fmt.Errorf("collect subtree for capability %s effective business domain update: %w", capabilityID, err)
	}

	for _, id := range subtreeIDs {
		if err := p.store.Upsert(ctx, readmodels.CMEffectiveBusinessDomainDTO{
			CapabilityID:       id,
			L1CapabilityID:     newL1,
			BusinessDomainID:   bdID,
			BusinessDomainName: bdName,
		}); err != nil {
			return fmt.Errorf("upsert effective business domain for capability %s in subtree rooted at %s: %w", id, capabilityID, err)
		}
	}

	return nil
}

type capAssignedToDomainEvent struct {
	BusinessDomainID string `json:"businessDomainId"`
	CapabilityID     string `json:"capabilityId"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event capAssignedToDomainEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityAssignedToDomain event data: %w", err)
		log.Printf("failed to unmarshal CapabilityAssignedToDomain event: %v", wrappedErr)
		return wrappedErr
	}

	bdName, err := p.lookupBusinessDomainName(ctx, event.BusinessDomainID)
	if err != nil {
		return fmt.Errorf("lookup business domain %s for capability %s assignment: %w", event.BusinessDomainID, event.CapabilityID, err)
	}

	return p.updateL1SubtreeBusinessDomain(ctx, event.CapabilityID, event.BusinessDomainID, bdName)
}

func (p *EffectiveBusinessDomainProjector) lookupBusinessDomainName(ctx context.Context, businessDomainID string) (string, error) {
	if p.bdProvider == nil {
		return "", nil
	}
	bd, err := p.bdProvider.GetByID(ctx, businessDomainID)
	if err != nil {
		return "", fmt.Errorf("load business domain %s: %w", businessDomainID, err)
	}
	if bd != nil {
		return bd.Name, nil
	}
	return "", nil
}

func (p *EffectiveBusinessDomainProjector) updateL1SubtreeBusinessDomain(ctx context.Context, capabilityID, bdID, bdName string) error {
	existing, err := p.store.GetByCapabilityID(ctx, capabilityID)
	if err != nil {
		return fmt.Errorf("load effective business domain for capability %s: %w", capabilityID, err)
	}
	if existing == nil {
		return nil
	}
	return p.store.UpdateBusinessDomainForL1Subtree(ctx, existing.L1CapabilityID, bdID, bdName)
}

type capUnassignedFromDomainEvent struct {
	CapabilityID string `json:"capabilityId"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event capUnassignedFromDomainEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		wrappedErr := fmt.Errorf("unmarshal CapabilityUnassignedFromDomain event data: %w", err)
		log.Printf("failed to unmarshal CapabilityUnassignedFromDomain event: %v", wrappedErr)
		return wrappedErr
	}
	return p.updateL1SubtreeBusinessDomain(ctx, event.CapabilityID, "", "")
}
