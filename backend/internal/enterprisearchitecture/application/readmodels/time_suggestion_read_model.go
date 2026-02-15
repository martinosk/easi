package readmodels

import (
	"context"
	"database/sql"

	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/domain/services"
	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
)

type TimeSuggestionDTO struct {
	CapabilityID   string   `json:"capabilityId"`
	CapabilityName string   `json:"capabilityName"`
	ComponentID    string   `json:"componentId"`
	ComponentName  string   `json:"componentName"`
	SuggestedTime  *string  `json:"suggestedTime"`
	TechnicalGap   *float64 `json:"technicalGap"`
	FunctionalGap  *float64 `json:"functionalGap"`
	Confidence     string   `json:"confidence"`
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
	capabilityID string
	componentID  string
}

func (rm *TimeSuggestionReadModel) GetAllSuggestions(ctx context.Context) ([]TimeSuggestionDTO, error) {
	return rm.getSuggestions(ctx, timeSuggestionFilter{})
}

func (rm *TimeSuggestionReadModel) GetByCapability(ctx context.Context, capabilityID string) ([]TimeSuggestionDTO, error) {
	return rm.getSuggestions(ctx, timeSuggestionFilter{capabilityID: capabilityID})
}

func (rm *TimeSuggestionReadModel) GetByComponent(ctx context.Context, componentID string) ([]TimeSuggestionDTO, error) {
	return rm.getSuggestions(ctx, timeSuggestionFilter{componentID: componentID})
}

func (rm *TimeSuggestionReadModel) getSuggestions(ctx context.Context, filter timeSuggestionFilter) ([]TimeSuggestionDTO, error) {
	pillars, err := rm.pillarsGateway.GetStrategyPillars(ctx)
	if err != nil {
		return nil, err
	}

	pillarFitTypes := rm.buildPillarFitTypeMap(pillars)

	realizationGaps, err := rm.queryRealizationGaps(ctx, filter.capabilityID, filter.componentID)
	if err != nil {
		return nil, err
	}

	return rm.calculateSuggestions(realizationGaps, pillarFitTypes), nil
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

func (rm *TimeSuggestionReadModel) queryRealizationGaps(ctx context.Context, capabilityID, componentID string) ([]realizationGaps, error) {
	tenantID, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return nil, err
	}

	realizationsMap := make(map[realizationKey]*realizationGaps)

	err = rm.db.WithReadOnlyTx(ctx, func(tx *sql.Tx) error {
		query := rm.buildGapsQuery(capabilityID, componentID)
		args := rm.buildQueryArgs(tenantID.Value(), capabilityID, componentID)

		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var capID, capName, compID, compName, pillarID string
			var importance, fitScore int

			if err := rows.Scan(&capID, &capName, &compID, &compName, &pillarID, &importance, &fitScore); err != nil {
				return err
			}

			key := realizationKey{
				capabilityID:   capID,
				capabilityName: capName,
				componentID:    compID,
				componentName:  compName,
			}

			if _, exists := realizationsMap[key]; !exists {
				realizationsMap[key] = &realizationGaps{key: key, gaps: []pillarGap{}}
			}

			gap := float64(importance - fitScore)
			realizationsMap[key].gaps = append(realizationsMap[key].gaps, pillarGap{pillarID: pillarID, gap: gap})
		}
		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	result := make([]realizationGaps, 0, len(realizationsMap))
	for _, rg := range realizationsMap {
		result = append(result, *rg)
	}
	return result, nil
}

func (rm *TimeSuggestionReadModel) buildGapsQuery(capabilityID, componentID string) string {
	baseQuery := `
		SELECT
			rc.capability_id,
			dcm.capability_name,
			rc.component_id,
			rc.component_name,
			ic.pillar_id,
			ic.effective_importance as importance,
			fs.score as fit_score
		FROM ea_realization_cache rc
		JOIN domain_capability_metadata dcm ON dcm.capability_id = rc.capability_id AND dcm.tenant_id = rc.tenant_id
		JOIN ea_importance_cache ic ON ic.capability_id = rc.capability_id
			AND ic.tenant_id = rc.tenant_id
			AND ic.business_domain_id = dcm.business_domain_id
		JOIN ea_fit_score_cache fs ON fs.component_id = rc.component_id
			AND fs.tenant_id = rc.tenant_id
			AND fs.pillar_id = ic.pillar_id
		WHERE rc.tenant_id = $1
			AND rc.origin = 'Direct'
			AND ic.effective_importance > 0
			AND fs.score > 0
	`

	argIndex := 2
	if capabilityID != "" {
		baseQuery += " AND rc.capability_id = $" + string(rune('0'+argIndex))
		argIndex++
	}
	if componentID != "" {
		baseQuery += " AND rc.component_id = $" + string(rune('0'+argIndex))
	}

	return baseQuery + " ORDER BY dcm.capability_name, rc.component_name"
}

func (rm *TimeSuggestionReadModel) buildQueryArgs(tenantID, capabilityID, componentID string) []any {
	args := []any{tenantID}
	if capabilityID != "" {
		args = append(args, capabilityID)
	}
	if componentID != "" {
		args = append(args, componentID)
	}
	return args
}

func (rm *TimeSuggestionReadModel) calculateSuggestions(realizations []realizationGaps, pillarFitTypes map[string]string) []TimeSuggestionDTO {
	result := make([]TimeSuggestionDTO, 0, len(realizations))
	for _, rg := range realizations {
		result = append(result, rm.calculateSingleSuggestion(rg, pillarFitTypes))
	}
	return result
}

func (rm *TimeSuggestionReadModel) calculateSingleSuggestion(rg realizationGaps, pillarFitTypes map[string]string) TimeSuggestionDTO {
	technicalGaps, functionalGaps := rm.separateGapsByFitType(rg.gaps, pillarFitTypes)
	calcResult := rm.calculator.Calculate(technicalGaps, functionalGaps)
	return rm.buildSuggestionDTO(rg.key, calcResult, technicalGaps, functionalGaps)
}

func (rm *TimeSuggestionReadModel) separateGapsByFitType(gaps []pillarGap, pillarFitTypes map[string]string) ([]float64, []float64) {
	var technicalGaps, functionalGaps []float64
	for _, pg := range gaps {
		fitType := pillarFitTypes[pg.pillarID]
		switch fitType {
		case "TECHNICAL":
			technicalGaps = append(technicalGaps, pg.gap)
		case "FUNCTIONAL":
			functionalGaps = append(functionalGaps, pg.gap)
		}
	}
	return technicalGaps, functionalGaps
}

func (rm *TimeSuggestionReadModel) buildSuggestionDTO(key realizationKey, calcResult services.TimeSuggestionResult, techGaps, funcGaps []float64) TimeSuggestionDTO {
	dto := TimeSuggestionDTO{
		CapabilityID:   key.capabilityID,
		CapabilityName: key.capabilityName,
		ComponentID:    key.componentID,
		ComponentName:  key.componentName,
		Confidence:     calcResult.Confidence,
	}
	if calcResult.SuggestedTime != "" {
		dto.SuggestedTime = &calcResult.SuggestedTime
	}
	if len(techGaps) > 0 {
		dto.TechnicalGap = &calcResult.TechnicalGap
	}
	if len(funcGaps) > 0 {
		dto.FunctionalGap = &calcResult.FunctionalGap
	}
	return dto
}
