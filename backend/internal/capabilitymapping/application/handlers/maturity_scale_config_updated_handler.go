package handlers

import (
	"context"
	"log"

	domain "easi/backend/internal/shared/eventsourcing"
)

type MaturityScaleCacheInvalidator interface {
	InvalidateCache(tenantID string)
}

type MaturityScaleConfigUpdatedHandler struct {
	cacheInvalidator MaturityScaleCacheInvalidator
}

func NewMaturityScaleConfigUpdatedHandler(cacheInvalidator MaturityScaleCacheInvalidator) *MaturityScaleConfigUpdatedHandler {
	return &MaturityScaleConfigUpdatedHandler{
		cacheInvalidator: cacheInvalidator,
	}
}

func (h *MaturityScaleConfigUpdatedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	data := event.EventData()

	tenantID, ok := data["tenantId"].(string)
	if !ok {
		log.Printf("MaturityScaleConfigUpdatedHandler: missing or invalid tenantId in event")
		return nil
	}

	log.Printf("Invalidating maturity scale cache for tenant %s", tenantID)
	h.cacheInvalidator.InvalidateCache(tenantID)

	return nil
}
