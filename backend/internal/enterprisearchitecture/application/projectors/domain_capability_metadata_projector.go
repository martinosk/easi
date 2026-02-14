package projectors

import (
	"context"
	"encoding/json"
	"log"
	"time"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type DomainCapabilityMetadataProjector struct {
	metadataReadModel    *readmodels.DomainCapabilityMetadataReadModel
	capabilityReadModel  *readmodels.EnterpriseCapabilityReadModel
	linkReadModel        *readmodels.EnterpriseCapabilityLinkReadModel
}

func NewDomainCapabilityMetadataProjector(
	metadataReadModel *readmodels.DomainCapabilityMetadataReadModel,
	capabilityReadModel *readmodels.EnterpriseCapabilityReadModel,
	linkReadModel *readmodels.EnterpriseCapabilityLinkReadModel,
) *DomainCapabilityMetadataProjector {
	return &DomainCapabilityMetadataProjector{
		metadataReadModel:   metadataReadModel,
		capabilityReadModel: capabilityReadModel,
		linkReadModel:       linkReadModel,
	}
}

func (p *DomainCapabilityMetadataProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		log.Printf("Failed to marshal event data: %v", err)
		return err
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

func (p *DomainCapabilityMetadataProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	handlers := map[string]func(context.Context, []byte) error{
		cmPL.CapabilityCreated:              p.handleCapabilityCreated,
		cmPL.CapabilityUpdated:              p.handleCapabilityUpdated,
		cmPL.CapabilityDeleted:              p.handleCapabilityDeleted,
		cmPL.CapabilityParentChanged:        p.handleCapabilityParentChanged,
		cmPL.CapabilityLevelChanged:         p.handleCapabilityLevelChanged,
		cmPL.CapabilityAssignedToDomain:     p.handleCapabilityAssignedToDomain,
		cmPL.CapabilityUnassignedFromDomain: p.handleCapabilityUnassignedFromDomain,
	}

	if handler, exists := handlers[eventType]; exists {
		return handler(ctx, eventData)
	}
	return nil
}

type capabilityCreatedEvent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    string    `json:"parentId"`
	Level       string    `json:"level"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityCreated(ctx context.Context, eventData []byte) error {
	var event capabilityCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityCreated event: %v", err)
		return err
	}

	l1CapabilityID := event.ID
	var businessDomainID, businessDomainName string

	if event.Level != "L1" && event.ParentID != "" {
		parentMeta, err := p.metadataReadModel.GetByID(ctx, event.ParentID)
		if err != nil {
			log.Printf("Failed to get parent metadata for %s: %v", event.ParentID, err)
		} else if parentMeta != nil {
			l1CapabilityID = parentMeta.L1CapabilityID
			businessDomainID = parentMeta.BusinessDomainID
			businessDomainName = parentMeta.BusinessDomainName
		}
	}

	dto := readmodels.DomainCapabilityMetadataDTO{
		CapabilityID:       event.ID,
		CapabilityName:     event.Name,
		CapabilityLevel:    event.Level,
		ParentID:           event.ParentID,
		L1CapabilityID:     l1CapabilityID,
		BusinessDomainID:   businessDomainID,
		BusinessDomainName: businessDomainName,
	}

	return p.metadataReadModel.Insert(ctx, dto)
}

type capabilityUpdatedEvent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	var event capabilityUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUpdated event: %v", err)
		return err
	}

	existing, err := p.metadataReadModel.GetByID(ctx, event.ID)
	if err != nil {
		log.Printf("Failed to get existing metadata for %s: %v", event.ID, err)
		return err
	}
	if existing == nil {
		return nil
	}

	existing.CapabilityName = event.Name
	return p.metadataReadModel.Insert(ctx, *existing)
}

type capabilityDeletedEvent struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var event capabilityDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityDeleted event: %v", err)
		return err
	}

	p.cleanupLinksForDeletedCapability(ctx, event.ID)

	return p.metadataReadModel.Delete(ctx, event.ID)
}

func (p *DomainCapabilityMetadataProjector) cleanupLinksForDeletedCapability(ctx context.Context, capabilityID string) {
	link, err := p.linkReadModel.GetByDomainCapabilityID(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to check link for deleted capability %s: %v", capabilityID, err)
		return
	}
	if link == nil {
		return
	}

	if err := p.linkReadModel.Delete(ctx, link.ID); err != nil {
		log.Printf("Failed to delete link for capability %s: %v", capabilityID, err)
	}
	if err := p.linkReadModel.DeleteBlockingByBlocker(ctx, capabilityID); err != nil {
		log.Printf("Failed to delete blocking records for capability %s: %v", capabilityID, err)
	}
	if err := p.capabilityReadModel.DecrementLinkCount(ctx, link.EnterpriseCapabilityID); err != nil {
		log.Printf("Failed to decrement link count: %v", err)
	}
	if err := p.capabilityReadModel.RecalculateDomainCount(ctx, link.EnterpriseCapabilityID); err != nil {
		log.Printf("Failed to recalculate domain count: %v", err)
	}
}

type capabilityParentChangedEvent struct {
	CapabilityID string    `json:"capabilityId"`
	OldParentID  string    `json:"oldParentId"`
	NewParentID  string    `json:"newParentId"`
	OldLevel     string    `json:"oldLevel"`
	NewLevel     string    `json:"newLevel"`
	Timestamp    time.Time `json:"timestamp"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityParentChanged(ctx context.Context, eventData []byte) error {
	var event capabilityParentChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityParentChanged event: %v", err)
		return err
	}

	if err := p.metadataReadModel.UpdateParentAndL1(ctx, readmodels.ParentL1Update{
		CapabilityID:      event.CapabilityID,
		NewParentID:       event.NewParentID,
		NewLevel:          event.NewLevel,
		NewL1CapabilityID: event.CapabilityID,
	}); err != nil {
		log.Printf("Failed to update parent for %s: %v", event.CapabilityID, err)
		return err
	}

	return p.recalculateSubtreeAndDomainCounts(ctx, event.CapabilityID)
}

type capabilityLevelChangedEvent struct {
	CapabilityID string `json:"capabilityId"`
	NewLevel     string `json:"newLevel"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityLevelChanged(ctx context.Context, eventData []byte) error {
	var event capabilityLevelChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityLevelChanged event: %v", err)
		return err
	}

	return p.metadataReadModel.UpdateLevel(ctx, event.CapabilityID, event.NewLevel)
}

type capabilityAssignedToDomainEvent struct {
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	AssignedAt       time.Time `json:"assignedAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	var event capabilityAssignedToDomainEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityAssignedToDomain event: %v", err)
		return err
	}

	domainName := p.lookupBusinessDomainName(ctx, event.BusinessDomainID)
	return p.updateBusinessDomainAndRecalculate(ctx, event.CapabilityID, event.BusinessDomainID, domainName)
}

type capabilityUnassignedFromDomainEvent struct {
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	UnassignedAt     time.Time `json:"unassignedAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	var event capabilityUnassignedFromDomainEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		log.Printf("Failed to unmarshal CapabilityUnassignedFromDomain event: %v", err)
		return err
	}

	return p.updateBusinessDomainAndRecalculate(ctx, event.CapabilityID, "", "")
}

func (p *DomainCapabilityMetadataProjector) updateBusinessDomainAndRecalculate(ctx context.Context, capabilityID, businessDomainID, businessDomainName string) error {
	meta, err := p.metadataReadModel.GetByID(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get metadata for %s: %v", capabilityID, err)
		return err
	}
	if meta != nil {
		if err := p.metadataReadModel.UpdateBusinessDomainForL1Subtree(ctx, meta.L1CapabilityID, businessDomainID, businessDomainName); err != nil {
			log.Printf("Failed to update business domain for L1 subtree %s: %v", meta.L1CapabilityID, err)
			return err
		}
	}

	return p.recalculateSubtreeAndDomainCounts(ctx, capabilityID)
}

func (p *DomainCapabilityMetadataProjector) recalculateSubtreeAndDomainCounts(ctx context.Context, capabilityID string) error {
	if err := p.metadataReadModel.RecalculateL1ForSubtree(ctx, capabilityID); err != nil {
		log.Printf("Failed to recalculate L1 for subtree %s: %v", capabilityID, err)
		return err
	}

	subtreeIDs, err := p.metadataReadModel.GetSubtreeCapabilityIDs(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get subtree for %s: %v", capabilityID, err)
		return err
	}

	return p.recalculateDomainCountsForLinkedCapabilities(ctx, subtreeIDs)
}

func (p *DomainCapabilityMetadataProjector) recalculateDomainCountsForLinkedCapabilities(ctx context.Context, capabilityIDs []string) error {
	enterpriseCapIDs, err := p.metadataReadModel.GetEnterpriseCapabilitiesLinkedToCapabilities(ctx, capabilityIDs)
	if err != nil {
		log.Printf("Failed to get enterprise capabilities linked to capabilities: %v", err)
		return err
	}

	for _, enterpriseCapID := range enterpriseCapIDs {
		if err := p.capabilityReadModel.RecalculateDomainCount(ctx, enterpriseCapID); err != nil {
			log.Printf("Failed to recalculate domain count for enterprise capability %s: %v", enterpriseCapID, err)
		}
	}

	return nil
}

func (p *DomainCapabilityMetadataProjector) lookupBusinessDomainName(ctx context.Context, businessDomainID string) string {
	name, err := p.metadataReadModel.LookupBusinessDomainName(ctx, businessDomainID)
	if err != nil {
		log.Printf("Failed to lookup business domain name for %s: %v", businessDomainID, err)
		return businessDomainID
	}
	return name
}
