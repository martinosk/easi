package readmodels

import (
	"context"
	"database/sql"
	"fmt"

	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ComponentStatisticsDTO struct {
	Unknown   int                    `json:"unknown"`
	Nominated int                    `json:"nominated"`
	Owned     int                    `json:"owned"`
	Managed   int                    `json:"managed"`
	Hosting   HostingDistributionDTO `json:"hosting"`
	Total     int                    `json:"total"`
	Links     types.Links            `json:"_links,omitempty"`
}

type HostingDistributionDTO struct {
	OnPremises       int `json:"on-premises"`
	Cloud            int `json:"cloud"`
	SaaS             int `json:"saas"`
	ThirdPartyHosted int `json:"third-party-hosted"`
	Unknown          int `json:"unknown"`
}

func (rm *ApplicationComponentReadModel) Statistics(ctx context.Context) (ComponentStatisticsDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return ComponentStatisticsDTO{}, fmt.Errorf("resolve tenant for component statistics: %w", err)
	}

	var stats ComponentStatisticsDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			"SELECT ownership_state, hosting, COUNT(*) FROM architecturemodeling.application_components WHERE tenant_id = $1 AND is_deleted = FALSE GROUP BY ownership_state, hosting",
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var state, hosting string
			var count int
			if err := rows.Scan(&state, &hosting, &count); err != nil {
				return err
			}
			stats.add(state, hosting, count)
		}
		return rows.Err()
	})
	if err != nil {
		return ComponentStatisticsDTO{}, fmt.Errorf("component statistics for tenant %s: %w", tenantID.Value(), err)
	}
	return stats, nil
}

func (s *ComponentStatisticsDTO) add(state, hosting string, count int) {
	switch state {
	case valueobjects.OwnershipStateUnknown:
		s.Unknown += count
	case valueobjects.OwnershipStateNominated:
		s.Nominated += count
	case valueobjects.OwnershipStateOwned:
		s.Owned += count
	case valueobjects.OwnershipStateManaged:
		s.Managed += count
	}
	s.Hosting.add(hosting, count)
	s.Total += count
}

func (d *HostingDistributionDTO) add(hosting string, count int) {
	switch hosting {
	case valueobjects.HostingOnPremises:
		d.OnPremises += count
	case valueobjects.HostingCloud:
		d.Cloud += count
	case valueobjects.HostingSaaS:
		d.SaaS += count
	case valueobjects.HostingThirdPartyHosted:
		d.ThirdPartyHosted += count
	case valueobjects.HostingUnknown:
		d.Unknown += count
	}
}
