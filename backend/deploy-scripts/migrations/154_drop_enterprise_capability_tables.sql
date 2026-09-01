-- Migration: Retire the Enterprise Capability
-- Spec: 213_RetireEnterpriseCapability
-- Description: Drop the read-model tables of the retired EnterpriseCapability,
-- EnterpriseStrategicImportance, Direction and StandardApplication aggregates,
-- plus capability_domain_cache, which only the direction read model read.
-- Stored events in infrastructure.events are untouched.

DROP TABLE IF EXISTS architecturedirection.direction_source_capabilities;
DROP TABLE IF EXISTS architecturedirection.directions;
DROP TABLE IF EXISTS architecturedirection.standard_application_history;
DROP TABLE IF EXISTS architecturedirection.standard_applications;
DROP TABLE IF EXISTS architecturedirection.enterprise_strategic_importance;
DROP TABLE IF EXISTS architecturedirection.enterprise_capabilities;
DROP TABLE IF EXISTS architecturedirection.capability_domain_cache;
