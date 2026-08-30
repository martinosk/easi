package api

import (
	"context"

	"easi/backend/internal/architecturedirection/application/readmodels"
	appservices "easi/backend/internal/architecturedirection/application/services"
	domainservices "easi/backend/internal/architecturedirection/domain/services"
)

type EnterpriseCapabilityLookup interface {
	GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityCacheDTO, error)
}

type directionSourcesProvider struct {
	readModel *readmodels.DirectionReadModel
}

func (p directionSourcesProvider) ActiveDirectionSources(ctx context.Context) ([]appservices.DirectionSources, error) {
	items, err := p.readModel.ActiveDirectionSourcesAcrossECs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]appservices.DirectionSources, len(items))
	for i, item := range items {
		out[i] = appservices.DirectionSources{
			EnterpriseCapabilityID: item.EnterpriseCapabilityID,
			Status:                 item.Status,
			SourceCapabilityIDs:    item.SourceCapabilityIDs,
		}
	}
	return out, nil
}

type sourceEligibilityService struct {
	composition *appservices.CompositionService
}

func (s sourceEligibilityService) FirstSourceConflict(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) (*domainservices.SourceConflict, error) {
	conflict, err := s.composition.FirstSourceConflict(ctx, enterpriseCapabilityID, sourceCapabilityIDs)
	if err != nil {
		return nil, err
	}
	if conflict == nil {
		return nil, nil
	}
	return &domainservices.SourceConflict{
		CapabilityID:             conflict.CapabilityID,
		CapabilityName:           conflict.CapabilityName,
		EnterpriseCapabilityID:   conflict.ConflictingEnterpriseCapabilityID,
		EnterpriseCapabilityName: conflict.ConflictingEnterpriseCapabilityName,
	}, nil
}

type compositionPreviewService struct {
	composition  *appservices.CompositionService
	capabilities EnterpriseCapabilityLookup
}

func (s compositionPreviewService) PreviewComposition(ctx context.Context, enterpriseCapabilityID string, sourceCapabilityIDs []string) (*CompositionPreviewData, error) {
	capability, err := s.capabilities.GetByID(ctx, enterpriseCapabilityID)
	if err != nil {
		return nil, err
	}
	if capability == nil {
		return nil, nil
	}
	preview, err := s.composition.Preview(ctx, enterpriseCapabilityID, sourceCapabilityIDs)
	if err != nil {
		return nil, err
	}
	return &CompositionPreviewData{
		IncludedCapabilities: previewItems(preview.Resolved),
		SourceEligibility:    previewEligibility(preview.SourceEligibility),
		Meta: CompositionPreviewMetaDTO{
			SourceCount:    preview.Counts.SourceCount,
			IncludedCount:  preview.Counts.IncludedCount,
			CarvedOutCount: preview.Counts.CarvedOutCount,
		},
	}, nil
}

func previewItems(resolved []domainservices.ResolvedCapability) []PreviewIncludedCapabilityDTO {
	items := make([]PreviewIncludedCapabilityDTO, len(resolved))
	for i, r := range resolved {
		items[i] = PreviewIncludedCapabilityDTO{
			CapabilityID:       r.Node.ID,
			Name:               r.Node.Name,
			Level:              r.Node.Level,
			BusinessDomainID:   optionalString(r.Node.BusinessDomainID),
			BusinessDomainName: optionalString(r.Node.BusinessDomainName),
			Role:               string(r.Role),
		}
		if r.CarvedOutBy != nil {
			items[i].CarvedOutBy = &CarvedOutByDTO{
				EnterpriseCapabilityID:   r.CarvedOutBy.EnterpriseCapabilityID,
				EnterpriseCapabilityName: r.CarvedOutBy.EnterpriseCapabilityName,
			}
		}
	}
	return items
}

func previewEligibility(eligibility []appservices.SourceEligibility) []SourceEligibilityDTO {
	out := make([]SourceEligibilityDTO, len(eligibility))
	for i, e := range eligibility {
		out[i] = SourceEligibilityDTO{
			CapabilityID:        e.CapabilityID,
			Eligible:            e.Eligible,
			IneligibilityReason: e.IneligibilityReason,
		}
		if e.ConflictingEnterpriseCapability != nil {
			out[i].ConflictingEnterpriseCapability = &ConflictingECDTO{
				ID:   e.ConflictingEnterpriseCapability.EnterpriseCapabilityID,
				Name: e.ConflictingEnterpriseCapability.EnterpriseCapabilityName,
			}
		}
	}
	return out
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
