#!/bin/bash

# Script to generate OpenAPI specification from Swagger annotations
# This script is used to create the API contract for frontend consumption

set -e

echo "Generating OpenAPI specification..."

# Ensure we're in the backend directory
cd "$(dirname "$0")/.."

# Generate swagger docs
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Create frontend directory if it doesn't exist
mkdir -p ../frontend

# Copy OpenAPI spec to frontend directory
cp docs/swagger.json ../frontend/openapi.json

echo "OpenAPI specification generated successfully!"
echo "  - Backend: docs/swagger.json"
echo "  - Frontend: ../frontend/openapi.json"
