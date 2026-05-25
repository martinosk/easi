package projectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"easi/backend/internal/architecturedirection/application/readmodels"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type capabilityDeletedPayload struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type nameChangePayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type domainAssignmentPayload struct {
	CapabilityID     string `json:"capabilityId"`
	BusinessDomainID string `json:"businessDomainId"`
}

type StaleReferenceStore interface {
	MarkSourceCapabilityStale(ctx context.Context, capabilityID readmodels.CapabilityID) error
	CacheReferenceName(ctx context.Context, entityType, entityID, name string) error
	UpdateCapabilityName(ctx context.Context, capabilityID readmodels.CapabilityID, name string) error
	CacheCapabilityDomain(ctx context.Context, capabilityID, businessDomainID string) error
	ClearCapabilityDomain(ctx context.Context, capabilityID string) error
	UpdateSourceCapabilityDomain(ctx context.Context, capabilityID readmodels.CapabilityID, businessDomainID string) error
	ClearSourceCapabilityDomain(ctx context.Context, capabilityID readmodels.CapabilityID) error
	UpdateBusinessDomainName(ctx context.Context, businessDomainID, name string) error
}

type StaleReferenceProjector struct {
	readModel StaleReferenceStore
}

func NewStaleReferenceProjector(readModel StaleReferenceStore) *StaleReferenceProjector {
	return &StaleReferenceProjector{readModel: readModel}
}

func (p *StaleReferenceProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		wrappedErr := fmt.Errorf("marshal %s event for aggregate %s: %w", event.EventType(), event.AggregateID(), err)
		log.Printf("failed to marshal event data: %v", wrappedErr)
		return wrappedErr
	}
	return p.ProjectEvent(ctx, event.EventType(), eventData)
}

type nameChangeSpec struct {
	entityType string
	update     func(ctx context.Context, id, name string) error
}

func (p *StaleReferenceProjector) ProjectEvent(ctx context.Context, eventType string, eventData []byte) error {
	switch eventType {
	case cmPL.CapabilityDeleted:
		return p.handleCapabilityDeleted(ctx, eventData)
	case cmPL.CapabilityCreated, cmPL.CapabilityUpdated:
		return p.handleNameChange(ctx, eventData, nameChangeSpec{
			entityType: "capability",
			update:     func(ctx context.Context, id, name string) error { return p.readModel.UpdateCapabilityName(ctx, readmodels.CapabilityID(id), name) },
		})
	case cmPL.BusinessDomainCreated, cmPL.BusinessDomainUpdated:
		return p.handleNameChange(ctx, eventData, nameChangeSpec{
			entityType: "business_domain",
			update:     func(ctx context.Context, id, name string) error { return p.readModel.UpdateBusinessDomainName(ctx, id, name) },
		})
	case cmPL.CapabilityAssignedToDomain:
		return p.handleDomainAssignment(ctx, eventData, false)
	case cmPL.CapabilityUnassignedFromDomain:
		return p.handleDomainAssignment(ctx, eventData, true)
	default:
		return nil
	}
}

func (p *StaleReferenceProjector) handleNameChange(ctx context.Context, eventData []byte, spec nameChangeSpec) error {
	var payload nameChangePayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal %s name-change payload: %w", spec.entityType, err)
	}
	if payload.ID == "" {
		return nil
	}
	if err := p.readModel.CacheReferenceName(ctx, spec.entityType, payload.ID, payload.Name); err != nil {
		return err
	}
	return spec.update(ctx, payload.ID, payload.Name)
}

func (p *StaleReferenceProjector) handleCapabilityDeleted(ctx context.Context, eventData []byte) error {
	var payload capabilityDeletedPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal CapabilityDeleted payload: %w", err)
	}
	if payload.ID == "" {
		return nil
	}
	return p.readModel.MarkSourceCapabilityStale(ctx, readmodels.CapabilityID(payload.ID))
}

func (p *StaleReferenceProjector) handleDomainAssignment(ctx context.Context, eventData []byte, clearing bool) error {
	var payload domainAssignmentPayload
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return fmt.Errorf("unmarshal domain-assignment payload: %w", err)
	}
	if payload.CapabilityID == "" {
		return nil
	}
	capID := readmodels.CapabilityID(payload.CapabilityID)
	if clearing {
		if err := p.readModel.ClearCapabilityDomain(ctx, payload.CapabilityID); err != nil {
			return err
		}
		return p.readModel.ClearSourceCapabilityDomain(ctx, capID)
	}
	if err := p.readModel.CacheCapabilityDomain(ctx, payload.CapabilityID, payload.BusinessDomainID); err != nil {
		return err
	}
	return p.readModel.UpdateSourceCapabilityDomain(ctx, capID, payload.BusinessDomainID)
}
