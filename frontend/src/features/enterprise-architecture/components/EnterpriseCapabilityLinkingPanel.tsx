import { EnterpriseCapabilityCard } from './EnterpriseCapabilityCard';
import type { EnterpriseCapability, EnterpriseCapabilityId } from '../types';
import type { Capability } from '../../../api/types';

export interface EnterpriseCapabilityLinkingPanelProps {
  capabilities: EnterpriseCapability[];
  isLoading: boolean;
  onLinkCapability: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => void;
}

export function EnterpriseCapabilityLinkingPanel({
  capabilities,
  isLoading,
  onLinkCapability,
}: EnterpriseCapabilityLinkingPanelProps) {
  if (isLoading) {
    return (
      <div style={{ padding: '1rem' }}>
        <div style={{ color: '#6b7280' }}>Loading enterprise capabilities...</div>
      </div>
    );
  }

  if (capabilities.length === 0) {
    return (
      <div style={{ padding: '1rem' }}>
        <div style={{ color: '#6b7280' }}>No enterprise capabilities available</div>
      </div>
    );
  }

  return (
    <div style={{ padding: '1rem' }}>
      <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '1rem' }}>
        Enterprise Capabilities
      </h2>
      <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1.5rem' }}>
        Drag domain capabilities here to create links
      </p>
      <div>
        {capabilities.map((capability) => (
          <EnterpriseCapabilityCard
            key={capability.id}
            capability={capability}
            onDrop={(domainCapability) => onLinkCapability(capability.id, domainCapability)}
          />
        ))}
      </div>
    </div>
  );
}
