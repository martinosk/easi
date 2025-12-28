package events

type StrategyPillarData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type PillarEventParams struct {
	ConfigID   string
	TenantID   string
	Version    int
	PillarID   string
	ModifiedBy string
}

type AddPillarParams struct {
	PillarEventParams
	Name        string
	Description string
}

type UpdatePillarParams struct {
	PillarEventParams
	NewName        string
	NewDescription string
}
