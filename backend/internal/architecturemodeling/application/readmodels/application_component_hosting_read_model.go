package readmodels

import (
	"context"
)

func (rm *ApplicationComponentReadModel) SetHosting(ctx context.Context, componentID, hosting string) error {
	return rm.execByID(ctx,
		"UPDATE architecturemodeling.application_components SET hosting = $3 WHERE tenant_id = $1 AND id = $2",
		componentID, hosting,
	)
}
