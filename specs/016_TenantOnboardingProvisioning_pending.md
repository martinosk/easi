# Tenant Onboarding & Provisioning

## Description
Implement tenant lifecycle management: automatic onboarding during OAuth login, infrastructure provisioning, settings management, and eventual offboarding. Handles SaaS business operations around multi-tenant infrastructure.

**Dependencies:** Spec 013 (Multi-Tenancy Infrastructure), Spec 015 (Authentication & Authorization)

## Core Requirements

### Tenant Lifecycle
1. **Discovery:** User logs in via OAuth, token contains org claim
2. **Onboarding:** Create tenant if not exists, provision infrastructure, log TENANT_CREATED
3. **Active:** Tenant performs operations, users invited, settings configured
4. **Offboarding (Future):** Grace period, archive data, delete infrastructure

### Tenant Aggregate
**Properties:**
- TenantId (immutable)
- DisplayName
- OAuthOrgID (external org ID from provider)
- Status: active, suspended, archived, deleted
- CreatedAt
- Metadata: display name, logo URL, website, timezone, subscription plan, max users, etc.

### API Endpoints

**POST /api/tenants**
- Creates new tenant (normally automatic during login)
- Validates ID pattern: `^[a-z0-9-]{3,50}$`
- Returns 409 if tenant exists
- Returns 201 Created with tenant object and HATEOAS links

**GET /api/tenants/{tenantId}**
- Returns tenant information with metadata
- User must belong to requested tenant (OAuth validation)
- HATEOAS links to settings and usage endpoints

**PATCH /api/tenants/{tenantId}**
- Updates tenant metadata (display name, logo, website, etc.)
- Only tenant admins can update
- Cannot change tenant ID or creation date
- Returns 403 if not admin

**DELETE /api/tenants/{tenantId}**
- Soft delete, sets status to "archived"
- Schedules data deletion after grace period (30 days)
- Notifies tenant admins
- Logs TENANT_DELETION_INITIATED event
- Admin-only operation

**GET /api/tenants/{tenantId}/usage**
- Returns usage metrics: components created, relations created, views created, API calls, storage used, active users
- Shows limits: max users, max components, max storage
- Calculates usage percentages

**GET /api/tenants/{tenantId}/settings**
- Returns tenant configuration: display name, timezone, date format, language, SSO, 2FA, session timeout, audit logging

**PATCH /api/tenants/{tenantId}/settings**
- Updates tenant settings
- Admin-only operation

### Automatic Onboarding
- Triggered during OAuth callback (Spec 015)
- Check if tenant exists by OAuth org ID
- If not exists, create tenant with ID from "org" claim
- Provision infrastructure (tenant record, default structures)
- Log TENANT_CREATED event
- Idempotent (safe to call multiple times)

### Infrastructure Provisioning
- Create tenant record in database
- All data uses existing tenant_id column (no per-tenant schemas)
- Initialize default structures (views, etc.)
- Log provisioning event
- Notify administrators

### Events

**TenantCreated:**
- Fields: tenantId, displayName, oauthOrgId, createdByUserID, createdAt
- Processors: send welcome email, initialize analytics, notify billing

**TenantSettingsUpdated:**
- Fields: tenantId, changes, updatedBy, updatedAt

**TenantDeletionInitiated:**
- Fields: tenantId, scheduledDeletionDate, deletedBy, initiatedAt
- Processors: send deletion warning, schedule final deletion, archive data

### Security
- Tenant ID immutable after creation
- Verify user belongs to OAuth org
- Admin-only operations for settings/deletion
- Audit trail for all lifecycle events
- Data archival for compliance (30 days)
- GDPR compliance support

## Checklist
- [ ] Tenant aggregate and repository
- [ ] CreateTenant command
- [ ] Automatic onboarding in OAuth callback
- [ ] Tenant provisioning function
- [ ] Tenant management endpoints (GET, PATCH, DELETE)
- [ ] Usage metrics tracking
- [ ] Settings management endpoints
- [ ] TenantCreated, TenantSettingsUpdated, TenantDeletionInitiated events
- [ ] Event processors
- [ ] Audit logging
- [ ] Cross-tenant isolation tests
- [ ] Automated cleanup for deleted tenants
- [ ] User sign-off
