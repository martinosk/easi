package services

import "context"

type CapabilityExists func(ctx context.Context, capabilityID string) (bool, error)

type ComponentExists func(ctx context.Context, componentID string) (bool, error)

type DomainExists func(ctx context.Context, domainID string) (bool, error)

type CapabilityEffectivelyInDomain func(ctx context.Context, capabilityID, domainID string) (bool, error)
