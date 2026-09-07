import type React from 'react';
import { useAppStore } from '../../../store/appStore';
import { ComponentDetailsPanel } from './ComponentDetailsPanel';
import { ComponentViewMembershipSection } from './ComponentViewMembershipSection';

interface ComponentDetailsProps {
  onRemoveFromView: () => void;
}

export const ComponentDetails: React.FC<ComponentDetailsProps> = ({ onRemoveFromView }) => {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  if (!selectedNodeId) return null;

  return (
    <ComponentDetailsPanel
      componentId={selectedNodeId}
      viewMembership={<ComponentViewMembershipSection componentId={selectedNodeId} onRemoveFromView={onRemoveFromView} />}
    />
  );
};
