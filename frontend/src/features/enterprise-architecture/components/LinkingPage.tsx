import { useState, useEffect, useCallback, useRef } from 'react';
import { EnterpriseCapabilityLinkingPanel } from './EnterpriseCapabilityLinkingPanel';
import { DomainCapabilityPanel } from './DomainCapabilityPanel';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { capabilitiesApi } from '../../capabilities/api/capabilitiesApi';
import type { EnterpriseCapability, EnterpriseCapabilityId, CapabilityLinkStatusResponse } from '../types';
import type { Capability } from '../../../api/types';

function useLinkingData() {
  const [enterpriseCapabilities, setEnterpriseCapabilities] = useState<EnterpriseCapability[]>([]);
  const [domainCapabilities, setDomainCapabilities] = useState<Capability[]>([]);
  const [linkStatuses, setLinkStatuses] = useState<Map<string, CapabilityLinkStatusResponse>>(new Map());
  const [isLoadingEnterprise, setIsLoadingEnterprise] = useState(true);
  const [isLoadingDomain, setIsLoadingDomain] = useState(true);
  const [isLoadingStatuses, setIsLoadingStatuses] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const initialLoadDone = useRef(false);

  const loadEnterpriseCapabilities = useCallback(async () => {
    try {
      setIsLoadingEnterprise(true);
      setError(null);
      const capabilities = await enterpriseArchApi.getAll();
      setEnterpriseCapabilities(capabilities);
    } catch (err) {
      console.error('Failed to load enterprise capabilities:', err);
      setError('Failed to load enterprise capabilities');
    } finally {
      setIsLoadingEnterprise(false);
    }
  }, []);

  const loadDomainCapabilities = useCallback(async () => {
    try {
      setIsLoadingDomain(true);
      setError(null);
      const capabilities = await capabilitiesApi.getAll();
      setDomainCapabilities(capabilities);
      return capabilities;
    } catch (err) {
      console.error('Failed to load domain capabilities:', err);
      setError('Failed to load domain capabilities');
      return [];
    } finally {
      setIsLoadingDomain(false);
    }
  }, []);

  const loadLinkStatuses = useCallback(async (capabilities: Capability[]) => {
    if (capabilities.length === 0) {
      setLinkStatuses(new Map());
      return;
    }
    try {
      setIsLoadingStatuses(true);
      const statuses = await enterpriseArchApi.getBatchLinkStatus(capabilities.map((c) => c.id));
      setLinkStatuses(new Map(statuses.map((s) => [s.capabilityId, s])));
    } catch (err) {
      console.error('Failed to load link statuses:', err);
    } finally {
      setIsLoadingStatuses(false);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    await loadEnterpriseCapabilities();
    const capabilities = await loadDomainCapabilities();
    await loadLinkStatuses(capabilities);
  }, [loadEnterpriseCapabilities, loadDomainCapabilities, loadLinkStatuses]);

  useEffect(() => {
    if (!initialLoadDone.current) {
      initialLoadDone.current = true;
      refreshAll();
    }
  }, []);

  const handleLinkCapability = useCallback(async (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => {
    try {
      await enterpriseArchApi.linkDomainCapability(enterpriseCapabilityId, { domainCapabilityId: domainCapability.id });
      await refreshAll();
    } catch (err) {
      console.error('Failed to link capability:', err);
      setError('Failed to link capability');
    }
  }, [refreshAll]);

  return {
    enterpriseCapabilities,
    domainCapabilities,
    linkStatuses,
    isLoadingEnterprise,
    isLoadingDomain: isLoadingDomain || isLoadingStatuses,
    error,
    handleLinkCapability,
  };
}

export function LinkingPage() {
  const {
    enterpriseCapabilities,
    domainCapabilities,
    linkStatuses,
    isLoadingEnterprise,
    isLoadingDomain,
    error,
    handleLinkCapability,
  } = useLinkingData();

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden' }}>
      <div
        style={{
          width: '50%',
          borderRight: '1px solid #e5e7eb',
          overflow: 'auto',
          backgroundColor: '#fafafa',
        }}
      >
        <EnterpriseCapabilityLinkingPanel
          capabilities={enterpriseCapabilities}
          isLoading={isLoadingEnterprise}
          onLinkCapability={handleLinkCapability}
        />
      </div>

      <div
        style={{
          width: '50%',
          overflow: 'auto',
          backgroundColor: '#ffffff',
        }}
      >
        <DomainCapabilityPanel
          capabilities={domainCapabilities}
          linkStatuses={linkStatuses}
          isLoading={isLoadingDomain}
        />
      </div>

      {error && (
        <div
          style={{
            position: 'fixed',
            bottom: '1rem',
            right: '1rem',
            backgroundColor: '#fee2e2',
            color: '#991b1b',
            padding: '1rem',
            borderRadius: '0.5rem',
            boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
          }}
        >
          {error}
        </div>
      )}
    </div>
  );
}
