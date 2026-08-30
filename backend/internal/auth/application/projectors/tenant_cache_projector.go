package projectors

import (
	"context"
	"encoding/json"
	"fmt"

	"easi/backend/internal/auth/application/readmodels"
	platformPL "easi/backend/internal/platform/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type TenantCacheStore interface {
	Save(ctx context.Context, entry readmodels.TenantCacheEntry) error
}

type TenantCacheProjector struct {
	cache TenantCacheStore
}

func NewTenantCacheProjector(cache TenantCacheStore) *TenantCacheProjector {
	return &TenantCacheProjector{cache: cache}
}

type tenantCreatedEventData struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	Domains      []string `json:"domains"`
	DiscoveryURL string   `json:"discoveryUrl"`
	IssuerURL    string   `json:"issuerUrl"`
	ClientID     string   `json:"clientId"`
	AuthMethod   string   `json:"authMethod"`
	Scopes       string   `json:"scopes"`
}

func (p *TenantCacheProjector) Handle(ctx context.Context, event domain.DomainEvent) error {
	if event.EventType() != platformPL.TenantCreated {
		return nil
	}

	payload, err := json.Marshal(event.EventData())
	if err != nil {
		return fmt.Errorf("marshal %s: %w", event.EventType(), err)
	}

	var data tenantCreatedEventData
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal %s: %w", event.EventType(), err)
	}

	return p.cache.Save(ctx, readmodels.TenantCacheEntry{
		TenantID:     data.ID,
		Name:         data.Name,
		Status:       data.Status,
		Domains:      data.Domains,
		DiscoveryURL: data.DiscoveryURL,
		IssuerURL:    data.IssuerURL,
		ClientID:     data.ClientID,
		AuthMethod:   data.AuthMethod,
		Scopes:       data.Scopes,
	})
}
