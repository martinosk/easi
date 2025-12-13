package aggregates

import (
	"testing"

	"easi/backend/internal/platform/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTenant_Valid(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	name, _ := valueobjects.NewTenantName("Acme Corporation")
	domains, _ := valueobjects.NewEmailDomainList([]string{"acme.com", "acme.co.uk"})
	oidcConfig, _ := valueobjects.NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		"client-secret",
		"openid email profile",
	)

	tenant, err := NewTenant(tenantID, name, domains, oidcConfig, "john.doe@acme.com")

	require.NoError(t, err)
	assert.Equal(t, "acme", tenant.ID())
	assert.Equal(t, name, tenant.Name())
	assert.Equal(t, valueobjects.TenantStatusActive, tenant.Status())
	assert.Len(t, tenant.Domains(), 2)
	assert.Equal(t, oidcConfig, tenant.OIDCConfig())
	assert.Equal(t, "john.doe@acme.com", tenant.FirstAdminEmail())
}

func TestNewTenant_RaisesTenantCreatedEvent(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	name, _ := valueobjects.NewTenantName("Acme Corporation")
	domains, _ := valueobjects.NewEmailDomainList([]string{"acme.com"})
	oidcConfig, _ := valueobjects.NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		"client-secret",
		"openid email profile",
	)

	tenant, err := NewTenant(tenantID, name, domains, oidcConfig, "admin@acme.com")

	require.NoError(t, err)
	events := tenant.GetUncommittedChanges()
	assert.Len(t, events, 1)
	assert.Equal(t, "TenantCreated", events[0].EventType())
}

func TestNewTenant_InvalidFirstAdminEmail(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	name, _ := valueobjects.NewTenantName("Acme Corporation")
	domains, _ := valueobjects.NewEmailDomainList([]string{"acme.com"})
	oidcConfig, _ := valueobjects.NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		"client-secret",
		"openid email profile",
	)

	_, err := NewTenant(tenantID, name, domains, oidcConfig, "")

	assert.Error(t, err)
	assert.Equal(t, ErrFirstAdminEmailRequired, err)
}

func TestNewTenant_FirstAdminEmailDomainMustMatch(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	name, _ := valueobjects.NewTenantName("Acme Corporation")
	domains, _ := valueobjects.NewEmailDomainList([]string{"acme.com"})
	oidcConfig, _ := valueobjects.NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		"client-secret",
		"openid email profile",
	)

	_, err := NewTenant(tenantID, name, domains, oidcConfig, "john@otherdomain.com")

	assert.Error(t, err)
	assert.Equal(t, ErrFirstAdminEmailDomainMismatch, err)
}

func TestLoadTenantFromHistory(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("acme")
	name, _ := valueobjects.NewTenantName("Acme Corporation")
	domains, _ := valueobjects.NewEmailDomainList([]string{"acme.com"})
	oidcConfig, _ := valueobjects.NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		"client-secret",
		"openid email profile",
	)

	originalTenant, _ := NewTenant(tenantID, name, domains, oidcConfig, "admin@acme.com")
	events := originalTenant.GetUncommittedChanges()

	loadedTenant, err := LoadTenantFromHistory(events)

	require.NoError(t, err)
	assert.Equal(t, "acme", loadedTenant.ID())
	assert.Equal(t, name.Value(), loadedTenant.Name().Value())
	assert.Equal(t, valueobjects.TenantStatusActive, loadedTenant.Status())
}
