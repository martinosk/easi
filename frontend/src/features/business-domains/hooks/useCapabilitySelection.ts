import { useState, useCallback } from 'react';
import type { Capability, CapabilityId } from '../../../api/types';

interface UseCapabilitySelectionResult {
  selectedCapabilities: Set<CapabilityId>;
  handleCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  selectAllL1Capabilities: () => void;
  clearSelection: () => void;
  setSelectedCapabilities: (value: Set<CapabilityId>) => void;
}

export function useCapabilitySelection(
  capabilities: Capability[],
  onRegularClick: (capability: Capability) => void
): UseCapabilitySelectionResult {
  const [selectedCapabilities, setSelectedCapabilities] = useState<Set<CapabilityId>>(new Set());

  const handleCapabilityClick = useCallback((capability: Capability, event: React.MouseEvent) => {
    if (event.shiftKey) {
      event.preventDefault();
      event.stopPropagation();
      setSelectedCapabilities(prev => {
        const next = new Set(prev);
        if (next.has(capability.id)) {
          next.delete(capability.id);
        } else {
          next.add(capability.id);
        }
        return next;
      });
    } else {
      setSelectedCapabilities(new Set());
      onRegularClick(capability);
    }
  }, [onRegularClick]);

  const selectAllL1Capabilities = useCallback(() => {
    const l1Capabilities = capabilities.filter(c => c.level === 'L1');
    setSelectedCapabilities(new Set(l1Capabilities.map(c => c.id)));
  }, [capabilities]);

  const clearSelection = useCallback(() => {
    setSelectedCapabilities(new Set());
  }, []);

  return {
    selectedCapabilities,
    handleCapabilityClick,
    selectAllL1Capabilities,
    clearSelection,
    setSelectedCapabilities,
  };
}
