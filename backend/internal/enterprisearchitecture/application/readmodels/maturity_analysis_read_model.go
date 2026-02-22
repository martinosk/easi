package readmodels

import (
	"context"
	"database/sql"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type MaturityDistributionDTO struct {
	Genesis     int `json:"genesis"`
	CustomBuild int `json:"customBuild"`
	Product     int `json:"product"`
	Commodity   int `json:"commodity"`
}

type MaturityAnalysisCandidateDTO struct {
	EnterpriseCapabilityID   string                  `json:"enterpriseCapabilityId"`
	EnterpriseCapabilityName string                  `json:"enterpriseCapabilityName"`
	Category                 string                  `json:"category,omitempty"`
	TargetMaturity           *int                    `json:"targetMaturity,omitempty"`
	TargetMaturitySection    string                  `json:"targetMaturitySection,omitempty"`
	ImplementationCount      int                     `json:"implementationCount"`
	DomainCount              int                     `json:"domainCount"`
	MaxMaturity              int                     `json:"maxMaturity"`
	MinMaturity              int                     `json:"minMaturity"`
	AverageMaturity          int                     `json:"averageMaturity"`
	MaxGap                   int                     `json:"maxGap"`
	MaturityDistribution     MaturityDistributionDTO `json:"maturityDistribution"`
	Links                    types.Links             `json:"_links,omitempty"`
}

type MaturityAnalysisSummaryDTO struct {
	CandidateCount       int `json:"candidateCount"`
	TotalImplementations int `json:"totalImplementations"`
	AverageGap           int `json:"averageGap"`
}

type ImplementationDetailDTO struct {
	DomainCapabilityID   string `json:"domainCapabilityId"`
	DomainCapabilityName string `json:"domainCapabilityName"`
	BusinessDomainID     string `json:"businessDomainId,omitempty"`
	BusinessDomainName   string `json:"businessDomainName,omitempty"`
	MaturityValue        int    `json:"maturityValue"`
	MaturitySection      string `json:"maturitySection"`
	Gap                  int    `json:"gap"`
	Priority             string `json:"priority"`
}

type InvestmentPrioritiesDTO struct {
	High     []ImplementationDetailDTO `json:"high"`
	Medium   []ImplementationDetailDTO `json:"medium"`
	Low      []ImplementationDetailDTO `json:"low"`
	OnTarget []ImplementationDetailDTO `json:"onTarget"`
}

type MaturityGapDetailDTO struct {
	EnterpriseCapabilityID   string                    `json:"enterpriseCapabilityId"`
	EnterpriseCapabilityName string                    `json:"enterpriseCapabilityName"`
	Category                 string                    `json:"category,omitempty"`
	TargetMaturity           *int                      `json:"targetMaturity,omitempty"`
	TargetMaturitySection    string                    `json:"targetMaturitySection,omitempty"`
	Implementations          []ImplementationDetailDTO `json:"implementations"`
	InvestmentPriorities     InvestmentPrioritiesDTO   `json:"investmentPriorities"`
	Links                    types.Links               `json:"_links,omitempty"`
}

type MaturityAnalysisReadModel struct {
	db *database.TenantAwareDB
}

func NewMaturityAnalysisReadModel(db *database.TenantAwareDB) *MaturityAnalysisReadModel {
	return &MaturityAnalysisReadModel{db: db}
}

func (rm *MaturityAnalysisReadModel) GetMaturityAnalysisCandidates(ctx context.Context, sortBy string) ([]MaturityAnalysisCandidateDTO, MaturityAnalysisSummaryDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, MaturityAnalysisSummaryDTO{}, err
	}

	query := rm.buildCandidatesQuery(sortBy)
	candidates, totalGap, err := rm.fetchCandidates(ctx, query, tenantID.Value())
	if err != nil {
		return nil, MaturityAnalysisSummaryDTO{}, err
	}

	rm.enrichCandidatesWithDistribution(ctx, candidates)
	summary := rm.buildSummary(candidates, totalGap)

	return candidates, summary, nil
}

func (rm *MaturityAnalysisReadModel) buildCandidatesQuery(sortBy string) string {
	orderBy := "max_gap DESC, impl_count DESC"
	if sortBy == "implementations" {
		orderBy = "impl_count DESC, max_gap DESC"
	}

	return `
		SELECT
			ec.id, ec.name, ec.category, ec.target_maturity,
			COUNT(DISTINCT ecl.domain_capability_id) as impl_count,
			COUNT(DISTINCT dcm.business_domain_id) as domain_count,
			COALESCE(MAX(dcm.maturity_value), 0) as max_maturity,
			COALESCE(MIN(dcm.maturity_value), 0) as min_maturity,
			COALESCE(AVG(dcm.maturity_value)::int, 0) as avg_maturity,
			GREATEST(0, COALESCE(ec.target_maturity, COALESCE(MAX(dcm.maturity_value), 0)) - COALESCE(MIN(dcm.maturity_value), 0)) as max_gap
		FROM enterprisearchitecture.enterprise_capabilities ec
		LEFT JOIN enterprisearchitecture.enterprise_capability_links ecl ON ec.id = ecl.enterprise_capability_id AND ec.tenant_id = ecl.tenant_id
		LEFT JOIN enterprisearchitecture.domain_capability_metadata dcm ON ecl.domain_capability_id = dcm.capability_id AND ecl.tenant_id = dcm.tenant_id
		WHERE ec.tenant_id = $1 AND ec.active = true
		GROUP BY ec.id, ec.name, ec.category, ec.target_maturity
		ORDER BY ` + orderBy
}

func (rm *MaturityAnalysisReadModel) fetchCandidates(ctx context.Context, query string, tenantID any) ([]MaturityAnalysisCandidateDTO, int, error) {
	var candidates []MaturityAnalysisCandidateDTO
	var totalGap int

	err := rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			dto, err := rm.scanCandidate(rows)
			if err != nil {
				return err
			}
			totalGap += dto.MaxGap
			candidates = append(candidates, dto)
		}
		return rows.Err()
	})

	return candidates, totalGap, err
}

func (rm *MaturityAnalysisReadModel) scanCandidate(rows *sql.Rows) (MaturityAnalysisCandidateDTO, error) {
	var dto MaturityAnalysisCandidateDTO
	var category sql.NullString
	var targetMaturity sql.NullInt64

	if err := rows.Scan(
		&dto.EnterpriseCapabilityID, &dto.EnterpriseCapabilityName, &category, &targetMaturity,
		&dto.ImplementationCount, &dto.DomainCount,
		&dto.MaxMaturity, &dto.MinMaturity, &dto.AverageMaturity, &dto.MaxGap,
	); err != nil {
		return dto, err
	}

	if category.Valid {
		dto.Category = category.String
	}
	if targetMaturity.Valid {
		tm := int(targetMaturity.Int64)
		dto.TargetMaturity = &tm
		dto.TargetMaturitySection = getMaturitySection(tm)
	}
	return dto, nil
}

func (rm *MaturityAnalysisReadModel) enrichCandidatesWithDistribution(ctx context.Context, candidates []MaturityAnalysisCandidateDTO) {
	for i := range candidates {
		dist, _ := rm.getMaturityDistribution(ctx, candidates[i].EnterpriseCapabilityID)
		candidates[i].MaturityDistribution = dist
	}
}

func (rm *MaturityAnalysisReadModel) buildSummary(candidates []MaturityAnalysisCandidateDTO, totalGap int) MaturityAnalysisSummaryDTO {
	var totalImplementations int
	for _, c := range candidates {
		totalImplementations += c.ImplementationCount
	}

	avgGap := 0
	if len(candidates) > 0 {
		avgGap = totalGap / len(candidates)
	}

	return MaturityAnalysisSummaryDTO{
		CandidateCount:       len(candidates),
		TotalImplementations: totalImplementations,
		AverageGap:           avgGap,
	}
}

func (rm *MaturityAnalysisReadModel) getMaturityDistribution(ctx context.Context, enterpriseCapabilityID string) (MaturityDistributionDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return MaturityDistributionDTO{}, err
	}

	query := `
		SELECT
			COUNT(*) FILTER (WHERE dcm.maturity_value <= 24) as genesis,
			COUNT(*) FILTER (WHERE dcm.maturity_value > 24 AND dcm.maturity_value <= 49) as custom_build,
			COUNT(*) FILTER (WHERE dcm.maturity_value > 49 AND dcm.maturity_value <= 74) as product,
			COUNT(*) FILTER (WHERE dcm.maturity_value > 74) as commodity
		FROM enterprisearchitecture.enterprise_capability_links ecl
		JOIN enterprisearchitecture.domain_capability_metadata dcm ON ecl.domain_capability_id = dcm.capability_id AND ecl.tenant_id = dcm.tenant_id
		WHERE ecl.tenant_id = $1 AND ecl.enterprise_capability_id = $2`

	var dist MaturityDistributionDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, tenantID.Value(), enterpriseCapabilityID).Scan(
			&dist.Genesis, &dist.CustomBuild, &dist.Product, &dist.Commodity,
		)
	})

	return dist, err
}

func (rm *MaturityAnalysisReadModel) GetMaturityGapDetail(ctx context.Context, enterpriseCapabilityID string) (*MaturityGapDetailDTO, error) {
	dto, maxMaturity, err := rm.fetchGapDetailHeader(ctx, enterpriseCapabilityID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil
	}

	target := maxMaturity
	if dto.TargetMaturity != nil {
		target = *dto.TargetMaturity
	}

	implementations, err := rm.getImplementations(ctx, enterpriseCapabilityID, target)
	if err != nil {
		return nil, err
	}

	dto.Implementations = implementations
	dto.InvestmentPriorities = categorizeByPriority(implementations)

	return dto, nil
}

func (rm *MaturityAnalysisReadModel) fetchGapDetailHeader(ctx context.Context, enterpriseCapabilityID string) (*MaturityGapDetailDTO, int, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, 0, err
	}

	var dto MaturityGapDetailDTO
	var category sql.NullString
	var targetMaturity sql.NullInt64
	var maxMaturity int
	var notFound bool

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT ec.id, ec.name, ec.category, ec.target_maturity,
			        (SELECT MAX(dcm.maturity_value) FROM enterprisearchitecture.enterprise_capability_links ecl
			         JOIN enterprisearchitecture.domain_capability_metadata dcm ON ecl.domain_capability_id = dcm.capability_id AND ecl.tenant_id = dcm.tenant_id
			         WHERE ecl.enterprise_capability_id = ec.id AND ecl.tenant_id = ec.tenant_id) as max_mat
			 FROM enterprisearchitecture.enterprise_capabilities ec
			 WHERE ec.tenant_id = $1 AND ec.id = $2 AND ec.active = true`,
			tenantID.Value(), enterpriseCapabilityID,
		).Scan(&dto.EnterpriseCapabilityID, &dto.EnterpriseCapabilityName, &category, &targetMaturity, &maxMaturity)

		if err == sql.ErrNoRows {
			notFound = true
			return nil
		}
		return err
	})

	if err != nil {
		return nil, 0, err
	}
	if notFound {
		return nil, 0, nil
	}

	if category.Valid {
		dto.Category = category.String
	}
	if targetMaturity.Valid {
		tm := int(targetMaturity.Int64)
		dto.TargetMaturity = &tm
		dto.TargetMaturitySection = getMaturitySection(tm)
	}

	return &dto, maxMaturity, nil
}

func (rm *MaturityAnalysisReadModel) getImplementations(ctx context.Context, enterpriseCapabilityID string, target int) ([]ImplementationDetailDTO, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT ecl.domain_capability_id, dcm.capability_name,
		       dcm.business_domain_id, dcm.business_domain_name,
		       dcm.maturity_value
		FROM enterprisearchitecture.enterprise_capability_links ecl
		JOIN enterprisearchitecture.domain_capability_metadata dcm ON ecl.domain_capability_id = dcm.capability_id AND ecl.tenant_id = dcm.tenant_id
		WHERE ecl.tenant_id = $1 AND ecl.enterprise_capability_id = $2
		ORDER BY dcm.maturity_value ASC`

	var implementations []ImplementationDetailDTO
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query, tenantID.Value(), enterpriseCapabilityID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			impl, err := rm.scanImplementation(rows, target)
			if err != nil {
				return err
			}
			implementations = append(implementations, impl)
		}

		return rows.Err()
	})

	return implementations, err
}

func (rm *MaturityAnalysisReadModel) scanImplementation(rows *sql.Rows, target int) (ImplementationDetailDTO, error) {
	var impl ImplementationDetailDTO
	var capabilityName, businessDomainID, businessDomainName sql.NullString

	if err := rows.Scan(
		&impl.DomainCapabilityID, &capabilityName,
		&businessDomainID, &businessDomainName,
		&impl.MaturityValue,
	); err != nil {
		return impl, err
	}

	if capabilityName.Valid {
		impl.DomainCapabilityName = capabilityName.String
	}
	if businessDomainID.Valid {
		impl.BusinessDomainID = businessDomainID.String
	}
	if businessDomainName.Valid {
		impl.BusinessDomainName = businessDomainName.String
	}

	impl.MaturitySection = getMaturitySection(impl.MaturityValue)
	impl.Gap = target - impl.MaturityValue
	if impl.Gap < 0 {
		impl.Gap = 0
	}
	impl.Priority = getPriority(impl.Gap)

	return impl, nil
}

func getMaturitySection(value int) string {
	switch {
	case value <= 24:
		return "Genesis"
	case value <= 49:
		return "Custom Build"
	case value <= 74:
		return "Product"
	default:
		return "Commodity"
	}
}

func getPriority(gap int) string {
	switch {
	case gap > 40:
		return "High"
	case gap >= 15:
		return "Medium"
	case gap >= 1:
		return "Low"
	default:
		return "None"
	}
}

func categorizeByPriority(implementations []ImplementationDetailDTO) InvestmentPrioritiesDTO {
	result := InvestmentPrioritiesDTO{
		High:     []ImplementationDetailDTO{},
		Medium:   []ImplementationDetailDTO{},
		Low:      []ImplementationDetailDTO{},
		OnTarget: []ImplementationDetailDTO{},
	}

	for _, impl := range implementations {
		switch impl.Priority {
		case "High":
			result.High = append(result.High, impl)
		case "Medium":
			result.Medium = append(result.Medium, impl)
		case "Low":
			result.Low = append(result.Low, impl)
		default:
			result.OnTarget = append(result.OnTarget, impl)
		}
	}

	return result
}
