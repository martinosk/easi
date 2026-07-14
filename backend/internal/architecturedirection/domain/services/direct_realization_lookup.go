package services

import "context"

type DirectRealizationLookup func(ctx context.Context, capabilityID, componentID string) (string, bool, error)
