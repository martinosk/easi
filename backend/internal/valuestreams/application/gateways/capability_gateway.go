package gateways

import "context"

type CapabilityInfo struct {
	ID   string
	Name string
}

type CapabilityGateway interface {
	CapabilityExists(ctx context.Context, capabilityID string) (bool, error)
	GetCapability(ctx context.Context, capabilityID string) (*CapabilityInfo, error)
}
