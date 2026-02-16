package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	domain "easi/backend/internal/shared/eventsourcing"
)

type MetadataStore interface {
	GetByID(ctx context.Context, capabilityID string) (*readmodels.DomainCapabilityMetadataDTO, error)
	Insert(ctx context.Context, dto readmodels.DomainCapabilityMetadataDTO) error
	Delete(ctx context.Context, capabilityID string) error
	UpdateParentAndL1(ctx context.Context, update readmodels.ParentL1Update) error
	UpdateLevel(ctx context.Context, capabilityID string, newLevel string) error
	UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, bd readmodels.BusinessDomainRef) error
	RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error
	GetSubtreeCapabilityIDs(ctx context.Context, rootID string) ([]string, error)
	GetEnterpriseCapabilitiesLinkedToCapabilities(ctx context.Context, capabilityIDs []string) ([]string, error)
	LookupBusinessDomainName(ctx context.Context, businessDomainID string) (string, error)
	UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error
}

type CapabilityCountUpdater interface {
	DecrementLinkCount(ctx context.Context, id string) error
	RecalculateDomainCount(ctx context.Context, enterpriseCapabilityID string) error
}

type CapabilityLinkStore interface {
	GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error)
	Delete(ctx context.Context, id string) error
	DeleteBlockingByBlocker(ctx context.Context, blockedByCapabilityID string) error
}

type DomainCapabilityMetadataProjector struct {
	metadataReadModel   MetadataStore
	capabilityReadModel CapabilityCountUpdater
	linkReadModel       CapabilityLinkStore
}

func NewDomainCapabilityMetadataProjector(
	metadataReadModel MetadataStore,
	capabilityReadModel CapabilityCountUpdater,
	linkReadModel CapabilityLinkStore,
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
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
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
		cmPL.CapabilityMetadataUpdated:      p.handleCapabilityMetadataUpdated,
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
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityCreatedEvent) error {
		l1CapabilityID := event.ID
		var businessDomainID, businessDomainName string

		if event.Level != "L1" && event.ParentID != "" {
			parentMeta, err := p.metadataReadModel.GetByID(ctx, event.ParentID)
			if err != nil {
				log.Printf("Failed to get parent metadata for %s: %v", event.ParentID, err)
				return fmt.Errorf("load parent metadata for capability %s parent %s: %w", event.ID, event.ParentID, err)
			} else if parentMeta != nil {
				l1CapabilityID = parentMeta.L1CapabilityID
				businessDomainID = parentMeta.BusinessDomainID
				businessDomainName = parentMeta.BusinessDomainName
			}
		}

		if err := p.metadataReadModel.Insert(ctx, readmodels.DomainCapabilityMetadataDTO{
			CapabilityID:       event.ID,
			CapabilityName:     event.Name,
			CapabilityLevel:    event.Level,
			ParentID:           event.ParentID,
			L1CapabilityID:     l1CapabilityID,
			BusinessDomainID:   businessDomainID,
			BusinessDomainName: businessDomainName,
		}); err != nil {
			return fmt.Errorf("project CapabilityCreated metadata insert for capability %s: %w", event.ID, err)
		}
		return nil
	})
}

type capabilityUpdatedEvent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityUpdatedEvent) error {
		existing, err := p.metadataReadModel.GetByID(ctx, event.ID)
		if err != nil {
			log.Printf("Failed to get existing metadata for %s: %v", event.ID, err)
			return fmt.Errorf("load existing metadata for capability %s: %w", event.ID, err)
		}
		if existing == nil {
			return nil
		}

		existing.CapabilityName = event.Name
		if err := p.metadataReadModel.Insert(ctx, *existing); err != nil {
			return fmt.Errorf("project CapabilityUpdated metadata upsert for capability %s: %w", event.ID, err)
		}
		return nil
	})
}

type capabilityDeletedEvent struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityDeletedEvent) error {
		if err := p.cleanupLinksForDeletedCapability(ctx, event.ID); err != nil {
			return fmt.Errorf("project CapabilityDeleted cleanup links for capability %s: %w", event.ID, err)
		}
		if err := p.metadataReadModel.Delete(ctx, event.ID); err != nil {
			return fmt.Errorf("project CapabilityDeleted metadata delete for capability %s: %w", event.ID, err)
		}
		return nil
	})
}

func (p *DomainCapabilityMetadataProjector) cleanupLinksForDeletedCapability(ctx context.Context, capabilityID string) error {
	link, err := p.linkReadModel.GetByDomainCapabilityID(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to check link for deleted capability %s: %v", capabilityID, err)
		return fmt.Errorf("cleanup enterprise links/counts for deleted capability %s: %w", capabilityID, err)
	}
	if link == nil {
		return nil
	}

	if err := p.linkReadModel.Delete(ctx, link.ID); err != nil {
		log.Printf("Failed to delete link for capability %s: %v", capabilityID, err)
		return fmt.Errorf("cleanup enterprise links/counts for deleted capability %s: %w", capabilityID, err)
	}
	if err := p.linkReadModel.DeleteBlockingByBlocker(ctx, capabilityID); err != nil {
		log.Printf("Failed to delete blocking records for capability %s: %v", capabilityID, err)
		return fmt.Errorf("cleanup enterprise links/counts for deleted capability %s: %w", capabilityID, err)
	}
	if err := p.capabilityReadModel.DecrementLinkCount(ctx, link.EnterpriseCapabilityID); err != nil {
		log.Printf("Failed to decrement link count: %v", err)
		return fmt.Errorf("cleanup enterprise links/counts for deleted capability %s: %w", capabilityID, err)
	}
	if err := p.capabilityReadModel.RecalculateDomainCount(ctx, link.EnterpriseCapabilityID); err != nil {
		log.Printf("Failed to recalculate domain count: %v", err)
		return fmt.Errorf("cleanup enterprise links/counts for deleted capability %s: %w", capabilityID, err)
	}

	return nil
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
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityParentChangedEvent) error {
		if err := p.metadataReadModel.UpdateParentAndL1(ctx, readmodels.ParentL1Update{
			CapabilityID:      event.CapabilityID,
			NewParentID:       event.NewParentID,
			NewLevel:          event.NewLevel,
			NewL1CapabilityID: event.CapabilityID,
		}); err != nil {
			log.Printf("Failed to update parent for %s: %v", event.CapabilityID, err)
			return fmt.Errorf("project CapabilityParentChanged parent/l1 update for capability %s: %w", event.CapabilityID, err)
		}
		return p.recalculateSubtreeAndDomainCounts(ctx, event.CapabilityID)
	})
}

type capabilityLevelChangedEvent struct {
	CapabilityID string `json:"capabilityId"`
	NewLevel     string `json:"newLevel"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityLevelChanged(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityLevelChangedEvent) error {
		if err := p.metadataReadModel.UpdateLevel(ctx, event.CapabilityID, event.NewLevel); err != nil {
			return fmt.Errorf("project CapabilityLevelChanged for capability %s: %w", event.CapabilityID, err)
		}
		return nil
	})
}

type capabilityAssignedToDomainEvent struct {
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	AssignedAt       time.Time `json:"assignedAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityAssignedToDomain(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityAssignedToDomainEvent) error {
		domainName, err := p.lookupBusinessDomainName(ctx, event.BusinessDomainID)
		if err != nil {
			return fmt.Errorf("project CapabilityAssignedToDomain lookup domain name for capability %s domain %s: %w", event.CapabilityID, event.BusinessDomainID, err)
		}
		return p.updateBusinessDomainAndRecalculate(ctx, event.CapabilityID, readmodels.BusinessDomainRef{ID: event.BusinessDomainID, Name: domainName})
	})
}

type capabilityUnassignedFromDomainEvent struct {
	ID               string    `json:"id"`
	BusinessDomainID string    `json:"businessDomainId"`
	CapabilityID     string    `json:"capabilityId"`
	UnassignedAt     time.Time `json:"unassignedAt"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityUnassignedFromDomain(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityUnassignedFromDomainEvent) error {
		return p.updateBusinessDomainAndRecalculate(ctx, event.CapabilityID, readmodels.BusinessDomainRef{})
	})
}

func (p *DomainCapabilityMetadataProjector) updateBusinessDomainAndRecalculate(ctx context.Context, capabilityID string, bd readmodels.BusinessDomainRef) error {
	meta, err := p.metadataReadModel.GetByID(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get metadata for %s: %v", capabilityID, err)
		return fmt.Errorf("load metadata for capability %s: %w", capabilityID, err)
	}
	if meta != nil {
		if err := p.metadataReadModel.UpdateBusinessDomainForL1Subtree(ctx, meta.L1CapabilityID, bd); err != nil {
			log.Printf("Failed to update business domain for L1 subtree %s: %v", meta.L1CapabilityID, err)
			return fmt.Errorf("update business domain for l1 subtree %s from capability %s: %w", meta.L1CapabilityID, capabilityID, err)
		}
	}

	return p.recalculateSubtreeAndDomainCounts(ctx, capabilityID)
}

func (p *DomainCapabilityMetadataProjector) recalculateSubtreeAndDomainCounts(ctx context.Context, capabilityID string) error {
	if err := p.metadataReadModel.RecalculateL1ForSubtree(ctx, capabilityID); err != nil {
		log.Printf("Failed to recalculate L1 for subtree %s: %v", capabilityID, err)
		return fmt.Errorf("recalculate l1 for subtree rooted at capability %s: %w", capabilityID, err)
	}

	subtreeIDs, err := p.metadataReadModel.GetSubtreeCapabilityIDs(ctx, capabilityID)
	if err != nil {
		log.Printf("Failed to get subtree for %s: %v", capabilityID, err)
		return fmt.Errorf("load subtree capability ids for capability %s: %w", capabilityID, err)
	}

	return p.recalculateDomainCountsForLinkedCapabilities(ctx, subtreeIDs)
}

func (p *DomainCapabilityMetadataProjector) recalculateDomainCountsForLinkedCapabilities(ctx context.Context, capabilityIDs []string) error {
	enterpriseCapIDs, err := p.metadataReadModel.GetEnterpriseCapabilitiesLinkedToCapabilities(ctx, capabilityIDs)
	if err != nil {
		log.Printf("Failed to get enterprise capabilities linked to capabilities: %v", err)
		return fmt.Errorf("load enterprise capabilities linked to %d capabilities: %w", len(capabilityIDs), err)
	}

	for _, enterpriseCapID := range enterpriseCapIDs {
		if err := p.capabilityReadModel.RecalculateDomainCount(ctx, enterpriseCapID); err != nil {
			log.Printf("Failed to recalculate domain count for enterprise capability %s: %v", enterpriseCapID, err)
			return fmt.Errorf("recalculate domain count for enterprise capability %s: %w", enterpriseCapID, err)
		}
	}

	return nil
}

func (p *DomainCapabilityMetadataProjector) lookupBusinessDomainName(ctx context.Context, businessDomainID string) (string, error) {
	name, err := p.metadataReadModel.LookupBusinessDomainName(ctx, businessDomainID)
	if err != nil {
		log.Printf("Failed to lookup business domain name for %s: %v", businessDomainID, err)
		return "", fmt.Errorf("lookup business domain name %s: %w", businessDomainID, err)
	}
	return name, nil
}

type capabilityMetadataUpdatedEvent struct {
	ID            string `json:"id"`
	MaturityValue int    `json:"maturityValue"`
}

func (p *DomainCapabilityMetadataProjector) handleCapabilityMetadataUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event capabilityMetadataUpdatedEvent) error {
		if err := p.metadataReadModel.UpdateMaturityValue(ctx, event.ID, event.MaturityValue); err != nil {
			return fmt.Errorf("project CapabilityMetadataUpdated maturity update for capability %s: %w", event.ID, err)
		}
		return nil
	})
}
