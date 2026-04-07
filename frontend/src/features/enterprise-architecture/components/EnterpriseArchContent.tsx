import React from 'react';
import type { Capability } from '../../../api/types';
import type { CapabilityLinkStatusResponse, EnterpriseCapability, EnterpriseCapabilityId } from '../types';
import { DomainCapabilityDockPanel } from './DomainCapabilityDockPanel';
import { EnterpriseCapabilitiesEmptyState } from './EnterpriseCapabilitiesEmptyState';
import { EnterpriseCapabilitiesTable } from './EnterpriseCapabilitiesTable';
import { EnterpriseCapabilityDetailPanel } from './EnterpriseCapabilityDetailPanel';

interface EnterpriseArchContentProps {
  isLoading: boolean;
  error: string | null;
  capabilities: EnterpriseCapability[];
  selectedCapability: EnterpriseCapability | null;
  canWrite: boolean;
  onSelect: (capability: EnterpriseCapability) => void;
  onDelete: (capability: EnterpriseCapability) => void;
  onCreateNew: () => void;
  isDockPanelOpen: boolean;
  domainCapabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoadingDomainCapabilities: boolean;
  onCloseDockPanel: () => void;
  onLinkCapability: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => void;
}

export const EnterpriseArchContent = React.memo<EnterpriseArchContentProps>(
  ({
    isLoading,
    error,
    capabilities,
    selectedCapability,
    canWrite,
    onSelect,
    onDelete,
    onCreateNew,
    isDockPanelOpen,
    domainCapabilities,
    linkStatuses,
    isLoadingDomainCapabilities,
    onCloseDockPanel,
    onLinkCapability,
  }) => {
    if (isLoading) {
      return (
        <div className="loading-state">
          <div className="loading-spinner" />
          <p>Loading enterprise capabilities...</p>
        </div>
      );
    }

    if (error) {
      return (
        <div className="error-message" data-testid="capabilities-error">
          {error}
        </div>
      );
    }

    if (capabilities.length === 0) {
      return <EnterpriseCapabilitiesEmptyState onCreateNew={onCreateNew} canWrite={canWrite} />;
    }

    const hasAnyPanel = selectedCapability || isDockPanelOpen;
    const hasBothPanels = selectedCapability && isDockPanelOpen;

    const getTableContainerClass = () => {
      if (hasBothPanels) return 'table-container with-both-panels';
      if (hasAnyPanel) return 'table-container with-panel';
      return 'table-container';
    };

    return (
      <div className="enterprise-arch-content-layout">
        <div className={getTableContainerClass()}>
          <EnterpriseCapabilitiesTable
            capabilities={capabilities}
            selectedId={selectedCapability?.id}
            onSelect={onSelect}
            onDelete={onDelete}
            isDockPanelOpen={isDockPanelOpen}
            onLinkCapability={onLinkCapability}
          />
        </div>
        {selectedCapability && (
          <EnterpriseCapabilityDetailPanel
            capability={selectedCapability}
            onClose={() => onSelect(selectedCapability)}
          />
        )}
        {isDockPanelOpen && (
          <DomainCapabilityDockPanel
            capabilities={domainCapabilities}
            linkStatuses={linkStatuses}
            isLoading={isLoadingDomainCapabilities}
            onClose={onCloseDockPanel}
          />
        )}
      </div>
    );
  },
);

EnterpriseArchContent.displayName = 'EnterpriseArchContent';
