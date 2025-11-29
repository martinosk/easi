import { useState } from 'react';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import type { Capability } from '../../../api/types';

interface CapabilityTagListProps {
  capabilities: Capability[];
  onRemove: (capability: Capability) => void;
}

export function CapabilityTagList({ capabilities, onRemove }: CapabilityTagListProps) {
  const [capabilityToRemove, setCapabilityToRemove] = useState<Capability | null>(null);

  const handleRemoveClick = (capability: Capability) => {
    setCapabilityToRemove(capability);
  };

  const handleConfirmRemove = () => {
    if (capabilityToRemove) {
      onRemove(capabilityToRemove);
      setCapabilityToRemove(null);
    }
  };

  const handleCancelRemove = () => {
    setCapabilityToRemove(null);
  };

  if (capabilities.length === 0) {
    return (
      <div className="empty-state" data-testid="capabilities-empty-state">
        <p>No capabilities associated yet.</p>
        <p>Add L1 capabilities to this domain.</p>
      </div>
    );
  }

  return (
    <>
      <div className="capability-tags" data-testid="capability-tag-list">
        {capabilities.map((capability) => (
          <div key={capability.id} className="capability-tag" data-testid={`capability-tag-${capability.id}`}>
            <span className="capability-tag-name">{capability.name}</span>
            {capability._links.dissociate && (
              <button
                type="button"
                className="capability-tag-remove"
                onClick={() => handleRemoveClick(capability)}
                aria-label={`Remove ${capability.name}`}
                data-testid={`capability-remove-${capability.id}`}
              >
                ×
              </button>
            )}
          </div>
        ))}
      </div>

      {capabilityToRemove && (
        <ConfirmationDialog
          title="Remove Capability"
          message={`Are you sure you want to remove "${capabilityToRemove.name}" from this domain?`}
          confirmText="Remove"
          cancelText="Cancel"
          onConfirm={handleConfirmRemove}
          onCancel={handleCancelRemove}
        />
      )}
    </>
  );
}
