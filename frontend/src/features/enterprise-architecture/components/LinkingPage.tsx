import { useCallback, useMemo } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { EnterpriseCapabilityLinkingPanel } from './EnterpriseCapabilityLinkingPanel';
import { DomainCapabilityPanel } from './DomainCapabilityPanel';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { capabilitiesApi } from '../../capabilities/api/capabilitiesApi';
import { queryKeys } from '../../../lib/queryClient';
import { useEnterpriseCapabilitiesQuery, useLinkDomainCapability } from '../hooks/useEnterpriseCapabilities';
import type { EnterpriseCapabilityId, CapabilityLinkStatusResponse } from '../types';
import type { Capability } from '../../../api/types';

function useLinkingData() {
  const queryClient = useQueryClient();

  const enterpriseQuery = useEnterpriseCapabilitiesQuery();

  const domainQuery = useQuery({
    queryKey: queryKeys.capabilities.lists(),
    queryFn: () => capabilitiesApi.getAll(),
  });

  const domainCapabilityIds = useMemo(
    () => domainQuery.data?.map((c) => c.id) ?? [],
    [domainQuery.data]
  );

  const linkStatusQuery = useQuery({
    queryKey: ['linkStatuses', domainCapabilityIds],
    queryFn: () => enterpriseArchApi.getBatchLinkStatus(domainCapabilityIds),
    enabled: domainCapabilityIds.length > 0,
  });

  const linkStatuses = useMemo(() => {
    if (!linkStatusQuery.data) return new Map<string, CapabilityLinkStatusResponse>();
    return new Map(linkStatusQuery.data.map((s) => [s.capabilityId, s]));
  }, [linkStatusQuery.data]);

  const linkMutation = useLinkDomainCapability();

  const handleLinkCapability = useCallback(
    async (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => {
      await linkMutation.mutateAsync({
        enterpriseCapabilityId,
        request: { domainCapabilityId: domainCapability.id },
      });
      queryClient.invalidateQueries({ queryKey: ['linkStatuses'] });
    },
    [linkMutation, queryClient]
  );

  return {
    enterpriseCapabilities: enterpriseQuery.data ?? [],
    domainCapabilities: domainQuery.data ?? [],
    linkStatuses,
    isLoadingEnterprise: enterpriseQuery.isLoading,
    isLoadingDomain: domainQuery.isLoading || linkStatusQuery.isLoading,
    error: enterpriseQuery.error?.message || domainQuery.error?.message || null,
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
