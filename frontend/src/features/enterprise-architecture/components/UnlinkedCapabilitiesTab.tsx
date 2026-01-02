import { useState, useCallback, useMemo } from 'react';
import { useUnlinkedCapabilities } from '../hooks/useMaturityAnalysis';
import { useBusinessDomains } from '../../business-domains/hooks/useBusinessDomains';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { UnlinkedCapability } from '../types';
import './UnlinkedCapabilitiesTab.css';

interface UnlinkedCapabilityItemProps {
  capability: UnlinkedCapability;
  getColorForValue: (value: number) => string;
}

function UnlinkedCapabilityItem({ capability, getColorForValue }: UnlinkedCapabilityItemProps) {
  return (
    <div className="unlinked-capability-item">
      <div className="capability-info">
        <span className="capability-name">{capability.capabilityName}</span>
        <div className="capability-meta">
          <span
            className="maturity-badge"
            style={{ backgroundColor: getColorForValue(capability.maturityValue) }}
          >
            {capability.maturitySection}
          </span>
          <span className="maturity-value">{capability.maturityValue}</span>
        </div>
      </div>
    </div>
  );
}

interface DomainGroupProps {
  domainName: string;
  capabilities: UnlinkedCapability[];
  getColorForValue: (value: number) => string;
}

function DomainGroup({ domainName, capabilities, getColorForValue }: DomainGroupProps) {
  return (
    <div className="domain-group">
      <h3 className="domain-group-header">{domainName}</h3>
      <div className="capability-list">
        {capabilities.map(capability => (
          <UnlinkedCapabilityItem
            key={capability.capabilityId}
            capability={capability}
            getColorForValue={getColorForValue}
          />
        ))}
      </div>
    </div>
  );
}

export function UnlinkedCapabilitiesTab() {
  const [businessDomainFilter, setBusinessDomainFilter] = useState<string>('');
  const [searchInput, setSearchInput] = useState<string>('');
  const [debouncedSearch, setDebouncedSearch] = useState<string>('');

  const { domains } = useBusinessDomains();
  const { getColorForValue } = useMaturityColorScale();
  const { capabilities, total, isLoading, error } = useUnlinkedCapabilities(
    businessDomainFilter || undefined,
    debouncedSearch || undefined
  );

  const handleDomainChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setBusinessDomainFilter(e.target.value);
  }, []);

  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchInput(e.target.value);
  }, []);

  const handleSearchSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    setDebouncedSearch(searchInput);
  }, [searchInput]);

  const handleSearchKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      setDebouncedSearch(searchInput);
    }
  }, [searchInput]);

  const groupedByDomain = useMemo(() => {
    const groups: Record<string, UnlinkedCapability[]> = {};

    capabilities.forEach(capability => {
      const domainName = capability.businessDomainName || 'No Domain';
      if (!groups[domainName]) {
        groups[domainName] = [];
      }
      groups[domainName].push(capability);
    });

    const sortedGroups = Object.entries(groups).sort(([a], [b]) => {
      if (a === 'No Domain') return 1;
      if (b === 'No Domain') return -1;
      return a.localeCompare(b);
    });

    return sortedGroups;
  }, [capabilities]);

  if (isLoading) {
    return (
      <div className="loading-state">
        <div className="loading-spinner" />
        <span>Loading unlinked capabilities...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-message">
        Failed to load unlinked capabilities: {error.message}
      </div>
    );
  }

  return (
    <div className="unlinked-capabilities-tab">
      <div className="unlinked-header">
        <div className="unlinked-summary">
          <span className="summary-count">{total}</span>
          <span className="summary-label">
            unlinked capabilities
            <HelpTooltip
              content="Domain capabilities not yet associated with any enterprise capability. Link them to enable maturity analysis and strategic planning."
              iconOnly
            />
          </span>
        </div>
        <div className="unlinked-filters">
          <form onSubmit={handleSearchSubmit} className="search-form">
            <input
              type="text"
              className="search-input"
              placeholder="Search by name..."
              value={searchInput}
              onChange={handleSearchChange}
              onKeyDown={handleSearchKeyDown}
            />
            <button type="submit" className="btn btn-sm btn-secondary">
              Search
            </button>
          </form>
          <select
            value={businessDomainFilter}
            onChange={handleDomainChange}
            className="domain-filter-select"
          >
            <option value="">All Domains</option>
            {domains.map(domain => (
              <option key={domain.id} value={domain.id}>
                {domain.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {capabilities.length === 0 ? (
        <div className="empty-state">
          <svg className="empty-state-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
          </svg>
          <h3 className="empty-state-title">All Capabilities Linked</h3>
          <p className="empty-state-description">
            All domain capabilities are currently linked to enterprise capabilities, or no capabilities match your filters.
          </p>
        </div>
      ) : (
        <div className="domain-groups">
          {groupedByDomain.map(([domainName, domainCapabilities]) => (
            <DomainGroup
              key={domainName}
              domainName={domainName}
              capabilities={domainCapabilities}
              getColorForValue={getColorForValue}
            />
          ))}
        </div>
      )}
    </div>
  );
}
