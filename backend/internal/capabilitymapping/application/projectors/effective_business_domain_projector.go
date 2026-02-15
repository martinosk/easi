package projectors

import (
	"context"
	"encoding/json"
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
		log.Printf("Failed to marshal event data: %v", err)
		return err
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
		log.Printf("Failed to unmarshal CapabilityCreated event: %v", err)
		return err
	}

	l1CapabilityID := event.ID
	var businessDomainID, businessDomainName string

	if event.Level != "L1" && event.ParentID != "" {
		parentBD, err := p.store.GetByCapabilityID(ctx, event.ParentID)
		if err != nil {
			log.Printf("Failed to get parent effective BD for %s: %v", event.ParentID, err)
		} else if parentBD != nil {
			l1CapabilityID = parentBD.L1CapabilityID
			businessDomainID = parentBD.BusinessDomainID
			businessDomainName = parentBD.BusinessDomainName
		}
	}

	return p.store.Upsert(ctx, readmodels.CMEffectiveBusinessDomainDTO{
		CapabilityID:       event.ID,
		L1CapabilityID:     l1CapabilityID,
		BusinessDomainID:   businessDomainID,
		BusinessDomainName: businessDomainName,
	})
}

type capDeletedEvent struct {
	ID string `json:"id"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event capDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}
	return p.store.Delete(ctx, event.ID)
}

type capParentChangedEvent struct {
	CapabilityID string `json:"capabilityId"`
	NewParentID  string `json:"newParentId"`
	NewLevel     string `json:"newLevel"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event capParentChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
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

func (p *EffectiveBusinessDomainProjector) collectSubtreeIDs(ctx context.Context, rootID string) []string {
	result := []string{rootID}
	if p.capProvider == nil {
		return result
	}

	children, err := p.capProvider.GetChildren(ctx, rootID)
	if err != nil {
		log.Printf("Failed to get children for %s: %v", rootID, err)
		return result
	}

	for _, child := range children {
		result = append(result, p.collectSubtreeIDs(ctx, child.ID)...)
	}

	return result
}

type capLevelChangedEvent struct {
	CapabilityID string `json:"capabilityId"`
	NewLevel     string `json:"newLevel"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityLevelChanged(ctx context.Context, eventData []byte) error {
	var event capLevelChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityLevelChanged event: %v", err)
		return err
	}

	existing, err := p.store.GetByCapabilityID(ctx, event.CapabilityID)
	if err != nil || existing == nil {
		return err
	}

	return p.updateSubtreeEffectiveBD(ctx, event.CapabilityID, "", event.NewLevel)
}

func (p *EffectiveBusinessDomainProjector) updateSubtreeEffectiveBD(ctx context.Context, capabilityID, parentID, newLevel string) error {
	newL1, bdID, bdName := p.resolveL1AndBD(ctx, capabilityID, parentID, newLevel)

	for _, id := range p.collectSubtreeIDs(ctx, capabilityID) {
		if err := p.store.Upsert(ctx, readmodels.CMEffectiveBusinessDomainDTO{
			CapabilityID:       id,
			L1CapabilityID:     newL1,
			BusinessDomainID:   bdID,
			BusinessDomainName: bdName,
		}); err != nil {
			return err
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
		log.Printf("Failed to unmarshal CapabilityAssignedToDomain event: %v", err)
		return err
	}

	var bdName string
	if p.bdProvider != nil {
		bd, err := p.bdProvider.GetByID(ctx, event.BusinessDomainID)
		if err != nil {
			log.Printf("Failed to look up business domain %s: %v", event.BusinessDomainID, err)
		} else if bd != nil {
			bdName = bd.Name
		}
	}

	existing, err := p.store.GetByCapabilityID(ctx, event.CapabilityID)
	if err != nil {
		log.Printf("Failed to get effective BD for %s: %v", event.CapabilityID, err)
		return err
	}
	if existing == nil {
		return nil
	}

	return p.store.UpdateBusinessDomainForL1Subtree(ctx, existing.L1CapabilityID, event.BusinessDomainID, bdName)
}

type capUnassignedFromDomainEvent struct {
	CapabilityID string `json:"capabilityId"`
}

func (p *EffectiveBusinessDomainProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event capUnassignedFromDomainEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUnassignedFromDomain event: %v", err)
		return err
	}

	existing, err := p.store.GetByCapabilityID(ctx, event.CapabilityID)
	if err != nil {
		log.Printf("Failed to get effective BD for %s: %v", event.CapabilityID, err)
		return err
	}
	if existing == nil {
		return nil
	}

	return p.store.UpdateBusinessDomainForL1Subtree(ctx, existing.L1CapabilityID, "", "")
}
