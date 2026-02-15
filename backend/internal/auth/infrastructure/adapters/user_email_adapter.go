package adapters

import (
	"context"

	"easi/backend/internal/auth/application/readmodels"
)

type UserEmailLookupAdapter struct {
	readModel *readmodels.UserReadModel
}

func NewUserEmailLookupAdapter(rm *readmodels.UserReadModel) *UserEmailLookupAdapter {
	return &UserEmailLookupAdapter{readModel: rm}
}

func (a *UserEmailLookupAdapter) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	user, err := a.readModel.GetByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

type InvitationCheckerAdapter struct {
	readModel *readmodels.InvitationReadModel
}

func NewInvitationCheckerAdapter(rm *readmodels.InvitationReadModel) *InvitationCheckerAdapter {
	return &InvitationCheckerAdapter{readModel: rm}
}

func (a *InvitationCheckerAdapter) HasPendingByEmail(ctx context.Context, email string) (bool, error) {
	inv, err := a.readModel.GetAnyPendingByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return inv != nil, nil
}

type DomainAllowlistCheckerAdapter struct {
	checker *readmodels.TenantDomainChecker
}

func NewDomainAllowlistCheckerAdapter(checker *readmodels.TenantDomainChecker) *DomainAllowlistCheckerAdapter {
	return &DomainAllowlistCheckerAdapter{checker: checker}
}

func (a *DomainAllowlistCheckerAdapter) IsDomainAllowed(ctx context.Context, email string) (bool, error) {
	return a.checker.IsDomainAllowed(ctx, email)
}
