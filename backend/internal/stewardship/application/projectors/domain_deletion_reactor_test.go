package projectors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/stewardship/application/commands"
)

type fakeLiveConcerns map[string][]string

func (f fakeLiveConcerns) LiveConcernsOfDomain(_ context.Context, domainID string) ([]string, error) {
	return f[domainID], nil
}

type recordingBus struct {
	dispatched []cqrs.Command
	calls      *[]string
	failWith   error
}

func (b *recordingBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	*b.calls = append(*b.calls, "release")
	return cqrs.EmptyResult(), b.failWith
}

func (b *recordingBus) Register(string, cqrs.CommandHandler) {}

type recordingForgetter struct {
	calls     *[]string
	forgotten []string
}

func (f *recordingForgetter) Delete(_ context.Context, domainID string) error {
	*f.calls = append(*f.calls, "forget")
	f.forgotten = append(f.forgotten, domainID)
	return nil
}

func TestDomainDeletionReactor_ReleasesEveryStewardshipThenForgetsTheDomain(t *testing.T) {
	var calls []string
	bus := &recordingBus{calls: &calls}
	domains := &recordingForgetter{calls: &calls}
	reactor := NewDomainDeletionReactor(fakeLiveConcerns{"ce": {"ownership", "assessment"}}, bus, domains)

	require.NoError(t, reactor.Handle(ctx, storedEvent(t, "ce", capPL.BusinessDomainDeleted, map[string]any{"id": "ce"})))

	require.Len(t, bus.dispatched, 2)
	for i, concern := range []string{"ownership", "assessment"} {
		cmd := bus.dispatched[i].(*commands.ReleaseSteward)
		assert.Equal(t, "ce", cmd.DomainID)
		assert.Equal(t, concern, cmd.Concern)
		assert.Equal(t, "system:domain-deleted", cmd.ReleasedBy)
	}
	assert.Equal(t, []string{"ce"}, domains.forgotten)
	assert.Equal(t, []string{"release", "release", "forget"}, calls)
}

func TestDomainDeletionReactor_KeepsTheDomainWhenAReleaseFails(t *testing.T) {
	var calls []string
	bus := &recordingBus{calls: &calls, failWith: errors.New("boom")}
	domains := &recordingForgetter{calls: &calls}
	reactor := NewDomainDeletionReactor(fakeLiveConcerns{"ce": {"ownership"}}, bus, domains)

	err := reactor.Handle(ctx, storedEvent(t, "ce", capPL.BusinessDomainDeleted, map[string]any{"id": "ce"}))

	assert.Error(t, err)
	assert.Empty(t, domains.forgotten)
}
