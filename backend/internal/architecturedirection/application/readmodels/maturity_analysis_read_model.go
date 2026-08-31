package readmodels

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"

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

type IncludedCapabilitiesProvider interface {
	IncludedCapabilityIDsByEC(ctx context.Context) (map[string][]string, error)
}

type MaturityAnalysisReadModel struct {
	db          *database.TenantAwareDB
	composition IncludedCapabilitiesProvider
}

func NewMaturityAnalysisReadModel(db *database.TenantAwareDB, composition IncludedCapabilitiesProvider) *MaturityAnalysisReadModel {
	return &MaturityAnalysisReadModel{db: db, composition: composition}
}

type ecHeaderRow struct {
	ID             string
	Name           string
	Category       string
	TargetMaturity *int
}

type capabilityMaturityRow struct {
	CapabilityID       string
	CapabilityName     string
	BusinessDomainID   string
	BusinessDomainName string
	MaturityValue      int
}

func (rm *MaturityAnalysisReadModel) GetMaturityAnalysisCandidates(ctx context.Context, sortBy string) ([]MaturityAnalysisCandidateDTO, MaturityAnalysisSummaryDTO, error) {
	headers, err := rm.loadActiveEnterpriseCapabilities(ctx)
	if err != nil {
		return nil, MaturityAnalysisSummaryDTO{}, err
	}
	meta, err := rm.loadCapabilityMaturity(ctx)
	if err != nil {
		return nil, MaturityAnalysisSummaryDTO{}, err
	}
	included, err := rm.composition.IncludedCapabilityIDsByEC(ctx)
	if err != nil {
		return nil, MaturityAnalysisSummaryDTO{}, fmt.Errorf("load included capabilities per enterprise capability: %w", err)
	}

	candidates := make([]MaturityAnalysisCandidateDTO, len(headers))
	for i, header := range headers {
		candidates[i] = buildMaturityCandidate(header, included[header.ID], meta)
	}
	sortMaturityCandidates(candidates, sortBy)

	return candidates, buildMaturitySummary(candidates), nil
}

func (rm *MaturityAnalysisReadModel) GetMaturityGapDetail(ctx context.Context, enterpriseCapabilityID string) (*MaturityGapDetailDTO, error) {
	header, err := rm.loadEnterpriseCapability(ctx, enterpriseCapabilityID)
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, nil
	}
	meta, err := rm.loadCapabilityMaturity(ctx)
	if err != nil {
		return nil, err
	}
	included, err := rm.composition.IncludedCapabilityIDsByEC(ctx)
	if err != nil {
		return nil, fmt.Errorf("load included capabilities for enterprise capability %s: %w", enterpriseCapabilityID, err)
	}

	dto := &MaturityGapDetailDTO{
		EnterpriseCapabilityID:   header.ID,
		EnterpriseCapabilityName: header.Name,
		Category:                 header.Category,
		TargetMaturity:           header.TargetMaturity,
	}
	if header.TargetMaturity != nil {
		dto.TargetMaturitySection = getMaturitySection(*header.TargetMaturity)
	}

	includedIDs := included[enterpriseCapabilityID]
	target := maxMaturityOf(includedIDs, meta)
	if header.TargetMaturity != nil {
		target = *header.TargetMaturity
	}
	dto.Implementations = buildGapImplementations(includedIDs, meta, target)
	dto.InvestmentPriorities = categorizeByPriority(dto.Implementations)

	return dto, nil
}

func (rm *MaturityAnalysisReadModel) loadActiveEnterpriseCapabilities(ctx context.Context) ([]ecHeaderRow, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var headers []ecHeaderRow
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT id, name, category, target_maturity FROM architecturedirection.enterprise_capability_cache
			 WHERE tenant_id = $1 AND active = true`,
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			header, err := scanECHeader(rows)
			if err != nil {
				return err
			}
			headers = append(headers, header)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("load active enterprise capabilities: %w", err)
	}
	return headers, nil
}

func (rm *MaturityAnalysisReadModel) loadEnterpriseCapability(ctx context.Context, id string) (*ecHeaderRow, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	var header *ecHeaderRow
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx,
			`SELECT id, name, category, target_maturity FROM architecturedirection.enterprise_capability_cache
			 WHERE tenant_id = $1 AND id = $2 AND active = true`,
			tenantID.Value(), id,
		)
		scanned, scanErr := scanECHeader(row)
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		header = &scanned
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load enterprise capability %s: %w", id, err)
	}
	return header, nil
}

type maturityRowScanner interface {
	Scan(dest ...any) error
}

func scanECHeader(row maturityRowScanner) (ecHeaderRow, error) {
	var header ecHeaderRow
	var category sql.NullString
	var targetMaturity sql.NullInt64
	if err := row.Scan(&header.ID, &header.Name, &category, &targetMaturity); err != nil {
		return header, err
	}
	header.Category = category.String
	if targetMaturity.Valid {
		target := int(targetMaturity.Int64)
		header.TargetMaturity = &target
	}
	return header, nil
}

func (rm *MaturityAnalysisReadModel) loadCapabilityMaturity(ctx context.Context) (map[string]capabilityMaturityRow, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}
	meta := map[string]capabilityMaturityRow{}
	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx,
			`SELECT capability_id, capability_name, business_domain_id, business_domain_name, maturity_value
			 FROM architecturedirection.capability_node_cache WHERE tenant_id = $1`,
			tenantID.Value(),
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var row capabilityMaturityRow
			var domainID, domainName sql.NullString
			if err := rows.Scan(&row.CapabilityID, &row.CapabilityName, &domainID, &domainName, &row.MaturityValue); err != nil {
				return err
			}
			row.BusinessDomainID = domainID.String
			row.BusinessDomainName = domainName.String
			meta[row.CapabilityID] = row
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("load capability maturity metadata: %w", err)
	}
	return meta, nil
}

func buildMaturityCandidate(header ecHeaderRow, includedIDs []string, meta map[string]capabilityMaturityRow) MaturityAnalysisCandidateDTO {
	candidate := MaturityAnalysisCandidateDTO{
		EnterpriseCapabilityID:   header.ID,
		EnterpriseCapabilityName: header.Name,
		Category:                 header.Category,
		TargetMaturity:           header.TargetMaturity,
	}
	if header.TargetMaturity != nil {
		candidate.TargetMaturitySection = getMaturitySection(*header.TargetMaturity)
	}

	stats := accumulateMaturityStats(includedIDs, meta)
	candidate.ImplementationCount = stats.count
	candidate.DomainCount = len(stats.domains)
	candidate.MaxMaturity = stats.maxValue
	candidate.MinMaturity = stats.minValue
	candidate.AverageMaturity = stats.average()
	candidate.MaturityDistribution = stats.distribution

	target := candidate.MaxMaturity
	if header.TargetMaturity != nil {
		target = *header.TargetMaturity
	}
	if gap := target - candidate.MinMaturity; gap > 0 {
		candidate.MaxGap = gap
	}
	return candidate
}

type maturityStats struct {
	count        int
	total        int
	maxValue     int
	minValue     int
	domains      map[string]struct{}
	distribution MaturityDistributionDTO
}

func accumulateMaturityStats(includedIDs []string, meta map[string]capabilityMaturityRow) maturityStats {
	stats := maturityStats{domains: map[string]struct{}{}}
	for _, id := range includedIDs {
		if row, ok := meta[id]; ok {
			stats.add(row)
		}
	}
	return stats
}

func (s *maturityStats) add(row capabilityMaturityRow) {
	s.count++
	s.total += row.MaturityValue
	if row.BusinessDomainID != "" {
		s.domains[row.BusinessDomainID] = struct{}{}
	}
	if s.count == 1 || row.MaturityValue > s.maxValue {
		s.maxValue = row.MaturityValue
	}
	if s.count == 1 || row.MaturityValue < s.minValue {
		s.minValue = row.MaturityValue
	}
	s.accumulateDistribution(row.MaturityValue)
}

func (s *maturityStats) average() int {
	if s.count == 0 {
		return 0
	}
	return int(math.Round(float64(s.total) / float64(s.count)))
}

func (s *maturityStats) accumulateDistribution(value int) {
	switch {
	case value <= 24:
		s.distribution.Genesis++
	case value <= 49:
		s.distribution.CustomBuild++
	case value <= 74:
		s.distribution.Product++
	default:
		s.distribution.Commodity++
	}
}

func sortMaturityCandidates(candidates []MaturityAnalysisCandidateDTO, sortBy string) {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if sortBy == "implementations" {
			if a.ImplementationCount != b.ImplementationCount {
				return a.ImplementationCount > b.ImplementationCount
			}
			return a.MaxGap > b.MaxGap
		}
		if a.MaxGap != b.MaxGap {
			return a.MaxGap > b.MaxGap
		}
		return a.ImplementationCount > b.ImplementationCount
	})
}

func buildMaturitySummary(candidates []MaturityAnalysisCandidateDTO) MaturityAnalysisSummaryDTO {
	summary := MaturityAnalysisSummaryDTO{CandidateCount: len(candidates)}
	totalGap := 0
	for _, c := range candidates {
		summary.TotalImplementations += c.ImplementationCount
		totalGap += c.MaxGap
	}
	if len(candidates) > 0 {
		summary.AverageGap = totalGap / len(candidates)
	}
	return summary
}

func maxMaturityOf(includedIDs []string, meta map[string]capabilityMaturityRow) int {
	maxValue := 0
	for _, id := range includedIDs {
		if row, ok := meta[id]; ok && row.MaturityValue > maxValue {
			maxValue = row.MaturityValue
		}
	}
	return maxValue
}

func buildGapImplementations(includedIDs []string, meta map[string]capabilityMaturityRow, target int) []ImplementationDetailDTO {
	implementations := []ImplementationDetailDTO{}
	for _, id := range includedIDs {
		row, ok := meta[id]
		if !ok {
			continue
		}
		impl := ImplementationDetailDTO{
			DomainCapabilityID:   row.CapabilityID,
			DomainCapabilityName: row.CapabilityName,
			BusinessDomainID:     row.BusinessDomainID,
			BusinessDomainName:   row.BusinessDomainName,
			MaturityValue:        row.MaturityValue,
			MaturitySection:      getMaturitySection(row.MaturityValue),
		}
		if gap := target - row.MaturityValue; gap > 0 {
			impl.Gap = gap
		}
		impl.Priority = getPriority(impl.Gap)
		implementations = append(implementations, impl)
	}
	sort.SliceStable(implementations, func(i, j int) bool {
		return implementations[i].MaturityValue < implementations[j].MaturityValue
	})
	return implementations
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
