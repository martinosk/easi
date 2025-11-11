# Synthetic Transaction Tenant for Production Monitoring

## Description
Implements synthetic transaction testing using a dedicated `synthetic-monitoring` tenant that executes real end-to-end transactions in production without polluting real customer data. This enables proactive health monitoring, alerting on production issues before customers are affected, and validation that all system components are functioning correctly.

Synthetic transactions run continuously in production, exercising critical user workflows and verifying system health from an end-user perspective. Unlike traditional health checks that only verify service availability, synthetic transactions validate the entire application stack including API, event processing, database writes, and read model projections.

**Dependencies:** Requires Spec 013 (Multi-Tenancy Infrastructure) to be completed first.

## Core Principles
- **Real transactions** - Execute actual commands/events, not mocked
- **Isolated data** - Synthetic tenant prevents pollution of production data
- **Self-cleaning** - Automated cleanup prevents data buildup
- **Continuous validation** - Runs on a schedule (e.g., every 5 minutes)
- **Full stack** - Tests API → Command → Event → Projection → Read Model
- **Observable** - Failures trigger alerts to on-call engineers

## Synthetic Transaction Scenarios

### Scenario 1: Component Creation Flow (Critical Path)
Tests the core workflow of creating a component and verifying it appears in read models.

**Steps:**
1. POST /api/application-component with synthetic data
2. Verify 201 Created response
3. Extract component ID from response
4. GET /api/application-component/{id}
5. Verify component data matches
6. GET /api/application-component
7. Verify component appears in list
8. Cleanup: DELETE synthetic component

**Expected Duration:** < 2 seconds
**Alert Threshold:** Failure or duration > 5 seconds

### Scenario 2: Component Relation Flow
Tests relationship creation between components.

**Steps:**
1. Create two synthetic components
2. POST /api/component-relation
3. Verify 201 Created response
4. GET /api/component-relation/from/{sourceId}
5. Verify relation appears
6. Cleanup: DELETE relation and components

**Expected Duration:** < 3 seconds
**Alert Threshold:** Failure or duration > 6 seconds

### Scenario 3: View Management Flow (Future)
Tests view creation and component-to-view association.

**Steps:**
1. Create synthetic view
2. Add synthetic components to view
3. Verify view contains components
4. Cleanup: DELETE view

**Expected Duration:** < 3 seconds
**Alert Threshold:** Failure or duration > 6 seconds

## API Endpoints

### POST /api/tenants/{tenantId}/synthetic/cleanup
Deletes all data for a synthetic tenant.

**Path Parameters:**
- `tenantId` (string, required): Must have prefix `synthetic-`

**Validation Rules:**
- Only tenants with prefix `synthetic-` can be cleaned up
- Returns 403 Forbidden for regular tenant IDs

**Response:** 204 No Content

**Error Responses:**
- 403 Forbidden: Tenant ID does not have `synthetic-` prefix
- 404 Not Found: Tenant does not exist

**Implementation:**
```go
func CleanupSyntheticTenant(tenantID TenantID) error {
    // Security: Only allow synthetic tenant cleanup
    if !strings.HasPrefix(tenantID.Value(), "synthetic-") {
        return ErrForbidden
    }

    // Delete from read models (cascading)
    componentRepo.DeleteByTenant(tenantID)
    relationRepo.DeleteByTenant(tenantID)
    viewRepo.DeleteByTenant(tenantID)

    // Archive events (don't delete - maintain audit trail)
    eventStore.ArchiveByTenant(tenantID)

    return nil
}
```

### GET /api/tenants/{tenantId}/synthetic/stats
Gets statistics about synthetic tenant data (for monitoring).

**Response:** 200 OK
```json
{
  "tenantId": "synthetic-monitoring",
  "componentCount": 42,
  "relationCount": 15,
  "eventCount": 238,
  "oldestEventAge": "4h32m",
  "lastCleanup": "2025-01-15T10:30:00Z",
  "_links": {
    "self": {
      "href": "/api/tenants/synthetic-monitoring/synthetic/stats"
    },
    "cleanup": {
      "href": "/api/tenants/synthetic-monitoring/synthetic/cleanup",
      "method": "POST"
    }
  }
}
```

## Synthetic Transaction Implementation

### Playwright E2E Tests as Synthetic Monitors

Reuse existing e2e tests as production synthetic monitors by:
1. Running against production URL with `synthetic-monitoring` tenant
2. Executing on schedule (cron job / GitHub Actions)
3. Reporting failures to monitoring system

**Configuration:**
```typescript
// playwright.config.synthetic.ts
export default defineConfig({
  testDir: './e2e/synthetic',
  workers: 1,
  retries: 2, // Retry transient failures
  timeout: 30000, // 30 second timeout per test

  use: {
    baseURL: process.env.PRODUCTION_URL || 'https://easi-prod.example.com',

    // Inject synthetic tenant header
    extraHTTPHeaders: {
      'X-Tenant-ID': 'synthetic-monitoring',
    },
  },

  // Report results to monitoring
  reporter: [
    ['list'],
    ['./reporters/datadog-reporter.ts'], // Send metrics to Datadog
  ],
});
```

**Synthetic Test Structure:**
```typescript
// e2e/synthetic/component-creation.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Synthetic: Component Creation', () => {
  test.beforeEach(async ({ request }) => {
    // Cleanup before test to prevent buildup
    await request.post('/api/tenants/synthetic-monitoring/synthetic/cleanup');
  });

  test('should create component and verify in read model', async ({ request, page }) => {
    const startTime = Date.now();

    // 1. Create component via API
    const createResponse = await request.post('/api/application-component', {
      headers: { 'X-Tenant-ID': 'synthetic-monitoring' },
      data: {
        name: `Synthetic-Component-${Date.now()}`,
        description: 'Synthetic monitoring test component',
      },
    });

    expect(createResponse.status()).toBe(201);
    const component = await createResponse.json();

    // 2. Verify in read model
    const getResponse = await request.get(`/api/application-component/${component.id}`, {
      headers: { 'X-Tenant-ID': 'synthetic-monitoring' },
    });

    expect(getResponse.status()).toBe(200);

    // 3. Verify in UI
    await page.goto('/');
    await page.waitForSelector('[data-testid="canvas-loaded"]');
    await expect(page.locator('.component-node-header')
      .filter({ hasText: component.name }))
      .toBeVisible();

    // 4. Report timing
    const duration = Date.now() - startTime;
    console.log(`Synthetic transaction completed in ${duration}ms`);

    // Alert if too slow
    expect(duration).toBeLessThan(5000);
  });

  test.afterEach(async ({ request }) => {
    // Cleanup after test
    await request.post('/api/tenants/synthetic-monitoring/synthetic/cleanup');
  });
});
```

### Scheduled Execution

**GitHub Actions Workflow:**
```yaml
# .github/workflows/synthetic-monitoring.yml
name: Synthetic Transaction Monitoring

on:
  schedule:
    # Run every 5 minutes
    - cron: '*/5 * * * *'
  workflow_dispatch: # Allow manual trigger

jobs:
  synthetic-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Run synthetic transactions
        env:
          PRODUCTION_URL: ${{ secrets.PRODUCTION_URL }}
        run: npm run test:synthetic

      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: synthetic-test-results
          path: playwright-report/

      - name: Notify on failure
        if: failure()
        uses: slackapi/slack-github-action@v1
        with:
          webhook-url: ${{ secrets.SLACK_WEBHOOK_URL }}
          payload: |
            {
              "text": "🚨 Synthetic transaction failed in production!",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Production Synthetic Transaction Failed*\nCheck GitHub Actions for details."
                  }
                }
              ]
            }
```

**Package.json Script:**
```json
{
  "scripts": {
    "test:synthetic": "playwright test --config=playwright.config.synthetic.ts"
  }
}
```

## Automated Cleanup Job

### Scheduled Cleanup Process

Prevent synthetic data buildup by running periodic cleanup:

**Backend Scheduler (using go-cron or similar):**
```go
func StartSyntheticCleanupScheduler() {
    c := cron.New()

    // Run cleanup every hour
    c.AddFunc("0 * * * *", func() {
        log.Info("Running scheduled synthetic tenant cleanup")

        syntheticTenants := []string{
            "synthetic-monitoring",
            "synthetic-load-test",
        }

        for _, tenantIDStr := range syntheticTenants {
            tenantID, err := NewTenantID(tenantIDStr)
            if err != nil {
                log.Error("Invalid synthetic tenant ID", "error", err)
                continue
            }

            if err := CleanupSyntheticTenant(tenantID); err != nil {
                log.Error("Failed to cleanup synthetic tenant",
                    "tenant", tenantID,
                    "error", err)
            } else {
                log.Info("Successfully cleaned up synthetic tenant",
                    "tenant", tenantID)
            }
        }
    })

    c.Start()
}
```

**Cleanup Metrics:**
```go
type CleanupMetrics struct {
    TenantID         string
    ComponentsDeleted int
    RelationsDeleted  int
    EventsArchived    int
    Duration          time.Duration
}

func RecordCleanupMetrics(metrics CleanupMetrics) {
    // Send to monitoring system (Prometheus, Datadog, etc.)
    prometheusHistogram.Observe(metrics.Duration.Seconds())
    prometheusGauge.Set(float64(metrics.ComponentsDeleted))
}
```

## Monitoring and Alerting

### Metrics to Track

**Transaction Success Rate:**
- `synthetic_transaction_success_total` (counter)
- `synthetic_transaction_failure_total` (counter)
- Alert if success rate < 95% over 15 minutes

**Transaction Duration:**
- `synthetic_transaction_duration_seconds` (histogram)
- Alert if p95 > 5 seconds

**Cleanup Efficiency:**
- `synthetic_cleanup_duration_seconds` (histogram)
- `synthetic_cleanup_records_deleted` (gauge)

**Data Buildup:**
- `synthetic_tenant_record_count` (gauge)
- Alert if count > 1000 (indicates cleanup failure)

### Alert Rules

**Critical Alerts (Page On-Call):**
- 3 consecutive synthetic transaction failures
- Transaction duration > 10 seconds
- Synthetic data count > 5000 records

**Warning Alerts (Slack/Email):**
- 1 synthetic transaction failure
- Transaction duration > 5 seconds
- Cleanup job failed

### Monitoring Dashboard

**Key Metrics to Display:**
1. Synthetic transaction success rate (last 24h)
2. Transaction duration over time (p50, p95, p99)
3. Current synthetic data record count
4. Last cleanup timestamp
5. Recent failures with stack traces

## Security and Safety

### Protections Against Accidental Production Data Deletion

**Tenant ID Validation:**
```go
func (h *SyntheticHandler) CleanupSyntheticTenant(w http.ResponseWriter, r *http.Request) {
    tenantIDStr := mux.Vars(r)["tenantId"]

    // CRITICAL: Only allow synthetic- prefix
    if !strings.HasPrefix(tenantIDStr, "synthetic-") {
        log.Warn("Attempted cleanup of non-synthetic tenant",
            "tenant", tenantIDStr,
            "ip", r.RemoteAddr)
        http.Error(w, "Only synthetic tenants can be cleaned up", http.StatusForbidden)
        return
    }

    // Additional safety: check against whitelist
    allowedTenants := []string{
        "synthetic-monitoring",
        "synthetic-load-test",
    }

    if !contains(allowedTenants, tenantIDStr) {
        http.Error(w, "Tenant not in cleanup whitelist", http.StatusForbidden)
        return
    }

    // Proceed with cleanup
    tenantID, _ := NewTenantID(tenantIDStr)
    if err := h.service.CleanupSyntheticTenant(tenantID); err != nil {
        http.Error(w, "Cleanup failed", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

**Audit Logging:**
```go
func AuditCleanup(tenantID TenantID, user string, result string) {
    auditLog := AuditEntry{
        Timestamp: time.Now(),
        Action:    "SYNTHETIC_CLEANUP",
        TenantID:  tenantID.Value(),
        User:      user,
        Result:    result,
    }

    auditStore.Record(auditLog)
}
```

### Event Archival (Not Deletion)

**Archive Table for Audit Trail:**
```sql
CREATE TABLE events_archive (
    id BIGSERIAL PRIMARY KEY,
    original_event_id BIGINT NOT NULL,
    tenant_id VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    archived_at TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_events_archive_tenant (tenant_id, archived_at)
);

-- Archive instead of delete
INSERT INTO events_archive
    (original_event_id, tenant_id, aggregate_id, event_type, event_data)
SELECT id, tenant_id, aggregate_id, event_type, event_data
FROM events
WHERE tenant_id = 'synthetic-monitoring';

-- Then delete from active table
DELETE FROM events WHERE tenant_id = 'synthetic-monitoring';
```

**Retention Policy:**
- Keep archived synthetic events for 30 days
- Purge after 30 days to prevent unbounded growth

## Configuration

### Environment Variables

```bash
# Enable synthetic monitoring
SYNTHETIC_MONITORING_ENABLED=true

# Synthetic tenant IDs (comma-separated)
SYNTHETIC_TENANT_IDS=synthetic-monitoring,synthetic-load-test

# Cleanup schedule (cron expression)
SYNTHETIC_CLEANUP_SCHEDULE="0 * * * *"  # Every hour

# Cleanup retention (how long to keep archived events)
SYNTHETIC_CLEANUP_RETENTION_DAYS=30

# Alert thresholds
SYNTHETIC_ALERT_FAILURE_THRESHOLD=3
SYNTHETIC_ALERT_DURATION_THRESHOLD_MS=5000
```

### Feature Flags

```go
type SyntheticConfig struct {
    Enabled              bool
    AllowedTenantIDs     []string
    CleanupSchedule      string
    RetentionDays        int
    AlertFailureThreshold int
    AlertDurationMs      int
}

func LoadSyntheticConfig() SyntheticConfig {
    return SyntheticConfig{
        Enabled: os.Getenv("SYNTHETIC_MONITORING_ENABLED") == "true",
        AllowedTenantIDs: strings.Split(
            os.Getenv("SYNTHETIC_TENANT_IDS"),
            ",",
        ),
        CleanupSchedule: getEnvOrDefault(
            "SYNTHETIC_CLEANUP_SCHEDULE",
            "0 * * * *",
        ),
        RetentionDays: getEnvIntOrDefault(
            "SYNTHETIC_CLEANUP_RETENTION_DAYS",
            30,
        ),
        AlertFailureThreshold: getEnvIntOrDefault(
            "SYNTHETIC_ALERT_FAILURE_THRESHOLD",
            3,
        ),
        AlertDurationMs: getEnvIntOrDefault(
            "SYNTHETIC_ALERT_DURATION_THRESHOLD_MS",
            5000,
        ),
    }
}
```

## Testing Strategy

### Unit Tests
- [ ] TenantID validation prevents non-synthetic cleanup
- [ ] Cleanup whitelist enforcement
- [ ] Event archival logic

### Integration Tests
- [ ] Create synthetic data and verify cleanup
- [ ] Verify archived events are preserved
- [ ] Test cleanup with multiple tenants

### E2E Tests
- [ ] Full synthetic transaction flow
- [ ] Verify read model updates
- [ ] Test cleanup doesn't affect other tenants

### Security Tests
- [ ] Attempt cleanup of production tenant (should fail)
- [ ] Attempt cleanup without `synthetic-` prefix (should fail)
- [ ] Cross-tenant data access attempts (should fail)

## Rollout Plan

### Phase 1: Infrastructure
- [ ] Implement cleanup endpoint
- [ ] Add tenant ID validation
- [ ] Create event archival mechanism

### Phase 2: Synthetic Tests
- [ ] Convert e2e tests to synthetic tests
- [ ] Add monitoring reporter
- [ ] Test locally against staging

### Phase 3: Scheduling
- [ ] Implement cleanup scheduler
- [ ] Configure GitHub Actions workflow
- [ ] Set up alerting rules

### Phase 4: Production Deployment
- [ ] Deploy to production with feature flag OFF
- [ ] Manual test synthetic transactions
- [ ] Enable scheduled execution
- [ ] Monitor for 1 week

### Phase 5: Full Automation
- [ ] Enable automatic alerting
- [ ] Document runbook for on-call
- [ ] Train team on synthetic monitoring

## Success Metrics

- **Coverage:** All critical user flows have synthetic tests
- **Reliability:** < 0.1% false positive alert rate
- **Performance:** Synthetic transactions complete in < 3 seconds (p95)
- **Detection:** Incidents detected by synthetic tests before user reports
- **Cleanup:** Zero data buildup (synthetic tenant stays < 100 records)

## Checklist
- [ ] Specification ready
- [ ] Cleanup endpoint implemented
- [ ] Tenant ID validation for cleanup
- [ ] Event archival mechanism created
- [ ] Scheduled cleanup job implemented
- [ ] Synthetic tenant e2e tests created
- [ ] GitHub Actions workflow configured
- [ ] Monitoring metrics instrumented
- [ ] Alert rules configured
- [ ] Audit logging implemented
- [ ] Unit tests implemented and passing
- [ ] Integration tests with cleanup passing
- [ ] E2E synthetic tests passing
- [ ] Security tests verify tenant isolation
- [ ] Documentation and runbook created
- [ ] Team trained on synthetic monitoring
- [ ] User sign-off
