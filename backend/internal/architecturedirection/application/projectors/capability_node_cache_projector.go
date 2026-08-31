package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/architecturedirection/application/readmodels"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type CapabilityNodeCacheStore interface {
	GetByID(ctx context.Context, capabilityID string) (*readmodels.CapabilityNodeDTO, error)
	Insert(ctx context.Context, dto readmodels.CapabilityNodeDTO) error
	Delete(ctx context.Context, capabilityID string) error
	UpdateParentAndL1(ctx context.Context, update readmodels.ParentL1Update) error
	UpdateLevel(ctx context.Context, capabilityID, newLevel string) error
	UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, domain readmodels.BusinessDomainRef) error
	UpdateBusinessDomainNameForDomain(ctx context.Context, businessDomainID, name string) error
	RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error
	UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error
}

type BusinessDomainNameLookup func(ctx context.Context, businessDomainID string) (string, error)

type CapabilityNodeCacheProjector struct {
	cache       CapabilityNodeCacheStore
	domainNames BusinessDomainNameLookup
}

func NewCapabilityNodeCacheProjector(cache CapabilityNodeCacheStore, domainNames BusinessDomainNameLookup) *CapabilityNodeCacheProjector {
	return &CapabilityNodeCacheProjector{cache: cache, domainNames: domainNames}
}

func (p *CapabilityNodeCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	return dispatchReferenceEvent(ctx, event, p.ProjectEvent)
}

const defaultCapabilityMaturityValue = 12

type capabilityNodeEvent struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ParentID         string `json:"parentId"`
	Level            string `json:"level"`
	CapabilityID     string `json:"capabilityId"`
	NewParentID      string `json:"newParentId"`
	NewLevel         string `json:"newLevel"`
	BusinessDomainID string `json:"businessDomainId"`
	MaturityValue    *int   `json:"maturityValue"`
}

func (e capabilityNodeEvent) maturityValueOrDefault() int {
	if e.MaturityValue == nil {
		return defaultCapabilityMaturityValue
	}
	return *e.MaturityValue
}

type nodeEventHandler func(ctx context.Context, event capabilityNodeEvent) error

func (p *CapabilityNodeCacheProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]nodeEventHandler{
		cmPL.CapabilityCreated:              p.projectCreated,
		cmPL.CapabilityUpdated:              p.projectUpdated,
		cmPL.CapabilityDeleted:              p.projectDeleted,
		cmPL.CapabilityParentChanged:        p.projectParentChanged,
		cmPL.CapabilityLevelChanged:         p.projectLevelChanged,
		cmPL.CapabilityAssignedToDomain:     p.projectAssignedToDomain,
		cmPL.CapabilityUnassignedFromDomain: p.projectUnassignedFromDomain,
		cmPL.CapabilityMetadataUpdated:      p.projectMetadataUpdated,
		cmPL.BusinessDomainUpdated:          p.projectBusinessDomainUpdated,
	}
	handler, ok := handlers[eventType]
	if !ok {
		return nil
	}
	var event capabilityNodeEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("unmarshal %s payload: %w", eventType, err)
	}
	if err := handler(ctx, event); err != nil {
		return fmt.Errorf("project %s into capability node cache: %w", eventType, err)
	}
	return nil
}

func (p *CapabilityNodeCacheProjector) projectCreated(ctx context.Context, event capabilityNodeEvent) error {
	node := readmodels.CapabilityNodeDTO{
		CapabilityID: event.ID, CapabilityName: event.Name, CapabilityLevel: event.Level,
		ParentID: event.ParentID, L1CapabilityID: event.ID, MaturityValue: event.maturityValueOrDefault(),
	}
	if event.Level != "L1" && event.ParentID != "" {
		parent, err := p.cache.GetByID(ctx, event.ParentID)
		if err != nil {
			return fmt.Errorf("load parent node %s of capability %s: %w", event.ParentID, event.ID, err)
		}
		if parent != nil {
			node.L1CapabilityID = parent.L1CapabilityID
			node.BusinessDomainID = parent.BusinessDomainID
			node.BusinessDomainName = parent.BusinessDomainName
		}
	}
	return p.cache.Insert(ctx, node)
}

func (p *CapabilityNodeCacheProjector) projectUpdated(ctx context.Context, event capabilityNodeEvent) error {
	existing, err := p.cache.GetByID(ctx, event.ID)
	if err != nil || existing == nil {
		return err
	}
	existing.CapabilityName = event.Name
	return p.cache.Insert(ctx, *existing)
}

func (p *CapabilityNodeCacheProjector) projectDeleted(ctx context.Context, event capabilityNodeEvent) error {
	return p.cache.Delete(ctx, event.ID)
}

func (p *CapabilityNodeCacheProjector) projectParentChanged(ctx context.Context, event capabilityNodeEvent) error {
	update := readmodels.ParentL1Update{
		CapabilityID: event.CapabilityID, NewParentID: event.NewParentID,
		NewLevel: event.NewLevel, NewL1CapabilityID: event.CapabilityID,
	}
	if err := p.cache.UpdateParentAndL1(ctx, update); err != nil {
		return err
	}
	return p.cache.RecalculateL1ForSubtree(ctx, event.CapabilityID)
}

func (p *CapabilityNodeCacheProjector) projectLevelChanged(ctx context.Context, event capabilityNodeEvent) error {
	return p.cache.UpdateLevel(ctx, event.CapabilityID, event.NewLevel)
}

func (p *CapabilityNodeCacheProjector) projectAssignedToDomain(ctx context.Context, event capabilityNodeEvent) error {
	name, err := p.domainNames(ctx, event.BusinessDomainID)
	if err != nil {
		return fmt.Errorf("look up business domain %s: %w", event.BusinessDomainID, err)
	}
	return p.applyDomainToL1Subtree(ctx, event.CapabilityID, readmodels.BusinessDomainRef{ID: event.BusinessDomainID, Name: name})
}

func (p *CapabilityNodeCacheProjector) projectUnassignedFromDomain(ctx context.Context, event capabilityNodeEvent) error {
	return p.applyDomainToL1Subtree(ctx, event.CapabilityID, readmodels.BusinessDomainRef{})
}

func (p *CapabilityNodeCacheProjector) applyDomainToL1Subtree(ctx context.Context, capabilityID string, domain readmodels.BusinessDomainRef) error {
	node, err := p.cache.GetByID(ctx, capabilityID)
	if err != nil {
		return err
	}
	if node != nil && node.CapabilityLevel != "L1" {
		return nil
	}
	if node != nil {
		if err := p.cache.UpdateBusinessDomainForL1Subtree(ctx, node.L1CapabilityID, domain); err != nil {
			return err
		}
	}
	return p.cache.RecalculateL1ForSubtree(ctx, capabilityID)
}

func (p *CapabilityNodeCacheProjector) projectMetadataUpdated(ctx context.Context, event capabilityNodeEvent) error {
	return p.cache.UpdateMaturityValue(ctx, event.ID, event.maturityValueOrDefault())
}

func (p *CapabilityNodeCacheProjector) projectBusinessDomainUpdated(ctx context.Context, event capabilityNodeEvent) error {
	return p.cache.UpdateBusinessDomainNameForDomain(ctx, event.ID, event.Name)
}
