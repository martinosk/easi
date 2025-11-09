# API Documentation

This directory contains auto-generated OpenAPI/Swagger documentation.

## Single Source of Truth

The API specification is **auto-generated** from Go code annotations using [swaggo/swag](https://github.com/swaggo/swag).

### Generated Files

- `docs.go` - Go package with embedded spec
- `swagger.json` - OpenAPI 2.0 specification (JSON format)
- ~~`swagger.yaml`~~ - Removed (redundant)

These files are **not committed to git** and are generated during build.

## Regenerating the Spec

Run the following command to regenerate the API documentation:

```bash
cd backend
swag init -g cmd/api/main.go -o docs
```

## Accessing the Spec

### Backend
The spec is served at:
- Swagger UI: `http://localhost:8080/swagger/`
- JSON spec: `http://localhost:8080/docs/swagger.json`

### Frontend Integration

The frontend should fetch the spec at build time or runtime:

**Option 1: Build-time (recommended)**
```bash
curl http://localhost:8080/docs/swagger.json > src/api/openapi.json
```

**Option 2: Runtime**
Fetch from `/docs/swagger.json` and use with code generators like `openapi-typescript` or `@rtk-query/codegen-openapi`.

## No Manual Copies

Do not create manual copies of the spec in the frontend or elsewhere. Always fetch from the backend to ensure consistency.
