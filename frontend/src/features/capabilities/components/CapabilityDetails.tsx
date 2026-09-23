import type React from 'react';
import type { CapabilityId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { CapabilityDetailsPanel } from './CapabilityDetailsPanel';
import { CapabilityViewMembershipSection, useCapabilityViewMembership } from './CapabilityViewMembershipSection';

interface CapabilityDetailsProps {
  onRemoveFromView: () => void;
}

interface SelectedCapabilityProps extends CapabilityDetailsProps {
  capabilityId: CapabilityId;
}

const SelectedCapabilityDetails: React.FC<SelectedCapabilityProps> = ({ capabilityId, onRemoveFromView }) => {
  const membership = useCapabilityViewMembership(capabilityId);

  return (
    <CapabilityDetailsPanel
      capabilityId={capabilityId}
      viewMembership={
        membership && <CapabilityViewMembershipSection {...membership} onRemoveFromView={onRemoveFromView} />
      }
    />
  );
};

export const CapabilityDetails: React.FC<CapabilityDetailsProps> = ({ onRemoveFromView }) => {
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  if (!selectedCapabilityId) return null;

  return (
    <SelectedCapabilityDetails
      capabilityId={selectedCapabilityId as CapabilityId}
      onRemoveFromView={onRemoveFromView}
    />
  );
};
