# Testing Guide

## Test Structure

The EASI backend has two types of tests:

### 1. Unit Tests
Fast, isolated tests that don't require external dependencies.

**Location**: `*_test.go` files without build tags

**Run:**
```bash
go test ./...
```

**Examples:**
- `internal/architecturemodeling/domain/valueobjects/component_name_test.go`
- `internal/architecturemodeling/domain/aggregates/application_component_test.go`
- `internal/infrastructure/eventstore/event_store_test.go`

### 2. Integration Tests
Tests that verify the full stack including database interactions.

**Location**: `*_integration_test.go` files with `// +build integration` tag

**Prerequisites:**
- PostgreSQL running on `localhost:5432`
- Database `easi` with user `easi` password `easi`

**Run:**
```bash
# Start database
cd /home/devuser/repos/easi
podman compose up -d

# Run integration tests
cd backend
./test_integration.sh

# Or manually:
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/... -count=1
```

## Integration Test Coverage

### Component API Tests ✅

**File:** `internal/architecturemodeling/infrastructure/api/component_handlers_integration_test.go`

**Test Cases:**
1. **TestCreateComponent_Integration**
   - Creates component via POST /api/v1/components
   - Verifies HTTP 201 Created response
   - Verifies event saved to event store
   - Verifies event data contains correct information
   - Tests read model population

2. **TestGetAllComponents_Integration**
   - Seeds test data in read model
   - Tests GET /api/v1/components
   - Verifies HTTP 200 OK response
   - Verifies all components returned
   - Verifies HATEOAS links present

3. **TestGetComponentByID_Integration**
   - Seeds specific component
   - Tests GET /api/v1/components/{id}
   - Verifies HTTP 200 OK response
   - Verifies correct component returned
   - Verifies HATEOAS links present

4. **TestGetComponentByID_NotFound_Integration**
   - Tests GET with non-existent ID
   - Verifies HTTP 404 Not Found response

5. **TestCreateComponent_ValidationError_Integration**
   - Tests POST with empty name
   - Verifies HTTP 400 Bad Request response
   - Verifies no event was created
   - Tests domain validation

## What Gets Tested

### Full Stack Verification ✅
- HTTP Request → Handler
- Handler → Command Bus → Command Handler
- Command Handler → Domain Validation (Value Objects)
- Domain Validation → Aggregate Creation
- Aggregate → Event Raising
- Event → Event Store Persistence
- Read Model Population (manual projection for testing)
- Read Model → HTTP Response

### Database Interactions ✅
- Event store schema initialization
- Event persistence to `events` table
- Event data serialization (JSONB)
- Read model schema initialization
- Read model queries
- Transaction handling
- Concurrent access (via optimistic concurrency in event store)

### API Compliance ✅
- Proper HTTP status codes (200, 201, 400, 404)
- Request body parsing and validation
- Response formatting
- HATEOAS link generation
- Error handling and error responses

## Running Specific Tests

### Run only integration tests:
```bash
go test -v -tags=integration ./...
```

### Run only unit tests (exclude integration):
```bash
go test -v ./...
```

### Run tests for specific package:
```bash
# Unit tests for domain layer
go test -v ./internal/architecturemodeling/domain/...

# Integration tests for API layer
go test -v -tags=integration ./internal/architecturemodeling/infrastructure/api/...
```

### Run with coverage:
```bash
# Unit tests
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Integration tests
go test -v -tags=integration -coverprofile=coverage_integration.out ./internal/architecturemodeling/infrastructure/api/...
go tool cover -html=coverage_integration.out -o coverage_integration.html
```

## Test Results

### Latest Test Run

**Unit Tests:**
```
ok  	github.com/easi/backend/internal/architecturemodeling/domain/aggregates
ok  	github.com/easi/backend/internal/architecturemodeling/domain/valueobjects
ok  	github.com/easi/backend/internal/infrastructure/eventstore
```
✅ All passing

**Integration Tests:**
```
=== RUN   TestCreateComponent_Integration
--- PASS: TestCreateComponent_Integration (0.12s)
=== RUN   TestGetAllComponents_Integration
--- PASS: TestGetAllComponents_Integration (0.11s)
=== RUN   TestGetComponentByID_Integration
--- PASS: TestGetComponentByID_Integration (0.12s)
=== RUN   TestGetComponentByID_NotFound_Integration
--- PASS: TestGetComponentByID_NotFound_Integration (0.10s)
=== RUN   TestCreateComponent_ValidationError_Integration
--- PASS: TestCreateComponent_ValidationError_Integration (0.09s)
PASS
ok  	github.com/easi/backend/internal/architecturemodeling/infrastructure/api	0.556s
```
✅ All 5 tests passing

## Writing New Integration Tests

### Template:

```go
// +build integration

package api

import (
	"testing"
	// ... imports
)

func TestYourFeature_Integration(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	handlers, readModel := setupHandlers(db)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/endpoint", body)
	w := httptest.NewRecorder()

	// Execute
	handlers.YourHandler(w, req)

	// Assert HTTP response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify database state
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM events WHERE event_type = 'YourEvent'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
```

## CI/CD Integration

For continuous integration, add to your pipeline:

```yaml
test:
  script:
    - podman compose up -d postgres
    - sleep 5  # Wait for PostgreSQL to be ready
    - go test -v ./...
    - go test -v -tags=integration ./...
  after_script:
    - podman compose down
```

## Troubleshooting

### "Connection refused" errors
- Ensure PostgreSQL is running: `podman ps | grep postgres`
- Start if needed: `podman compose up -d`

### "Database does not exist" errors
- The database is created automatically by docker-compose
- Check connection: `podman exec easi-postgres psql -U easi -d easi -c '\dt'`

### "Table already exists" errors
- Tests should clean up after themselves
- Manual cleanup: `podman exec easi-postgres psql -U easi -d easi -c 'DROP TABLE IF EXISTS events CASCADE; DROP TABLE IF EXISTS application_components CASCADE;'`

### Tests hanging
- Check for unclosed database connections
- Verify cleanup functions are called with `defer`

## Future Test Coverage

To be added:
- [ ] Integration tests for Relations API
- [ ] Integration tests for Views API
- [ ] End-to-end tests with frontend
- [ ] Performance/load tests
- [ ] Concurrency tests for event store
- [ ] Event projection verification tests

---

*Last updated: 2025-11-08*
