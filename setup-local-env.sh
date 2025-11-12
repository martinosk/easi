#!/bin/bash

# EASI Local Development Environment Setup Script
# This script helps developers set up environment variables for local development

set -e

ENV_FILE=".env"
ENV_EXAMPLE=".env.example"

echo "==================================================================="
echo "EASI Local Development Environment Setup"
echo "==================================================================="
echo ""

# Check if .env.example exists
if [ ! -f "$ENV_EXAMPLE" ]; then
    echo "Error: $ENV_EXAMPLE file not found!"
    echo "Please ensure $ENV_EXAMPLE exists in the project root."
    exit 1
fi

# Check if .env already exists
if [ -f "$ENV_FILE" ]; then
    echo "Warning: $ENV_FILE already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled. Existing $ENV_FILE preserved."
        exit 0
    fi
    echo "Backing up existing $ENV_FILE to ${ENV_FILE}.backup"
    cp "$ENV_FILE" "${ENV_FILE}.backup"
fi

# Copy .env.example to .env
cp "$ENV_EXAMPLE" "$ENV_FILE"
echo "Created $ENV_FILE from $ENV_EXAMPLE"
echo ""

# Inform user about customization
echo "==================================================================="
echo "Setup Complete!"
echo "==================================================================="
echo ""
echo "The $ENV_FILE file has been created with default values."
echo ""
echo "IMPORTANT:"
echo "- These are LOCAL DEVELOPMENT passwords only!"
echo "- Review and customize $ENV_FILE if needed"
echo "- Never commit $ENV_FILE to version control"
echo "- The $ENV_FILE is already in .gitignore"
echo ""
echo "To start the development environment:"
echo "  docker-compose up -d"
echo ""
echo "To run integration tests:"
echo "  cd backend && ./test_integration.sh"
echo ""
echo "==================================================================="
