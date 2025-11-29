import { useState } from 'react';
import toast from 'react-hot-toast';
import { CapabilityTagList } from './CapabilityTagList';
import { CapabilitySelectorModal } from './CapabilitySelectorModal';
import { useDomainCapabilities } from '../hooks/useDomainCapabilities';
import type { Capability } from '../../../api/types';

interface CapabilityAssociationManagerProps {
  capabilitiesLink?: string;
  associateLink?: string;
}

export function CapabilityAssociationManager({ capabilitiesLink, associateLink }: CapabilityAssociationManagerProps) {
  const [isSelectorOpen, setIsSelectorOpen] = useState(false);
  const { capabilities, isLoading, error, associateCapability, dissociateCapability } = useDomainCapabilities(
    capabilitiesLink,
    associateLink
  );

  const handleAddCapabilities = async (selectedCapabilities: Capability[]) => {
    try {
      for (const capability of selectedCapabilities) {
        await associateCapability(capability.id, capability);
      }
      toast.success(`Added ${selectedCapabilities.length} ${selectedCapabilities.length === 1 ? 'capability' : 'capabilities'}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to add capabilities');
      throw err;
    }
  };

  const handleRemoveCapability = async (capability: Capability) => {
    try {
      await dissociateCapability(capability);
      toast.success(`Removed ${capability.name}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to remove capability');
    }
  };

  if (isLoading) {
    return <div className="loading-message">Loading capabilities...</div>;
  }

  if (error) {
    return (
      <div className="error-message" data-testid="capability-association-error">
        {error.message}
      </div>
    );
  }

  return (
    <div className="capability-association-manager" data-testid="capability-association-manager">
      <div className="capability-association-header">
        <h3>Associated Capabilities</h3>
        {associateLink && (
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => setIsSelectorOpen(true)}
            data-testid="add-capabilities-button"
          >
            Add Capabilities
          </button>
        )}
      </div>

      <CapabilityTagList capabilities={capabilities} onRemove={handleRemoveCapability} />

      {isSelectorOpen && (
        <CapabilitySelectorModal
          isOpen={isSelectorOpen}
          onClose={() => setIsSelectorOpen(false)}
          currentAssociations={capabilities}
          onSave={handleAddCapabilities}
        />
      )}
    </div>
  );
}
