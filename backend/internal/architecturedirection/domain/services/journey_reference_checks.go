package services

import (
	"context"
	"errors"
)

var (
	ErrReferencedEntityNotFound      = errors.New("a referenced entity does not exist or is not accessible in this tenant")
	ErrTargetParentNotInTargetDomain = errors.New("the target parent capability does not effectively belong to the target business domain")
)

type CapabilityExists func(ctx context.Context, capabilityID string) (bool, error)

type ComponentExists func(ctx context.Context, componentID string) (bool, error)

type DomainExists func(ctx context.Context, domainID string) (bool, error)

type CapabilityEffectivelyInDomain func(ctx context.Context, capabilityID, domainID string) (bool, error)
