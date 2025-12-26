# Capability Maturity Data Migration

## Description
Migrate capability maturity storage from string-based section names to numeric values (0-99), enabling dynamic mapping to configurable sections.

## Purpose
Enable capabilities to retain their maturity values when the maturity scale configuration changes, allowing automatic remapping to new section names based on numeric values.

## Dependencies
- Spec 090: MetaModel Bounded Context
- Spec 091: Maturity Scale Configuration Aggregate

## Current State

### Domain Model
- `MaturityLevel` is a string-based value (Genesis, Custom Build, Product, Commodity)
- Returns ordinal numeric value (1-4) for ordering

### Database (Read Model)
- `maturity_level VARCHAR(50)` column stores section name

### Events
- `CapabilityMetadataUpdated` stores maturity as string field `maturityLevel`

## Target State

### Domain Model
- `MaturityLevel` stores numeric value (0-99)
- Constructor validates range 0-99
- Factory method to create from section name using current configuration
- Methods to derive section name/order from current configuration

### Database (Read Model)
- `maturity_value INT` column stores numeric value (0-99)
- Default: 12 (center of Genesis section in default config)

### Events (Backward Compatible)
- `CapabilityMetadataUpdated` includes both:
  - `maturityValue` (int, 0-99) - new format
  - `maturityLevel` (string) - for backward compatibility with existing events

## Migration Strategy

### Phase 1: Add Numeric Column (Non-Breaking)

**Migration**: `0XX_add_maturity_value_column.sql`
```sql
ALTER TABLE capabilities ADD COLUMN maturity_value INT;

UPDATE capabilities SET maturity_value =
    CASE maturity_level
        WHEN 'Genesis' THEN 12        -- Midpoint of 0-24
        WHEN 'Custom Build' THEN 37   -- Midpoint of 25-49
        WHEN 'Product' THEN 62        -- Midpoint of 50-74
        WHEN 'Commodity' THEN 87      -- Midpoint of 75-99
        ELSE 12                       -- Default to Genesis midpoint
    END;

ALTER TABLE capabilities ALTER COLUMN maturity_value SET NOT NULL;
ALTER TABLE capabilities ALTER COLUMN maturity_value SET DEFAULT 12;
```

### Phase 2: Update Domain Model

1. Change `MaturityLevel` value object to store `int` (0-99)
2. Update `Capability` aggregate to use new MaturityLevel
3. Update `CapabilityMetadataUpdated` event to include both formats
4. Update projector to write maturity_value instead of maturity_level

### Phase 3: Update API Responses

Capability responses include both formats for backward compatibility:
```json
{
  "maturityLevel": "Genesis",      // Derived from maturity_value using current config
  "maturityValue": 12,             // Actual stored value
  "maturitySection": {
    "name": "Genesis",
    "order": 1,
    "range": { "min": 0, "max": 24 }
  }
}
```

### Phase 4: Remove String Column (Breaking - Later)

After frontend migration is complete and tested:

**Migration**: `0XX_remove_maturity_level_column.sql`
```sql
ALTER TABLE capabilities DROP COLUMN maturity_level;
```

## Event Replay Handling

When replaying events from history, the aggregate must handle both formats:
1. If `maturityValue` is present and > 0, use it directly
2. If only `maturityLevel` (string) is present, convert to numeric using legacy conversion

### Legacy Conversion
Legacy string values are converted to numeric midpoints using the default configuration from MetaModel context:
- Genesis → midpoint of section 1 (default: 12)
- Custom Build → midpoint of section 2 (default: 37)
- Product → midpoint of section 3 (default: 62)
- Commodity → midpoint of section 4 (default: 87)

The conversion references the default configuration from MetaModel, ensuring consistency if default values ever change.

## Cross-Context Event Handling

### MaturityScaleConfigUpdated Event

When the MetaModel context publishes `MaturityScaleConfigUpdated`, the CapabilityMapping context may need to update capabilities:

**Scenario 1: Section boundaries change**
- No action needed - capabilities keep their numeric values
- Section names are derived dynamically at read time

**Scenario 2: Section names change**
- No action needed - names are derived dynamically

**Benefit**: Capabilities don't need to be updated when configuration changes. The read model can derive section names at query time.

## API Changes

### Update Capability Metadata Request

**Current**:
```json
{
  "maturityLevel": "Genesis"
}
```

**New** (accepts either):
```json
{
  "maturityLevel": "Genesis"    // Will be converted to value 12
}
```
OR
```json
{
  "maturityValue": 42           // Direct numeric value
}
```

Validation: If `maturityValue` is provided, it must be 0-99. If `maturityLevel` is provided, it must match a section name in the current configuration.

### Get Capability Response

```json
{
  "id": "cap-123",
  "name": "Customer Onboarding",
  "maturityValue": 42,
  "maturitySection": {
    "name": "Custom Built",
    "order": 2,
    "range": { "min": 25, "max": 49 }
  }
}
```

## Testing Strategy

1. **Unit Tests**:
   - New MaturityLevel value object with 0-99 validation
   - Legacy string-to-value conversion
   - Section lookup based on numeric value

2. **Integration Tests**:
   - Event replay with mixed legacy/new event formats
   - API accepts both string and numeric formats
   - Read model correctly stores numeric values

3. **Migration Tests**:
   - Verify existing data is converted correctly
   - Verify constraints are maintained

## Rollback Plan

If issues arise:
1. Keep both columns during transition period
2. Fall back to reading maturity_level if maturity_value is problematic
3. Restore original value object if needed (strings only)

## Checklist
- [ ] Specification ready
- [ ] Database migration script created
- [ ] MaturityLevel value object updated to use int
- [ ] Capability aggregate updated
- [ ] CapabilityMetadataUpdated event updated (dual format)
- [ ] Legacy event conversion implemented
- [ ] Projector updated
- [ ] Read model updated
- [ ] API handlers updated (accept both formats)
- [ ] Unit tests for new value object
- [ ] Unit tests for legacy conversion
- [ ] Integration tests for migration
- [ ] Migration tested on staging data
- [ ] User sign-off
