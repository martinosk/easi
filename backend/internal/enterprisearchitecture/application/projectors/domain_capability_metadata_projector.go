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
	UpdateBusinessDomainNameForDomain(ctx context.Context, businessDomainID, name string) error
	RecalculateL1ForSubtree(ctx context.Context, capabilityID string) error
	UpdateMaturityValue(ctx context.Context, capabilityID string, maturityValue int) error
}

type BusinessDomainNameLookup func(ctx context.Context, businessDomainID string) (string, error)

type DomainCapabilityMetadataProjector struct {
	metadataReadModel MetadataStore
	domainNames       BusinessDomainNameLookup
}

func NewDomainCapabilityMetadataProjector(metadataReadModel MetadataStore, domainNames BusinessDomainNameLookup) *DomainCapabilityMetadataProjector {
	return &DomainCapabilityMetadataProjector{metadataReadModel: metadataReadModel, domainNames: domainNames}
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
		cmPL.BusinessDomainUpdated:          p.handleBusinessDomainUpdated,
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
		if err := p.metadataReadModel.Delete(ctx, event.ID); err != nil {
			return fmt.Errorf("project CapabilityDeleted metadata delete for capability %s: %w", event.ID, err)
		}
		return nil
	})
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
		return p.recalculateL1ForSubtree(ctx, event.CapabilityID)
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
		domainName, err := p.domainNames(ctx, event.BusinessDomainID)
		if err != nil {
			return fmt.Errorf("project CapabilityAssignedToDomain lookup domain name for capability %s domain %s: %w", event.CapabilityID, event.BusinessDomainID, err)
		}
		return p.updateBusinessDomainAndRecalculate(ctx, event.CapabilityID, readmodels.BusinessDomainRef{ID: event.BusinessDomainID, Name: domainName})
	})
}

type businessDomainUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *DomainCapabilityMetadataProjector) handleBusinessDomainUpdated(ctx context.Context, eventData []byte) error {
	return handleProjection(ctx, eventData, func(ctx context.Context, event businessDomainUpdatedEvent) error {
		if err := p.metadataReadModel.UpdateBusinessDomainNameForDomain(ctx, event.ID, event.Name); err != nil {
			return fmt.Errorf("project BusinessDomainUpdated rename for domain %s: %w", event.ID, err)
		}
		return nil
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

	return p.recalculateL1ForSubtree(ctx, capabilityID)
}

func (p *DomainCapabilityMetadataProjector) recalculateL1ForSubtree(ctx context.Context, capabilityID string) error {
	if err := p.metadataReadModel.RecalculateL1ForSubtree(ctx, capabilityID); err != nil {
		log.Printf("Failed to recalculate L1 for subtree %s: %v", capabilityID, err)
		return fmt.Errorf("recalculate l1 for subtree rooted at capability %s: %w", capabilityID, err)
	}
	return nil
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
