package gateways

import "context"

type CapabilityGateway interface {
	CapabilityExists(ctx context.Context, capabilityID string) (bool, error)
}
