-- Migration: Add Release 0.14.0
-- Description: Adds release notes for version 0.14.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.14.0', '2025-12-13', '## What''s New in 0.14.0

### Multi-Tenant Authentication
- **Tenant login flow**: Users can now log in through their organization''s identity provider via OIDC
- **OIDC configuration per tenant**: Each tenant can configure their own identity provider (discovery URL, client ID, scopes)
- **Domain-based tenant resolution**: Automatic tenant identification based on email domain

### Tenant Provisioning
- **Tenant management**: Platform-level tenant creation with status lifecycle (active, suspended, archived)
- **User management**: Tenant-scoped users with role-based access (admin, architect, stakeholder)
- **Invitation system**: Invite users via email with role assignment and expiration

### Security
- **Row-level security**: All tenant-scoped tables (users, invitations) enforce strict tenant isolation
- **HTTP session management**: Secure session storage using SCS with PostgreSQL backend
- **Security review**: Comprehensive security hardening across the authentication flow

### Developer Experience
- **Local OIDC with Dex**: Development environment now includes Dex as a local identity provider
- **Test tenant seeding**: Automatic provisioning of test tenant (ACME) for local development
', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
