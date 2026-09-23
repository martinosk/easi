import type React from 'react';
import { useAppStore } from '../../../store/appStore';
import { ComponentDetailsPanel } from './ComponentDetailsPanel';
import { ComponentViewMembershipSection, useComponentViewMembership } from './ComponentViewMembershipSection';

interface ComponentDetailsProps {
  onRemoveFromView: () => void;
}

interface SelectedComponentProps extends ComponentDetailsProps {
  componentId: string;
}

const SelectedComponentDetails: React.FC<SelectedComponentProps> = ({ componentId, onRemoveFromView }) => {
  const membership = useComponentViewMembership(componentId);

  return (
    <ComponentDetailsPanel
      componentId={componentId}
      viewMembership={
        membership && <ComponentViewMembershipSection {...membership} onRemoveFromView={onRemoveFromView} />
      }
    />
  );
};

export const ComponentDetails: React.FC<ComponentDetailsProps> = ({ onRemoveFromView }) => {
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  if (!selectedNodeId) return null;

  return <SelectedComponentDetails componentId={selectedNodeId} onRemoveFromView={onRemoveFromView} />;
};
