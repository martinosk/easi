package adapters

import (
	"context"
)

// ExistenceCheck reports whether an upstream entity with the given ID is
// visible to the caller's tenant. Implementations are constructed in the
// wiring layer from the upstream read models so this package stays free of
// cross-context imports.
type ExistenceCheck func(ctx context.Context, id string) (bool, error)

type ReferenceCheckerAdapter struct {
	enterpriseCapability ExistenceCheck
	physicalCapability   ExistenceCheck
	businessDomain       ExistenceCheck
}

func NewReferenceCheckerAdapter(enterpriseCapability, physicalCapability, businessDomain ExistenceCheck) *ReferenceCheckerAdapter {
	return &ReferenceCheckerAdapter{
		enterpriseCapability: enterpriseCapability,
		physicalCapability:   physicalCapability,
		businessDomain:       businessDomain,
	}
}

func (a *ReferenceCheckerAdapter) EnterpriseCapabilityExists(ctx context.Context, id string) (bool, error) {
	return a.enterpriseCapability(ctx, id)
}

func (a *ReferenceCheckerAdapter) PhysicalCapabilityExists(ctx context.Context, id string) (bool, error) {
	return a.physicalCapability(ctx, id)
}

func (a *ReferenceCheckerAdapter) BusinessDomainExists(ctx context.Context, id string) (bool, error) {
	return a.businessDomain(ctx, id)
}
