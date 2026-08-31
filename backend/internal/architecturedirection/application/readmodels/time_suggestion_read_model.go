package readmodels

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/infrastructure/database"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
)

type TimeSuggestionDTO struct {
	Grade         *string  `json:"grade"`
	Confidence    string   `json:"confidence"`
	TechnicalGap  *float64 `json:"technicalGap"`
	FunctionalGap *float64 `json:"functionalGap"`
}

type RealizationPair struct {
	CapabilityID string
	ComponentID  string
}

type SuggestedRealization struct {
	Pair           RealizationPair
	CapabilityName string
	ComponentName  string
	Suggestion     TimeSuggestionDTO
}

type TimeSuggestionReadModel struct {
	db             *database.TenantAwareDB
	pillarsGateway mmPL.StrategyPillarsGateway
	calculator     *services.TimeSuggestionCalculator
}

func NewTimeSuggestionReadModel(
	db *database.TenantAwareDB,
	pillarsGateway mmPL.StrategyPillarsGateway,
) *TimeSuggestionReadModel {
	return &TimeSuggestionReadModel{
		db:             db,
		pillarsGateway: pillarsGateway,
		calculator:     services.NewTimeSuggestionCalculator(services.DefaultGapThreshold),
	}
}

type timeSuggestionFilter struct {
	capabilityIDs []string
	componentID   string
}

func (rm *TimeSuggestionReadModel) All(ctx context.Context) ([]SuggestedRealization, error) {
	return rm.suggestions(ctx, timeSuggestionFilter{})
}

func (rm *TimeSuggestionReadModel) ForCapabilities(ctx context.Context, capabilityIDs []string) ([]SuggestedRealization, error) {
	if len(capabilityIDs) == 0 {
		return []SuggestedRealization{}, nil
	}
	return rm.suggestions(ctx, timeSuggestionFilter{capabilityIDs: capabilityIDs})
}

func (rm *TimeSuggestionReadModel) ForPair(ctx context.Context, capabilityID, componentID string) (*TimeSuggestionDTO, error) {
	found, err := rm.suggestions(ctx, timeSuggestionFilter{capabilityIDs: []string{capabilityID}, componentID: componentID})
	if err != nil || len(found) == 0 {
		return nil, err
	}
	suggestion := found[0].Suggestion
	return &suggestion, nil
}

func (rm *TimeSuggestionReadModel) suggestions(ctx context.Context, filter timeSuggestionFilter) ([]SuggestedRealization, error) {
	pillars, err := rm.pillarsGateway.GetStrategyPillars(ctx)
	if err != nil {
		return nil, err
	}

	realizationGaps, err := rm.queryRealizationGaps(ctx, filter)
	if err != nil {
		return nil, err
	}

	return rm.calculateSuggestions(realizationGaps, rm.buildPillarFitTypeMap(pillars)), nil
}

func (rm *TimeSuggestionReadModel) buildPillarFitTypeMap(pillars *mmPL.StrategyPillarsConfigDTO) map[string]string {
	result := make(map[string]string)
	for _, pillar := range pillars.Pillars {
		if pillar.FitType != "" && pillar.FitScoringEnabled {
			result[pillar.ID] = pillar.FitType
		}
	}
	return result
}

type realizationKey struct {
	capabilityID   string
	capabilityName string
	componentID    string
	componentName  string
}

type pillarGap struct {
	pillarID string
	gap      float64
}

type realizationGaps struct {
	key  realizationKey
	gaps []pillarGap
}

const realizationGapsQuery = `
	SELECT
		rc.capability_id,
		cnc.capability_name,
		rc.component_id,
		COALESCE(names.name, '') AS component_name,
		ic.pillar_id,
		ic.effective_importance,
		fs.score
	FROM architecturedirection.realization_cache rc
	JOIN architecturedirection.capability_node_cache cnc
		ON cnc.tenant_id = rc.tenant_id AND cnc.capability_id = rc.capability_id
	JOIN architecturedirection.ea_importance_cache ic
		ON ic.tenant_id = rc.tenant_id
		AND ic.capability_id = rc.capability_id
		AND ic.business_domain_id = cnc.business_domain_id
	JOIN architecturedirection.ea_fit_score_cache fs
		ON fs.tenant_id = rc.tenant_id
		AND fs.component_id = rc.component_id
		AND fs.pillar_id = ic.pillar_id
	LEFT JOIN architecturedirection.reference_name_cache names
		ON names.tenant_id = rc.tenant_id
		AND names.entity_type = 'application'
		AND names.entity_id = rc.component_id
	WHERE rc.tenant_id = $1
		AND ic.effective_importance > 0
		AND fs.score > 0
		AND ($2::text[] IS NULL OR rc.capability_id = ANY($2::text[]))
		AND ($3::text = '' OR rc.component_id = $3::text)
	ORDER BY cnc.capability_name, component_name`

func (rm *TimeSuggestionReadModel) queryRealizationGaps(ctx context.Context, filter timeSuggestionFilter) ([]realizationGaps, error) {
	tenantID, err := tenantOf(ctx)
	if err != nil {
		return nil, err
	}

	ordered := []*realizationGaps{}
	byKey := map[realizationKey]*realizationGaps{}

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, realizationGapsQuery,
			tenantID, capabilityIDArg(filter.capabilityIDs), filter.componentID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			key, gap, scanErr := scanRealizationGap(rows)
			if scanErr != nil {
				return scanErr
			}
			entry, exists := byKey[key]
			if !exists {
				entry = &realizationGaps{key: key}
				byKey[key] = entry
				ordered = append(ordered, entry)
			}
			entry.gaps = append(entry.gaps, gap)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	result := make([]realizationGaps, 0, len(ordered))
	for _, rg := range ordered {
		result = append(result, *rg)
	}
	return result, nil
}

func capabilityIDArg(capabilityIDs []string) any {
	if len(capabilityIDs) == 0 {
		return nil
	}
	return pq.Array(capabilityIDs)
}

func scanRealizationGap(rows *sql.Rows) (realizationKey, pillarGap, error) {
	var key realizationKey
	var pillarID string
	var importance, fitScore int
	err := rows.Scan(&key.capabilityID, &key.capabilityName, &key.componentID, &key.componentName,
		&pillarID, &importance, &fitScore)
	return key, pillarGap{pillarID: pillarID, gap: float64(importance - fitScore)}, err
}

func (rm *TimeSuggestionReadModel) calculateSuggestions(realizations []realizationGaps, pillarFitTypes map[string]string) []SuggestedRealization {
	result := make([]SuggestedRealization, 0, len(realizations))
	for _, rg := range realizations {
		result = append(result, rm.calculateSingleSuggestion(rg, pillarFitTypes))
	}
	return result
}

func (rm *TimeSuggestionReadModel) calculateSingleSuggestion(rg realizationGaps, pillarFitTypes map[string]string) SuggestedRealization {
	technicalGaps, functionalGaps := separateGapsByFitType(rg.gaps, pillarFitTypes)
	calcResult := rm.calculator.Calculate(technicalGaps, functionalGaps)
	return SuggestedRealization{
		Pair:           RealizationPair{CapabilityID: rg.key.capabilityID, ComponentID: rg.key.componentID},
		CapabilityName: rg.key.capabilityName,
		ComponentName:  rg.key.componentName,
		Suggestion:     buildSuggestionDTO(calcResult, technicalGaps, functionalGaps),
	}
}

func separateGapsByFitType(gaps []pillarGap, pillarFitTypes map[string]string) ([]float64, []float64) {
	var technicalGaps, functionalGaps []float64
	for _, pg := range gaps {
		switch pillarFitTypes[pg.pillarID] {
		case "TECHNICAL":
			technicalGaps = append(technicalGaps, pg.gap)
		case "FUNCTIONAL":
			functionalGaps = append(functionalGaps, pg.gap)
		}
	}
	return technicalGaps, functionalGaps
}

func buildSuggestionDTO(calcResult services.TimeSuggestionResult, techGaps, funcGaps []float64) TimeSuggestionDTO {
	dto := TimeSuggestionDTO{Confidence: calcResult.Confidence}
	if calcResult.SuggestedGrade != "" {
		dto.Grade = &calcResult.SuggestedGrade
	}
	if len(techGaps) > 0 {
		dto.TechnicalGap = &calcResult.TechnicalGap
	}
	if len(funcGaps) > 0 {
		dto.FunctionalGap = &calcResult.FunctionalGap
	}
	return dto
}
