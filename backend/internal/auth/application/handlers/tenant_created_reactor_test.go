package handlers

import (
	"context"
	"testing"
	"time"

	"easi/backend/internal/auth/application/commands"
	authPL "easi/backend/internal/auth/publishedlanguage"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingCommandBus struct {
	dispatched []cqrs.Command
	tenantIDs  []string
}

func (b *recordingCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	tenant, err := sharedctx.GetTenant(ctx)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	b.tenantIDs = append(b.tenantIDs, tenant.Value())
	return cqrs.NewResult("invitation-1"), nil
}

type supplierEvent struct {
	aggregateID string
	eventType   string
	data        map[string]interface{}
}

func (e supplierEvent) AggregateID() string               { return e.aggregateID }
func (e supplierEvent) EventType() string                 { return e.eventType }
func (e supplierEvent) OccurredAt() time.Time             { return time.Now() }
func (e supplierEvent) EventData() map[string]interface{} { return e.data }

func tenantCreated(firstAdminEmail string) domain.DomainEvent {
	return supplierEvent{
		aggregateID: "acme",
		eventType:   authPL.TenantCreated,
		data: map[string]interface{}{
			"id":              "acme",
			"name":            "Acme Corporation",
			"status":          "active",
			"domains":         []string{"acme.com"},
			"firstAdminEmail": firstAdminEmail,
		},
	}
}

func TestTenantCreatedReactor_InvitesFirstAdminInTenantContext(t *testing.T) {
	bus := &recordingCommandBus{}
	reactor := NewTenantCreatedReactor(bus)

	err := reactor.Handle(context.Background(), tenantCreated("admin@acme.com"))

	require.NoError(t, err)
	require.Len(t, bus.dispatched, 1)
	invitation, ok := bus.dispatched[0].(*commands.CreateInvitation)
	require.True(t, ok)
	assert.Equal(t, "admin@acme.com", invitation.Email)
	assert.Equal(t, "admin", invitation.Role)
	assert.Equal(t, []string{"acme"}, bus.tenantIDs)
}

func TestTenantCreatedReactor_SkipsWhenNoFirstAdminEmail(t *testing.T) {
	bus := &recordingCommandBus{}
	reactor := NewTenantCreatedReactor(bus)

	err := reactor.Handle(context.Background(), tenantCreated("   "))

	require.NoError(t, err)
	assert.Empty(t, bus.dispatched)
}
