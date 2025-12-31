import React from 'react';
import { DomainCapabilityPanel } from './DomainCapabilityPanel';
import type { Capability } from '../../../api/types';
import type { CapabilityLinkStatusResponse } from '../types';

interface DomainCapabilityDockPanelProps {
  capabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoading: boolean;
  onClose: () => void;
}

export const DomainCapabilityDockPanel = React.memo<DomainCapabilityDockPanelProps>(({
  capabilities,
  linkStatuses,
  isLoading,
  onClose,
}) => {
  return (
    <div className="dock-panel">
      <div className="dock-panel-header">
        <h3 className="dock-panel-title">Link Capabilities</h3>
        <button
          type="button"
          className="btn-close"
          onClick={onClose}
          aria-label="Close dock panel"
        >
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="20" height="20">
            <path d="M18 6L6 18M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </div>
      <DomainCapabilityPanel
        capabilities={capabilities}
        linkStatuses={linkStatuses}
        isLoading={isLoading}
      />
    </div>
  );
});

DomainCapabilityDockPanel.displayName = 'DomainCapabilityDockPanel';
