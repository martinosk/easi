#!/bin/bash

# E2E Test Environment Setup Script
# This script manages the isolated Docker environment for E2E tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[E2E Setup]${NC} $1"
}

log_error() {
    echo -e "${RED}[E2E Setup]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[E2E Setup]${NC} $1"
}

# Start the e2e environment
start_e2e_env() {
    log_info "Starting E2E test environment..."

    cd "$PROJECT_ROOT"

    # Stop any existing e2e containers
    docker compose -f docker-compose.e2e.yml down -v 2>/dev/null || true

    # Start the e2e environment
    log_info "Building and starting containers..."
    docker compose -f docker-compose.e2e.yml up -d --build

    # Wait for backend to be healthy
    log_info "Waiting for backend to be ready..."
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -f http://localhost:8081/health > /dev/null 2>&1; then
            log_info "Backend is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        echo -n "."
        sleep 1
    done

    log_error "Backend failed to start within timeout"
    docker compose -f docker-compose.e2e.yml logs backend-e2e
    return 1
}

# Stop the e2e environment
stop_e2e_env() {
    log_info "Stopping E2E test environment..."
    cd "$PROJECT_ROOT"
    docker compose -f docker-compose.e2e.yml down -v
    log_info "E2E environment stopped and cleaned up"
}

# Show logs
show_logs() {
    cd "$PROJECT_ROOT"
    docker compose -f docker-compose.e2e.yml logs -f
}

# Main command handling
case "${1:-}" in
    start)
        start_e2e_env
        ;;
    stop)
        stop_e2e_env
        ;;
    restart)
        stop_e2e_env
        start_e2e_env
        ;;
    logs)
        show_logs
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|logs}"
        echo ""
        echo "Commands:"
        echo "  start    - Start the E2E test environment"
        echo "  stop     - Stop and clean up the E2E test environment"
        echo "  restart  - Restart the E2E test environment"
        echo "  logs     - Show logs from the E2E environment"
        exit 1
        ;;
esac
