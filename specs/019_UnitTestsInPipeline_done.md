# Unit Tests in Azure Pipeline

## Description
Add unit test execution for both backend and frontend in the Azure DevOps pipeline. The pipeline must fail if any tests don't pass, preventing deployment of broken code.

## Requirements
- Run backend unit tests (Go) before building Docker images
- Run frontend unit tests (npm/vitest) before building Docker images
- Pipeline fails if any tests fail
- Test results are visible in Azure DevOps test results UI
- Tests run in parallel with each other when possible for efficiency
- Test execution happens as early as possible in the pipeline

## Technical Details
Backend tests:
- Command: `go test ./...`
- Working directory: `backend`
- No external dependencies required for unit tests

Frontend tests:
- Command: `npm ci && npm run test`
- Working directory: `frontend`
- Uses vitest for unit tests

## Checklist
- [x] Specification ready
- [x] Implementation done
- [x] Pipeline successfully runs tests and fails on test failure
- [x] Test results visible in Azure DevOps
- [ ] User sign-off
