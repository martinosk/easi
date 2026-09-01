-- Migration: Retire the enterprise-capability one-pager configuration
-- Spec: 213_RetireEnterpriseCapability
-- Description: The enterprise-capability subject type is retired. Its configuration
-- read-model row must go before the deploy-time archival sweep runs, so the
-- completeness recompute triggered by each OnePagerFactsArchived event resolves the
-- retired type to zero required fields instead of consulting a retired catalog.
-- One-pager facts are archived through ArchiveOnePagerFacts, never by SQL.

DELETE FROM onepagers.one_pager_configurations WHERE subject_type = 'enterprise-capability';
