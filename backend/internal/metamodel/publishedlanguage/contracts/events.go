package contracts

import "time"

type StrategyPillarData struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Active            bool   `json:"active"`
	FitScoringEnabled bool   `json:"fitScoringEnabled"`
	FitCriteria       string `json:"fitCriteria"`
	FitType           string `json:"fitType"`
}

type MetaModelConfigurationCreatedPayload struct {
	ID        string               `json:"id"`
	TenantID  string               `json:"tenantId"`
	Pillars   []StrategyPillarData `json:"pillars"`
	CreatedAt time.Time            `json:"createdAt"`
	CreatedBy string               `json:"createdBy"`
}

type StrategyPillarAddedPayload struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	Version     int       `json:"version"`
	PillarID    string    `json:"pillarId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	ModifiedBy  string    `json:"modifiedBy"`
}

type StrategyPillarUpdatedPayload struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	Version        int       `json:"version"`
	PillarID       string    `json:"pillarId"`
	NewName        string    `json:"newName"`
	NewDescription string    `json:"newDescription"`
	ModifiedAt     time.Time `json:"modifiedAt"`
	ModifiedBy     string    `json:"modifiedBy"`
}

type StrategyPillarRemovedPayload struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenantId"`
	Version    int       `json:"version"`
	PillarID   string    `json:"pillarId"`
	ModifiedAt time.Time `json:"modifiedAt"`
	ModifiedBy string    `json:"modifiedBy"`
}

type PillarFitConfigurationUpdatedPayload struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenantId"`
	Version           int       `json:"version"`
	PillarID          string    `json:"pillarId"`
	FitScoringEnabled bool      `json:"fitScoringEnabled"`
	FitCriteria       string    `json:"fitCriteria"`
	FitType           string    `json:"fitType"`
	ModifiedAt        time.Time `json:"modifiedAt"`
	ModifiedBy        string    `json:"modifiedBy"`
}
