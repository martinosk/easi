import type React from 'react';
import { useAppStore } from '../../../store/appStore';
import { CapabilityDetailsPanel } from './CapabilityDetailsPanel';
import { CapabilityViewMembershipSection } from './CapabilityViewMembershipSection';

interface CapabilityDetailsProps {
  onRemoveFromView: () => void;
}

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromView }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  if (!selectedCapabilityId) return null;

  return (
    <CapabilityDetailsPanel
      capabilityId={selectedCapabilityId}
      viewMembership={
        <CapabilityViewMembershipSection capabilityId={selectedCapabilityId} onRemoveFromView={onRemoveFromView} />
      }
    />
  );
};
