package api

import (
	"context"

	"easi/backend/internal/architecturedirection/application/readmodels"
)

type capabilityNodeLookup interface {
	GetByID(ctx context.Context, capabilityID string) (*readmodels.CapabilityNodeDTO, error)
}

type referenceExistenceLookup interface {
	Exists(ctx context.Context, entity readmodels.ReferenceEntity, entityID string) (bool, error)
}

type directRealizationSource interface {
	DirectRealizationID(ctx context.Context, capabilityID readmodels.CapabilityID, componentID readmodels.ComponentID) (readmodels.RealizationID, bool, error)
}

type referenceLookups struct {
	capabilityExists              func(ctx context.Context, capabilityID string) (bool, error)
	componentExists               func(ctx context.Context, componentID string) (bool, error)
	domainExists                  func(ctx context.Context, domainID string) (bool, error)
	capabilityEffectivelyInDomain func(ctx context.Context, capabilityID, domainID string) (bool, error)
	directRealization             func(ctx context.Context, capabilityID, componentID string) (string, bool, error)
	capabilityMaturity            func(ctx context.Context, capabilityID string) (int, error)
}

func newReferenceLookups(nodes capabilityNodeLookup, references referenceExistenceLookup, realizations directRealizationSource) referenceLookups {
	return referenceLookups{
		capabilityExists: func(ctx context.Context, capabilityID string) (bool, error) {
			node, err := nodes.GetByID(ctx, capabilityID)
			return node != nil, err
		},
		componentExists: existsInReferenceCache(references, readmodels.ReferenceEntityApplication),
		domainExists:    existsInReferenceCache(references, readmodels.ReferenceEntityBusinessDomain),
		capabilityEffectivelyInDomain: func(ctx context.Context, capabilityID, domainID string) (bool, error) {
			node, err := nodes.GetByID(ctx, capabilityID)
			if err != nil || node == nil {
				return false, err
			}
			return node.BusinessDomainID == domainID, nil
		},
		directRealization: func(ctx context.Context, capabilityID, componentID string) (string, bool, error) {
			realizationID, found, err := realizations.DirectRealizationID(ctx, readmodels.CapabilityID(capabilityID), readmodels.ComponentID(componentID))
			return string(realizationID), found, err
		},
		capabilityMaturity: func(ctx context.Context, capabilityID string) (int, error) {
			node, err := nodes.GetByID(ctx, capabilityID)
			if err != nil || node == nil {
				return 0, err
			}
			return node.MaturityValue, nil
		},
	}
}

func existsInReferenceCache(references referenceExistenceLookup, entity readmodels.ReferenceEntity) func(context.Context, string) (bool, error) {
	return func(ctx context.Context, entityID string) (bool, error) {
		return references.Exists(ctx, entity, entityID)
	}
}
