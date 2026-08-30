package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type TenantOIDC struct {
	DiscoveryURL string
	ClientID     string
	AuthMethod   string
	Scopes       string
}

type TenantCreated struct {
	domain.BaseEvent
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Domains         []string  `json:"domains"`
	FirstAdminEmail string    `json:"firstAdminEmail"`
	DiscoveryURL    string    `json:"discoveryUrl"`
	ClientID        string    `json:"clientId"`
	AuthMethod      string    `json:"authMethod"`
	Scopes          string    `json:"scopes"`
	CreatedAt       time.Time `json:"createdAt"`
}

type TenantDetails struct {
	ID              string
	Name            string
	Status          string
	Domains         []string
	FirstAdminEmail string
	OIDC            TenantOIDC
}

func NewTenantCreated(details TenantDetails) TenantCreated {
	return TenantCreated{
		BaseEvent:       domain.NewBaseEvent(details.ID),
		ID:              details.ID,
		Name:            details.Name,
		Status:          details.Status,
		Domains:         details.Domains,
		FirstAdminEmail: details.FirstAdminEmail,
		DiscoveryURL:    details.OIDC.DiscoveryURL,
		ClientID:        details.OIDC.ClientID,
		AuthMethod:      details.OIDC.AuthMethod,
		Scopes:          details.OIDC.Scopes,
		CreatedAt:       time.Now().UTC(),
	}
}

func (e TenantCreated) EventType() string {
	return "TenantCreated"
}

func (e TenantCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":              e.ID,
		"name":            e.Name,
		"status":          e.Status,
		"domains":         e.Domains,
		"firstAdminEmail": e.FirstAdminEmail,
		"discoveryUrl":    e.DiscoveryURL,
		"clientId":        e.ClientID,
		"authMethod":      e.AuthMethod,
		"scopes":          e.Scopes,
		"createdAt":       e.CreatedAt,
	}
}
