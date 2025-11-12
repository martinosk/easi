-- Migration: Initial Schema
-- Description: Creates base database schema for event sourcing and CQRS read models
-- This migration creates all tables needed for a fresh database installation

-- ============================================================================
-- Event Store Tables
-- ============================================================================

-- Events table - stores all domain events
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    event_data JSONB NOT NULL,
    version INT NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(aggregate_id, version)
);

-- Indexes for efficient event retrieval
CREATE INDEX IF NOT EXISTS idx_events_aggregate_id ON events(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_occurred_at ON events(occurred_at);

-- Snapshots table - stores aggregate snapshots for performance
CREATE TABLE IF NOT EXISTS snapshots (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id VARCHAR(255) NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    version INT NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(aggregate_id, version)
);

-- Indexes for efficient snapshot retrieval
CREATE INDEX IF NOT EXISTS idx_snapshots_aggregate_id ON snapshots(aggregate_id);

-- ============================================================================
-- Architecture Modeling Read Models
-- ============================================================================

-- Application Components read model
CREATE TABLE IF NOT EXISTS application_components (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_application_components_name ON application_components(name);
CREATE INDEX IF NOT EXISTS idx_application_components_created_at ON application_components(created_at);

-- Component Relations read model
CREATE TABLE IF NOT EXISTS component_relations (
    id VARCHAR(255) PRIMARY KEY,
    source_component_id VARCHAR(255) NOT NULL,
    target_component_id VARCHAR(255) NOT NULL,
    relation_type VARCHAR(50) NOT NULL,
    name VARCHAR(500),
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_component_relations_source ON component_relations(source_component_id);
CREATE INDEX IF NOT EXISTS idx_component_relations_target ON component_relations(target_component_id);
CREATE INDEX IF NOT EXISTS idx_component_relations_type ON component_relations(relation_type);
CREATE INDEX IF NOT EXISTS idx_component_relations_created_at ON component_relations(created_at);

-- ============================================================================
-- Architecture Views Read Models
-- ============================================================================

-- Architecture Views read model
CREATE TABLE IF NOT EXISTS architecture_views (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_architecture_views_name ON architecture_views(name);
CREATE INDEX IF NOT EXISTS idx_architecture_views_created_at ON architecture_views(created_at);
CREATE INDEX IF NOT EXISTS idx_architecture_views_is_default ON architecture_views(is_default);

-- View Component Positions junction table
CREATE TABLE IF NOT EXISTS view_component_positions (
    view_id VARCHAR(255) NOT NULL,
    component_id VARCHAR(255) NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (view_id, component_id),
    FOREIGN KEY (view_id) REFERENCES architecture_views(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_view_component_positions_view_id ON view_component_positions(view_id);

-- ============================================================================
-- Migration complete
-- ============================================================================
