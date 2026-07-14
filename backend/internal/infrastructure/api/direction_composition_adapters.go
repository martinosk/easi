package api

import (
	"context"

	adReadModels "easi/backend/internal/architecturedirection/application/readmodels"
	directionServices "easi/backend/internal/architecturedirection/domain/services"
	directionAPI "easi/backend/internal/architecturedirection/infrastructure/api"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaProjectors "easi/backend/internal/enterprisearchitecture/application/projectors"
	eaReadModels "easi/backend/internal/enterprisearchitecture/application/readmodels"
	eaServices "easi/backend/internal/enterprisearchitecture/application/services"
	eaDomainServices "easi/backend/internal/enterprisearchitecture/domain/services"
)

type directionSourcesAdapter struct {
	readModel *adReadModels.DirectionReadModel
}

func (a directionSourcesAdapter) ActiveDirectionSources(ctx context.Context) ([]eaServices.DirectionSources, error) {
	items, err := a.readModel.ActiveDirectionSourcesAcrossECs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]eaServices.DirectionSources, len(items))
	for i, item := range items {
		out[i] = eaServices.DirectionSources{
			EnterpriseCapabilityID: item.EnterpriseCapabilityID,
			Status:                 item.Status,
			SourceCapabilityIDs:    item.SourceCapabilityIDs,
		}
	}
	return out, nil
}

type directionEligibilityAdapter struct {
	service *eaServices.CompositionService
}

func (a directionEligibilityAdapter) FirstSourceConflict(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) (*directionServices.SourceConflict, error) {
	conflict, err := a.service.FirstSourceConflict(ctx, enterpriseCapabilityID, sourceCapabilityIDs)
	if err != nil {
		return nil, err
	}
	if conflict == nil {
		return nil, nil
	}
	return &directionServices.SourceConflict{
		CapabilityID:             conflict.CapabilityID,
		CapabilityName:           conflict.CapabilityName,
		EnterpriseCapabilityID:   conflict.ConflictingEnterpriseCapabilityID,
		EnterpriseCapabilityName: conflict.ConflictingEnterpriseCapabilityName,
	}, nil
}

type compositionPreviewAdapter struct {
	service      *eaServices.CompositionService
	capabilities *eaReadModels.EnterpriseCapabilityReadModel
}

func (a compositionPreviewAdapter) PreviewComposition(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) (*directionAPI.CompositionPreviewData, error) {
	capability, err := a.capabilities.GetByID(ctx, enterpriseCapabilityID)
	if err != nil {
		return nil, err
	}
	if capability == nil {
		return nil, nil
	}
	preview, err := a.service.Preview(ctx, enterpriseCapabilityID, sourceCapabilityIDs)
	if err != nil {
		return nil, err
	}
	return &directionAPI.CompositionPreviewData{
		IncludedCapabilities: previewItems(preview.Resolved),
		SourceEligibility:    previewEligibility(preview.SourceEligibility),
		Meta: directionAPI.CompositionPreviewMetaDTO{
			SourceCount:    preview.Counts.SourceCount,
			IncludedCount:  preview.Counts.IncludedCount,
			CarvedOutCount: preview.Counts.CarvedOutCount,
		},
	}, nil
}

func previewItems(resolved []eaDomainServices.ResolvedCapability) []directionAPI.PreviewIncludedCapabilityDTO {
	items := make([]directionAPI.PreviewIncludedCapabilityDTO, len(resolved))
	for i, r := range resolved {
		items[i] = directionAPI.PreviewIncludedCapabilityDTO{
			CapabilityID:       r.Node.ID,
			Name:               r.Node.Name,
			Level:              r.Node.Level,
			BusinessDomainID:   optionalString(r.Node.BusinessDomainID),
			BusinessDomainName: optionalString(r.Node.BusinessDomainName),
			Role:               string(r.Role),
		}
		if r.CarvedOutBy != nil {
			items[i].CarvedOutBy = &directionAPI.CarvedOutByDTO{
				EnterpriseCapabilityID:   r.CarvedOutBy.EnterpriseCapabilityID,
				EnterpriseCapabilityName: r.CarvedOutBy.EnterpriseCapabilityName,
			}
		}
	}
	return items
}

func previewEligibility(eligibility []eaServices.SourceEligibility) []directionAPI.SourceEligibilityDTO {
	out := make([]directionAPI.SourceEligibilityDTO, len(eligibility))
	for i, e := range eligibility {
		out[i] = directionAPI.SourceEligibilityDTO{
			CapabilityID:        e.CapabilityID,
			Eligible:            e.Eligible,
			IneligibilityReason: e.IneligibilityReason,
		}
		if e.ConflictingEnterpriseCapability != nil {
			out[i].ConflictingEnterpriseCapability = &directionAPI.ConflictingECDTO{
				ID:   e.ConflictingEnterpriseCapability.EnterpriseCapabilityID,
				Name: e.ConflictingEnterpriseCapability.EnterpriseCapabilityName,
			}
		}
	}
	return out
}

func businessDomainNameLookup(readModel *capReadModels.BusinessDomainReadModel) eaProjectors.BusinessDomainNameLookup {
	return func(ctx context.Context, businessDomainID string) (string, error) {
		domain, err := readModel.GetByID(ctx, businessDomainID)
		if err != nil {
			return "", err
		}
		if domain == nil {
			return "", nil
		}
		return domain.Name, nil
	}
}

func enterpriseCapabilityIsActive(readModel *eaReadModels.EnterpriseCapabilityReadModel) directionServices.ExistenceCheck {
	return func(ctx context.Context, id string) (bool, error) {
		capability, err := readModel.GetByID(ctx, id)
		if err != nil {
			return false, err
		}
		return capability != nil && capability.Active, nil
	}
}

func capabilityEffectivelyInDomain(readModel *capReadModels.CMEffectiveBusinessDomainReadModel) directionServices.CapabilityEffectivelyInDomain {
	return func(ctx context.Context, capabilityID, domainID string) (bool, error) {
		effective, err := readModel.GetByCapabilityID(ctx, capabilityID)
		if err != nil {
			return false, err
		}
		if effective == nil {
			return false, nil
		}
		return effective.BusinessDomainID == domainID, nil
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
