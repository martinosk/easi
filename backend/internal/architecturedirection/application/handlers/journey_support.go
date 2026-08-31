package handlers

import (
	"context"
	"errors"

	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
)

var ErrTargetPeriodRequiresBoth = errors.New("target period requires both year and quarter, or neither")

type JourneyReferenceChecks struct {
	CapabilityExists              services.CapabilityExists
	ComponentExists               services.ComponentExists
	DomainExists                  services.DomainExists
	CapabilityEffectivelyInDomain services.CapabilityEffectivelyInDomain
}

func buildTargetPeriod(year, quarter *int) (*valueobjects.TargetPeriod, error) {
	if year == nil && quarter == nil {
		return nil, nil
	}
	if year == nil || quarter == nil {
		return nil, ErrTargetPeriodRequiresBoth
	}
	tp, err := valueobjects.NewTargetPeriod(*year, *quarter)
	if err != nil {
		return nil, err
	}
	return &tp, nil
}

type entityRef interface {
	Value() string
}

func requireCapabilityEffectivelyInDomain(
	ctx context.Context,
	check services.CapabilityEffectivelyInDomain,
	parent valueobjects.PhysicalCapabilityRef,
	domain valueobjects.BusinessDomainRef,
) error {
	if check == nil {
		return nil
	}
	ok, err := check(ctx, parent.Value(), domain.Value())
	if err != nil {
		return err
	}
	if !ok {
		return services.ErrTargetParentNotInTargetDomain
	}
	return nil
}

func requireReferenceExists(ctx context.Context, check func(context.Context, string) (bool, error), ref entityRef) error {
	if check == nil {
		return nil
	}
	exists, err := check(ctx, ref.Value())
	if err != nil {
		return err
	}
	if !exists {
		return services.ErrReferencedEntityNotFound
	}
	return nil
}

func verifyComponentsExist(ctx context.Context, check services.ComponentExists, refs []valueobjects.ApplicationRef) error {
	for _, ref := range refs {
		if err := requireReferenceExists(ctx, check, ref); err != nil {
			return err
		}
	}
	return nil
}

func parseApplicationRefs(ids []string) ([]valueobjects.ApplicationRef, error) {
	refs := make([]valueobjects.ApplicationRef, len(ids))
	for i, id := range ids {
		ref, err := valueobjects.NewApplicationRef(id)
		if err != nil {
			return nil, err
		}
		refs[i] = ref
	}
	return refs, nil
}

func parseOptionalBusinessDomainRef(id string) (*valueobjects.BusinessDomainRef, error) {
	return parseOptionalRef(id, valueobjects.NewBusinessDomainRef)
}

func parseOptionalPhysicalCapabilityRef(id string) (*valueobjects.PhysicalCapabilityRef, error) {
	return parseOptionalRef(id, valueobjects.NewPhysicalCapabilityRef)
}

func parseOptionalRef[T any](id string, newRef func(string) (T, error)) (*T, error) {
	if id == "" {
		return nil, nil
	}
	ref, err := newRef(id)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}
