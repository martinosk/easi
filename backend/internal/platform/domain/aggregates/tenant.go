package aggregates

import (
	"errors"
	"strings"
	"time"

	"easi/backend/internal/platform/domain/events"
	"easi/backend/internal/platform/domain/valueobjects"
	"easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrFirstAdminEmailRequired       = errors.New("first admin email is required")
	ErrFirstAdminEmailDomainMismatch = errors.New("first admin email domain must match one of the tenant domains")
)

type Tenant struct {
	domain.AggregateRoot
	name            valueobjects.TenantName
	status          valueobjects.TenantStatus
	domains         []valueobjects.EmailDomain
	oidcConfig      valueobjects.OIDCConfig
	firstAdminEmail string
	createdAt       time.Time
}

func NewTenant(
	id sharedvo.TenantID,
	name valueobjects.TenantName,
	domains []valueobjects.EmailDomain,
	oidcConfig valueobjects.OIDCConfig,
	firstAdminEmail string,
) (*Tenant, error) {
	firstAdminEmail = strings.TrimSpace(firstAdminEmail)
	if firstAdminEmail == "" {
		return nil, ErrFirstAdminEmailRequired
	}

	if !emailDomainMatchesTenantDomains(firstAdminEmail, domains) {
		return nil, ErrFirstAdminEmailDomainMismatch
	}

	tenant := &Tenant{
		AggregateRoot: domain.NewAggregateRootWithID(id.Value()),
		oidcConfig:    oidcConfig,
	}

	domainStrings := make([]string, len(domains))
	for i, d := range domains {
		domainStrings[i] = d.Value()
	}

	event := events.NewTenantCreated(
		id.Value(),
		name.Value(),
		valueobjects.TenantStatusActive.Value(),
		domainStrings,
		firstAdminEmail,
	)

	tenant.apply(event)
	tenant.RaiseEvent(event)

	return tenant, nil
}

func LoadTenantFromHistory(evts []domain.DomainEvent) (*Tenant, error) {
	tenant := &Tenant{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	tenant.LoadFromHistory(evts, func(event domain.DomainEvent) {
		tenant.apply(event)
	})

	return tenant, nil
}

func (t *Tenant) apply(event domain.DomainEvent) {
	if e, ok := event.(events.TenantCreated); ok {
		t.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		t.name, _ = valueobjects.NewTenantName(e.Name)
		t.status, _ = valueobjects.NewTenantStatus(e.Status)
		t.domains = make([]valueobjects.EmailDomain, len(e.Domains))
		for i, d := range e.Domains {
			t.domains[i], _ = valueobjects.NewEmailDomain(d)
		}
		t.firstAdminEmail = e.FirstAdminEmail
		t.createdAt = e.CreatedAt
	}
}

func (t *Tenant) Name() valueobjects.TenantName {
	return t.name
}

func (t *Tenant) Status() valueobjects.TenantStatus {
	return t.status
}

func (t *Tenant) Domains() []valueobjects.EmailDomain {
	return t.domains
}

func (t *Tenant) OIDCConfig() valueobjects.OIDCConfig {
	return t.oidcConfig
}

func (t *Tenant) FirstAdminEmail() string {
	return t.firstAdminEmail
}

func (t *Tenant) CreatedAt() time.Time {
	return t.createdAt
}

func emailDomainMatchesTenantDomains(email string, domains []valueobjects.EmailDomain) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	emailDomain := strings.ToLower(parts[1])

	for _, d := range domains {
		if d.Value() == emailDomain {
			return true
		}
	}
	return false
}
