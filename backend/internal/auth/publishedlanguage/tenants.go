package publishedlanguage

import "context"

type TenantDirectory interface {
	TenantIDs(ctx context.Context) ([]string, error)
}
