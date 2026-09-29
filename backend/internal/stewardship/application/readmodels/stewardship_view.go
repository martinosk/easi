package readmodels

import (
	"context"
	"time"

	"easi/backend/internal/shared/types"
	"easi/backend/internal/stewardship/domain/valueobjects"
)

type PersonDTO struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}

type ConcernStewardshipDTO struct {
	Concern     string      `json:"concern"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Steward     *PersonDTO  `json:"steward"`
	AssignedBy  *string     `json:"assignedBy"`
	AssignedAt  *time.Time  `json:"assignedAt"`
	Links       types.Links `json:"_links,omitempty"`
}

type DomainStewardshipsDTO struct {
	DomainID   string
	DomainName string
	Fallback   *PersonDTO
	Concerns   []ConcernStewardshipDTO
}

type domainSource interface {
	getWithArchitect(ctx context.Context, domainID string) (*cachedDomainWithArchitect, error)
}

type stewardshipRowSource interface {
	ForDomain(ctx context.Context, domainID string) ([]StewardshipRow, error)
}

type StewardshipView struct {
	domains domainSource
	rows    stewardshipRowSource
}

func NewStewardshipView(domains *DomainCacheReadModel, rows *StewardshipReadModel) *StewardshipView {
	return newStewardshipView(domains, rows)
}

func newStewardshipView(domains domainSource, rows stewardshipRowSource) *StewardshipView {
	return &StewardshipView{domains: domains, rows: rows}
}

func (v *StewardshipView) DomainStewardships(ctx context.Context, domainID string) (*DomainStewardshipsDTO, error) {
	domain, err := v.domains.getWithArchitect(ctx, domainID)
	if err != nil || domain == nil {
		return nil, err
	}
	rows, err := v.rows.ForDomain(ctx, domainID)
	if err != nil {
		return nil, err
	}
	return &DomainStewardshipsDTO{
		DomainID:   domain.ID,
		DomainName: domain.Name,
		Fallback:   fallbackOf(domain),
		Concerns:   concernStewardships(rows),
	}, nil
}

func fallbackOf(domain *cachedDomainWithArchitect) *PersonDTO {
	if domain.DomainArchitectID == nil {
		return nil
	}
	return &PersonDTO{ID: *domain.DomainArchitectID, Name: domain.DomainArchitectName}
}

func concernStewardships(rows []StewardshipRow) []ConcernStewardshipDTO {
	byConcern := make(map[string]StewardshipRow, len(rows))
	for _, row := range rows {
		byConcern[row.Concern] = row
	}
	concerns := valueobjects.AllConcerns()
	result := make([]ConcernStewardshipDTO, len(concerns))
	for i, concern := range concerns {
		result[i] = ConcernStewardshipDTO{Concern: concern.Value(), Label: concern.Label(), Description: concern.Description()}
		if row, assigned := byConcern[concern.Value()]; assigned {
			result[i].Steward = &PersonDTO{ID: row.StewardID, Name: row.StewardName}
			result[i].AssignedBy = &row.AssignedBy
			result[i].AssignedAt = &row.AssignedAt
		}
	}
	return result
}
