package handlers

import (
	"testing"

	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/require"
)

func mustNewEnterpriseCapabilityRef(t *testing.T, id string) valueobjects.EnterpriseCapabilityRef {
	t.Helper()
	ref, err := valueobjects.NewEnterpriseCapabilityRef(id)
	require.NoError(t, err)
	return ref
}

func mustNewApplicationRef(t *testing.T, id string) valueobjects.ApplicationRef {
	t.Helper()
	ref, err := valueobjects.NewApplicationRef(id)
	require.NoError(t, err)
	return ref
}

func mustNewNarrative(t *testing.T, v string) sharedvo.Description {
	t.Helper()
	n, err := sharedvo.NewDescription(v)
	require.NoError(t, err)
	return n
}
