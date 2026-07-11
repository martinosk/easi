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

type builtInFieldFetch func(ctx context.Context, subjectID string) (*ports.SubjectSnapshot, error)

func (f builtInFieldFetch) FetchSubject(ctx context.Context, subjectID string) (*ports.SubjectSnapshot, error) {
	return f(ctx, subjectID)
}

func builtInFieldSource[T any](subjectType string, getByID func(context.Context, string) (*T, error), toSnapshot func(*T) *ports.SubjectSnapshot) ports.BuiltInFieldSource {
	return builtInFieldFetch(func(ctx context.Context, subjectID string) (*ports.SubjectSnapshot, error) {
		dto, err := getByID(ctx, subjectID)
		if err != nil {
			return nil, fmt.Errorf("fetch %s %s for one-pager: %w", subjectType, subjectID, err)
		}
		return toSnapshot(dto), nil
	})
}

func newOnePagerBuiltInFieldSources(db *database.TenantAwareDB) map[string]ports.BuiltInFieldSource {
	return map[string]ports.BuiltInFieldSource{
		"capability":            builtInFieldSource("capability", capReadModels.NewCapabilityReadModel(db).GetByID, capabilitySnapshot),
		"enterprise-capability": builtInFieldSource("enterprise capability", eaReadModels.NewEnterpriseCapabilityReadModel(db).GetByID, enterpriseCapabilitySnapshot),
		"application":           builtInFieldSource("application", archReadModels.NewApplicationComponentReadModel(db).GetByID, applicationSnapshot),
		"acquired-entity":       builtInFieldSource("acquired entity", archReadModels.NewAcquiredEntityReadModel(db).GetByID, acquiredEntitySnapshot),
		"vendor":                builtInFieldSource("vendor", archReadModels.NewVendorReadModel(db).GetByID, vendorSnapshot),
		"internal-team":         builtInFieldSource("internal team", archReadModels.NewInternalTeamReadModel(db).GetByID, internalTeamSnapshot),
	}
}

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
