# MetaModel Security Hardening

## Description
Address critical security vulnerabilities identified in the MetaModel bounded context implementation, including missing Row-Level Security (RLS), absent RBAC enforcement, and audit trail deficiencies.

## Priority
**CRITICAL** - Must be resolved before production deployment.

## Findings Summary

| Issue | Severity | Component |
|-------|----------|-----------|
| Missing RLS on read model table | CRITICAL | Database |
| No RBAC on write endpoints | CRITICAL | API |
| Hardcoded user identity | HIGH | API/Audit |

## Requirements

### 1. Database Row-Level Security (RLS)

**Problem**: The `meta_model_configurations` read model table lacks RLS policy, creating a single point of failure if application-level filtering is bypassed.

**Requirement**: Add RLS policy to `meta_model_configurations` table matching the pattern used for other tables (events, snapshots, capabilities, etc.).

**Acceptance Criteria**:
- RLS enabled on `meta_model_configurations` table
- Policy uses `app.current_tenant` session variable
- Policy applies to SELECT, INSERT, UPDATE, DELETE operations
- Applied to `easi_app` role

### 2. RBAC Enforcement on Write Endpoints

**Problem**: Any authenticated user can modify the metamodel configuration. Per requirements, only admins should be able to modify the meta-data model.

**Endpoints Requiring Protection**:
- `PUT /api/v1/metamodel/maturity-scale`
- `PUT /api/v1/metamodel/maturity-scale/reset`

**Requirement**: Add authorization middleware requiring admin permission for metamodel write operations.

**Acceptance Criteria**:
- New permission `metamodel:write` defined in permission value object
- Admin role includes `metamodel:write` permission
- Write endpoints protected with `RequirePermission` middleware
- Non-admin users receive 403 Forbidden response
- Read endpoints remain accessible to all authenticated users

### 3. Audit Trail User Identity

**Problem**: API handlers use hardcoded `system@easi.io` instead of extracting the authenticated user's email, breaking audit trail integrity.

**Requirement**: Extract authenticated user email from session context for all write operations.

**Acceptance Criteria**:
- `UpdateMaturityScale` handler extracts user email from session
- `ResetMaturityScale` handler extracts user email from session
- Events record actual user who made the change
- Read model reflects actual user in `modifiedBy` field

## Dependencies
- Spec 090: MetaModel Bounded Context (base implementation)

## Checklist
- [x] Specification ready
- [x] Migration file for RLS policy created
- [x] Permission value object updated
- [x] Admin role permissions updated
- [x] Routes updated with authorization middleware
- [x] Handlers updated to extract user identity
- [x] MetaModel routes connected to main router
- [x] Unit tests for authorization
- [x] Integration tests for session/authentication
- [ ] User sign-off
