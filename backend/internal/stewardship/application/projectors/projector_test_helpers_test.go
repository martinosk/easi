package projectors

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	domain "easi/backend/internal/shared/eventsourcing"
)

func storedEvent(t *testing.T, aggregateID, eventType string, data map[string]any) domain.DomainEvent {
	t.Helper()
	payload, err := json.Marshal(data)
	require.NoError(t, err)
	return domain.NewGenericDomainEvent(aggregateID, eventType, payload, time.Now().UTC())
}

var ctx = context.Background()
