package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hostingWrite struct {
	componentID string
	hosting     string
}

type fakeHostingWriter struct {
	writes []hostingWrite
}

func (f *fakeHostingWriter) SetHosting(_ context.Context, componentID, hosting string) error {
	f.writes = append(f.writes, hostingWrite{componentID, hosting})
	return nil
}

func TestApplicationHostingProjector_WritesHosting(t *testing.T) {
	writer := &fakeHostingWriter{}
	projector := NewApplicationHostingProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewApplicationHostingClassified("comp-1", valueobjects.HostingSaaS)))

	assert.Equal(t, []hostingWrite{{"comp-1", valueobjects.HostingSaaS}}, writer.writes)
}

func TestApplicationHostingProjector_IgnoresOtherEvents(t *testing.T) {
	writer := &fakeHostingWriter{}
	projector := NewApplicationHostingProjector(writer)

	require.NoError(t, projector.Handle(context.Background(), events.NewApplicationComponentCreated("comp-1", "Billing", "")))

	assert.Empty(t, writer.writes)
}
