package aggregates

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"easi/backend/internal/platform/domain/events"
	"easi/backend/internal/platform/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
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

	if err := tenant.apply(event); err != nil {
		return nil, err
	}
	tenant.RaiseEvent(event)

	return tenant, nil
}

func LoadTenantFromHistory(evts []domain.DomainEvent) (*Tenant, error) {
	tenant := &Tenant{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	tenant.LoadFromHistory(evts, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = tenant.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return tenant, nil
}

func (t *Tenant) apply(event domain.DomainEvent) error {
	if e, ok := event.(events.TenantCreated); ok {
		return t.applyCreated(e)
	}
	return nil
}

func (t *Tenant) applyCreated(e events.TenantCreated) error {
	name, err := valueobjects.NewTenantName(e.Name)
	if err != nil {
		return fmt.Errorf("%w: tenant name %q: %v", domain.ErrCorruptedEvent, e.Name, err)
	}
	status, err := valueobjects.NewTenantStatus(e.Status)
	if err != nil {
		return fmt.Errorf("%w: tenant status %q: %v", domain.ErrCorruptedEvent, e.Status, err)
	}
	domains := make([]valueobjects.EmailDomain, len(e.Domains))
	for i, d := range e.Domains {
		domains[i], err = valueobjects.NewEmailDomain(d)
		if err != nil {
			return fmt.Errorf("%w: email domain %q: %v", domain.ErrCorruptedEvent, d, err)
		}
	}
	t.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	t.name = name
	t.status = status
	t.domains = domains
	t.firstAdminEmail = e.FirstAdminEmail
	t.createdAt = e.CreatedAt
	return nil
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
