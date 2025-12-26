# MetaModel Cross-Context Integration

## Description
Define the integration patterns between the MetaModel bounded context and other contexts (Platform, CapabilityMapping) using event-driven architecture.

## Purpose
Ensure loose coupling between bounded contexts while enabling automatic provisioning of default configuration and dynamic derivation of maturity section names.

## Dependencies
- Spec 090: MetaModel Bounded Context
- Spec 091: Maturity Scale Configuration Aggregate
- Spec 065: TenantProvisioning

## Integration 1: Platform → MetaModel (TenantCreated)

### Event Source
Platform context publishes `TenantCreated` when a new tenant is provisioned.

### Handler: TenantCreatedHandler

**Location**: `backend/internal/metamodel/application/handlers/tenant_created_handler.go`

**Behavior**:
1. Receive `TenantCreated` event from Platform context
2. Check if MetaModelConfiguration already exists for tenant (idempotency via `GetByTenantID`)
3. If not exists:
   - Generate new `MetaModelConfigurationID` (UUID)
   - Create default `MaturityScaleConfig`
   - Create `MetaModelConfiguration` aggregate with both IDs
4. Save aggregate (triggers `MetaModelConfigurationCreated` event)

### Event Subscription
Register subscription to `TenantCreated` event during MetaModel context initialization.

## Integration 2: MetaModel → CapabilityMapping (Configuration Read)

### Pattern: Anti-Corruption Layer with REST API

CapabilityMapping context needs the current maturity scale configuration to:
1. Validate maturity values in UpdateCapabilityMetadata command
2. Derive section names for API responses
3. Populate the `/api/v1/capabilities/metadata/maturity-levels` endpoint

### MaturityScaleGateway (Anti-Corruption Layer)

**Location**: `backend/internal/capabilitymapping/infrastructure/metamodel/maturity_scale_gateway.go`

**Interface**: `GetMaturityScaleConfig(ctx) → MaturityScaleConfigDTO`

**DTO Structure**:
- `Sections`: array of section DTOs (order, name, minValue, maxValue)

### Implementation: REST API

**Architectural Decision:** Use REST API calls to maintain loose coupling between bounded contexts. Direct database queries are NOT allowed as they create tight coupling and violate context boundaries.

**Why REST API over Direct Database Query:**
- Maintains bounded context isolation
- MetaModel can evolve its internal schema without breaking CapabilityMapping
- Clear Published Language contract via API specification
- Enables future deployment separation if needed

### Caching Strategy

The maturity scale configuration changes infrequently. Implement tenant-aware caching:
- Cache key: tenant ID
- TTL: 5 minutes
- Thread-safe access with read/write locking
- Automatic expiry and refresh on access

## Integration 3: MetaModel → CapabilityMapping (Configuration Changed)

### Event: MaturityScaleConfigUpdated

When the maturity scale configuration changes, CapabilityMapping invalidates its cache.

**Note**: Capabilities don't need to be updated because they store numeric values (0-99). Section names are derived at read time.

### Handler: MaturityScaleConfigUpdatedHandler

**Location**: `backend/internal/capabilitymapping/application/handlers/maturity_scale_config_updated_handler.go`

**Behavior**: Invalidate cached maturity scale config for the affected tenant.

## Updated Maturity Levels Endpoint

### Location
`backend/internal/capabilitymapping/infrastructure/api/maturity_level_handlers.go`

### Behavior
1. Fetch configuration via MaturityScaleGateway
2. If unavailable, fall back to hardcoded defaults
3. Transform sections to maturity level DTOs with ranges
4. Include HATEOAS link to configuration endpoint

## Fallback Behavior

If the MetaModel context is unavailable or configuration is missing:

1. **During tenant provisioning failure**: Log error, allow operation with defaults
2. **During API requests**: Fall back to hardcoded defaults (Genesis 0-24, Custom Built 25-49, Product 50-74, Commodity 75-99)
3. **During event handling**: Retry with exponential backoff

## Testing Cross-Context Integration

### Unit Tests
- Mock gateway for CapabilityMapping handlers
- Test cache invalidation on events

### Integration Tests
- End-to-end: Create tenant → verify MetaModelConfiguration exists
- End-to-end: Update maturity scale → verify maturity-levels endpoint reflects changes
- Fallback behavior when MetaModel unavailable

## Checklist
- [ ] Specification ready
- [ ] TenantCreatedHandler implemented in MetaModel context
- [ ] Event subscription registered in MetaModel routes
- [ ] MaturityScaleGateway interface defined
- [ ] MaturityScaleGateway REST implementation
- [ ] Caching layer implemented
- [ ] MaturityScaleConfigUpdatedHandler for cache invalidation
- [ ] Maturity levels endpoint updated to use gateway
- [ ] Fallback to defaults implemented
- [ ] Unit tests for handlers
- [ ] Integration tests for cross-context flows
- [ ] User sign-off
