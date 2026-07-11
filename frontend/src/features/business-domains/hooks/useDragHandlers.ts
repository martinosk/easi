import { useCallback, useState } from 'react';
import toast from 'react-hot-toast';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';

interface UseDragHandlersProps {
  associateCapability: (domainId: BusinessDomainId, capabilityId: CapabilityId) => Promise<void>;
  isCapabilityAssignedToDomain: (domainId: BusinessDomainId, capabilityId: CapabilityId) => boolean;
  refetchDomain: (domainId: BusinessDomainId) => Promise<void>;
}

export function useDragHandlers({
  associateCapability,
  isCapabilityAssignedToDomain,
  refetchDomain,
}: UseDragHandlersProps) {
  const [activeCapability, setActiveCapability] = useState<Capability | null>(null);
  const [dragOverDomainId, setDragOverDomainId] = useState<BusinessDomainId | null>(null);

  const handleDragStart = useCallback((capability: Capability) => {
    setActiveCapability(capability);
  }, []);

  const handleDragEnd = useCallback(() => {
    setActiveCapability(null);
    setDragOverDomainId(null);
  }, []);

  const handleDragOver = useCallback((domainId: BusinessDomainId, e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setDragOverDomainId(domainId);
  }, []);

  const handleDragLeave = useCallback(() => {
    setDragOverDomainId(null);
  }, []);

  const handleDrop = useCallback(
    async (domainId: BusinessDomainId, e: React.DragEvent) => {
      e.preventDefault();
      setDragOverDomainId(null);

      const capabilityJson = e.dataTransfer.getData('application/json');
      if (!capabilityJson) {
        setActiveCapability(null);
        return;
      }

      try {
        const capability = JSON.parse(capabilityJson) as Capability;

        if (capability.level !== 'L1') {
          setActiveCapability(null);
          return;
        }

        if (isCapabilityAssignedToDomain(domainId, capability.id)) {
          setActiveCapability(null);
          return;
        }

        await associateCapability(domainId, capability.id);
        await refetchDomain(domainId);
      } catch {
        toast.error('Failed to assign capability');
      } finally {
        setActiveCapability(null);
      }
    },
    [associateCapability, isCapabilityAssignedToDomain, refetchDomain],
  );

  return {
    activeCapability,
    dragOverDomainId,
    handleDragStart,
    handleDragEnd,
    handleDragOver,
    handleDragLeave,
    handleDrop,
  };
}
