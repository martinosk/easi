package projectors

import (
	"context"
	"fmt"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const retiredSubjectType = "enterprise-capability"

type TenantDirectory interface {
	TenantIDs(ctx context.Context) ([]string, error)
}

type SubjectIDLister interface {
	SubjectIDs(ctx context.Context, subjectType string) ([]string, error)
}

type RetiredSubjectArchival struct {
	tenants  TenantDirectory
	subjects SubjectIDLister
	reactor  *SubjectDeletedReactor
}

func NewRetiredSubjectArchival(tenants TenantDirectory, subjects SubjectIDLister, facts FactsFinder, commands CommandDispatcher) *RetiredSubjectArchival {
	return &RetiredSubjectArchival{
		tenants:  tenants,
		subjects: subjects,
		reactor:  NewSubjectDeletedReactor(facts, commands),
	}
}

func (a *RetiredSubjectArchival) Run(ctx context.Context) error {
	tenantIDs, err := a.tenants.TenantIDs(ctx)
	if err != nil {
		return fmt.Errorf("list tenants for retired %s archival: %w", retiredSubjectType, err)
	}
	for _, tenantID := range tenantIDs {
		if err := a.archiveTenant(ctx, tenantID); err != nil {
			return err
		}
	}
	return nil
}

func (a *RetiredSubjectArchival) archiveTenant(ctx context.Context, tenantID string) error {
	tenant, err := sharedvo.NewTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("parse tenant %s for retired %s archival: %w", tenantID, retiredSubjectType, err)
	}
	tenantCtx := sharedctx.WithTenant(ctx, tenant)

	subjectIDs, err := a.subjects.SubjectIDs(tenantCtx, retiredSubjectType)
	if err != nil {
		return fmt.Errorf("list retired %s subjects for tenant %s: %w", retiredSubjectType, tenantID, err)
	}
	for _, subjectID := range subjectIDs {
		if err := a.reactor.archiveFacts(tenantCtx, retiredSubjectType, subjectID); err != nil {
			return err
		}
	}
	return nil
}
