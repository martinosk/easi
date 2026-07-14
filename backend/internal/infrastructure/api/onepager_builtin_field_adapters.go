package api

import (
	"context"
	"fmt"
	"time"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaReadModels "easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	metaReadModels "easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/onepagers/application/ports"
)

type builtInFieldAdapter struct {
	fetch    func(ctx context.Context, subjectID string, includedEntryIDs []string) (*ports.SubjectSnapshot, error)
	count    func(ctx context.Context) (int, error)
	fill     func(ctx context.Context, subjectIDs, entryIDs []string) (map[string]map[string]bool, error)
	countVal func(ctx context.Context, entryID string) (int, error)
}

func (a builtInFieldAdapter) FetchSubject(ctx context.Context, subjectID string, includedEntryIDs []string) (*ports.SubjectSnapshot, error) {
	return a.fetch(ctx, subjectID, includedEntryIDs)
}

func (a builtInFieldAdapter) CountSubjects(ctx context.Context) (int, error) {
	return a.count(ctx)
}

func (a builtInFieldAdapter) FilledBuiltInFields(ctx context.Context, subjectIDs, entryIDs []string) (map[string]map[string]bool, error) {
	return a.fill(ctx, subjectIDs, entryIDs)
}

func (a builtInFieldAdapter) CountSubjectsWithBuiltInValue(ctx context.Context, entryID string) (int, error) {
	return a.countVal(ctx, entryID)
}

type relationBinding[T any] struct {
	entryID string
	resolve func(context.Context, *T) (ports.ReferenceListValue, error)
}

type builtInSourceConfig[T any] struct {
	subjectType   string
	getByID       func(context.Context, string) (*T, error)
	getByIDs      func(context.Context, []string) ([]T, error)
	getAll        func(context.Context) ([]T, error)
	toSnapshot    func(*T) *ports.SubjectSnapshot
	idOf          func(*T) string
	countSubjects func(context.Context) (int, error)
	relations     []relationBinding[T]
}

func builtInFieldSource[T any](cfg builtInSourceConfig[T]) ports.BuiltInFieldSource {
	return builtInFieldAdapter{
		fetch:    cfg.fetchSubject,
		count:    cfg.countAllSubjects,
		fill:     cfg.filledBuiltInFields,
		countVal: cfg.countSubjectsWithValue,
	}
}

func (cfg builtInSourceConfig[T]) fetchSubject(ctx context.Context, subjectID string, includedEntryIDs []string) (*ports.SubjectSnapshot, error) {
	dto, err := cfg.getByID(ctx, subjectID)
	if err != nil {
		return nil, fmt.Errorf("fetch %s %s for one-pager: %w", cfg.subjectType, subjectID, err)
	}
	snapshot := cfg.toSnapshot(dto)
	if snapshot == nil {
		return nil, nil
	}
	if err := cfg.resolveRelations(ctx, dto, snapshot, includedEntryIDs); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (cfg builtInSourceConfig[T]) resolveRelations(ctx context.Context, dto *T, snapshot *ports.SubjectSnapshot, includedEntryIDs []string) error {
	if len(cfg.relations) == 0 {
		return nil
	}
	included := make(map[string]struct{}, len(includedEntryIDs))
	for _, entryID := range includedEntryIDs {
		included[entryID] = struct{}{}
	}
	for _, binding := range cfg.relations {
		if _, ok := included[binding.entryID]; !ok {
			continue
		}
		value, err := binding.resolve(ctx, dto)
		if err != nil {
			return fmt.Errorf("resolve %s relation %s for one-pager: %w", cfg.subjectType, binding.entryID, err)
		}
		snapshot.Fields[binding.entryID] = value
	}
	return nil
}

func (cfg builtInSourceConfig[T]) countAllSubjects(ctx context.Context) (int, error) {
	count, err := cfg.countSubjects(ctx)
	if err != nil {
		return 0, fmt.Errorf("count %s subjects for one-pager: %w", cfg.subjectType, err)
	}
	return count, nil
}

func (cfg builtInSourceConfig[T]) filledBuiltInFields(ctx context.Context, subjectIDs, entryIDs []string) (map[string]map[string]bool, error) {
	if len(subjectIDs) == 0 || len(entryIDs) == 0 {
		return map[string]map[string]bool{}, nil
	}
	dtos, err := cfg.getByIDs(ctx, subjectIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch %s subjects for one-pager completeness: %w", cfg.subjectType, err)
	}
	filled := make(map[string]map[string]bool, len(dtos))
	for i := range dtos {
		snapshot := cfg.toSnapshot(&dtos[i])
		filled[cfg.idOf(&dtos[i])] = filledEntries(snapshot, entryIDs)
	}
	return filled, nil
}

func (cfg builtInSourceConfig[T]) countSubjectsWithValue(ctx context.Context, entryID string) (int, error) {
	dtos, err := cfg.getAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetch %s subjects for one-pager impact preview: %w", cfg.subjectType, err)
	}
	count := 0
	for i := range dtos {
		if ports.ValueFilled(snapshotField(cfg.toSnapshot(&dtos[i]), entryID)) {
			count++
		}
	}
	return count, nil
}

func filledEntries(snapshot *ports.SubjectSnapshot, entryIDs []string) map[string]bool {
	filled := make(map[string]bool, len(entryIDs))
	for _, entryID := range entryIDs {
		filled[entryID] = ports.ValueFilled(snapshotField(snapshot, entryID))
	}
	return filled
}

func snapshotField(snapshot *ports.SubjectSnapshot, entryID string) ports.BuiltInFieldValue {
	if snapshot == nil {
		return nil
	}
	return snapshot.Fields[entryID]
}

func newOnePagerBuiltInFieldSources(db *database.TenantAwareDB) map[string]ports.BuiltInFieldSource {
	capabilities := capReadModels.NewCapabilityReadModel(db)
	enterpriseCapabilities := eaReadModels.NewEnterpriseCapabilityReadModel(db)
	applications := archReadModels.NewApplicationComponentReadModel(db)
	acquiredEntities := archReadModels.NewAcquiredEntityReadModel(db)
	vendors := archReadModels.NewVendorReadModel(db)
	internalTeams := archReadModels.NewInternalTeamReadModel(db)
	relations := newOnePagerRelationModels(db)

	return map[string]ports.BuiltInFieldSource{
		"capability": builtInFieldSource(builtInSourceConfig[capReadModels.CapabilityDTO]{
			subjectType: "capability", getByID: capabilities.GetByID, getByIDs: capabilities.GetByIDs, getAll: capabilities.GetAll,
			toSnapshot: capabilitySnapshot, idOf: capabilityID, countSubjects: capabilities.Count,
			relations: relations.capabilityRelations(),
		}),
		"enterprise-capability": builtInFieldSource(builtInSourceConfig[eaReadModels.EnterpriseCapabilityDTO]{
			subjectType: "enterprise capability", getByID: enterpriseCapabilities.GetByID, getByIDs: enterpriseCapabilities.GetByIDs, getAll: enterpriseCapabilities.GetAll,
			toSnapshot: enterpriseCapabilitySnapshot, idOf: enterpriseCapabilityID, countSubjects: enterpriseCapabilities.Count,
			relations: relations.enterpriseCapabilityRelations(),
		}),
		"application": builtInFieldSource(builtInSourceConfig[archReadModels.ApplicationComponentDTO]{
			subjectType: "application", getByID: applications.GetByID, getByIDs: applications.GetByIDs, getAll: applications.GetAll,
			toSnapshot: applicationSnapshot, idOf: applicationID, countSubjects: applications.Count,
			relations: relations.applicationRelations(),
		}),
		"acquired-entity": builtInFieldSource(builtInSourceConfig[archReadModels.AcquiredEntityDTO]{
			subjectType: "acquired entity", getByID: acquiredEntities.GetByID, getByIDs: acquiredEntities.GetByIDs, getAll: acquiredEntities.GetAll,
			toSnapshot: acquiredEntitySnapshot, idOf: acquiredEntityID, countSubjects: acquiredEntities.Count,
			relations: relations.acquiredEntityRelations(),
		}),
		"vendor": builtInFieldSource(builtInSourceConfig[archReadModels.VendorDTO]{
			subjectType: "vendor", getByID: vendors.GetByID, getByIDs: vendors.GetByIDs, getAll: vendors.GetAll,
			toSnapshot: vendorSnapshot, idOf: vendorID, countSubjects: vendors.Count,
			relations: relations.vendorRelations(),
		}),
		"internal-team": builtInFieldSource(builtInSourceConfig[archReadModels.InternalTeamDTO]{
			subjectType: "internal team", getByID: internalTeams.GetByID, getByIDs: internalTeams.GetByIDs, getAll: internalTeams.GetAll,
			toSnapshot: internalTeamSnapshot, idOf: internalTeamID, countSubjects: internalTeams.Count,
			relations: relations.internalTeamRelations(),
		}),
	}
}

func capabilityID(dto *capReadModels.CapabilityDTO) string                    { return dto.ID }
func enterpriseCapabilityID(dto *eaReadModels.EnterpriseCapabilityDTO) string { return dto.ID }
func applicationID(dto *archReadModels.ApplicationComponentDTO) string        { return dto.ID }
func acquiredEntityID(dto *archReadModels.AcquiredEntityDTO) string           { return dto.ID }
func vendorID(dto *archReadModels.VendorDTO) string                           { return dto.ID }
func internalTeamID(dto *archReadModels.InternalTeamDTO) string               { return dto.ID }

func capabilitySnapshot(dto *capReadModels.CapabilityDTO) *ports.SubjectSnapshot {
	if dto == nil {
		return nil
	}
	return buildSnapshot(dto.Name,
		namedField{"description", textOrNil(dto.Description)},
		namedField{"maturity", ports.MaturityValue{Value: dto.MaturityValue}},
		namedField{"experts", expertsValue(dto.Experts, capabilityExpert)},
	)
}

func enterpriseCapabilitySnapshot(dto *eaReadModels.EnterpriseCapabilityDTO) *ports.SubjectSnapshot {
	if dto == nil {
		return nil
	}
	return buildSnapshot(dto.Name,
		namedField{"description", textOrNil(dto.Description)},
		namedField{"category", textOrNil(dto.Category)},
	)
}

func applicationSnapshot(dto *archReadModels.ApplicationComponentDTO) *ports.SubjectSnapshot {
	if dto == nil {
		return nil
	}
	return buildSnapshot(dto.Name,
		namedField{"description", textOrNil(dto.Description)},
		namedField{"experts", expertsValue(dto.Experts, applicationExpert)},
	)
}

func acquiredEntitySnapshot(dto *archReadModels.AcquiredEntityDTO) *ports.SubjectSnapshot {
	if dto == nil {
		return nil
	}
	return buildSnapshot(dto.Name,
		namedField{"acquisition-date", dateOrNil(dto.AcquisitionDate)},
		namedField{"integration-status", textOrNil(dto.IntegrationStatus)},
	)
}

func vendorSnapshot(dto *archReadModels.VendorDTO) *ports.SubjectSnapshot {
	if dto == nil {
		return nil
	}
	return buildSnapshot(dto.Name,
		namedField{"implementation-partner", textOrNil(dto.ImplementationPartner)},
		namedField{"notes", textOrNil(dto.Notes)},
	)
}

func internalTeamSnapshot(dto *archReadModels.InternalTeamDTO) *ports.SubjectSnapshot {
	if dto == nil {
		return nil
	}
	return buildSnapshot(dto.Name,
		namedField{"department", textOrNil(dto.Department)},
		namedField{"contact-person", textOrNil(dto.ContactPerson)},
	)
}

type namedField struct {
	key   string
	value ports.BuiltInFieldValue
}

func buildSnapshot(name string, extraFields ...namedField) *ports.SubjectSnapshot {
	fields := map[string]ports.BuiltInFieldValue{"name": ports.TextValue{Text: name}}
	for _, field := range extraFields {
		fields[field.key] = field.value
	}
	return &ports.SubjectSnapshot{Name: name, Fields: fields}
}

func textOrNil(text string) ports.BuiltInFieldValue {
	if text == "" {
		return nil
	}
	return ports.TextValue{Text: text}
}

func dateOrNil(date *time.Time) ports.BuiltInFieldValue {
	if date == nil {
		return nil
	}
	return ports.DateValue{Date: *date}
}

func capabilityExpert(expert capReadModels.ExpertDTO) ports.Expert {
	return ports.Expert{Name: expert.Name, Role: expert.Role, Contact: expert.Contact}
}

func applicationExpert(expert archReadModels.ExpertDTO) ports.Expert {
	return ports.Expert{Name: expert.Name, Role: expert.Role, Contact: expert.Contact}
}

func expertsValue[T any](experts []T, toExpert func(T) ports.Expert) ports.BuiltInFieldValue {
	if len(experts) == 0 {
		return nil
	}
	mapped := make([]ports.Expert, len(experts))
	for i, expert := range experts {
		mapped[i] = toExpert(expert)
	}
	return ports.ExpertsValue{Experts: mapped}
}

type onePagerMaturityScaleAdapter struct {
	configurations *metaReadModels.MetaModelConfigurationReadModel
}

func newOnePagerMaturityScaleAdapter(db *database.TenantAwareDB) ports.MaturityScaleSource {
	return onePagerMaturityScaleAdapter{configurations: metaReadModels.NewMetaModelConfigurationReadModel(db)}
}

func (a onePagerMaturityScaleAdapter) Sections(ctx context.Context) ([]ports.MaturitySection, error) {
	config, err := a.configurations.GetByTenantID(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch metamodel configuration for one-pager maturity scale: %w", err)
	}
	return maturitySections(config), nil
}

func maturitySections(config *metaReadModels.MetaModelConfigurationDTO) []ports.MaturitySection {
	if config == nil {
		return nil
	}
	sections := make([]ports.MaturitySection, len(config.Sections))
	for i, section := range config.Sections {
		sections[i] = ports.MaturitySection{Name: section.Name, MinValue: section.MinValue, MaxValue: section.MaxValue}
	}
	return sections
}
