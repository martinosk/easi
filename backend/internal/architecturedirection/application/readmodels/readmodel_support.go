package readmodels

import (
	"context"

	sharedctx "easi/backend/internal/shared/context"
)

const pgUniqueViolation = "23505"

type CapabilityID string

func tenantOf(ctx context.Context) (string, error) {
	t, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return "", err
	}
	return t.Value(), nil
}
